package store

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{pool: pool}
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

func (s *PostgresStore) SendMessage(conversationID, content string) (SendMessageResult, error) {
	ctx := context.Background()
	now := time.Now().UTC()
	if conversationID == "" {
		conv, err := s.CreateConversation("访客", "Web")
		if err != nil {
			return SendMessageResult{}, err
		}
		conversationID = conv.ID
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
	}

	userMessage := Message{ID: fmt.Sprintf("msg_%d_user", now.UnixNano()), ConversationID: conversationID, Role: "user", Content: content, Engine: "customer", Safe: true, CreatedAt: now.Format(time.RFC3339)}
	evidence, err := s.SearchKnowledge(content)
	if err != nil {
		return SendMessageResult{}, err
	}
	agentMessage, gap := postgresAgentReply(conversationID, content, evidence, now)

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return SendMessageResult{}, fmt.Errorf("begin send message: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
		insert into conversation_messages (id, conversation_id, role, content, engine, safe, fallback_reason, evidence_ids, created_at)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, userMessage.ID, userMessage.ConversationID, userMessage.Role, userMessage.Content, userMessage.Engine, userMessage.Safe, userMessage.FallbackReason, userMessage.EvidenceIDs, now); err != nil {
		return SendMessageResult{}, fmt.Errorf("insert user message: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		insert into conversation_messages (id, conversation_id, role, content, engine, safe, fallback_reason, evidence_ids, created_at)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, agentMessage.ID, agentMessage.ConversationID, agentMessage.Role, agentMessage.Content, agentMessage.Engine, agentMessage.Safe, agentMessage.FallbackReason, agentMessage.EvidenceIDs, now); err != nil {
		return SendMessageResult{}, fmt.Errorf("insert agent message: %w", err)
	}
	if gap != nil {
		if _, err := tx.Exec(ctx, `
			insert into knowledge_gaps (id, conversation_id, question, reason, status, priority, created_at)
			values ($1, $2, $3, $4, $5, $6, $7)
		`, gap.ID, gap.ConversationID, gap.Question, gap.Reason, gap.Status, gap.Priority, now); err != nil {
			return SendMessageResult{}, fmt.Errorf("insert knowledge gap: %w", err)
		}
	}

	conv := postgresConversationState(conversationID, content, evidence, gap, now)
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

	openGaps := 0
	for _, gap := range gaps {
		if gap.Status == "OPEN" {
			openGaps++
		}
	}

	return Dashboard{
		Metrics: []Metric{
			{Label: "Active sessions", Value: fmt.Sprintf("%d", len(conversations)), Note: "PostgreSQL runtime"},
			{Label: "Knowledge items", Value: fmt.Sprintf("%d", len(knowledge)), Note: "trusted articles in PostgreSQL"},
			{Label: "Open gaps", Value: fmt.Sprintf("%d", openGaps), Note: "created by no-evidence fallback"},
			{Label: "Enabled rules", Value: fmt.Sprintf("%d", enabledRules(rules)), Note: "guardrail and transfer policies"},
		},
		Conversations: conversations,
		KnowledgeGaps: gaps,
		Rules:         rules,
	}, nil
}

func (s *PostgresStore) conversationExists(ctx context.Context, id string) (bool, error) {
	var exists bool
	if err := s.pool.QueryRow(ctx, "select exists(select 1 from conversations where id = $1)", id).Scan(&exists); err != nil {
		return false, fmt.Errorf("check conversation exists: %w", err)
	}
	return exists, nil
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

func postgresAgentReply(conversationID, content string, evidence []KnowledgeArticle, now time.Time) (Message, *KnowledgeGap) {
	if shouldTransfer(content) {
		return Message{
			ID: fmt.Sprintf("msg_%d_agent", now.UnixNano()), ConversationID: conversationID, Role: "assistant",
			Content: "我已经为你转接人工客服，并保留当前对话上下文。人工客服接手前，我不会编造处理结果。",
			Engine:  "rule", Safe: true, FallbackReason: "TRANSFER_THRESHOLD", CreatedAt: now.Format(time.RFC3339),
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
		}, &gap
	}
	ids := make([]string, 0, len(evidence))
	for _, item := range evidence {
		ids = append(ids, item.ID)
	}
	return Message{
		ID: fmt.Sprintf("msg_%d_agent", now.UnixNano()), ConversationID: conversationID, Role: "assistant",
		Content: fmt.Sprintf("根据知识库《%s》：%s", evidence[0].Title, evidence[0].Content),
		Engine:  "rag+rule", Safe: true, EvidenceIDs: ids, CreatedAt: now.Format(time.RFC3339),
	}, nil
}

func postgresConversationState(conversationID, content string, evidence []KnowledgeArticle, gap *KnowledgeGap, now time.Time) Conversation {
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
		ID: conversationID, Customer: "访客", Channel: "Web", Intent: intent,
		Status: status, RiskLevel: risk, StartedAt: now.Format(time.RFC3339), LastMessage: content,
	}
}
