package store

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct {
	pool      *pgxpool.Pool
	generator ReplyGenerator
}

type PostgresOption func(*PostgresStore)

func WithPostgresReplyGenerator(generator ReplyGenerator) PostgresOption {
	return func(s *PostgresStore) {
		s.generator = generator
	}
}

func NewPostgresStore(pool *pgxpool.Pool, options ...PostgresOption) *PostgresStore {
	st := &PostgresStore{pool: pool}
	for _, option := range options {
		option(st)
	}
	return st
}

func (s *PostgresStore) ListConversations() ([]Conversation, error) {
	rows, err := s.pool.Query(context.Background(), `
		select id, customer, channel, intent, status, risk_level, started_at, last_message
		from conversations
		order by started_at desc
		limit 50
	`)
	if err != nil {
		return nil, fmt.Errorf("list conversations: %w", err)
	}
	defer rows.Close()

	items, err := scanConversations(rows)
	if err != nil {
		return nil, fmt.Errorf("scan conversations: %w", err)
	}
	return items, nil
}

func (s *PostgresStore) CreateConversation(customer, channel string) (Conversation, error) {
	now := time.Now().UTC()
	conv := Conversation{
		ID: fmt.Sprintf("conv_%d", now.UnixNano()), Customer: fallback(customer, "访客"),
		Channel: fallback(channel, "Web"), Intent: "待识别", Status: "Active",
		RiskLevel: "LOW", StartedAt: now.Format(time.RFC3339),
	}
	if _, err := s.pool.Exec(context.Background(), `
		insert into conversations (id, customer, channel, intent, status, risk_level, started_at, last_message)
		values ($1, $2, $3, $4, $5, $6, $7, $8)
	`, conv.ID, conv.Customer, conv.Channel, conv.Intent, conv.Status, conv.RiskLevel, now, conv.LastMessage); err != nil {
		return Conversation{}, fmt.Errorf("create conversation: %w", err)
	}
	return conv, nil
}

func (s *PostgresStore) ListMessages(conversationID string) ([]Message, error) {
	rows, err := s.pool.Query(context.Background(), `
		select id, conversation_id, role, content, engine, safe, fallback_reason, evidence_ids, created_at, trace
		from conversation_messages
		where conversation_id = $1
		order by created_at asc, id asc
	`, conversationID)
	if err != nil {
		return nil, fmt.Errorf("list messages: %w", err)
	}
	defer rows.Close()

	items := make([]Message, 0)
	for rows.Next() {
		item, err := scanMessage(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate messages: %w", err)
	}
	return items, nil
}

func (s *PostgresStore) SendMessage(conversationID, content string) (SendMessageResult, error) {
	ctx := context.Background()
	now := time.Now().UTC()
	channel := "Web"
	if conversationID == "" {
		conv, err := s.CreateConversation("访客", "Web")
		if err != nil {
			return SendMessageResult{}, err
		}
		conversationID = conv.ID
		channel = conv.Channel
	}
	exists, err := s.conversationExists(ctx, conversationID)
	if err != nil {
		return SendMessageResult{}, err
	}
	if !exists {
		if _, err := s.pool.Exec(ctx, `
			insert into conversations (id, customer, channel, intent, status, risk_level, started_at, last_message)
			values ($1, '访客', 'Web', '待识别', 'Active', 'LOW', $2, '')
			on conflict (id) do nothing
		`, conversationID, now); err != nil {
			return SendMessageResult{}, fmt.Errorf("ensure conversation: %w", err)
		}
	} else {
		channel, err = s.conversationChannel(ctx, conversationID)
		if err != nil {
			return SendMessageResult{}, err
		}
	}

	userMessage := Message{ID: fmt.Sprintf("msg_%d_user", now.UnixNano()), ConversationID: conversationID, Role: "user", Content: content, Engine: "customer", Safe: true, CreatedAt: now.Format(time.RFC3339)}
	evidence, err := s.SearchKnowledge(content)
	if err != nil {
		return SendMessageResult{}, err
	}
	rules, err := s.listRules()
	if err != nil {
		return SendMessageResult{}, err
	}
	ruleResult := evaluateRules(content, evidence, rulesByStage(rules, "active"))
	history, err := s.ListMessages(conversationID)
	if err != nil {
		return SendMessageResult{}, err
	}
	agentMessage, gap := postgresAgentReply(s.generator, conversationID, content, evidence, history, ruleResult, now)

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return SendMessageResult{}, fmt.Errorf("begin send message: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
		insert into conversation_messages (id, conversation_id, role, content, engine, safe, fallback_reason, evidence_ids, created_at, trace)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, userMessage.ID, userMessage.ConversationID, userMessage.Role, userMessage.Content, userMessage.Engine, userMessage.Safe, userMessage.FallbackReason, userMessage.EvidenceIDs, now, messageTraceJSON(userMessage.Trace)); err != nil {
		return SendMessageResult{}, fmt.Errorf("insert user message: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		insert into conversation_messages (id, conversation_id, role, content, engine, safe, fallback_reason, evidence_ids, created_at, trace)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, agentMessage.ID, agentMessage.ConversationID, agentMessage.Role, agentMessage.Content, agentMessage.Engine, agentMessage.Safe, agentMessage.FallbackReason, agentMessage.EvidenceIDs, now, messageTraceJSON(agentMessage.Trace)); err != nil {
		return SendMessageResult{}, fmt.Errorf("insert agent message: %w", err)
	}
	if shouldCreateTransfer(agentMessage.FallbackReason) {
		ticket := newTransferTicket(conversationID, channel, content, now.Format(time.RFC3339))
		if _, err := tx.Exec(ctx, `
			insert into transfer_tickets (id, conversation_id, channel, question, reason, priority, status, created_at)
			values ($1, $2, $3, $4, $5, $6, $7, $8)
		`, ticket.ID, ticket.ConversationID, ticket.Channel, ticket.Question, ticket.Reason, ticket.Priority, ticket.Status, now); err != nil {
			return SendMessageResult{}, fmt.Errorf("insert transfer ticket: %w", err)
		}
	}
	reviewTask := newReviewTask(agentMessage.ID, conversationID, channel, reviewPriority(agentMessage), reviewReason(agentMessage), now.Format(time.RFC3339))
	if _, err := tx.Exec(ctx, `
		insert into review_tasks (id, message_id, conversation_id, channel, assignee, status, priority, reason, created_at)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		on conflict (id) do nothing
	`, reviewTask.ID, reviewTask.MessageID, reviewTask.ConversationID, reviewTask.Channel, reviewTask.Assignee, reviewTask.Status, reviewTask.Priority, reviewTask.Reason, now); err != nil {
		return SendMessageResult{}, fmt.Errorf("insert review task: %w", err)
	}
	if gap != nil {
		if _, err := tx.Exec(ctx, `
			insert into knowledge_gaps (id, conversation_id, question, reason, status, priority, created_at)
			values ($1, $2, $3, $4, $5, $6, $7)
		`, gap.ID, gap.ConversationID, gap.Question, gap.Reason, gap.Status, gap.Priority, now); err != nil {
			return SendMessageResult{}, fmt.Errorf("insert knowledge gap: %w", err)
		}
	}
	if strings.TrimSpace(ruleResult.RuleCode) != "" {
		if _, err := tx.Exec(ctx, `
			update agent_rules
			set hit_count = hit_count + 1,
			    last_hit_at = now(),
			    updated_at = now()
			where code = $1
			  and stage = 'active'
		`, ruleResult.RuleCode); err != nil {
			return SendMessageResult{}, fmt.Errorf("record rule hit: %w", err)
		}
	}

	conv := postgresConversationState(conversationID, channel, content, evidence, gap, now)
	if _, err := tx.Exec(ctx, `
		insert into conversations (id, customer, channel, intent, status, risk_level, started_at, last_message)
		values ($1, $2, $3, $4, $5, $6, $7, $8)
		on conflict (id) do update
		set intent = excluded.intent,
		    status = excluded.status,
		    risk_level = excluded.risk_level,
		    last_message = excluded.last_message
	`, conv.ID, conv.Customer, conv.Channel, conv.Intent, conv.Status, conv.RiskLevel, now, conv.LastMessage); err != nil {
		return SendMessageResult{}, fmt.Errorf("update conversation state: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return SendMessageResult{}, fmt.Errorf("commit send message: %w", err)
	}
	return SendMessageResult{Conversation: conv, UserMessage: userMessage, AgentMessage: agentMessage, Evidence: evidence, Gap: gap}, nil
}

func (s *PostgresStore) ReceiveChannelMessage(message ChannelInboundMessage) (SendMessageResult, error) {
	now := time.Now().UTC()
	channel := fallback(message.Channel, "Web")
	externalID := fallback(message.ExternalConversationID, fmt.Sprintf("anon_%d", now.UnixNano()))
	conversationID := inboundConversationID(channel, externalID)
	if _, err := s.pool.Exec(context.Background(), `
		insert into conversations (id, customer, channel, intent, status, risk_level, started_at, last_message)
		values ($1, $2, $3, '待识别', 'Active', 'LOW', $4, '')
		on conflict (id) do nothing
	`, conversationID, fallback(message.Customer, "访客"), channel, now); err != nil {
		return SendMessageResult{}, fmt.Errorf("ensure inbound conversation: %w", err)
	}
	return s.SendMessage(conversationID, message.Content)
}

func (s *PostgresStore) RecordChannelInbound(receipt ChannelInboundReceipt) (bool, error) {
	tag, err := s.pool.Exec(context.Background(), `
		insert into channel_inbound_events (
			replay_key, channel, external_conversation_id, external_message_id, payload_timestamp, signature, content_hash
		)
		values ($1, $2, $3, $4, $5, $6, $7)
		on conflict (replay_key) do nothing
	`, receipt.ReplayKey, receipt.Channel, receipt.ExternalConversationID, receipt.ExternalMessageID, receipt.Timestamp, receipt.Signature, receipt.ContentHash)
	if err != nil {
		return false, fmt.Errorf("record channel inbound: %w", err)
	}
	return tag.RowsAffected() == 1, nil
}

func (s *PostgresStore) RecordChannelFailure(event ChannelFailureEvent) error {
	if strings.TrimSpace(event.Channel) == "" {
		event.Channel = "Unknown"
	}
	if strings.TrimSpace(event.Code) == "" {
		event.Code = "channel_failure"
	}
	createdAt := time.Now().UTC()
	if parsed, err := time.Parse(time.RFC3339, event.CreatedAt); err == nil {
		createdAt = parsed.UTC()
	}
	ctx := context.Background()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin record channel failure: %w", err)
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `
		insert into channel_failure_events (
			channel, code, reason, external_conversation_id, external_message_id, origin, created_at
		)
		values ($1, $2, $3, $4, $5, $6, $7)
	`, event.Channel, event.Code, event.Reason, event.ExternalConversationID, event.ExternalMessageID, event.Origin, createdAt); err != nil {
		return fmt.Errorf("record channel failure: %w", err)
	}
	alerts, err := s.listChannelAlertsTx(ctx, tx)
	if err != nil {
		return err
	}
	policies, err := s.listChannelAlertPoliciesTx(ctx, tx, alerts)
	if err != nil {
		return err
	}
	if err := insertTriggeredNotifications(ctx, tx, policies, createdAt.Format(time.RFC3339)); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit channel failure: %w", err)
	}
	return nil
}

func (s *PostgresStore) AcknowledgeChannelNotification(id, actor, note string) (ChannelNotification, error) {
	var item ChannelNotification
	var createdAt time.Time
	var ackedAt *time.Time
	err := s.pool.QueryRow(context.Background(), `
		update channel_notifications
		set status = 'ACKED',
		    acked_by = $2,
		    ack_note = $3,
		    acked_at = now()
		where id = $1
		returning id, channel, severity, target, status, reason, count, created_at, acked_by, ack_note, acked_at
	`, id, fallback(actor, "operator"), fallback(note, "已确认渠道告警")).Scan(
		&item.ID, &item.Channel, &item.Severity, &item.Target, &item.Status, &item.Reason, &item.Count,
		&createdAt, &item.AckedBy, &item.AckNote, &ackedAt,
	)
	if err != nil {
		return ChannelNotification{}, fmt.Errorf("ack channel notification: %w", err)
	}
	item.CreatedAt = createdAt.UTC().Format(time.RFC3339)
	if ackedAt != nil {
		item.AckedAt = ackedAt.UTC().Format(time.RFC3339)
	}
	return item, nil
}

func (s *PostgresStore) ChannelIntegration(channel string) (ChannelIntegration, error) {
	var item ChannelIntegration
	var updatedAt time.Time
	var rotatesAt *time.Time
	err := s.pool.QueryRow(context.Background(), `
		select channel, display_name, enabled, secret_source, secret_ref, next_secret_ref, signature_window_seconds, replay_protection, allowed_origins, rate_limit_per_minute, rotation_hint, rotates_at, updated_at
		from channel_integrations
		where lower(channel) = lower($1)
	`, channel).Scan(
		&item.Channel,
		&item.DisplayName,
		&item.Enabled,
		&item.SecretSource,
		&item.SecretRef,
		&item.NextSecretRef,
		&item.SignatureWindowSeconds,
		&item.ReplayProtection,
		&item.AllowedOrigins,
		&item.RateLimitPerMinute,
		&item.RotationHint,
		&rotatesAt,
		&updatedAt,
	)
	if err == nil {
		if rotatesAt != nil {
			item.RotatesAt = rotatesAt.UTC().Format(time.RFC3339)
		}
		item.UpdatedAt = updatedAt.UTC().Format(time.RFC3339)
		return item, nil
	}
	if err == pgx.ErrNoRows {
		return channelIntegrationFor(defaultChannelIntegrations(time.Now().UTC().Format(time.RFC3339)), channel), nil
	}
	return ChannelIntegration{}, fmt.Errorf("load channel integration: %w", err)
}

func (s *PostgresStore) RecordChannelRateLimit(channel string, windowStart time.Time, limit int) (bool, int, error) {
	if limit <= 0 {
		return true, 0, nil
	}
	var count int
	err := s.pool.QueryRow(context.Background(), `
		insert into channel_rate_limit_windows (channel, window_start, request_count)
		values ($1, $2, 1)
		on conflict (channel, window_start)
		do update set
			request_count = channel_rate_limit_windows.request_count + 1,
			updated_at = now()
		where channel_rate_limit_windows.request_count < $3
		returning request_count
	`, channel, windowStart.UTC(), limit).Scan(&count)
	if err == nil {
		return true, count, nil
	}
	if err == pgx.ErrNoRows {
		return false, limit, nil
	}
	return false, 0, fmt.Errorf("record channel rate limit: %w", err)
}

func (s *PostgresStore) ListKnowledge() ([]KnowledgeArticle, error) {
	rows, err := s.pool.Query(context.Background(), `
		select id, title, category, content, tags, trust_level, updated_at
		from knowledge_articles
		order by updated_at desc, id
	`)
	if err != nil {
		return nil, fmt.Errorf("list knowledge: %w", err)
	}
	defer rows.Close()

	items, err := scanKnowledge(rows)
	if err != nil {
		return nil, fmt.Errorf("scan knowledge: %w", err)
	}
	return items, nil
}

func (s *PostgresStore) SearchKnowledge(query string) ([]KnowledgeArticle, error) {
	items, err := s.ListKnowledge()
	if err != nil {
		return nil, err
	}
	query = strings.ToLower(query)
	matches := make([]KnowledgeArticle, 0)
	for _, item := range items {
		haystack := strings.ToLower(item.Title + " " + item.Category + " " + item.Content + " " + strings.Join(item.Tags, " "))
		if strings.Contains(query, strings.ToLower(item.Title)) || strings.Contains(query, strings.ToLower(item.Category)) {
			matches = append(matches, item)
			continue
		}
		matched := false
		for _, tag := range item.Tags {
			if strings.Contains(query, strings.ToLower(tag)) {
				matches = append(matches, item)
				matched = true
				break
			}
		}
		if matched {
			continue
		}
		for _, token := range strings.Fields(query) {
			if strings.Contains(haystack, token) {
				matches = append(matches, item)
				break
			}
		}
	}
	return matches, nil
}

func (s *PostgresStore) ResolveKnowledgeGap(id string) (KnowledgeGap, error) {
	var item KnowledgeGap
	var createdAt time.Time
	err := s.pool.QueryRow(context.Background(), `
		update knowledge_gaps
		set status = 'RESOLVED'
		where id = $1
		returning id, conversation_id, question, reason, status, priority, created_at
	`, id).Scan(&item.ID, &item.ConversationID, &item.Question, &item.Reason, &item.Status, &item.Priority, &createdAt)
	if err != nil {
		return KnowledgeGap{}, fmt.Errorf("resolve knowledge gap: %w", err)
	}
	item.CreatedAt = createdAt.UTC().Format(time.RFC3339)
	return item, nil
}

func (s *PostgresStore) CreateArticleFromGap(gapID, title, category, content string, tags []string) (KnowledgeArticle, error) {
	ctx := context.Background()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return KnowledgeArticle{}, fmt.Errorf("begin create article from gap: %w", err)
	}
	defer tx.Rollback(ctx)

	var gap KnowledgeGap
	var gapCreatedAt time.Time
	if err := tx.QueryRow(ctx, `
		select id, conversation_id, question, reason, status, priority, created_at
		from knowledge_gaps
		where id = $1
	`, gapID).Scan(&gap.ID, &gap.ConversationID, &gap.Question, &gap.Reason, &gap.Status, &gap.Priority, &gapCreatedAt); err != nil {
		return KnowledgeArticle{}, fmt.Errorf("find knowledge gap: %w", err)
	}

	now := time.Now().UTC()
	article := KnowledgeArticle{
		ID:         fmt.Sprintf("kb_%d", now.UnixNano()),
		Title:      fallback(title, gap.Question),
		Category:   fallback(category, "运营补充"),
		Content:    fallback(content, gap.Question),
		Tags:       normalizeTags(tags, gap.Question),
		TrustLevel: "HIGH",
		UpdatedAt:  now.Format(time.RFC3339),
	}
	if _, err := tx.Exec(ctx, `
		insert into knowledge_articles (id, title, category, content, tags, trust_level, updated_at)
		values ($1, $2, $3, $4, $5, $6, $7)
	`, article.ID, article.Title, article.Category, article.Content, article.Tags, article.TrustLevel, now); err != nil {
		return KnowledgeArticle{}, fmt.Errorf("insert knowledge article: %w", err)
	}
	if _, err := tx.Exec(ctx, "update knowledge_gaps set status = 'RESOLVED' where id = $1", gapID); err != nil {
		return KnowledgeArticle{}, fmt.Errorf("resolve knowledge gap: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return KnowledgeArticle{}, fmt.Errorf("commit create article from gap: %w", err)
	}
	return article, nil
}

func (s *PostgresStore) TestRule(content string) (RuleTestResult, error) {
	evidence, err := s.SearchKnowledge(content)
	if err != nil {
		return RuleTestResult{}, err
	}
	rules, err := s.listRules()
	if err != nil {
		return RuleTestResult{}, err
	}
	result := evaluateRules(content, evidence, rulesByStage(rules, "active"))
	if strings.TrimSpace(result.RuleCode) != "" {
		if err := s.recordRuleHit(result.RuleCode); err != nil {
			return RuleTestResult{}, err
		}
	}
	return result, nil
}

func (s *PostgresStore) CompareRuleVersions(content string) (RuleComparison, error) {
	evidence, err := s.SearchKnowledge(content)
	if err != nil {
		return RuleComparison{}, err
	}
	rules, err := s.listRules()
	if err != nil {
		return RuleComparison{}, err
	}
	current := evaluateRules(content, evidence, rulesByStage(rules, "active"))
	canary := evaluateRules(content, evidence, rulesForComparison(rules))
	return compareRuleResults(content, current, canary), nil
}

func (s *PostgresStore) SubmitRuleApproval(code, approver, riskLevel, note string, sampleIDs []string) (RuleApproval, error) {
	ctx := context.Background()
	var exists bool
	if err := s.pool.QueryRow(ctx, `
		select exists(select 1 from agent_rules where code = $1 and stage = 'canary')
	`, code).Scan(&exists); err != nil {
		return RuleApproval{}, fmt.Errorf("check canary rule: %w", err)
	}
	if !exists {
		return RuleApproval{}, fmt.Errorf("canary rule %s not found", code)
	}
	approval := newRuleApproval(code, approver, riskLevel, note, sampleIDs, time.Now().UTC().Format(time.RFC3339))
	createdAt, err := time.Parse(time.RFC3339, approval.CreatedAt)
	if err != nil {
		createdAt = time.Now().UTC()
	}
	if _, err := s.pool.Exec(ctx, `
		insert into rule_approvals (id, rule_code, approver, risk_level, sample_ids, sample_count, status, note, created_at)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, approval.ID, approval.RuleCode, approval.Approver, approval.RiskLevel, approval.SampleIDs, approval.SampleCount, approval.Status, approval.Note, createdAt); err != nil {
		return RuleApproval{}, fmt.Errorf("submit rule approval: %w", err)
	}
	return approval, nil
}

func (s *PostgresStore) PublishCanaryRule(code, actor, note string) (RuleReleaseEvent, error) {
	ctx := context.Background()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return RuleReleaseEvent{}, fmt.Errorf("begin publish canary rule: %w", err)
	}
	defer tx.Rollback(ctx)

	var version string
	if err := tx.QueryRow(ctx, `
		select version
		from agent_rules
		where code = $1
		  and stage = 'canary'
		  and enabled = true
	`, code).Scan(&version); err != nil {
		return RuleReleaseEvent{}, fmt.Errorf("load canary rule: %w", err)
	}
	var approved bool
	if err := tx.QueryRow(ctx, `
		select exists(
			select 1
			from rule_approvals
			where rule_code = $1
			  and status = 'APPROVED'
			  and sample_count >= 3
		)
	`, code).Scan(&approved); err != nil {
		return RuleReleaseEvent{}, fmt.Errorf("check rule approval: %w", err)
	}
	if !approved {
		return RuleReleaseEvent{}, fmt.Errorf("rule %s requires approved gate before publish", code)
	}
	publishedVersion := nextPublishedVersion(version)
	if _, err := tx.Exec(ctx, `
		update agent_rules
		set stage = 'archived',
		    enabled = false,
		    updated_at = now()
		where code = $1
		  and stage = 'active'
	`, code); err != nil {
		return RuleReleaseEvent{}, fmt.Errorf("archive active rule: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		update agent_rules
		set stage = 'active',
		    version = $2,
		    enabled = true,
		    updated_at = now()
		where code = $1
		  and stage = 'canary'
	`, code, publishedVersion); err != nil {
		return RuleReleaseEvent{}, fmt.Errorf("publish canary rule: %w", err)
	}
	event := newRuleReleaseEvent(code, publishedVersion, "PUBLISH", actor, note, time.Now().UTC().Format(time.RFC3339))
	if err := insertRuleEvent(ctx, tx, event); err != nil {
		return RuleReleaseEvent{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return RuleReleaseEvent{}, fmt.Errorf("commit publish canary rule: %w", err)
	}
	return event, nil
}

func (s *PostgresStore) RollbackRule(code, actor, note string) (RuleReleaseEvent, error) {
	ctx := context.Background()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return RuleReleaseEvent{}, fmt.Errorf("begin rollback rule: %w", err)
	}
	defer tx.Rollback(ctx)

	var version string
	if err := tx.QueryRow(ctx, `
		select version
		from agent_rules
		where code = $1
		  and stage = 'active'
	`, code).Scan(&version); err != nil {
		return RuleReleaseEvent{}, fmt.Errorf("load active rule: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		update agent_rules
		set stage = 'canary',
		    enabled = false,
		    updated_at = now()
		where code = $1
		  and stage = 'active'
	`, code); err != nil {
		return RuleReleaseEvent{}, fmt.Errorf("rollback active rule: %w", err)
	}
	event := newRuleReleaseEvent(code, version, "ROLLBACK", actor, note, time.Now().UTC().Format(time.RFC3339))
	if err := insertRuleEvent(ctx, tx, event); err != nil {
		return RuleReleaseEvent{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return RuleReleaseEvent{}, fmt.Errorf("commit rollback rule: %w", err)
	}
	return event, nil
}

func (s *PostgresStore) ResolveTransferTicket(id, assignee, note string) (TransferTicket, error) {
	var item TransferTicket
	var createdAt time.Time
	var resolvedAt *time.Time
	err := s.pool.QueryRow(context.Background(), `
		update transfer_tickets
		set status = 'RESOLVED',
		    assignee = $2,
		    resolution_note = $3,
		    resolved_at = now()
		where id = $1
		returning id, conversation_id, question, reason, priority, status, assignee, resolution_note, created_at, resolved_at
	`, id, fallback(assignee, "operator"), fallback(note, "人工已处理")).Scan(
		&item.ID, &item.ConversationID, &item.Question, &item.Reason, &item.Priority, &item.Status,
		&item.Assignee, &item.ResolutionNote, &createdAt, &resolvedAt,
	)
	if err != nil {
		return TransferTicket{}, fmt.Errorf("resolve transfer ticket: %w", err)
	}
	item.CreatedAt = createdAt.UTC().Format(time.RFC3339)
	if resolvedAt != nil {
		item.ResolvedAt = resolvedAt.UTC().Format(time.RFC3339)
	}
	item.Events = transferEvents(item)
	return item, nil
}

func (s *PostgresStore) SubmitAnnotation(messageID, reviewer, verdict, note string, dimensions AnnotationDimensions, tags []string) (Annotation, error) {
	createdAt := time.Now().UTC()
	annotation := newAnnotation(messageID, reviewer, verdict, note, dimensions, tags, createdAt.Format(time.RFC3339))
	ctx := context.Background()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Annotation{}, fmt.Errorf("begin submit annotation: %w", err)
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `
		insert into message_annotations (id, message_id, reviewer, verdict, note, groundedness, safety, helpfulness, tags, score, created_at)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`, annotation.ID, annotation.MessageID, annotation.Reviewer, annotation.Verdict, annotation.Note,
		annotation.Dimensions.Groundedness, annotation.Dimensions.Safety, annotation.Dimensions.Helpfulness,
		annotation.Tags, annotation.Score, createdAt); err != nil {
		return Annotation{}, fmt.Errorf("submit annotation: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		update review_tasks
		set status = 'COMPLETED',
		    assignee = coalesce(nullif(assignee, ''), $2),
		    completed_at = $3
		where message_id = $1
		  and status <> 'COMPLETED'
	`, messageID, annotation.Reviewer, createdAt); err != nil {
		return Annotation{}, fmt.Errorf("complete review task from annotation: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return Annotation{}, fmt.Errorf("commit submit annotation: %w", err)
	}
	return annotation, nil
}

func (s *PostgresStore) AssignReviewTask(id, assignee string) (ReviewTask, error) {
	task, err := s.updateReviewTask(`
		update review_tasks
		set status = case when status = 'COMPLETED' then status else 'ASSIGNED' end,
		    assignee = case when status = 'COMPLETED' then assignee else $2 end
		where id = $1
		returning id, message_id, conversation_id, channel, assignee, status, priority, reason, created_at, completed_at
	`, id, fallback(assignee, "qa-operator"))
	if err != nil {
		return ReviewTask{}, fmt.Errorf("assign review task: %w", err)
	}
	return task, nil
}

func (s *PostgresStore) CompleteReviewTask(id string) (ReviewTask, error) {
	task, err := s.updateReviewTask(`
		update review_tasks
		set status = 'COMPLETED',
		    assignee = coalesce(nullif(assignee, ''), 'qa-operator'),
		    completed_at = coalesce(completed_at, now())
		where id = $1
		returning id, message_id, conversation_id, channel, assignee, status, priority, reason, created_at, completed_at
	`, id)
	if err != nil {
		return ReviewTask{}, fmt.Errorf("complete review task: %w", err)
	}
	return task, nil
}

func (s *PostgresStore) ExportTrainingSamples(maxScore int) ([]TrainingSample, error) {
	maxScore = normalizeMaxScore(maxScore)
	rows, err := s.pool.Query(context.Background(), `
		select
			a.id,
			m.conversation_id,
			m.id,
			c.channel,
			coalesce((
				select um.content
				from conversation_messages um
				where um.conversation_id = m.conversation_id
				  and um.role = 'user'
				  and um.created_at <= m.created_at
				order by um.created_at desc, um.id desc
				limit 1
			), '') as prompt,
			m.content,
			m.engine,
			m.evidence_ids,
			a.reviewer,
			a.verdict,
			a.score,
			a.groundedness,
			a.safety,
			a.helpfulness,
			a.note,
			a.tags,
			a.created_at
		from message_annotations a
		join conversation_messages m on m.id = a.message_id
		join conversations c on c.id = m.conversation_id
		where m.role = 'assistant'
		  and (a.score <= $1 or upper(a.verdict) in ('FAIL', 'REVIEW'))
		order by a.created_at desc, a.id desc
		limit 100
	`, maxScore)
	if err != nil {
		return nil, fmt.Errorf("export training samples: %w", err)
	}
	defer rows.Close()

	samples := make([]TrainingSample, 0)
	for rows.Next() {
		var annotationID string
		var createdAt time.Time
		var sample TrainingSample
		if err := rows.Scan(
			&annotationID,
			&sample.ConversationID,
			&sample.MessageID,
			&sample.Channel,
			&sample.Prompt,
			&sample.Answer,
			&sample.Engine,
			&sample.EvidenceIDs,
			&sample.Reviewer,
			&sample.Verdict,
			&sample.Score,
			&sample.Dimensions.Groundedness,
			&sample.Dimensions.Safety,
			&sample.Dimensions.Helpfulness,
			&sample.Note,
			&sample.Tags,
			&createdAt,
		); err != nil {
			return nil, err
		}
		sample.ID = fmt.Sprintf("sample_%s", annotationID)
		sample.CreatedAt = createdAt.UTC().Format(time.RFC3339)
		samples = append(samples, sample)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return samples, nil
}

func (s *PostgresStore) Dashboard() (Dashboard, error) {
	conversations, err := s.ListConversations()
	if err != nil {
		return Dashboard{}, err
	}
	knowledge, err := s.ListKnowledge()
	if err != nil {
		return Dashboard{}, err
	}
	gaps, err := s.listGaps()
	if err != nil {
		return Dashboard{}, err
	}
	rules, err := s.listRules()
	if err != nil {
		return Dashboard{}, err
	}
	ruleApprovals, err := s.listRuleApprovals()
	if err != nil {
		return Dashboard{}, err
	}
	ruleEvents, err := s.listRuleEvents()
	if err != nil {
		return Dashboard{}, err
	}
	transfers, err := s.listTransferTickets()
	if err != nil {
		return Dashboard{}, err
	}
	channelPolicies, err := s.listChannelPolicies()
	if err != nil {
		return Dashboard{}, err
	}
	integrations, err := s.listChannelIntegrations()
	if err != nil {
		return Dashboard{}, err
	}
	messages, err := s.listRecentMessages()
	if err != nil {
		return Dashboard{}, err
	}
	annotations, err := s.listAnnotations()
	if err != nil {
		return Dashboard{}, err
	}
	reviewTasks, err := s.listReviewTasks()
	if err != nil {
		return Dashboard{}, err
	}
	channelAlerts, err := s.listChannelAlerts()
	if err != nil {
		return Dashboard{}, err
	}
	channelTrends, err := s.listChannelFailureTrends()
	if err != nil {
		return Dashboard{}, err
	}
	alertPolicies, err := s.listChannelAlertPolicies(channelAlerts)
	if err != nil {
		return Dashboard{}, err
	}
	notifications, err := s.listChannelNotifications()
	if err != nil {
		return Dashboard{}, err
	}

	openGaps := 0
	for _, gap := range gaps {
		if gap.Status == "OPEN" {
			openGaps++
		}
	}
	openTransfers := 0
	for _, ticket := range transfers {
		if ticket.Status == "OPEN" {
			openTransfers++
		}
	}
	openReviews := 0
	for _, task := range reviewTasks {
		if task.Status != "COMPLETED" {
			openReviews++
		}
	}
	transfers = withTransferSLAs(transfers, channelPolicies, time.Now().UTC())

	return Dashboard{
		Metrics: []Metric{
			{Label: "Active sessions", Value: fmt.Sprintf("%d", len(conversations)), Note: "PostgreSQL runtime"},
			{Label: "Knowledge items", Value: fmt.Sprintf("%d", len(knowledge)), Note: "trusted articles in PostgreSQL"},
			{Label: "Open gaps", Value: fmt.Sprintf("%d", openGaps), Note: "created by no-evidence fallback"},
			{Label: "Channels", Value: fmt.Sprintf("%d", activeChannelCount(conversations)), Note: "conversation sources with policies"},
			{Label: "Open transfers", Value: fmt.Sprintf("%d", openTransfers), Note: "waiting for human agents"},
			{Label: "SLA escalations", Value: fmt.Sprintf("%d", escalatedTransferCount(transfers)), Note: "open tickets past response SLA"},
			{Label: "Enabled rules", Value: fmt.Sprintf("%d", enabledRules(rules)), Note: "guardrail and transfer policies"},
			{Label: "Channel failures", Value: fmt.Sprintf("%d", channelAlertCount(channelAlerts)), Note: "rejected inbound requests"},
			{Label: "Active alerts", Value: fmt.Sprintf("%d", activeAlertPolicies(alertPolicies)), Note: "notification policies triggered"},
			{Label: "Review tasks", Value: fmt.Sprintf("%d", openReviews), Note: "assistant replies awaiting QA"},
		},
		Conversations:    conversations,
		KnowledgeGaps:    gaps,
		Rules:            rules,
		Transfers:        transfers,
		ChannelPolicies:  channelPolicies,
		Integrations:     integrations,
		ChannelAlerts:    channelAlerts,
		ChannelTrends:    channelTrends,
		AlertPolicies:    alertPolicies,
		Notifications:    notifications,
		Quality:          qualitySummary(messages, gaps, transfers, annotations),
		Annotations:      annotations,
		ReviewTasks:      reviewTasks,
		RuleApprovals:    ruleApprovals,
		RuleEvents:       ruleEvents,
		RuleObservations: ruleReleaseObservations(ruleEvents, rules, transfers, annotations, time.Now().UTC()),
	}, nil
}

func (s *PostgresStore) conversationExists(ctx context.Context, id string) (bool, error) {
	var exists bool
	if err := s.pool.QueryRow(ctx, "select exists(select 1 from conversations where id = $1)", id).Scan(&exists); err != nil {
		return false, fmt.Errorf("check conversation exists: %w", err)
	}
	return exists, nil
}

func (s *PostgresStore) conversationChannel(ctx context.Context, id string) (string, error) {
	var channel string
	if err := s.pool.QueryRow(ctx, "select channel from conversations where id = $1", id).Scan(&channel); err != nil {
		return "", fmt.Errorf("load conversation channel: %w", err)
	}
	return fallback(channel, "Web"), nil
}

func (s *PostgresStore) listGaps() ([]KnowledgeGap, error) {
	rows, err := s.pool.Query(context.Background(), `
		select id, conversation_id, question, reason, status, priority, created_at
		from knowledge_gaps
		order by created_at desc
		limit 50
	`)
	if err != nil {
		return nil, fmt.Errorf("list knowledge gaps: %w", err)
	}
	defer rows.Close()

	gaps := make([]KnowledgeGap, 0)
	for rows.Next() {
		var item KnowledgeGap
		var createdAt time.Time
		if err := rows.Scan(&item.ID, &item.ConversationID, &item.Question, &item.Reason, &item.Status, &item.Priority, &createdAt); err != nil {
			return nil, err
		}
		item.CreatedAt = createdAt.UTC().Format(time.RFC3339)
		gaps = append(gaps, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return gaps, nil
}

func (s *PostgresStore) listRules() ([]Rule, error) {
	rows, err := s.pool.Query(context.Background(), `
		select id, code, name, trigger_expr, action, enabled, version, stage, hit_count, last_hit_at
		from agent_rules
		order by stage, code
	`)
	if err != nil {
		return nil, fmt.Errorf("list agent rules: %w", err)
	}
	defer rows.Close()

	rules := make([]Rule, 0)
	for rows.Next() {
		var item Rule
		var lastHitAt *time.Time
		if err := rows.Scan(
			&item.ID, &item.Code, &item.Name, &item.Trigger, &item.Action, &item.Enabled,
			&item.Version, &item.Stage, &item.HitCount, &lastHitAt,
		); err != nil {
			return nil, err
		}
		if lastHitAt != nil {
			item.LastHitAt = lastHitAt.UTC().Format(time.RFC3339)
		}
		rules = append(rules, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return rules, nil
}

func (s *PostgresStore) recordRuleHit(code string) error {
	if _, err := s.pool.Exec(context.Background(), `
		update agent_rules
		set hit_count = hit_count + 1,
		    last_hit_at = now(),
		    updated_at = now()
		where code = $1
		  and stage = 'active'
	`, code); err != nil {
		return fmt.Errorf("record rule hit: %w", err)
	}
	return nil
}

func (s *PostgresStore) listRuleApprovals() ([]RuleApproval, error) {
	rows, err := s.pool.Query(context.Background(), `
		select id, rule_code, approver, risk_level, sample_ids, sample_count, status, note, created_at
		from rule_approvals
		order by created_at desc, id desc
		limit 20
	`)
	if err != nil {
		return nil, fmt.Errorf("list rule approvals: %w", err)
	}
	defer rows.Close()

	approvals := make([]RuleApproval, 0)
	for rows.Next() {
		var item RuleApproval
		var createdAt time.Time
		if err := rows.Scan(&item.ID, &item.RuleCode, &item.Approver, &item.RiskLevel, &item.SampleIDs, &item.SampleCount, &item.Status, &item.Note, &createdAt); err != nil {
			return nil, err
		}
		item.CreatedAt = createdAt.UTC().Format(time.RFC3339)
		approvals = append(approvals, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return approvals, nil
}

func insertRuleEvent(ctx context.Context, tx pgx.Tx, event RuleReleaseEvent) error {
	createdAt, err := time.Parse(time.RFC3339, event.CreatedAt)
	if err != nil {
		createdAt = time.Now().UTC()
	}
	if _, err := tx.Exec(ctx, `
		insert into rule_release_events (id, rule_code, version, action, actor, note, created_at)
		values ($1, $2, $3, $4, $5, $6, $7)
	`, event.ID, event.RuleCode, event.Version, event.Action, event.Actor, event.Note, createdAt); err != nil {
		return fmt.Errorf("insert rule release event: %w", err)
	}
	return nil
}

func (s *PostgresStore) listRuleEvents() ([]RuleReleaseEvent, error) {
	rows, err := s.pool.Query(context.Background(), `
		select id, rule_code, version, action, actor, note, created_at
		from rule_release_events
		order by created_at desc, id desc
		limit 20
	`)
	if err != nil {
		return nil, fmt.Errorf("list rule release events: %w", err)
	}
	defer rows.Close()

	events := make([]RuleReleaseEvent, 0)
	for rows.Next() {
		var item RuleReleaseEvent
		var createdAt time.Time
		if err := rows.Scan(&item.ID, &item.RuleCode, &item.Version, &item.Action, &item.Actor, &item.Note, &createdAt); err != nil {
			return nil, err
		}
		item.CreatedAt = createdAt.UTC().Format(time.RFC3339)
		events = append(events, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return events, nil
}

func (s *PostgresStore) listTransferTickets() ([]TransferTicket, error) {
	rows, err := s.pool.Query(context.Background(), `
		select id, conversation_id, channel, question, reason, priority, status, assignee, resolution_note, created_at, resolved_at
		from transfer_tickets
		order by created_at desc
		limit 50
	`)
	if err != nil {
		return nil, fmt.Errorf("list transfer tickets: %w", err)
	}
	defer rows.Close()

	tickets := make([]TransferTicket, 0)
	for rows.Next() {
		var item TransferTicket
		var createdAt time.Time
		var resolvedAt *time.Time
		if err := rows.Scan(
			&item.ID, &item.ConversationID, &item.Channel, &item.Question, &item.Reason, &item.Priority, &item.Status,
			&item.Assignee, &item.ResolutionNote, &createdAt, &resolvedAt,
		); err != nil {
			return nil, err
		}
		item.CreatedAt = createdAt.UTC().Format(time.RFC3339)
		if resolvedAt != nil {
			item.ResolvedAt = resolvedAt.UTC().Format(time.RFC3339)
		}
		item.Events = transferEvents(item)
		tickets = append(tickets, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tickets, nil
}

func (s *PostgresStore) listChannelPolicies() ([]ChannelPolicy, error) {
	rows, err := s.pool.Query(context.Background(), `
		select channel, display_name, tone, sla_minutes, risk_boost, escalation_note, enabled
		from channel_policies
		order by sla_minutes asc, channel asc
	`)
	if err != nil {
		return nil, fmt.Errorf("list channel policies: %w", err)
	}
	defer rows.Close()

	items := make([]ChannelPolicy, 0)
	for rows.Next() {
		var item ChannelPolicy
		if err := rows.Scan(&item.Channel, &item.DisplayName, &item.Tone, &item.SLAMinutes, &item.RiskBoost, &item.EscalationNote, &item.Enabled); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return defaultChannelPolicies(), nil
	}
	return items, nil
}

func (s *PostgresStore) listChannelIntegrations() ([]ChannelIntegration, error) {
	rows, err := s.pool.Query(context.Background(), `
		select channel, display_name, enabled, secret_source, secret_ref, next_secret_ref, signature_window_seconds, replay_protection, allowed_origins, rate_limit_per_minute, rotation_hint, rotates_at, updated_at
		from channel_integrations
		order by channel asc
	`)
	if err != nil {
		return nil, fmt.Errorf("list channel integrations: %w", err)
	}
	defer rows.Close()

	items := make([]ChannelIntegration, 0)
	for rows.Next() {
		var item ChannelIntegration
		var updatedAt time.Time
		var rotatesAt *time.Time
		if err := rows.Scan(
			&item.Channel,
			&item.DisplayName,
			&item.Enabled,
			&item.SecretSource,
			&item.SecretRef,
			&item.NextSecretRef,
			&item.SignatureWindowSeconds,
			&item.ReplayProtection,
			&item.AllowedOrigins,
			&item.RateLimitPerMinute,
			&item.RotationHint,
			&rotatesAt,
			&updatedAt,
		); err != nil {
			return nil, err
		}
		if rotatesAt != nil {
			item.RotatesAt = rotatesAt.UTC().Format(time.RFC3339)
		}
		item.UpdatedAt = updatedAt.UTC().Format(time.RFC3339)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return defaultChannelIntegrations(time.Now().UTC().Format(time.RFC3339)), nil
	}
	return items, nil
}

func (s *PostgresStore) listChannelAlerts() ([]ChannelAlert, error) {
	return s.listChannelAlertsTx(context.Background(), s.pool)
}

type queryer interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func (s *PostgresStore) listChannelAlertsTx(ctx context.Context, q queryer) ([]ChannelAlert, error) {
	rows, err := q.Query(ctx, `
		select channel, code, count(*)::int, max(created_at)
		from channel_failure_events
		where created_at >= now() - interval '24 hours'
		group by channel, code
		order by max(created_at) desc
		limit 20
	`)
	if err != nil {
		return nil, fmt.Errorf("list channel alerts: %w", err)
	}
	defer rows.Close()

	alerts := make([]ChannelAlert, 0)
	for rows.Next() {
		var alert ChannelAlert
		var lastSeenAt time.Time
		if err := rows.Scan(&alert.Channel, &alert.Code, &alert.Count, &lastSeenAt); err != nil {
			return nil, err
		}
		alert.LastSeenAt = lastSeenAt.UTC().Format(time.RFC3339)
		if err := q.QueryRow(ctx, `
			select reason, origin
			from channel_failure_events
			where channel = $1 and code = $2
			order by created_at desc
			limit 1
		`, alert.Channel, alert.Code).Scan(&alert.LastReason, &alert.LastOrigin); err != nil {
			return nil, err
		}
		alerts = append(alerts, alert)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return alerts, nil
}

func (s *PostgresStore) listChannelFailureTrends() ([]ChannelFailureTrend, error) {
	rows, err := s.pool.Query(context.Background(), `
		select channel, date_trunc('hour', created_at) as bucket_start, count(*)::int
		from channel_failure_events
		where created_at >= now() - interval '4 hours'
		group by channel, bucket_start
		order by bucket_start desc, channel
		limit 40
	`)
	if err != nil {
		return nil, fmt.Errorf("list channel failure trends: %w", err)
	}
	defer rows.Close()

	trends := make([]ChannelFailureTrend, 0)
	for rows.Next() {
		var item ChannelFailureTrend
		var bucketStart time.Time
		if err := rows.Scan(&item.Channel, &bucketStart, &item.Count); err != nil {
			return nil, err
		}
		item.BucketStart = bucketStart.UTC().Format(time.RFC3339)
		trends = append(trends, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return trends, nil
}

func (s *PostgresStore) listChannelAlertPolicies(alerts []ChannelAlert) ([]ChannelAlertPolicy, error) {
	return s.listChannelAlertPoliciesTx(context.Background(), s.pool, alerts)
}

func (s *PostgresStore) listChannelAlertPoliciesTx(ctx context.Context, q queryer, alerts []ChannelAlert) ([]ChannelAlertPolicy, error) {
	rows, err := q.Query(ctx, `
		select channel, severity, threshold, window_minutes, notify_target, enabled
		from channel_alert_policies
		order by channel
	`)
	if err != nil {
		return nil, fmt.Errorf("list channel alert policies: %w", err)
	}
	defer rows.Close()

	policies := make([]ChannelAlertPolicy, 0)
	for rows.Next() {
		var item ChannelAlertPolicy
		if err := rows.Scan(&item.Channel, &item.Severity, &item.Threshold, &item.WindowMinutes, &item.NotifyTarget, &item.Enabled); err != nil {
			return nil, err
		}
		policies = append(policies, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(policies) == 0 {
		policies = defaultChannelAlertPolicies()
	}
	return channelAlertPolicies(policies, alerts), nil
}

func insertTriggeredNotifications(ctx context.Context, tx pgx.Tx, policies []ChannelAlertPolicy, createdAt string) error {
	for _, policy := range policies {
		if !policy.Active {
			continue
		}
		var exists bool
		if err := tx.QueryRow(ctx, `
			select exists(
				select 1
				from channel_notifications
				where channel = $1
				  and status = 'OPEN'
			)
		`, policy.Channel).Scan(&exists); err != nil {
			return fmt.Errorf("check open channel notification: %w", err)
		}
		if exists {
			continue
		}
		notification := newChannelNotification(policy, createdAt)
		parsedCreatedAt, err := time.Parse(time.RFC3339, notification.CreatedAt)
		if err != nil {
			parsedCreatedAt = time.Now().UTC()
		}
		if _, err := tx.Exec(ctx, `
			insert into channel_notifications (id, channel, severity, target, status, reason, count, created_at)
			values ($1, $2, $3, $4, $5, $6, $7, $8)
		`, notification.ID, notification.Channel, notification.Severity, notification.Target, notification.Status, notification.Reason, notification.Count, parsedCreatedAt); err != nil {
			return fmt.Errorf("insert channel notification: %w", err)
		}
	}
	return nil
}

func (s *PostgresStore) listChannelNotifications() ([]ChannelNotification, error) {
	rows, err := s.pool.Query(context.Background(), `
		select id, channel, severity, target, status, reason, count, created_at, acked_by, ack_note, acked_at
		from channel_notifications
		order by created_at desc, id desc
		limit 30
	`)
	if err != nil {
		return nil, fmt.Errorf("list channel notifications: %w", err)
	}
	defer rows.Close()

	items := make([]ChannelNotification, 0)
	for rows.Next() {
		var item ChannelNotification
		var createdAt time.Time
		var ackedAt *time.Time
		if err := rows.Scan(
			&item.ID, &item.Channel, &item.Severity, &item.Target, &item.Status, &item.Reason, &item.Count,
			&createdAt, &item.AckedBy, &item.AckNote, &ackedAt,
		); err != nil {
			return nil, err
		}
		item.CreatedAt = createdAt.UTC().Format(time.RFC3339)
		if ackedAt != nil {
			item.AckedAt = ackedAt.UTC().Format(time.RFC3339)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (s *PostgresStore) listRecentMessages() ([]Message, error) {
	rows, err := s.pool.Query(context.Background(), `
		select id, conversation_id, role, content, engine, safe, fallback_reason, evidence_ids, created_at, trace
		from conversation_messages
		order by created_at desc, id desc
		limit 200
	`)
	if err != nil {
		return nil, fmt.Errorf("list recent messages: %w", err)
	}
	defer rows.Close()
	return scanMessages(rows)
}

func (s *PostgresStore) listAnnotations() ([]Annotation, error) {
	rows, err := s.pool.Query(context.Background(), `
		select id, message_id, reviewer, verdict, note, groundedness, safety, helpfulness, tags, score, created_at
		from message_annotations
		order by created_at desc, id desc
		limit 50
	`)
	if err != nil {
		return nil, fmt.Errorf("list annotations: %w", err)
	}
	defer rows.Close()

	items := make([]Annotation, 0)
	for rows.Next() {
		var item Annotation
		var createdAt time.Time
		if err := rows.Scan(
			&item.ID, &item.MessageID, &item.Reviewer, &item.Verdict, &item.Note,
			&item.Dimensions.Groundedness, &item.Dimensions.Safety, &item.Dimensions.Helpfulness,
			&item.Tags, &item.Score, &createdAt,
		); err != nil {
			return nil, err
		}
		item.CreatedAt = createdAt.UTC().Format(time.RFC3339)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (s *PostgresStore) listReviewTasks() ([]ReviewTask, error) {
	rows, err := s.pool.Query(context.Background(), `
		select id, message_id, conversation_id, channel, assignee, status, priority, reason, created_at, completed_at
		from review_tasks
		order by
		  case when status = 'COMPLETED' then 1 else 0 end,
		  created_at desc,
		  id desc
		limit 50
	`)
	if err != nil {
		return nil, fmt.Errorf("list review tasks: %w", err)
	}
	defer rows.Close()

	items := make([]ReviewTask, 0)
	for rows.Next() {
		item, err := scanReviewTask(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (s *PostgresStore) updateReviewTask(query string, args ...any) (ReviewTask, error) {
	row := s.pool.QueryRow(context.Background(), query, args...)
	item, err := scanReviewTask(row)
	if err != nil {
		return ReviewTask{}, err
	}
	return item, nil
}

type reviewTaskScanner interface {
	Scan(dest ...any) error
}

func scanReviewTask(row reviewTaskScanner) (ReviewTask, error) {
	var item ReviewTask
	var createdAt time.Time
	var completedAt *time.Time
	if err := row.Scan(
		&item.ID, &item.MessageID, &item.ConversationID, &item.Channel, &item.Assignee,
		&item.Status, &item.Priority, &item.Reason, &createdAt, &completedAt,
	); err != nil {
		return ReviewTask{}, err
	}
	item.CreatedAt = createdAt.UTC().Format(time.RFC3339)
	if completedAt != nil {
		item.CompletedAt = completedAt.UTC().Format(time.RFC3339)
	}
	return item, nil
}

func scanMessages(rows pgx.Rows) ([]Message, error) {
	items := make([]Message, 0)
	for rows.Next() {
		item, err := scanMessage(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func scanMessage(rows pgx.Rows) (Message, error) {
	var item Message
	var createdAt time.Time
	var traceBytes []byte
	if err := rows.Scan(
		&item.ID, &item.ConversationID, &item.Role, &item.Content, &item.Engine,
		&item.Safe, &item.FallbackReason, &item.EvidenceIDs, &createdAt, &traceBytes,
	); err != nil {
		return Message{}, fmt.Errorf("scan message: %w", err)
	}
	item.CreatedAt = createdAt.UTC().Format(time.RFC3339)
	if len(traceBytes) > 0 && string(traceBytes) != "{}" {
		var trace AgentTrace
		if err := json.Unmarshal(traceBytes, &trace); err != nil {
			return Message{}, fmt.Errorf("decode message trace: %w", err)
		}
		item.Trace = &trace
	}
	return item, nil
}

func scanConversations(rows pgx.Rows) ([]Conversation, error) {
	items := make([]Conversation, 0)
	for rows.Next() {
		var item Conversation
		var startedAt time.Time
		if err := rows.Scan(&item.ID, &item.Customer, &item.Channel, &item.Intent, &item.Status, &item.RiskLevel, &startedAt, &item.LastMessage); err != nil {
			return nil, err
		}
		item.StartedAt = startedAt.UTC().Format(time.RFC3339)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func scanKnowledge(rows pgx.Rows) ([]KnowledgeArticle, error) {
	items := make([]KnowledgeArticle, 0)
	for rows.Next() {
		var item KnowledgeArticle
		var updatedAt time.Time
		if err := rows.Scan(&item.ID, &item.Title, &item.Category, &item.Content, &item.Tags, &item.TrustLevel, &updatedAt); err != nil {
			return nil, err
		}
		item.UpdatedAt = updatedAt.UTC().Format(time.RFC3339)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func postgresAgentReply(generator ReplyGenerator, conversationID, content string, evidence []KnowledgeArticle, history []Message, ruleResult RuleTestResult, now time.Time) (Message, *KnowledgeGap) {
	if ruleResult.Action == "recommend_human_transfer" {
		return Message{
			ID: fmt.Sprintf("msg_%d_agent", now.UnixNano()), ConversationID: conversationID, Role: "assistant",
			Content: "我已经为你转接人工客服，并保留当前对话上下文。人工客服接手前，我不会编造处理结果。",
			Engine:  "rule", Safe: true, FallbackReason: fallback(ruleResult.RuleCode, "TRANSFER_THRESHOLD"), CreatedAt: now.Format(time.RFC3339),
			Trace: newAgentTrace("human_transfer", evidence, history),
		}, nil
	}
	if len(evidence) == 0 {
		gap := KnowledgeGap{
			ID: fmt.Sprintf("gap_%d", now.UnixNano()), ConversationID: conversationID,
			Question: content, Reason: "NO_EVIDENCE", Status: "OPEN", Priority: "MEDIUM", CreatedAt: now.Format(time.RFC3339),
		}
		return Message{
			ID: fmt.Sprintf("msg_%d_agent", now.UnixNano()), ConversationID: conversationID, Role: "assistant",
			Content: "这个问题我还没有找到可靠知识依据，已记录给运营补充知识。为了避免误导，我先不直接下结论。",
			Engine:  "rule", Safe: true, FallbackReason: "NO_EVIDENCE", CreatedAt: now.Format(time.RFC3339),
			Trace: newAgentTrace("no_evidence_fallback", evidence, history),
		}, &gap
	}
	ids := make([]string, 0, len(evidence))
	for _, item := range evidence {
		ids = append(ids, item.ID)
	}
	trace := newAgentTrace("evidence_answer", evidence, history)
	if generator != nil {
		trace.ModelAttempted = true
		startedAt := time.Now()
		generated, err := generator.GenerateReply(context.Background(), ReplyRequest{
			ConversationID: conversationID,
			Question:       content,
			Evidence:       evidence,
			History:        history,
		})
		trace.ModelDurationMs = time.Since(startedAt).Milliseconds()
		trace.Model = strings.TrimSpace(generated.Model)
		if err == nil && strings.TrimSpace(generated.Content) != "" {
			return Message{
				ID: fmt.Sprintf("msg_%d_agent", now.UnixNano()), ConversationID: conversationID, Role: "assistant",
				Content: strings.TrimSpace(generated.Content), Engine: "llm+rag", Safe: true, EvidenceIDs: ids, CreatedAt: now.Format(time.RFC3339), Trace: trace,
			}, nil
		}
		trace.ModelFallback = true
		if err != nil {
			trace.ModelFallbackReason = err.Error()
		} else {
			trace.ModelFallbackReason = "empty model response"
		}
	}
	return Message{
		ID: fmt.Sprintf("msg_%d_agent", now.UnixNano()), ConversationID: conversationID, Role: "assistant",
		Content: fmt.Sprintf("根据知识库《%s》：%s", evidence[0].Title, evidence[0].Content),
		Engine:  "rag+rule", Safe: true, EvidenceIDs: ids, CreatedAt: now.Format(time.RFC3339), Trace: trace,
	}, nil
}

func messageTraceJSON(trace *AgentTrace) []byte {
	if trace == nil {
		return []byte("{}")
	}
	payload, err := json.Marshal(trace)
	if err != nil {
		return []byte("{}")
	}
	return payload
}

func postgresConversationState(conversationID, channel, content string, evidence []KnowledgeArticle, gap *KnowledgeGap, now time.Time) Conversation {
	status := "Active"
	risk := "LOW"
	intent := "知识问答"
	if gap != nil {
		status = "KnowledgeGap"
		risk = "MEDIUM"
		intent = "待补知识"
	}
	if shouldTransfer(content) {
		status = "NeedsHuman"
		risk = "HIGH"
		intent = "人工协助"
	}
	if len(evidence) > 0 {
		intent = evidence[0].Category
	}
	return Conversation{
		ID: conversationID, Customer: "访客", Channel: fallback(channel, "Web"), Intent: intent,
		Status: status, RiskLevel: risk, StartedAt: now.Format(time.RFC3339), LastMessage: content,
	}
}
