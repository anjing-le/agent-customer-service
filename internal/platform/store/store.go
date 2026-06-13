package store

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

type Store struct {
	mu sync.RWMutex

	conversations []Conversation
	messages      []Message
	knowledge     []KnowledgeArticle
	gaps          []KnowledgeGap
	rules         []Rule
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

type Dashboard struct {
	Metrics       []Metric       `json:"metrics"`
	Conversations []Conversation `json:"conversations"`
	KnowledgeGaps []KnowledgeGap `json:"knowledgeGaps"`
	Rules         []Rule         `json:"rules"`
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

func NewSeedStore() *Store {
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
	return st
}

func (s *Store) ListConversations() []Conversation {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]Conversation(nil), s.conversations...)
}

func (s *Store) CreateConversation(customer, channel string) Conversation {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC().Format(time.RFC3339)
	conv := Conversation{
		ID: fmt.Sprintf("conv_%d", time.Now().UnixNano()), Customer: customer,
		Channel: fallback(channel, "Web"), Intent: "待识别", Status: "Active",
		RiskLevel: "LOW", StartedAt: now,
	}
	s.conversations = append([]Conversation{conv}, s.conversations...)
	return conv
}

func (s *Store) SendMessage(conversationID, content string) SendMessageResult {
	s.mu.Lock()
	defer s.mu.Unlock()

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
	agentMessage, gap := s.agentReplyLocked(conversationID, content, evidence, now)
	s.messages = append(s.messages, userMessage, agentMessage)

	conv := s.touchConversationLocked(conversationID, content, evidence, gap)
	return SendMessageResult{Conversation: conv, UserMessage: userMessage, AgentMessage: agentMessage, Evidence: evidence, Gap: gap}
}

func (s *Store) ListKnowledge() []KnowledgeArticle {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]KnowledgeArticle(nil), s.knowledge...)
}

func (s *Store) SearchKnowledge(query string) []KnowledgeArticle {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.searchLocked(query)
}

func (s *Store) Dashboard() Dashboard {
	s.mu.RLock()
	defer s.mu.RUnlock()
	openGaps := 0
	for _, gap := range s.gaps {
		if gap.Status == "OPEN" {
			openGaps++
		}
	}
	return Dashboard{
		Metrics: []Metric{
			{Label: "Active sessions", Value: fmt.Sprintf("%d", len(s.conversations)), Note: "in-memory V1 runtime"},
			{Label: "Knowledge items", Value: fmt.Sprintf("%d", len(s.knowledge)), Note: "seeded trusted articles"},
			{Label: "Open gaps", Value: fmt.Sprintf("%d", openGaps), Note: "created by no-evidence fallback"},
			{Label: "Enabled rules", Value: fmt.Sprintf("%d", enabledRules(s.rules)), Note: "guardrail and transfer policies"},
		},
		Conversations: append([]Conversation(nil), s.conversations...),
		KnowledgeGaps: append([]KnowledgeGap(nil), s.gaps...),
		Rules:         append([]Rule(nil), s.rules...),
	}
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

func (s *Store) agentReplyLocked(conversationID, content string, evidence []KnowledgeArticle, now string) (Message, *KnowledgeGap) {
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
		s.gaps = append([]KnowledgeGap{gap}, s.gaps...)
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
	return Message{
		ID: fmt.Sprintf("msg_%d_agent", time.Now().UnixNano()), ConversationID: conversationID, Role: "assistant",
		Content: fmt.Sprintf("根据知识库《%s》：%s", evidence[0].Title, evidence[0].Content),
		Engine:  "rag+rule", Safe: true, EvidenceIDs: ids, CreatedAt: now,
	}, nil
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
