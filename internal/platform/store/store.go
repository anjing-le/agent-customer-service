package store

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

type ReplyGenerator interface {
	GenerateReply(ctx context.Context, req ReplyRequest) (ReplyGeneration, error)
}

type ReplyRequest struct {
	ConversationID string
	Question       string
	Evidence       []KnowledgeArticle
	History        []Message
}

type ReplyGeneration struct {
	Content string
	Model   string
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
	ReceiveChannelMessage(message ChannelInboundMessage) (SendMessageResult, error)
	ListKnowledge() ([]KnowledgeArticle, error)
	SearchKnowledge(query string) ([]KnowledgeArticle, error)
	ResolveKnowledgeGap(id string) (KnowledgeGap, error)
	CreateArticleFromGap(gapID, title, category, content string, tags []string) (KnowledgeArticle, error)
	TestRule(content string) (RuleTestResult, error)
	ResolveTransferTicket(id, assignee, note string) (TransferTicket, error)
	SubmitAnnotation(messageID, reviewer, verdict, note string, dimensions AnnotationDimensions, tags []string) (Annotation, error)
	ExportTrainingSamples(maxScore int) ([]TrainingSample, error)
	Dashboard() (Dashboard, error)
}

type Store struct {
	mu sync.RWMutex

	conversations   []Conversation
	messages        []Message
	knowledge       []KnowledgeArticle
	gaps            []KnowledgeGap
	rules           []Rule
	tickets         []TransferTicket
	channelPolicies []ChannelPolicy
	annotations     []Annotation
	generator       ReplyGenerator
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
	ID             string      `json:"id"`
	ConversationID string      `json:"conversationId"`
	Role           string      `json:"role"`
	Content        string      `json:"content"`
	Engine         string      `json:"engine"`
	Safe           bool        `json:"safe"`
	FallbackReason string      `json:"fallbackReason,omitempty"`
	EvidenceIDs    []string    `json:"evidenceIds,omitempty"`
	CreatedAt      string      `json:"createdAt"`
	Trace          *AgentTrace `json:"trace,omitempty"`
}

type AgentTrace struct {
	Strategy            string `json:"strategy"`
	EvidenceCount       int    `json:"evidenceCount"`
	HistoryCount        int    `json:"historyCount"`
	ModelAttempted      bool   `json:"modelAttempted"`
	Model               string `json:"model,omitempty"`
	ModelDurationMs     int64  `json:"modelDurationMs,omitempty"`
	ModelFallback       bool   `json:"modelFallback"`
	ModelFallbackReason string `json:"modelFallbackReason,omitempty"`
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
	ID             string          `json:"id"`
	ConversationID string          `json:"conversationId"`
	Channel        string          `json:"channel"`
	Question       string          `json:"question"`
	Reason         string          `json:"reason"`
	Priority       string          `json:"priority"`
	Status         string          `json:"status"`
	SLAMinutes     int             `json:"slaMinutes"`
	WaitMinutes    int             `json:"waitMinutes"`
	SLAStatus      string          `json:"slaStatus"`
	Escalated      bool            `json:"escalated"`
	Assignee       string          `json:"assignee,omitempty"`
	ResolutionNote string          `json:"resolutionNote,omitempty"`
	CreatedAt      string          `json:"createdAt"`
	ResolvedAt     string          `json:"resolvedAt,omitempty"`
	Events         []TransferEvent `json:"events,omitempty"`
}

type ChannelPolicy struct {
	Channel        string `json:"channel"`
	DisplayName    string `json:"displayName"`
	Tone           string `json:"tone"`
	SLAMinutes     int    `json:"slaMinutes"`
	RiskBoost      string `json:"riskBoost"`
	EscalationNote string `json:"escalationNote"`
	Enabled        bool   `json:"enabled"`
}

type TransferEvent struct {
	Type      string `json:"type"`
	Actor     string `json:"actor"`
	Note      string `json:"note"`
	CreatedAt string `json:"createdAt"`
}

type Annotation struct {
	ID         string               `json:"id"`
	MessageID  string               `json:"messageId"`
	Reviewer   string               `json:"reviewer"`
	Verdict    string               `json:"verdict"`
	Note       string               `json:"note"`
	Dimensions AnnotationDimensions `json:"dimensions"`
	Tags       []string             `json:"tags"`
	Score      int                  `json:"score"`
	CreatedAt  string               `json:"createdAt"`
}

type AnnotationDimensions struct {
	Groundedness int `json:"groundedness"`
	Safety       int `json:"safety"`
	Helpfulness  int `json:"helpfulness"`
}

type TrainingSample struct {
	ID             string               `json:"id"`
	ConversationID string               `json:"conversationId"`
	MessageID      string               `json:"messageId"`
	Channel        string               `json:"channel"`
	Prompt         string               `json:"prompt"`
	Answer         string               `json:"answer"`
	Engine         string               `json:"engine"`
	EvidenceIDs    []string             `json:"evidenceIds"`
	Reviewer       string               `json:"reviewer"`
	Verdict        string               `json:"verdict"`
	Score          int                  `json:"score"`
	Dimensions     AnnotationDimensions `json:"dimensions"`
	Note           string               `json:"note"`
	Tags           []string             `json:"tags"`
	CreatedAt      string               `json:"createdAt"`
}

type Dashboard struct {
	Metrics         []Metric         `json:"metrics"`
	Conversations   []Conversation   `json:"conversations"`
	KnowledgeGaps   []KnowledgeGap   `json:"knowledgeGaps"`
	Rules           []Rule           `json:"rules"`
	Transfers       []TransferTicket `json:"transfers"`
	ChannelPolicies []ChannelPolicy  `json:"channelPolicies"`
	Quality         QualitySummary   `json:"quality"`
	Annotations     []Annotation     `json:"annotations"`
}

type Metric struct {
	Label string `json:"label"`
	Value string `json:"value"`
	Note  string `json:"note"`
}

type QualitySummary struct {
	Score            int      `json:"score"`
	ReviewedMessages int      `json:"reviewedMessages"`
	EvidenceAnswers  int      `json:"evidenceAnswers"`
	SafeFallbacks    int      `json:"safeFallbacks"`
	HumanTransfers   int      `json:"humanTransfers"`
	AnnotationCount  int      `json:"annotationCount"`
	AverageReview    int      `json:"averageReview"`
	Notes            []string `json:"notes"`
}

type SendMessageResult struct {
	Conversation Conversation       `json:"conversation"`
	UserMessage  Message            `json:"userMessage"`
	AgentMessage Message            `json:"agentMessage"`
	Evidence     []KnowledgeArticle `json:"evidence"`
	Gap          *KnowledgeGap      `json:"gap,omitempty"`
}

type ChannelInboundMessage struct {
	Channel                string `json:"channel"`
	ExternalConversationID string `json:"externalConversationId"`
	Customer               string `json:"customer"`
	Content                string `json:"content"`
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
		channelPolicies: defaultChannelPolicies(),
	}
	st.messages = []Message{
		{ID: "msg_demo_1", ConversationID: "conv_demo_refund", Role: "user", Content: "7 天无理由退货的运费怎么计算？", Engine: "customer", Safe: true, CreatedAt: now},
		{ID: "msg_demo_2", ConversationID: "conv_demo_refund", Role: "assistant", Content: "根据售后知识库，签收 7 天内可申请无理由退货；非质量问题寄回运费通常由用户承担，质量问题由商家承担。", Engine: "rag+rule", Safe: true, EvidenceIDs: []string{"kb_refund_7d"}, CreatedAt: now},
	}
	oldTransferAt := time.Now().UTC().Add(-45 * time.Minute).Format(time.RFC3339)
	st.tickets = []TransferTicket{
		demoTransferTicket("ticket_demo_transfer", "conv_demo_transfer", "WeChat", "我已经催了三次，必须马上找人工处理。", oldTransferAt),
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
	channel := s.conversationChannelLocked(conversationID)

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
		s.tickets = append([]TransferTicket{newTransferTicket(conversationID, channel, content, now)}, s.tickets...)
	}

	conv := s.touchConversationLocked(conversationID, content, evidence, gap)
	return SendMessageResult{Conversation: conv, UserMessage: userMessage, AgentMessage: agentMessage, Evidence: evidence, Gap: gap}, nil
}

func (s *Store) ReceiveChannelMessage(message ChannelInboundMessage) (SendMessageResult, error) {
	channel := fallback(message.Channel, "Web")
	externalID := fallback(message.ExternalConversationID, fmt.Sprintf("anon_%d", time.Now().UnixNano()))
	conversationID := inboundConversationID(channel, externalID)
	s.mu.Lock()
	if !s.conversationExistsLocked(conversationID) {
		now := time.Now().UTC().Format(time.RFC3339)
		s.conversations = append([]Conversation{{
			ID:        conversationID,
			Customer:  fallback(message.Customer, "访客"),
			Channel:   channel,
			Intent:    "待识别",
			Status:    "Active",
			RiskLevel: "LOW",
			StartedAt: now,
		}}, s.conversations...)
	}
	s.mu.Unlock()
	return s.SendMessage(conversationID, message.Content)
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
			s.tickets[idx].Events = transferEvents(s.tickets[idx])
			return withTransferSLA(s.tickets[idx], s.channelPolicies, time.Now().UTC()), nil
		}
	}
	return TransferTicket{}, fmt.Errorf("transfer ticket %s not found", id)
}

func (s *Store) SubmitAnnotation(messageID, reviewer, verdict, note string, dimensions AnnotationDimensions, tags []string) (Annotation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.messageExistsLocked(messageID) {
		return Annotation{}, fmt.Errorf("message %s not found", messageID)
	}
	annotation := newAnnotation(messageID, reviewer, verdict, note, dimensions, tags, time.Now().UTC().Format(time.RFC3339))
	s.annotations = append([]Annotation{annotation}, s.annotations...)
	return annotation, nil
}

func (s *Store) ExportTrainingSamples(maxScore int) ([]TrainingSample, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return trainingSamples(s.annotations, s.messages, s.conversations, normalizeMaxScore(maxScore)), nil
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
	transfers := withTransferSLAs(s.tickets, s.channelPolicies, time.Now().UTC())
	quality := qualitySummary(s.messages, s.gaps, transfers, s.annotations)
	return Dashboard{
		Metrics: []Metric{
			{Label: "Active sessions", Value: fmt.Sprintf("%d", len(s.conversations)), Note: "in-memory V1 runtime"},
			{Label: "Knowledge items", Value: fmt.Sprintf("%d", len(s.knowledge)), Note: "seeded trusted articles"},
			{Label: "Open gaps", Value: fmt.Sprintf("%d", openGaps), Note: "created by no-evidence fallback"},
			{Label: "Channels", Value: fmt.Sprintf("%d", activeChannelCount(s.conversations)), Note: "conversation sources with policies"},
			{Label: "Open transfers", Value: fmt.Sprintf("%d", openTransfers), Note: "waiting for human agents"},
			{Label: "SLA escalations", Value: fmt.Sprintf("%d", escalatedTransferCount(transfers)), Note: "open tickets past response SLA"},
			{Label: "Enabled rules", Value: fmt.Sprintf("%d", enabledRules(s.rules)), Note: "guardrail and transfer policies"},
		},
		Conversations:   append([]Conversation(nil), s.conversations...),
		KnowledgeGaps:   append([]KnowledgeGap(nil), s.gaps...),
		Rules:           append([]Rule(nil), s.rules...),
		Transfers:       transfers,
		ChannelPolicies: append([]ChannelPolicy(nil), s.channelPolicies...),
		Quality:         quality,
		Annotations:     append([]Annotation(nil), s.annotations...),
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
			Trace: newAgentTrace("human_transfer", evidence, history),
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
				ID: fmt.Sprintf("msg_%d_agent", time.Now().UnixNano()), ConversationID: conversationID, Role: "assistant",
				Content: strings.TrimSpace(generated.Content), Engine: "llm+rag", Safe: true, EvidenceIDs: ids, CreatedAt: now, Trace: trace,
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
		ID: fmt.Sprintf("msg_%d_agent", time.Now().UnixNano()), ConversationID: conversationID, Role: "assistant",
		Content: fmt.Sprintf("根据知识库《%s》：%s", evidence[0].Title, evidence[0].Content),
		Engine:  "rag+rule", Safe: true, EvidenceIDs: ids, CreatedAt: now, Trace: trace,
	}, nil
}

func newAgentTrace(strategy string, evidence []KnowledgeArticle, history []Message) *AgentTrace {
	return &AgentTrace{
		Strategy:      strategy,
		EvidenceCount: len(evidence),
		HistoryCount:  len(history),
	}
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

func (s *Store) messageExistsLocked(messageID string) bool {
	for _, message := range s.messages {
		if message.ID == messageID {
			return true
		}
	}
	return false
}

func (s *Store) conversationExistsLocked(conversationID string) bool {
	for _, conversation := range s.conversations {
		if conversation.ID == conversationID {
			return true
		}
	}
	return false
}

func (s *Store) conversationChannelLocked(conversationID string) string {
	for _, item := range s.conversations {
		if item.ID == conversationID {
			return fallback(item.Channel, "Web")
		}
	}
	return "Web"
}

func inboundConversationID(channel, externalID string) string {
	return fmt.Sprintf("conv_%s_%s", normalizeIDPart(channel), normalizeIDPart(externalID))
}

func normalizeIDPart(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return "unknown"
	}
	var builder strings.Builder
	for _, char := range value {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') {
			builder.WriteRune(char)
			continue
		}
		builder.WriteRune('_')
	}
	normalized := strings.Trim(builder.String(), "_")
	if normalized == "" {
		return "unknown"
	}
	return normalized
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

func newTransferTicket(conversationID, channel, question, createdAt string) TransferTicket {
	return demoTransferTicket(fmt.Sprintf("ticket_%d", time.Now().UnixNano()), conversationID, channel, question, createdAt)
}

func demoTransferTicket(id, conversationID, channel, question, createdAt string) TransferTicket {
	if id == "" {
		id = fmt.Sprintf("ticket_%d", time.Now().UnixNano())
	}
	ticket := TransferTicket{
		ID:             id,
		ConversationID: conversationID,
		Channel:        fallback(channel, "Web"),
		Question:       question,
		Reason:         "TRANSFER_THRESHOLD",
		Priority:       "HIGH",
		Status:         "OPEN",
		CreatedAt:      createdAt,
	}
	ticket.Events = transferEvents(ticket)
	return withTransferSLA(ticket, defaultChannelPolicies(), time.Now().UTC())
}

func transferEvents(ticket TransferTicket) []TransferEvent {
	events := []TransferEvent{{
		Type:      "CREATED",
		Actor:     "agent",
		Note:      "Agent 命中高风险边界并创建人工接管工单。",
		CreatedAt: ticket.CreatedAt,
	}}
	if ticket.Status == "RESOLVED" {
		events = append(events, TransferEvent{
			Type:      "RESOLVED",
			Actor:     fallback(ticket.Assignee, "operator"),
			Note:      fallback(ticket.ResolutionNote, "人工已处理"),
			CreatedAt: ticket.ResolvedAt,
		})
	}
	return events
}

func withTransferSLAs(tickets []TransferTicket, policies []ChannelPolicy, now time.Time) []TransferTicket {
	items := make([]TransferTicket, 0, len(tickets))
	for _, ticket := range tickets {
		items = append(items, withTransferSLA(ticket, policies, now))
	}
	return items
}

func withTransferSLA(ticket TransferTicket, policies []ChannelPolicy, now time.Time) TransferTicket {
	policy := channelPolicyFor(policies, ticket.Channel)
	ticket.Channel = policy.Channel
	ticket.SLAMinutes = policy.SLAMinutes
	ticket.WaitMinutes = transferWaitMinutes(ticket, now)
	if ticket.Status == "RESOLVED" {
		ticket.SLAStatus = "RESOLVED"
		return ticket
	}
	if ticket.WaitMinutes >= ticket.SLAMinutes {
		ticket.SLAStatus = "BREACHED"
		ticket.Escalated = true
		ticket.Priority = "CRITICAL"
		return ticket
	}
	ticket.SLAStatus = "ON_TRACK"
	return ticket
}

func defaultChannelPolicies() []ChannelPolicy {
	return []ChannelPolicy{
		{Channel: "Web", DisplayName: "Web 客服", Tone: "标准、清晰、可追溯", SLAMinutes: 30, RiskBoost: "NORMAL", EscalationNote: "网页渠道按标准客服 SLA 处理。", Enabled: true},
		{Channel: "WeChat", DisplayName: "微信客服", Tone: "简洁、安抚、快速接管", SLAMinutes: 15, RiskBoost: "HIGH", EscalationNote: "微信投诉和催办要更快进入人工队列。", Enabled: true},
		{Channel: "App", DisplayName: "App 客服", Tone: "直接、产品化、引导自助", SLAMinutes: 20, RiskBoost: "NORMAL", EscalationNote: "App 内问题优先引导订单和售后入口。", Enabled: true},
		{Channel: "Marketplace", DisplayName: "平台店铺客服", Tone: "谨慎、合规、避免承诺", SLAMinutes: 10, RiskBoost: "HIGH", EscalationNote: "平台投诉可能影响店铺指标，优先升级。", Enabled: true},
	}
}

func channelPolicyFor(policies []ChannelPolicy, channel string) ChannelPolicy {
	normalized := strings.ToLower(strings.TrimSpace(channel))
	for _, policy := range policies {
		if strings.ToLower(policy.Channel) == normalized && policy.Enabled {
			return policy
		}
	}
	for _, policy := range defaultChannelPolicies() {
		if strings.ToLower(policy.Channel) == normalized {
			return policy
		}
	}
	return defaultChannelPolicies()[0]
}

func activeChannelCount(conversations []Conversation) int {
	seen := map[string]bool{}
	for _, conv := range conversations {
		channel := fallback(conv.Channel, "Web")
		seen[channel] = true
	}
	return len(seen)
}

func transferWaitMinutes(ticket TransferTicket, now time.Time) int {
	startedAt, err := time.Parse(time.RFC3339, ticket.CreatedAt)
	if err != nil {
		return 0
	}
	endedAt := now
	if ticket.ResolvedAt != "" {
		if parsed, err := time.Parse(time.RFC3339, ticket.ResolvedAt); err == nil {
			endedAt = parsed
		}
	}
	minutes := int(endedAt.Sub(startedAt).Minutes())
	if minutes < 0 {
		return 0
	}
	return minutes
}

func newAnnotation(messageID, reviewer, verdict, note string, dimensions AnnotationDimensions, tags []string, createdAt string) Annotation {
	annotation := Annotation{
		ID:         fmt.Sprintf("ann_%d", time.Now().UnixNano()),
		MessageID:  messageID,
		Reviewer:   fallback(reviewer, "operator"),
		Verdict:    fallback(verdict, "PASS"),
		Note:       fallback(note, "人工复核通过"),
		Dimensions: normalizeAnnotationDimensions(dimensions),
		Tags:       normalizeTags(tags, verdict),
		CreatedAt:  createdAt,
	}
	annotation.Score = annotationScore(annotation.Dimensions)
	return annotation
}

func normalizeAnnotationDimensions(dimensions AnnotationDimensions) AnnotationDimensions {
	return AnnotationDimensions{
		Groundedness: clampScore(dimensions.Groundedness),
		Safety:       clampScore(dimensions.Safety),
		Helpfulness:  clampScore(dimensions.Helpfulness),
	}
}

func clampScore(value int) int {
	if value < 1 {
		return 1
	}
	if value > 5 {
		return 5
	}
	return value
}

func annotationScore(dimensions AnnotationDimensions) int {
	return (dimensions.Groundedness + dimensions.Safety + dimensions.Helpfulness) * 100 / 15
}

func normalizeMaxScore(maxScore int) int {
	if maxScore <= 0 {
		return 80
	}
	if maxScore > 100 {
		return 100
	}
	return maxScore
}

func trainingSamples(annotations []Annotation, messages []Message, conversations []Conversation, maxScore int) []TrainingSample {
	samples := make([]TrainingSample, 0)
	for _, annotation := range annotations {
		if !needsReview(annotation, maxScore) {
			continue
		}
		answer, ok := messageByID(messages, annotation.MessageID)
		if !ok || answer.Role != "assistant" {
			continue
		}
		conversation := conversationByID(conversations, answer.ConversationID)
		samples = append(samples, TrainingSample{
			ID:             fmt.Sprintf("sample_%s", annotation.ID),
			ConversationID: answer.ConversationID,
			MessageID:      answer.ID,
			Channel:        fallback(conversation.Channel, "Web"),
			Prompt:         previousUserPrompt(messages, answer),
			Answer:         answer.Content,
			Engine:         answer.Engine,
			EvidenceIDs:    append([]string(nil), answer.EvidenceIDs...),
			Reviewer:       annotation.Reviewer,
			Verdict:        annotation.Verdict,
			Score:          annotation.Score,
			Dimensions:     annotation.Dimensions,
			Note:           annotation.Note,
			Tags:           append([]string(nil), annotation.Tags...),
			CreatedAt:      annotation.CreatedAt,
		})
	}
	return samples
}

func needsReview(annotation Annotation, maxScore int) bool {
	verdict := strings.ToUpper(strings.TrimSpace(annotation.Verdict))
	return annotation.Score <= maxScore || verdict == "FAIL" || verdict == "REVIEW"
}

func messageByID(messages []Message, id string) (Message, bool) {
	for _, message := range messages {
		if message.ID == id {
			return message, true
		}
	}
	return Message{}, false
}

func conversationByID(conversations []Conversation, id string) Conversation {
	for _, conversation := range conversations {
		if conversation.ID == id {
			return conversation
		}
	}
	return Conversation{ID: id, Customer: "访客", Channel: "Web", Status: "Active", RiskLevel: "LOW"}
}

func previousUserPrompt(messages []Message, answer Message) string {
	prompt := ""
	for _, message := range messages {
		if message.ConversationID != answer.ConversationID {
			continue
		}
		if message.ID == answer.ID {
			break
		}
		if message.Role == "user" {
			prompt = message.Content
		}
	}
	return prompt
}

func qualitySummary(messages []Message, gaps []KnowledgeGap, tickets []TransferTicket, annotations []Annotation) QualitySummary {
	summary := QualitySummary{}
	for _, message := range messages {
		if message.Role != "assistant" {
			continue
		}
		summary.ReviewedMessages++
		if message.Safe {
			summary.Score += 50
		}
		switch message.Engine {
		case "rag+rule", "llm+rag":
			summary.EvidenceAnswers++
			summary.Score += 30
		}
		switch message.FallbackReason {
		case "NO_EVIDENCE":
			summary.SafeFallbacks++
			summary.Score += 20
		case "TRANSFER_THRESHOLD":
			summary.HumanTransfers++
			summary.Score += 20
		}
	}
	for _, annotation := range annotations {
		summary.AnnotationCount++
		summary.AverageReview += annotation.Score
	}
	if summary.AnnotationCount > 0 {
		summary.AverageReview = summary.AverageReview / summary.AnnotationCount
	}
	if summary.ReviewedMessages > 0 {
		summary.Score = summary.Score / summary.ReviewedMessages
	}
	if summary.AverageReview > 0 {
		summary.Score = (summary.Score + summary.AverageReview) / 2
	}
	if summary.Score > 100 {
		summary.Score = 100
	}
	summary.Notes = []string{
		fmt.Sprintf("%d evidence-backed replies", summary.EvidenceAnswers),
		fmt.Sprintf("%d safe no-evidence fallbacks", summary.SafeFallbacks),
		fmt.Sprintf("%d human transfer decisions", summary.HumanTransfers),
		fmt.Sprintf("%d open knowledge gaps", openGapCount(gaps)),
		fmt.Sprintf("%d open transfer tickets", openTransferCount(tickets)),
		fmt.Sprintf("%d escalated transfer tickets", escalatedTransferCount(tickets)),
		fmt.Sprintf("%d human annotations", summary.AnnotationCount),
	}
	return summary
}

func openGapCount(gaps []KnowledgeGap) int {
	count := 0
	for _, gap := range gaps {
		if gap.Status == "OPEN" {
			count++
		}
	}
	return count
}

func openTransferCount(tickets []TransferTicket) int {
	count := 0
	for _, ticket := range tickets {
		if ticket.Status == "OPEN" {
			count++
		}
	}
	return count
}

func escalatedTransferCount(tickets []TransferTicket) int {
	count := 0
	for _, ticket := range tickets {
		if ticket.Escalated {
			count++
		}
	}
	return count
}
