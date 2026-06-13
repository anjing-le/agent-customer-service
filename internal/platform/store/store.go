package store

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

type ReplyGenerator interface {
	GenerateReply(ctx context.Context, req ReplyRequest) (string, error)
}

type ReplyRequest struct {
	ConversationID string
	Question       string
	Evidence       []KnowledgeArticle
	History        []Message
}

type Option func(*Store)

func WithReplyGenerator(generator ReplyGenerator) Option {
	return func(s *Store) {
		s.generator = generator
	}
}

type Runtime interface {
	ListConversations() ([]Conversation, error)
	CreateConversation(customer, channel string) (Conversation, error)
	ListMessages(conversationID string) ([]Message, error)
	SendMessage(conversationID, content string) (SendMessageResult, error)
	ListKnowledge() ([]KnowledgeArticle, error)
	SearchKnowledge(query string) ([]KnowledgeArticle, error)
	ResolveKnowledgeGap(id string) (KnowledgeGap, error)
	CreateArticleFromGap(gapID, title, category, content string, tags []string) (KnowledgeArticle, error)
	TestRule(content string) (RuleTestResult, error)
	ResolveTransferTicket(id, assignee, note string) (TransferTicket, error)
	Dashboard() (Dashboard, error)
}

type Store struct {
	mu sync.RWMutex

	conversations []Conversation
	messages      []Message
	knowledge     []KnowledgeArticle
	gaps          []KnowledgeGap
	rules         []Rule
	tickets       []TransferTicket
	generator     ReplyGenerator
}

type Conversation struct {
	ID          string `json:"id"`
	Customer    string `json:"customer"`
	Channel     string `json:"channel"`
	Intent      string `json:"intent"`
	Status      string `json:"status"`
	RiskLevel   string `json:"riskLevel"`
	StartedAt   string `json:"startedAt"`
	LastMessage string `json:"lastMessage"`
}

type Message struct {
	ID             string   `json:"id"`
	ConversationID string   `json:"conversationId"`
	Role           string   `json:"role"`
	Content        string   `json:"content"`
	Engine         string   `json:"engine"`
	Safe           bool     `json:"safe"`
	FallbackReason string   `json:"fallbackReason,omitempty"`
	EvidenceIDs    []string `json:"evidenceIds,omitempty"`
	CreatedAt      string   `json:"createdAt"`
}

type KnowledgeArticle struct {
	ID         string   `json:"id"`
	Title      string   `json:"title"`
	Category   string   `json:"category"`
	Content    string   `json:"content"`
	Tags       []string `json:"tags"`
	TrustLevel string   `json:"trustLevel"`
	UpdatedAt  string   `json:"updatedAt"`
}

type KnowledgeGap struct {
	ID             string `json:"id"`
	ConversationID string `json:"conversationId"`
	Question       string `json:"question"`
	Reason         string `json:"reason"`
	Status         string `json:"status"`
	Priority       string `json:"priority"`
	CreatedAt      string `json:"createdAt"`
}

type Rule struct {
	ID      string `json:"id"`
	Code    string `json:"code"`
	Name    string `json:"name"`
	Trigger string `json:"trigger"`
	Action  string `json:"action"`
	Enabled bool   `json:"enabled"`
}

type RuleTestResult struct {
	Input       string `json:"input"`
	Matched     bool   `json:"matched"`
	RuleCode    string `json:"ruleCode,omitempty"`
	Action      string `json:"action"`
	RiskLevel   string `json:"riskLevel"`
	Fallback    bool   `json:"fallback"`
	Reason      string `json:"reason"`
	Recommended string `json:"recommended"`
}

type TransferTicket struct {
	ID             string `json:"id"`
	ConversationID string `json:"conversationId"`
	Question       string `json:"question"`
	Reason         string `json:"reason"`
	Priority       string `json:"priority"`
	Status         string `json:"status"`
	Assignee       string `json:"assignee,omitempty"`
	ResolutionNote string `json:"resolutionNote,omitempty"`
	CreatedAt      string `json:"createdAt"`
	ResolvedAt     string `json:"resolvedAt,omitempty"`
}

type Dashboard struct {
	Metrics       []Metric         `json:"metrics"`
	Conversations []Conversation   `json:"conversations"`
	KnowledgeGaps []KnowledgeGap   `json:"knowledgeGaps"`
	Rules         []Rule           `json:"rules"`
	Transfers     []TransferTicket `json:"transfers"`
}

type Metric struct {
	Label string `json:"label"`
	Value string `json:"value"`
	Note  string `json:"note"`
}

type SendMessageResult struct {
	Conversation Conversation       `json:"conversation"`
	UserMessage  Message            `json:"userMessage"`
	AgentMessage Message            `json:"agentMessage"`
	Evidence     []KnowledgeArticle `json:"evidence"`
	Gap          *KnowledgeGap      `json:"gap,omitempty"`
}

func NewSeedStore(options ...Option) *Store {
	now := time.Now().UTC().Format(time.RFC3339)
	st := &Store{
		conversations: []Conversation{
			{
				ID: "conv_demo_refund", Customer: "林夏", Channel: "Web", Intent: "退款规则",
				Status: "Active", RiskLevel: "LOW", StartedAt: now,
				LastMessage: "7 天无理由退货的运费怎么计算？",
			},
			{
				ID: "conv_demo_transfer", Customer: "周辰", Channel: "WeChat", Intent: "人工协助",
				Status: "NeedsHuman", RiskLevel: "HIGH", StartedAt: now,
				LastMessage: "我已经催了三次，必须马上找人工处理。",
			},
		},
		knowledge: []KnowledgeArticle{
			{ID: "kb_refund_7d", Title: "7 天无理由退货", Category: "售后", Content: "签收后 7 天内可申请无理由退货；非质量问题由用户承担寄回运费，质量问题由商家承担。", Tags: []string{"退款", "退货", "运费"}, TrustLevel: "HIGH", UpdatedAt: now},
			{ID: "kb_invoice", Title: "电子发票开具", Category: "订单", Content: "订单完成后可在订单详情申请电子发票，通常 24 小时内发送到预留邮箱。", Tags: []string{"发票", "订单"}, TrustLevel: "HIGH", UpdatedAt: now},
			{ID: "kb_vip_transfer", Title: "高风险投诉转人工", Category: "服务规则", Content: "出现强烈投诉、法律风险、连续催办等信号时，Agent 应停止自由生成并转人工。", Tags: []string{"投诉", "人工", "风险"}, TrustLevel: "HIGH", UpdatedAt: now},
		},
		rules: []Rule{
			{ID: "rule_low_evidence", Code: "NO_EVIDENCE_FALLBACK", Name: "无可靠证据兜底", Trigger: "knowledge evidence empty", Action: "safe_fallback_and_create_gap", Enabled: true},
			{ID: "rule_human_transfer", Code: "TRANSFER_THRESHOLD", Name: "转人工阈值", Trigger: "投诉/催办/法律风险", Action: "recommend_human_transfer", Enabled: true},
		},
	}
	st.messages = []Message{
		{ID: "msg_demo_1", ConversationID: "conv_demo_refund", Role: "user", Content: "7 天无理由退货的运费怎么计算？", Engine: "customer", Safe: true, CreatedAt: now},
		{ID: "msg_demo_2", ConversationID: "conv_demo_refund", Role: "assistant", Content: "根据售后知识库，签收 7 天内可申请无理由退货；非质量问题寄回运费通常由用户承担，质量问题由商家承担。", Engine: "rule+rAG", Safe: true, EvidenceIDs: []string{"kb_refund_7d"}, CreatedAt: now},
	}
	st.tickets = []TransferTicket{
		{ID: "ticket_demo_transfer", ConversationID: "conv_demo_transfer", Question: "我已经催了三次，必须马上找人工处理。", Reason: "TRANSFER_THRESHOLD", Priority: "HIGH", Status: "OPEN", CreatedAt: now},
	}
	for _, option := range options {
		option(st)
	}
	return st
}

func (s *Store) ListConversations() ([]Conversation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]Conversation(nil), s.conversations...), nil
}

func (s *Store) CreateConversation(customer, channel string) (Conversation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC().Format(time.RFC3339)
	conv := Conversation{
		ID: fmt.Sprintf("conv_%d", time.Now().UnixNano()), Customer: customer,
		Channel: fallback(channel, "Web"), Intent: "待识别", Status: "Active",
		RiskLevel: "LOW", StartedAt: now,
	}
	s.conversations = append([]Conversation{conv}, s.conversations...)
	return conv, nil
}

func (s *Store) ListMessages(conversationID string) ([]Message, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]Message, 0)
	for _, item := range s.messages {
		if item.ConversationID == conversationID {
			items = append(items, item)
		}
	}
	return items, nil
}

func (s *Store) SendMessage(conversationID, content string) (SendMessageResult, error) {
	s.mu.Lock()
	now := time.Now().UTC().Format(time.RFC3339)
	if conversationID == "" {
		conversationID = fmt.Sprintf("conv_%d", time.Now().UnixNano())
		s.conversations = append([]Conversation{{
			ID: conversationID, Customer: "访客", Channel: "Web", Intent: "待识别",
			Status: "Active", RiskLevel: "LOW", StartedAt: now,
		}}, s.conversations...)
	}

	userMessage := Message{ID: fmt.Sprintf("msg_%d_user", time.Now().UnixNano()), ConversationID: conversationID, Role: "user", Content: content, Engine: "customer", Safe: true, CreatedAt: now}
	evidence := s.searchLocked(content)
	history := s.messagesLocked(conversationID)
	generator := s.generator
	s.mu.Unlock()

	agentMessage, gap := agentReply(generator, conversationID, content, evidence, history, now)

	s.mu.Lock()
	defer s.mu.Unlock()
	if gap != nil {
		s.gaps = append([]KnowledgeGap{*gap}, s.gaps...)
	}
	s.messages = append(s.messages, userMessage, agentMessage)
	if agentMessage.FallbackReason == "TRANSFER_THRESHOLD" {
		s.tickets = append([]TransferTicket{newTransferTicket(conversationID, content, now)}, s.tickets...)
	}

	conv := s.touchConversationLocked(conversationID, content, evidence, gap)
	return SendMessageResult{Conversation: conv, UserMessage: userMessage, AgentMessage: agentMessage, Evidence: evidence, Gap: gap}, nil
}

func (s *Store) ListKnowledge() ([]KnowledgeArticle, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]KnowledgeArticle(nil), s.knowledge...), nil
}

func (s *Store) SearchKnowledge(query string) ([]KnowledgeArticle, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.searchLocked(query), nil
}

func (s *Store) ResolveKnowledgeGap(id string) (KnowledgeGap, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for idx := range s.gaps {
		if s.gaps[idx].ID == id {
			s.gaps[idx].Status = "RESOLVED"
			return s.gaps[idx], nil
		}
	}
	return KnowledgeGap{}, fmt.Errorf("knowledge gap %s not found", id)
}

func (s *Store) CreateArticleFromGap(gapID, title, category, content string, tags []string) (KnowledgeArticle, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	gapIndex := -1
	for idx := range s.gaps {
		if s.gaps[idx].ID == gapID {
			gapIndex = idx
			break
		}
	}
	if gapIndex < 0 {
		return KnowledgeArticle{}, fmt.Errorf("knowledge gap %s not found", gapID)
	}
	now := time.Now().UTC().Format(time.RFC3339)
	article := KnowledgeArticle{
		ID:         fmt.Sprintf("kb_%d", time.Now().UnixNano()),
		Title:      fallback(title, s.gaps[gapIndex].Question),
		Category:   fallback(category, "运营补充"),
		Content:    fallback(content, s.gaps[gapIndex].Question),
		Tags:       normalizeTags(tags, s.gaps[gapIndex].Question),
		TrustLevel: "HIGH",
		UpdatedAt:  now,
	}
	s.knowledge = append([]KnowledgeArticle{article}, s.knowledge...)
	s.gaps[gapIndex].Status = "RESOLVED"
	return article, nil
}

func (s *Store) TestRule(content string) (RuleTestResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return evaluateRules(content, s.searchLocked(content), s.rules), nil
}

func (s *Store) ResolveTransferTicket(id, assignee, note string) (TransferTicket, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for idx := range s.tickets {
		if s.tickets[idx].ID == id {
			s.tickets[idx].Status = "RESOLVED"
			s.tickets[idx].Assignee = fallback(assignee, "operator")
			s.tickets[idx].ResolutionNote = fallback(note, "人工已处理")
			s.tickets[idx].ResolvedAt = time.Now().UTC().Format(time.RFC3339)
			return s.tickets[idx], nil
		}
	}
	return TransferTicket{}, fmt.Errorf("transfer ticket %s not found", id)
}

func (s *Store) Dashboard() (Dashboard, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	openGaps := 0
	for _, gap := range s.gaps {
		if gap.Status == "OPEN" {
			openGaps++
		}
	}
	openTransfers := 0
	for _, ticket := range s.tickets {
		if ticket.Status == "OPEN" {
			openTransfers++
		}
	}
	return Dashboard{
		Metrics: []Metric{
			{Label: "Active sessions", Value: fmt.Sprintf("%d", len(s.conversations)), Note: "in-memory V1 runtime"},
			{Label: "Knowledge items", Value: fmt.Sprintf("%d", len(s.knowledge)), Note: "seeded trusted articles"},
			{Label: "Open gaps", Value: fmt.Sprintf("%d", openGaps), Note: "created by no-evidence fallback"},
			{Label: "Open transfers", Value: fmt.Sprintf("%d", openTransfers), Note: "waiting for human agents"},
			{Label: "Enabled rules", Value: fmt.Sprintf("%d", enabledRules(s.rules)), Note: "guardrail and transfer policies"},
		},
		Conversations: append([]Conversation(nil), s.conversations...),
		KnowledgeGaps: append([]KnowledgeGap(nil), s.gaps...),
		Rules:         append([]Rule(nil), s.rules...),
		Transfers:     append([]TransferTicket(nil), s.tickets...),
	}, nil
}

func (s *Store) searchLocked(query string) []KnowledgeArticle {
	query = strings.ToLower(query)
	if strings.TrimSpace(query) == "" {
		return nil
	}
	matches := make([]KnowledgeArticle, 0)
	for _, item := range s.knowledge {
		haystack := strings.ToLower(item.Title + " " + item.Category + " " + item.Content + " " + strings.Join(item.Tags, " "))
		if strings.Contains(query, strings.ToLower(item.Title)) || strings.Contains(query, strings.ToLower(item.Category)) {
			matches = append(matches, item)
			continue
		}
		tagMatched := false
		for _, tag := range item.Tags {
			if strings.Contains(query, strings.ToLower(tag)) {
				matches = append(matches, item)
				tagMatched = true
				break
			}
		}
		if tagMatched {
			continue
		}
		for _, token := range strings.Fields(query) {
			if strings.Contains(haystack, token) {
				matches = append(matches, item)
				break
			}
		}
	}
	return matches
}

func agentReply(generator ReplyGenerator, conversationID, content string, evidence []KnowledgeArticle, history []Message, now string) (Message, *KnowledgeGap) {
	if shouldTransfer(content) {
		return Message{
			ID: fmt.Sprintf("msg_%d_agent", time.Now().UnixNano()), ConversationID: conversationID, Role: "assistant",
			Content: "我已经为你转接人工客服，并保留当前对话上下文。人工客服接手前，我不会编造处理结果。",
			Engine:  "rule", Safe: true, FallbackReason: "TRANSFER_THRESHOLD", CreatedAt: now,
		}, nil
	}
	if len(evidence) == 0 {
		gap := KnowledgeGap{
			ID: fmt.Sprintf("gap_%d", time.Now().UnixNano()), ConversationID: conversationID,
			Question: content, Reason: "NO_EVIDENCE", Status: "OPEN", Priority: "MEDIUM", CreatedAt: now,
		}
		return Message{
			ID: fmt.Sprintf("msg_%d_agent", time.Now().UnixNano()), ConversationID: conversationID, Role: "assistant",
			Content: "这个问题我还没有找到可靠知识依据，已记录给运营补充知识。为了避免误导，我先不直接下结论。",
			Engine:  "rule", Safe: true, FallbackReason: "NO_EVIDENCE", CreatedAt: now,
		}, &gap
	}
	ids := make([]string, 0, len(evidence))
	for _, item := range evidence {
		ids = append(ids, item.ID)
	}
	if generator != nil {
		reply, err := generator.GenerateReply(context.Background(), ReplyRequest{
			ConversationID: conversationID,
			Question:       content,
			Evidence:       evidence,
			History:        history,
		})
		if err == nil && strings.TrimSpace(reply) != "" {
			return Message{
				ID: fmt.Sprintf("msg_%d_agent", time.Now().UnixNano()), ConversationID: conversationID, Role: "assistant",
				Content: strings.TrimSpace(reply), Engine: "llm+rag", Safe: true, EvidenceIDs: ids, CreatedAt: now,
			}, nil
		}
	}
	return Message{
		ID: fmt.Sprintf("msg_%d_agent", time.Now().UnixNano()), ConversationID: conversationID, Role: "assistant",
		Content: fmt.Sprintf("根据知识库《%s》：%s", evidence[0].Title, evidence[0].Content),
		Engine:  "rag+rule", Safe: true, EvidenceIDs: ids, CreatedAt: now,
	}, nil
}

func (s *Store) messagesLocked(conversationID string) []Message {
	items := make([]Message, 0)
	for _, item := range s.messages {
		if item.ConversationID == conversationID {
			items = append(items, item)
		}
	}
	return items
}

func (s *Store) touchConversationLocked(conversationID, content string, evidence []KnowledgeArticle, gap *KnowledgeGap) Conversation {
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
	for idx := range s.conversations {
		if s.conversations[idx].ID == conversationID {
			s.conversations[idx].Status = status
			s.conversations[idx].RiskLevel = risk
			s.conversations[idx].Intent = intent
			s.conversations[idx].LastMessage = content
			return s.conversations[idx]
		}
	}
	conv := Conversation{ID: conversationID, Customer: "访客", Channel: "Web", Intent: intent, Status: status, RiskLevel: risk, StartedAt: time.Now().UTC().Format(time.RFC3339), LastMessage: content}
	s.conversations = append([]Conversation{conv}, s.conversations...)
	return conv
}

func shouldTransfer(content string) bool {
	text := strings.ToLower(content)
	return strings.Contains(text, "人工") || strings.Contains(text, "投诉") || strings.Contains(text, "法律") || strings.Contains(text, "催")
}

func enabledRules(rules []Rule) int {
	total := 0
	for _, rule := range rules {
		if rule.Enabled {
			total++
		}
	}
	return total
}

func fallback(value, defaultValue string) string {
	if strings.TrimSpace(value) == "" {
		return defaultValue
	}
	return value
}

func normalizeTags(tags []string, fallbackTag string) []string {
	result := make([]string, 0, len(tags)+1)
	seen := map[string]bool{}
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" || seen[tag] {
			continue
		}
		seen[tag] = true
		result = append(result, tag)
	}
	if len(result) == 0 {
		result = append(result, fallbackTag)
	}
	return result
}

func evaluateRules(content string, evidence []KnowledgeArticle, rules []Rule) RuleTestResult {
	result := RuleTestResult{
		Input:       content,
		Matched:     false,
		Action:      "allow_answer",
		RiskLevel:   "LOW",
		Fallback:    false,
		Reason:      "EVIDENCE_OK",
		Recommended: "可以基于可靠知识回答。",
	}
	if strings.TrimSpace(content) == "" {
		result.Matched = true
		result.RuleCode = "EMPTY_INPUT"
		result.Action = "reject_request"
		result.RiskLevel = "LOW"
		result.Fallback = true
		result.Reason = "EMPTY_INPUT"
		result.Recommended = "请输入用户问题后再测试规则。"
		return result
	}
	if shouldTransfer(content) && ruleEnabled(rules, "TRANSFER_THRESHOLD") {
		result.Matched = true
		result.RuleCode = "TRANSFER_THRESHOLD"
		result.Action = "recommend_human_transfer"
		result.RiskLevel = "HIGH"
		result.Fallback = true
		result.Reason = "用户问题触发投诉、催办、法律风险或人工诉求。"
		result.Recommended = "停止自由生成，保留上下文并转人工处理。"
		return result
	}
	if len(evidence) == 0 && ruleEnabled(rules, "NO_EVIDENCE_FALLBACK") {
		result.Matched = true
		result.RuleCode = "NO_EVIDENCE_FALLBACK"
		result.Action = "safe_fallback_and_create_gap"
		result.RiskLevel = "MEDIUM"
		result.Fallback = true
		result.Reason = "当前问题没有命中可信知识证据。"
		result.Recommended = "返回无答案兜底话术，并沉淀知识缺口。"
		return result
	}
	return result
}

func ruleEnabled(rules []Rule, code string) bool {
	for _, rule := range rules {
		if rule.Code == code && rule.Enabled {
			return true
		}
	}
	return false
}

func newTransferTicket(conversationID, question, createdAt string) TransferTicket {
	return TransferTicket{
		ID:             fmt.Sprintf("ticket_%d", time.Now().UnixNano()),
		ConversationID: conversationID,
		Question:       question,
		Reason:         "TRANSFER_THRESHOLD",
		Priority:       "HIGH",
		Status:         "OPEN",
		CreatedAt:      createdAt,
	}
}
