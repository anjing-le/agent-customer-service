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
	history, err := s.ListMessages(conversationID)
	if err != nil {
		return SendMessageResult{}, err
	}
	agentMessage, gap := postgresAgentReply(s.generator, conversationID, content, evidence, history, now)

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
	if agentMessage.FallbackReason == "TRANSFER_THRESHOLD" {
		ticket := newTransferTicket(conversationID, channel, content, now.Format(time.RFC3339))
		if _, err := tx.Exec(ctx, `
			insert into transfer_tickets (id, conversation_id, channel, question, reason, priority, status, created_at)
			values ($1, $2, $3, $4, $5, $6, $7, $8)
		`, ticket.ID, ticket.ConversationID, ticket.Channel, ticket.Question, ticket.Reason, ticket.Priority, ticket.Status, now); err != nil {
			return SendMessageResult{}, fmt.Errorf("insert transfer ticket: %w", err)
		}
	}
	if gap != nil {
		if _, err := tx.Exec(ctx, `
			insert into knowledge_gaps (id, conversation_id, question, reason, status, priority, created_at)
			values ($1, $2, $3, $4, $5, $6, $7)
		`, gap.ID, gap.ConversationID, gap.Question, gap.Reason, gap.Status, gap.Priority, now); err != nil {
			return SendMessageResult{}, fmt.Errorf("insert knowledge gap: %w", err)
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
	return evaluateRules(content, evidence, rules), nil
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
	if _, err := s.pool.Exec(ctx, `
		insert into message_annotations (id, message_id, reviewer, verdict, note, groundedness, safety, helpfulness, tags, score, created_at)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`, annotation.ID, annotation.MessageID, annotation.Reviewer, annotation.Verdict, annotation.Note,
		annotation.Dimensions.Groundedness, annotation.Dimensions.Safety, annotation.Dimensions.Helpfulness,
		annotation.Tags, annotation.Score, createdAt); err != nil {
		return Annotation{}, fmt.Errorf("submit annotation: %w", err)
	}
	return annotation, nil
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
	transfers, err := s.listTransferTickets()
	if err != nil {
		return Dashboard{}, err
	}
	channelPolicies, err := s.listChannelPolicies()
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
		},
		Conversations:   conversations,
		KnowledgeGaps:   gaps,
		Rules:           rules,
		Transfers:       transfers,
		ChannelPolicies: channelPolicies,
		Quality:         qualitySummary(messages, gaps, transfers, annotations),
		Annotations:     annotations,
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
		select id, code, name, trigger_expr, action, enabled
		from agent_rules
		order by code
	`)
	if err != nil {
		return nil, fmt.Errorf("list agent rules: %w", err)
	}
	defer rows.Close()

	rules := make([]Rule, 0)
	for rows.Next() {
		var item Rule
		if err := rows.Scan(&item.ID, &item.Code, &item.Name, &item.Trigger, &item.Action, &item.Enabled); err != nil {
			return nil, err
		}
		rules = append(rules, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return rules, nil
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

func postgresAgentReply(generator ReplyGenerator, conversationID, content string, evidence []KnowledgeArticle, history []Message, now time.Time) (Message, *KnowledgeGap) {
	if shouldTransfer(content) {
		return Message{
			ID: fmt.Sprintf("msg_%d_agent", now.UnixNano()), ConversationID: conversationID, Role: "assistant",
			Content: "我已经为你转接人工客服，并保留当前对话上下文。人工客服接手前，我不会编造处理结果。",
			Engine:  "rule", Safe: true, FallbackReason: "TRANSFER_THRESHOLD", CreatedAt: now.Format(time.RFC3339),
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
