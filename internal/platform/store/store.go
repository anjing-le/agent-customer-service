package store

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
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

type NotificationDeliveryClient interface {
	DeliverChannelNotification(ctx context.Context, req NotificationDeliveryRequest) (NotificationDeliveryResult, error)
}

type NotificationDeliveryRequest struct {
	Notification  ChannelNotification
	Outcome       string
	SignedPayload string
}

type NotificationDeliveryResult struct {
	Accepted      bool
	ReceiptStatus string
	ReceiptBody   string
	Error         string
}

type NotificationDeliveryAudit struct {
	Attempt          int    `json:"attempt"`
	TargetURL        string `json:"targetUrl"`
	SecretRef        string `json:"secretRef"`
	SignaturePreview string `json:"signaturePreview"`
	PayloadHash      string `json:"payloadHash"`
	RequestSummary   string `json:"requestSummary"`
	ResponseSummary  string `json:"responseSummary"`
	CreatedAt        string `json:"createdAt"`
}

type Option func(*Store)

func WithReplyGenerator(generator ReplyGenerator) Option {
	return func(s *Store) {
		s.generator = generator
	}
}

func WithChannelIntegrations(integrations []ChannelIntegration) Option {
	return func(s *Store) {
		s.integrations = append([]ChannelIntegration(nil), integrations...)
	}
}

func WithNotificationDeliveryClient(client NotificationDeliveryClient) Option {
	return func(s *Store) {
		s.delivery = client
	}
}

type Runtime interface {
	ListConversations() ([]Conversation, error)
	CreateConversation(customer, channel string) (Conversation, error)
	ListMessages(conversationID string) ([]Message, error)
	SendMessage(conversationID, content string) (SendMessageResult, error)
	ChannelIntegration(channel string) (ChannelIntegration, error)
	RecordChannelRateLimit(channel string, windowStart time.Time, limit int) (bool, int, error)
	RecordChannelInbound(receipt ChannelInboundReceipt) (bool, error)
	RecordChannelInboundAudit(audit ChannelInboundAudit) error
	ListChannelInboundAudits(limit int) ([]ChannelInboundAudit, error)
	ListChannelInboundAuditQualityEvents(limit int) ([]ChannelInboundAuditQualityEvent, error)
	RecordChannelFailure(event ChannelFailureEvent) error
	UpdateChannelAlertPolicy(channel string, targetURL string, secretRef string, maxAttempts int, backoffSeconds int, inboundAuditMinSamples int, inboundAuditMinAcceptanceRate int, inboundAuditMaxErrorCount int, actor string, note string) (ChannelAlertPolicy, error)
	ApproveNotificationPolicyChange(id string, approver string, note string, confirmation string) (ChannelAlertPolicy, error)
	RejectNotificationPolicyChange(id string, reviewer string, note string) (NotificationPolicyChange, error)
	CancelNotificationPolicyChange(id string, actor string, note string) (NotificationPolicyChange, error)
	RollbackChannelAlertPolicy(channel string, actor string, note string, confirmation string) (ChannelAlertPolicy, error)
	DispatchChannelNotification(id, outcome string) (ChannelNotification, error)
	AcknowledgeChannelNotification(id, actor, note string) (ChannelNotification, error)
	CompleteChannelRunbookCheck(check ChannelRunbookCheck) (ChannelRunbookCheck, error)
	ReceiveChannelMessage(message ChannelInboundMessage) (SendMessageResult, error)
	ListKnowledge() ([]KnowledgeArticle, error)
	SearchKnowledge(query string) ([]KnowledgeArticle, error)
	ResolveKnowledgeGap(id string) (KnowledgeGap, error)
	CreateArticleFromGap(gapID, title, category, content string, tags []string) (KnowledgeArticle, error)
	TestRule(content string) (RuleTestResult, error)
	CompareRuleVersions(content string) (RuleComparison, error)
	SubmitRuleApproval(code, approver, riskLevel, note string, sampleIDs []string) (RuleApproval, error)
	PublishCanaryRule(code, actor, note string) (RuleReleaseEvent, error)
	RollbackRule(code, actor, note string) (RuleReleaseEvent, error)
	ResolveTransferTicket(id, assignee, note string) (TransferTicket, error)
	AssignReviewTask(id, assignee string) (ReviewTask, error)
	CompleteReviewTask(id string) (ReviewTask, error)
	SubmitAnnotation(messageID, reviewer, verdict, note string, dimensions AnnotationDimensions, tags []string) (Annotation, error)
	ExportTrainingSamples(maxScore int) ([]TrainingSample, error)
	SaveChannelOpsReport(report ChannelOpsReport) (ChannelOpsReport, error)
	ListChannelOpsReports(limit int) ([]ChannelOpsReport, error)
	ChannelOpsReport(id string) (ChannelOpsReport, error)
	PruneChannelOpsReports(retain int) (int, error)
	RecordChannelOpsReportEvent(event ChannelOpsReportEvent) (ChannelOpsReportEvent, error)
	ListChannelOpsReportEvents(limit int) ([]ChannelOpsReportEvent, error)
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
	integrations    []ChannelIntegration
	alertPolicies   []ChannelAlertPolicy
	notifications   []ChannelNotification
	runbookChecks   []ChannelRunbookCheck
	channelReports  []ChannelOpsReport
	reportEvents    []ChannelOpsReportEvent
	auditEvents     []ChannelInboundAuditQualityEvent
	policyEvents    []NotificationPolicyEvent
	policyChanges   []NotificationPolicyChange
	ruleApprovals   []RuleApproval
	ruleEvents      []RuleReleaseEvent
	annotations     []Annotation
	reviewTasks     []ReviewTask
	inboundReplay   map[string]ChannelInboundReceipt
	inboundAudits   []ChannelInboundAudit
	rateWindows     map[string]int
	channelFailures []ChannelFailureEvent
	generator       ReplyGenerator
	delivery        NotificationDeliveryClient
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
	ID        string `json:"id"`
	Code      string `json:"code"`
	Name      string `json:"name"`
	Trigger   string `json:"trigger"`
	Action    string `json:"action"`
	Enabled   bool   `json:"enabled"`
	Version   string `json:"version"`
	Stage     string `json:"stage"`
	HitCount  int    `json:"hitCount"`
	LastHitAt string `json:"lastHitAt,omitempty"`
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

type RuleComparison struct {
	Input          string         `json:"input"`
	Current        RuleTestResult `json:"current"`
	Canary         RuleTestResult `json:"canary"`
	Changed        bool           `json:"changed"`
	Recommendation string         `json:"recommendation"`
}

type RuleApproval struct {
	ID          string   `json:"id"`
	RuleCode    string   `json:"ruleCode"`
	Approver    string   `json:"approver"`
	RiskLevel   string   `json:"riskLevel"`
	SampleIDs   []string `json:"sampleIds"`
	SampleCount int      `json:"sampleCount"`
	Status      string   `json:"status"`
	Note        string   `json:"note"`
	CreatedAt   string   `json:"createdAt"`
}

type RuleReleaseEvent struct {
	ID        string `json:"id"`
	RuleCode  string `json:"ruleCode"`
	Version   string `json:"version"`
	Action    string `json:"action"`
	Actor     string `json:"actor"`
	Note      string `json:"note"`
	CreatedAt string `json:"createdAt"`
}

type RuleReleaseObservation struct {
	RuleCode            string `json:"ruleCode"`
	Version             string `json:"version"`
	ReleasedAt          string `json:"releasedAt"`
	Window              string `json:"window"`
	RuleHits            int    `json:"ruleHits"`
	TransferTickets     int    `json:"transferTickets"`
	LowScoreSamples     int    `json:"lowScoreSamples"`
	AverageReview       int    `json:"averageReview"`
	RiskLevel           string `json:"riskLevel"`
	Recommendation      string `json:"recommendation"`
	RollbackRecommended bool   `json:"rollbackRecommended"`
}

type NotificationPolicyEvent struct {
	ID        string `json:"id"`
	Channel   string `json:"channel"`
	Action    string `json:"action"`
	Actor     string `json:"actor"`
	Before    string `json:"before"`
	After     string `json:"after"`
	Note      string `json:"note"`
	CreatedAt string `json:"createdAt"`
}

type NotificationPolicyChange struct {
	ID                    string                   `json:"id"`
	Channel               string                   `json:"channel"`
	TargetURL             string                   `json:"targetUrl"`
	SecretRef             string                   `json:"secretRef"`
	MaxAttempts           int                      `json:"maxAttempts"`
	BackoffSeconds        int                      `json:"backoffSeconds"`
	CurrentTargetURL      string                   `json:"currentTargetUrl"`
	CurrentSecretRef      string                   `json:"currentSecretRef"`
	CurrentMaxAttempts    int                      `json:"currentMaxAttempts"`
	CurrentBackoffSeconds int                      `json:"currentBackoffSeconds"`
	Diff                  []NotificationPolicyDiff `json:"diff"`
	ConfirmationText      string                   `json:"confirmationText"`
	RequestedBy           string                   `json:"requestedBy"`
	Status                string                   `json:"status"`
	Note                  string                   `json:"note"`
	CreatedAt             string                   `json:"createdAt"`
	ExpiresAt             string                   `json:"expiresAt"`
	ApprovedBy            string                   `json:"approvedBy,omitempty"`
	ApprovedAt            string                   `json:"approvedAt,omitempty"`
}

type NotificationPolicyDiff struct {
	Field  string `json:"field"`
	Before string `json:"before"`
	After  string `json:"after"`
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

type ChannelIntegration struct {
	Channel                string   `json:"channel"`
	DisplayName            string   `json:"displayName"`
	Enabled                bool     `json:"enabled"`
	SecretSource           string   `json:"secretSource"`
	SecretRef              string   `json:"secretRef"`
	NextSecretRef          string   `json:"nextSecretRef"`
	SignatureWindowSeconds int      `json:"signatureWindowSeconds"`
	ReplayProtection       bool     `json:"replayProtection"`
	AllowedOrigins         []string `json:"allowedOrigins"`
	RateLimitPerMinute     int      `json:"rateLimitPerMinute"`
	RotationHint           string   `json:"rotationHint"`
	RotatesAt              string   `json:"rotatesAt,omitempty"`
	UpdatedAt              string   `json:"updatedAt"`
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

type ReviewTask struct {
	ID             string `json:"id"`
	MessageID      string `json:"messageId"`
	ConversationID string `json:"conversationId"`
	Channel        string `json:"channel"`
	Assignee       string `json:"assignee,omitempty"`
	Status         string `json:"status"`
	Priority       string `json:"priority"`
	Reason         string `json:"reason"`
	CreatedAt      string `json:"createdAt"`
	CompletedAt    string `json:"completedAt,omitempty"`
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
	Metrics          []Metric                          `json:"metrics"`
	Conversations    []Conversation                    `json:"conversations"`
	KnowledgeGaps    []KnowledgeGap                    `json:"knowledgeGaps"`
	Rules            []Rule                            `json:"rules"`
	Transfers        []TransferTicket                  `json:"transfers"`
	ChannelPolicies  []ChannelPolicy                   `json:"channelPolicies"`
	Integrations     []ChannelIntegration              `json:"integrations"`
	ChannelAlerts    []ChannelAlert                    `json:"channelAlerts"`
	ChannelTrends    []ChannelFailureTrend             `json:"channelFailureTrends"`
	ChannelAudits    []ChannelInboundAudit             `json:"channelInboundAudits"`
	AuditEvents      []ChannelInboundAuditQualityEvent `json:"channelInboundAuditQualityEvents"`
	AlertPolicies    []ChannelAlertPolicy              `json:"channelAlertPolicies"`
	Notifications    []ChannelNotification             `json:"channelNotifications"`
	ChannelRunbooks  []ChannelRunbook                  `json:"channelRunbooks"`
	PolicyEvents     []NotificationPolicyEvent         `json:"notificationPolicyEvents"`
	PolicyChanges    []NotificationPolicyChange        `json:"notificationPolicyChanges"`
	Quality          QualitySummary                    `json:"quality"`
	Annotations      []Annotation                      `json:"annotations"`
	ReviewTasks      []ReviewTask                      `json:"reviewTasks"`
	RuleApprovals    []RuleApproval                    `json:"ruleApprovals"`
	RuleEvents       []RuleReleaseEvent                `json:"ruleEvents"`
	RuleObservations []RuleReleaseObservation          `json:"ruleObservations"`
}

type ChannelOpsReport struct {
	ID          string                  `json:"id"`
	Format      string                  `json:"format"`
	ContentType string                  `json:"contentType"`
	Content     string                  `json:"content,omitempty"`
	Summary     ChannelOpsReportSummary `json:"summary"`
	GeneratedAt string                  `json:"generatedAt"`
}

type ChannelOpsReportSummary struct {
	FailureCount        int                               `json:"failureCount"`
	ActiveRunbooks      int                               `json:"activeRunbooks"`
	OpenNotifications   int                               `json:"openNotifications"`
	Retrying            int                               `json:"retrying"`
	DeadLetters         int                               `json:"deadLetters"`
	Channels            []string                          `json:"channels"`
	InboundAudit        ChannelInboundAuditSummary        `json:"inboundAudit"`
	InboundAuditQuality ChannelInboundAuditQualitySummary `json:"inboundAuditQuality"`
	HandoffPriorities   []ChannelOpsHandoffPriority       `json:"handoffPriorities"`
}

type ChannelOpsHandoffPriority struct {
	Rank              int    `json:"rank"`
	Channel           string `json:"channel"`
	Severity          string `json:"severity"`
	Source            string `json:"source"`
	Reason            string `json:"reason"`
	RecommendedAction string `json:"recommendedAction"`
	Count             int    `json:"count"`
	ActionType        string `json:"actionType"`
	ActionRef         string `json:"actionRef,omitempty"`
	ActionLabel       string `json:"actionLabel"`
	NotificationID    string `json:"notificationId,omitempty"`
	RunbookStatus     string `json:"runbookStatus,omitempty"`
}

type ChannelInboundAuditSummary struct {
	Total          int                            `json:"total"`
	Accepted       int                            `json:"accepted"`
	Rejected       int                            `json:"rejected"`
	AcceptanceRate int                            `json:"acceptanceRate"`
	TopErrorCodes  []ChannelInboundAuditCodeCount `json:"topErrorCodes"`
}

type ChannelInboundAuditCodeCount struct {
	Code  string `json:"code"`
	Count int    `json:"count"`
}

type ChannelInboundAuditQualitySummary struct {
	EventCount        int      `json:"eventCount"`
	Active            int      `json:"active"`
	Watch             int      `json:"watch"`
	Recovered         int      `json:"recovered"`
	ActiveChannels    []string `json:"activeChannels"`
	WatchChannels     []string `json:"watchChannels"`
	RecoveredChannels []string `json:"recoveredChannels"`
}

type ChannelInboundAuditQualityEvent struct {
	ID                string `json:"id"`
	Channel           string `json:"channel"`
	Severity          string `json:"severity"`
	Status            string `json:"status"`
	FailureCode       string `json:"failureCode"`
	Total             int    `json:"total"`
	Accepted          int    `json:"accepted"`
	Rejected          int    `json:"rejected"`
	AcceptanceRate    int    `json:"acceptanceRate"`
	MinSamples        int    `json:"minSamples"`
	MinAcceptanceRate int    `json:"minAcceptanceRate"`
	MaxErrorCount     int    `json:"maxErrorCount"`
	Reason            string `json:"reason"`
	CreatedAt         string `json:"createdAt"`
}

type ChannelOpsReportEvent struct {
	ID        string `json:"id"`
	Action    string `json:"action"`
	Actor     string `json:"actor"`
	Status    string `json:"status"`
	ReportID  string `json:"reportId,omitempty"`
	Format    string `json:"format"`
	Pruned    int    `json:"pruned"`
	Note      string `json:"note,omitempty"`
	Error     string `json:"error,omitempty"`
	CreatedAt string `json:"createdAt"`
}

type Metric struct {
	Label string `json:"label"`
	Value string `json:"value"`
	Note  string `json:"note"`
}

type ChannelRunbook struct {
	Channel           string                `json:"channel"`
	Severity          string                `json:"severity"`
	Status            string                `json:"status"`
	FailureCode       string                `json:"failureCode"`
	Owner             string                `json:"owner"`
	NextAction        string                `json:"nextAction"`
	Escalation        string                `json:"escalation"`
	NotificationID    string                `json:"notificationId,omitempty"`
	NotificationState string                `json:"notificationState,omitempty"`
	Steps             []string              `json:"steps"`
	Checks            []ChannelRunbookCheck `json:"checks"`
}

type ChannelRunbookCheck struct {
	ID            string `json:"id"`
	Channel       string `json:"channel"`
	RunbookStatus string `json:"runbookStatus"`
	Step          string `json:"step"`
	StepIndex     int    `json:"stepIndex"`
	ActionRef     string `json:"actionRef,omitempty"`
	ReportID      string `json:"reportId,omitempty"`
	Actor         string `json:"actor"`
	Note          string `json:"note,omitempty"`
	CompletedAt   string `json:"completedAt"`
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

type ChannelInboundReceipt struct {
	ReplayKey              string `json:"replayKey"`
	Channel                string `json:"channel"`
	ExternalConversationID string `json:"externalConversationId"`
	ExternalMessageID      string `json:"externalMessageId"`
	Timestamp              string `json:"timestamp"`
	Signature              string `json:"signature"`
	ContentHash            string `json:"contentHash"`
}

type ChannelFailureEvent struct {
	Channel                string `json:"channel"`
	Code                   string `json:"code"`
	Reason                 string `json:"reason"`
	ExternalConversationID string `json:"externalConversationId"`
	ExternalMessageID      string `json:"externalMessageId"`
	Origin                 string `json:"origin"`
	CreatedAt              string `json:"createdAt"`
}

type ChannelInboundAudit struct {
	ID                     string `json:"id"`
	Channel                string `json:"channel"`
	ExternalConversationID string `json:"externalConversationId"`
	ExternalMessageID      string `json:"externalMessageId,omitempty"`
	Origin                 string `json:"origin,omitempty"`
	Status                 string `json:"status"`
	Code                   string `json:"code"`
	Reason                 string `json:"reason,omitempty"`
	ReplayKey              string `json:"replayKey,omitempty"`
	SignaturePreview       string `json:"signaturePreview,omitempty"`
	ContentHash            string `json:"contentHash,omitempty"`
	CreatedAt              string `json:"createdAt"`
}

type ChannelAlert struct {
	Channel    string `json:"channel"`
	Code       string `json:"code"`
	Count      int    `json:"count"`
	LastReason string `json:"lastReason"`
	LastOrigin string `json:"lastOrigin"`
	LastSeenAt string `json:"lastSeenAt"`
}

type ChannelFailureTrend struct {
	Channel     string `json:"channel"`
	BucketStart string `json:"bucketStart"`
	Count       int    `json:"count"`
}

type ChannelAlertPolicy struct {
	Channel                       string `json:"channel"`
	Severity                      string `json:"severity"`
	Threshold                     int    `json:"threshold"`
	WindowMinutes                 int    `json:"windowMinutes"`
	NotifyTarget                  string `json:"notifyTarget"`
	TargetURL                     string `json:"targetUrl"`
	SecretRef                     string `json:"secretRef"`
	MaxAttempts                   int    `json:"maxAttempts"`
	BackoffSeconds                int    `json:"backoffSeconds"`
	InboundAuditMinSamples        int    `json:"inboundAuditMinSamples"`
	InboundAuditMinAcceptanceRate int    `json:"inboundAuditMinAcceptanceRate"`
	InboundAuditMaxErrorCount     int    `json:"inboundAuditMaxErrorCount"`
	Enabled                       bool   `json:"enabled"`
	Active                        bool   `json:"active"`
	CurrentCount                  int    `json:"currentCount"`
	LastTriggeredAt               string `json:"lastTriggeredAt,omitempty"`
}

type ChannelNotification struct {
	ID               string                      `json:"id"`
	Channel          string                      `json:"channel"`
	Severity         string                      `json:"severity"`
	Target           string                      `json:"target"`
	TargetURL        string                      `json:"targetUrl"`
	SecretRef        string                      `json:"secretRef"`
	Status           string                      `json:"status"`
	Reason           string                      `json:"reason"`
	Count            int                         `json:"count"`
	Attempts         int                         `json:"attempts"`
	MaxAttempts      int                         `json:"maxAttempts"`
	BackoffSeconds   int                         `json:"backoffSeconds"`
	NextRetryAt      string                      `json:"nextRetryAt,omitempty"`
	Signature        string                      `json:"signature,omitempty"`
	LastDispatchAt   string                      `json:"lastDispatchAt,omitempty"`
	LastError        string                      `json:"lastError,omitempty"`
	ReceiptStatus    string                      `json:"receiptStatus,omitempty"`
	ReceiptBody      string                      `json:"receiptBody,omitempty"`
	DeliveryAudit    []NotificationDeliveryAudit `json:"deliveryAudit,omitempty"`
	DeadLetterReason string                      `json:"deadLetterReason,omitempty"`
	CreatedAt        string                      `json:"createdAt"`
	AckedBy          string                      `json:"ackedBy,omitempty"`
	AckNote          string                      `json:"ackNote,omitempty"`
	AckedAt          string                      `json:"ackedAt,omitempty"`
}

func NewSeedStore(options ...Option) *Store {
	now := time.Now().UTC().Format(time.RFC3339)
	st := &Store{
		inboundReplay: make(map[string]ChannelInboundReceipt),
		rateWindows:   make(map[string]int),
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
			{ID: "rule_low_evidence", Code: "NO_EVIDENCE_FALLBACK", Name: "无可靠证据兜底", Trigger: "knowledge evidence empty", Action: "safe_fallback_and_create_gap", Enabled: true, Version: "2026-06-active", Stage: "active"},
			{ID: "rule_human_transfer", Code: "TRANSFER_THRESHOLD", Name: "转人工阈值", Trigger: "投诉/催办/法律风险", Action: "recommend_human_transfer", Enabled: true, Version: "2026-06-active", Stage: "active"},
			{ID: "rule_cancel_canary", Code: "CANCEL_RISK_TRANSFER", Name: "取消/退订灰度转人工", Trigger: "取消订单/退订服务/退款争议", Action: "recommend_human_transfer", Enabled: true, Version: "2026-06-canary", Stage: "canary"},
		},
		channelPolicies: defaultChannelPolicies(),
		integrations:    defaultChannelIntegrations(now),
		alertPolicies:   defaultChannelAlertPolicies(),
	}
	st.messages = []Message{
		{ID: "msg_demo_1", ConversationID: "conv_demo_refund", Role: "user", Content: "7 天无理由退货的运费怎么计算？", Engine: "customer", Safe: true, CreatedAt: now},
		{ID: "msg_demo_2", ConversationID: "conv_demo_refund", Role: "assistant", Content: "根据售后知识库，签收 7 天内可申请无理由退货；非质量问题寄回运费通常由用户承担，质量问题由商家承担。", Engine: "rag+rule", Safe: true, EvidenceIDs: []string{"kb_refund_7d"}, CreatedAt: now},
	}
	st.reviewTasks = []ReviewTask{
		newReviewTask("msg_demo_2", "conv_demo_refund", "Web", "HIGH", "种子回复需验证证据完整性", now),
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
	ruleResult := evaluateRules(content, evidence, rulesByStage(s.rules, "active"))
	generator := s.generator
	s.mu.Unlock()

	agentMessage, gap := agentReply(generator, conversationID, content, evidence, history, ruleResult, now)

	s.mu.Lock()
	defer s.mu.Unlock()
	if gap != nil {
		s.gaps = append([]KnowledgeGap{*gap}, s.gaps...)
	}
	s.messages = append(s.messages, userMessage, agentMessage)
	if shouldCreateTransfer(agentMessage.FallbackReason) {
		s.tickets = append([]TransferTicket{newTransferTicket(conversationID, channel, content, now)}, s.tickets...)
	}
	s.recordRuleHitLocked(ruleResult.RuleCode, now)
	s.reviewTasks = append([]ReviewTask{newReviewTask(agentMessage.ID, conversationID, channel, reviewPriority(agentMessage), reviewReason(agentMessage), now)}, s.reviewTasks...)

	conv := s.touchConversationLocked(conversationID, content, evidence, gap)
	return SendMessageResult{Conversation: conv, UserMessage: userMessage, AgentMessage: agentMessage, Evidence: evidence, Gap: gap}, nil
}

func (s *Store) RecordChannelInbound(receipt ChannelInboundReceipt) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if receipt.ReplayKey == "" {
		return false, fmt.Errorf("channel inbound replay key is required")
	}
	if _, exists := s.inboundReplay[receipt.ReplayKey]; exists {
		return false, nil
	}
	s.inboundReplay[receipt.ReplayKey] = receipt
	return true, nil
}

func (s *Store) RecordChannelInboundAudit(audit ChannelInboundAudit) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	normalizeChannelInboundAudit(&audit)
	s.inboundAudits = append([]ChannelInboundAudit{audit}, s.inboundAudits...)
	if len(s.inboundAudits) > 100 {
		s.inboundAudits = s.inboundAudits[:100]
	}
	policies := channelAlertPolicies(s.alertPolicies, channelAlerts(s.channelFailures))
	if event, ok := channelInboundAuditQualityEvent(s.inboundAudits, policies); ok && !sameChannelInboundAuditQualityEvent(s.auditEvents, event) {
		s.auditEvents = append([]ChannelInboundAuditQualityEvent{event}, s.auditEvents...)
		if len(s.auditEvents) > 100 {
			s.auditEvents = s.auditEvents[:100]
		}
	}
	return nil
}

func (s *Store) ListChannelInboundAudits(limit int) ([]ChannelInboundAudit, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	limit = normalizeChannelInboundAuditLimit(limit)
	items := append([]ChannelInboundAudit(nil), s.inboundAudits...)
	if len(items) > limit {
		items = items[:limit]
	}
	return items, nil
}

func (s *Store) ListChannelInboundAuditQualityEvents(limit int) ([]ChannelInboundAuditQualityEvent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	limit = normalizeChannelAuditQualityEventLimit(limit)
	items := append([]ChannelInboundAuditQualityEvent(nil), s.auditEvents...)
	if len(items) > limit {
		items = items[:limit]
	}
	return items, nil
}

func (s *Store) RecordChannelFailure(event ChannelFailureEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if strings.TrimSpace(event.Channel) == "" {
		event.Channel = "Unknown"
	}
	if strings.TrimSpace(event.Code) == "" {
		event.Code = "channel_failure"
	}
	if strings.TrimSpace(event.CreatedAt) == "" {
		event.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}
	s.channelFailures = append([]ChannelFailureEvent{event}, s.channelFailures...)
	if len(s.channelFailures) > 100 {
		s.channelFailures = s.channelFailures[:100]
	}
	alerts := channelAlerts(s.channelFailures)
	policies := channelAlertPolicies(s.alertPolicies, alerts)
	s.notifications = appendTriggeredNotifications(s.notifications, policies, event.CreatedAt)
	return nil
}

func (s *Store) UpdateChannelAlertPolicy(channel string, targetURL string, secretRef string, maxAttempts int, backoffSeconds int, inboundAuditMinSamples int, inboundAuditMinAcceptanceRate int, inboundAuditMaxErrorCount int, actor string, note string) (ChannelAlertPolicy, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	normalized := strings.TrimSpace(channel)
	if normalized == "" {
		return ChannelAlertPolicy{}, fmt.Errorf("channel is required")
	}
	targetURL = strings.TrimSpace(targetURL)
	secretRef = strings.TrimSpace(secretRef)
	maxAttempts = normalizeMaxAttempts(maxAttempts)
	backoffSeconds = normalizeBackoffSeconds(backoffSeconds)
	for idx := range s.alertPolicies {
		if strings.EqualFold(s.alertPolicies[idx].Channel, normalized) {
			before := s.alertPolicies[idx]
			normalizeChannelAlertPolicy(&before)
			inboundAuditMinSamples, inboundAuditMinAcceptanceRate, inboundAuditMaxErrorCount = normalizeInboundAuditThresholdInputs(before, inboundAuditMinSamples, inboundAuditMinAcceptanceRate, inboundAuditMaxErrorCount)
			deliveryChanged := channelAlertPolicyDeliveryChanged(before, targetURL, secretRef, maxAttempts, backoffSeconds)
			thresholdChanged := channelAlertPolicyInboundAuditThresholdChanged(before, inboundAuditMinSamples, inboundAuditMinAcceptanceRate, inboundAuditMaxErrorCount)
			if requiresNotificationPolicyApproval(before) && deliveryChanged {
				if thresholdChanged {
					applyChannelAlertPolicyInboundAuditThresholds(&s.alertPolicies[idx], inboundAuditMinSamples, inboundAuditMinAcceptanceRate, inboundAuditMaxErrorCount)
					afterThresholds := s.alertPolicies[idx]
					normalizeChannelAlertPolicy(&afterThresholds)
					event := newNotificationPolicyEvent(afterThresholds.Channel, "UPDATE_AUDIT_THRESHOLDS", actor, notificationPolicySummary(before), notificationPolicySummary(afterThresholds), note, time.Now().UTC().Format(time.RFC3339))
					s.policyEvents = append([]NotificationPolicyEvent{event}, s.policyEvents...)
				}
				change := newNotificationPolicyChange(before, targetURL, secretRef, maxAttempts, backoffSeconds, actor, note, time.Now().UTC().Format(time.RFC3339))
				s.policyChanges = append([]NotificationPolicyChange{change}, s.policyChanges...)
				event := newNotificationPolicyEvent(before.Channel, "REQUEST_APPROVAL", actor, notificationPolicySummary(before), notificationPolicyChangeSummary(change), note, time.Now().UTC().Format(time.RFC3339))
				s.policyEvents = append([]NotificationPolicyEvent{event}, s.policyEvents...)
				if len(s.policyEvents) > 30 {
					s.policyEvents = s.policyEvents[:30]
				}
				return channelAlertPolicies([]ChannelAlertPolicy{s.alertPolicies[idx]}, channelAlerts(s.channelFailures))[0], nil
			}
			s.alertPolicies[idx].TargetURL = fallback(targetURL, notificationTargetURL(s.alertPolicies[idx].NotifyTarget))
			s.alertPolicies[idx].SecretRef = fallback(secretRef, notificationSecretRef(s.alertPolicies[idx].NotifyTarget))
			s.alertPolicies[idx].MaxAttempts = maxAttempts
			s.alertPolicies[idx].BackoffSeconds = backoffSeconds
			applyChannelAlertPolicyInboundAuditThresholds(&s.alertPolicies[idx], inboundAuditMinSamples, inboundAuditMinAcceptanceRate, inboundAuditMaxErrorCount)
			after := s.alertPolicies[idx]
			normalizeChannelAlertPolicy(&after)
			event := newNotificationPolicyEvent(after.Channel, "UPDATE", actor, notificationPolicySummary(before), notificationPolicySummary(after), note, time.Now().UTC().Format(time.RFC3339))
			s.policyEvents = append([]NotificationPolicyEvent{event}, s.policyEvents...)
			if len(s.policyEvents) > 30 {
				s.policyEvents = s.policyEvents[:30]
			}
			alerts := channelAlerts(s.channelFailures)
			return channelAlertPolicies([]ChannelAlertPolicy{s.alertPolicies[idx]}, alerts)[0], nil
		}
	}
	return ChannelAlertPolicy{}, fmt.Errorf("channel alert policy %s not found", channel)
}

func (s *Store) ApproveNotificationPolicyChange(id string, approver string, note string, confirmation string) (ChannelAlertPolicy, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.expireNotificationPolicyChangesLocked(time.Now().UTC())
	for changeIdx := range s.policyChanges {
		if s.policyChanges[changeIdx].ID != id {
			continue
		}
		if s.policyChanges[changeIdx].Status != "PENDING" {
			return ChannelAlertPolicy{}, fmt.Errorf("notification policy change %s is not pending", id)
		}
		approvedBy := fallback(approver, "ops-lead")
		if err := requireNotificationPolicyApprover(approvedBy); err != nil {
			return ChannelAlertPolicy{}, err
		}
		if err := requireNotificationPolicyConfirmation("APPROVE", s.policyChanges[changeIdx].Channel, confirmation); err != nil {
			return ChannelAlertPolicy{}, err
		}
		for policyIdx := range s.alertPolicies {
			if !strings.EqualFold(s.alertPolicies[policyIdx].Channel, s.policyChanges[changeIdx].Channel) {
				continue
			}
			before := s.alertPolicies[policyIdx]
			normalizeChannelAlertPolicy(&before)
			s.alertPolicies[policyIdx].TargetURL = s.policyChanges[changeIdx].TargetURL
			s.alertPolicies[policyIdx].SecretRef = s.policyChanges[changeIdx].SecretRef
			s.alertPolicies[policyIdx].MaxAttempts = s.policyChanges[changeIdx].MaxAttempts
			s.alertPolicies[policyIdx].BackoffSeconds = s.policyChanges[changeIdx].BackoffSeconds
			after := s.alertPolicies[policyIdx]
			normalizeChannelAlertPolicy(&after)
			now := time.Now().UTC().Format(time.RFC3339)
			s.policyChanges[changeIdx].Status = "APPROVED"
			s.policyChanges[changeIdx].ApprovedBy = approvedBy
			s.policyChanges[changeIdx].ApprovedAt = now
			event := newNotificationPolicyEvent(after.Channel, "APPROVE", approvedBy, notificationPolicySummary(before), notificationPolicySummary(after), fallback(note, s.policyChanges[changeIdx].Note), now)
			s.policyEvents = append([]NotificationPolicyEvent{event}, s.policyEvents...)
			return channelAlertPolicies([]ChannelAlertPolicy{s.alertPolicies[policyIdx]}, channelAlerts(s.channelFailures))[0], nil
		}
	}
	return ChannelAlertPolicy{}, fmt.Errorf("notification policy change %s not found", id)
}

func (s *Store) RejectNotificationPolicyChange(id string, reviewer string, note string) (NotificationPolicyChange, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.expireNotificationPolicyChangesLocked(time.Now().UTC())
	for changeIdx := range s.policyChanges {
		if s.policyChanges[changeIdx].ID != id {
			continue
		}
		if s.policyChanges[changeIdx].Status != "PENDING" {
			return NotificationPolicyChange{}, fmt.Errorf("notification policy change %s is not pending", id)
		}
		reviewedBy := fallback(reviewer, "ops-lead")
		if err := requireNotificationPolicyApprover(reviewedBy); err != nil {
			return NotificationPolicyChange{}, err
		}
		now := time.Now().UTC().Format(time.RFC3339)
		s.policyChanges[changeIdx].Status = "REJECTED"
		s.policyChanges[changeIdx].ApprovedBy = reviewedBy
		s.policyChanges[changeIdx].ApprovedAt = now
		before := notificationPolicyFor(s.alertPolicies, s.policyChanges[changeIdx].Channel)
		event := newNotificationPolicyEvent(s.policyChanges[changeIdx].Channel, "REJECT", reviewedBy, notificationPolicySummary(before), notificationPolicyChangeSummary(s.policyChanges[changeIdx]), fallback(note, s.policyChanges[changeIdx].Note), now)
		s.policyEvents = append([]NotificationPolicyEvent{event}, s.policyEvents...)
		return s.policyChanges[changeIdx], nil
	}
	return NotificationPolicyChange{}, fmt.Errorf("notification policy change %s not found", id)
}

func (s *Store) CancelNotificationPolicyChange(id string, actor string, note string) (NotificationPolicyChange, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.expireNotificationPolicyChangesLocked(time.Now().UTC())
	for changeIdx := range s.policyChanges {
		if s.policyChanges[changeIdx].ID != id {
			continue
		}
		if s.policyChanges[changeIdx].Status != "PENDING" {
			return NotificationPolicyChange{}, fmt.Errorf("notification policy change %s is not pending", id)
		}
		now := time.Now().UTC().Format(time.RFC3339)
		s.policyChanges[changeIdx].Status = "CANCELED"
		s.policyChanges[changeIdx].ApprovedBy = fallback(actor, "ops-a")
		s.policyChanges[changeIdx].ApprovedAt = now
		before := notificationPolicyFor(s.alertPolicies, s.policyChanges[changeIdx].Channel)
		event := newNotificationPolicyEvent(s.policyChanges[changeIdx].Channel, "CANCEL", actor, notificationPolicySummary(before), notificationPolicyChangeSummary(s.policyChanges[changeIdx]), fallback(note, s.policyChanges[changeIdx].Note), now)
		s.policyEvents = append([]NotificationPolicyEvent{event}, s.policyEvents...)
		return s.policyChanges[changeIdx], nil
	}
	return NotificationPolicyChange{}, fmt.Errorf("notification policy change %s not found", id)
}

func (s *Store) RollbackChannelAlertPolicy(channel string, actor string, note string, confirmation string) (ChannelAlertPolicy, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	normalized := strings.TrimSpace(channel)
	if normalized == "" {
		return ChannelAlertPolicy{}, fmt.Errorf("channel is required")
	}
	rolledBackBy := fallback(actor, "ops-lead")
	if err := requireNotificationPolicyApprover(rolledBackBy); err != nil {
		return ChannelAlertPolicy{}, err
	}
	if err := requireNotificationPolicyConfirmation("ROLLBACK", normalized, confirmation); err != nil {
		return ChannelAlertPolicy{}, err
	}
	for policyIdx := range s.alertPolicies {
		if !strings.EqualFold(s.alertPolicies[policyIdx].Channel, normalized) {
			continue
		}
		rollback, ok := latestApprovedNotificationPolicyChange(s.policyChanges, s.alertPolicies[policyIdx].Channel)
		if !ok {
			return ChannelAlertPolicy{}, fmt.Errorf("channel alert policy %s has no approved change to rollback", channel)
		}
		before := s.alertPolicies[policyIdx]
		normalizeChannelAlertPolicy(&before)
		s.alertPolicies[policyIdx].TargetURL = rollback.CurrentTargetURL
		s.alertPolicies[policyIdx].SecretRef = rollback.CurrentSecretRef
		s.alertPolicies[policyIdx].MaxAttempts = normalizeMaxAttempts(rollback.CurrentMaxAttempts)
		s.alertPolicies[policyIdx].BackoffSeconds = normalizeBackoffSeconds(rollback.CurrentBackoffSeconds)
		after := s.alertPolicies[policyIdx]
		normalizeChannelAlertPolicy(&after)
		now := time.Now().UTC().Format(time.RFC3339)
		event := newNotificationPolicyEvent(after.Channel, "ROLLBACK", rolledBackBy, notificationPolicySummary(before), notificationPolicySummary(after), fallback(note, "通知目标已回滚到上一版"), now)
		s.policyEvents = append([]NotificationPolicyEvent{event}, s.policyEvents...)
		if len(s.policyEvents) > 30 {
			s.policyEvents = s.policyEvents[:30]
		}
		return channelAlertPolicies([]ChannelAlertPolicy{s.alertPolicies[policyIdx]}, channelAlerts(s.channelFailures))[0], nil
	}
	return ChannelAlertPolicy{}, fmt.Errorf("channel alert policy %s not found", channel)
}

func (s *Store) DispatchChannelNotification(id, outcome string) (ChannelNotification, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delivery := s.delivery
	if delivery == nil {
		delivery = demoNotificationDeliveryClient{}
	}
	for idx := range s.notifications {
		if s.notifications[idx].ID == id {
			s.notifications[idx] = dispatchChannelNotification(s.notifications[idx], outcome, delivery, time.Now().UTC())
			return s.notifications[idx], nil
		}
	}
	return ChannelNotification{}, fmt.Errorf("channel notification %s not found", id)
}

func (s *Store) AcknowledgeChannelNotification(id, actor, note string) (ChannelNotification, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for idx := range s.notifications {
		if s.notifications[idx].ID == id {
			s.notifications[idx].Status = "ACKED"
			s.notifications[idx].AckedBy = fallback(actor, "operator")
			s.notifications[idx].AckNote = fallback(note, "已确认渠道告警")
			s.notifications[idx].AckedAt = time.Now().UTC().Format(time.RFC3339)
			return s.notifications[idx], nil
		}
	}
	return ChannelNotification{}, fmt.Errorf("channel notification %s not found", id)
}

func (s *Store) CompleteChannelRunbookCheck(check ChannelRunbookCheck) (ChannelRunbookCheck, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	normalizeChannelRunbookCheck(&check)
	for idx, item := range s.runbookChecks {
		if sameChannelRunbookCheck(item, check) {
			check.ID = item.ID
			s.runbookChecks[idx] = check
			return check, nil
		}
	}
	s.runbookChecks = append([]ChannelRunbookCheck{check}, s.runbookChecks...)
	return check, nil
}

func (s *Store) ChannelIntegration(channel string) (ChannelIntegration, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return channelIntegrationFor(s.integrations, channel), nil
}

func (s *Store) RecordChannelRateLimit(channel string, windowStart time.Time, limit int) (bool, int, error) {
	if limit <= 0 {
		return true, 0, nil
	}
	key := strings.ToLower(strings.TrimSpace(channel)) + ":" + windowStart.UTC().Format(time.RFC3339)
	s.mu.Lock()
	defer s.mu.Unlock()
	next := s.rateWindows[key] + 1
	if next > limit {
		return false, s.rateWindows[key], nil
	}
	s.rateWindows[key] = next
	return true, next, nil
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
	s.mu.Lock()
	defer s.mu.Unlock()
	result := evaluateRules(content, s.searchLocked(content), rulesByStage(s.rules, "active"))
	s.recordRuleHitLocked(result.RuleCode, time.Now().UTC().Format(time.RFC3339))
	return result, nil
}

func (s *Store) CompareRuleVersions(content string) (RuleComparison, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	evidence := s.searchLocked(content)
	current := evaluateRules(content, evidence, rulesByStage(s.rules, "active"))
	canary := evaluateRules(content, evidence, rulesForComparison(s.rules))
	return compareRuleResults(content, current, canary), nil
}

func (s *Store) SubmitRuleApproval(code, approver, riskLevel, note string, sampleIDs []string) (RuleApproval, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.ruleIndexLocked(code, "canary") < 0 {
		return RuleApproval{}, fmt.Errorf("canary rule %s not found", code)
	}
	approval := newRuleApproval(code, approver, riskLevel, note, sampleIDs, time.Now().UTC().Format(time.RFC3339))
	s.ruleApprovals = append([]RuleApproval{approval}, s.ruleApprovals...)
	return approval, nil
}

func (s *Store) PublishCanaryRule(code, actor, note string) (RuleReleaseEvent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	canaryIndex := s.ruleIndexLocked(code, "canary")
	if canaryIndex < 0 {
		return RuleReleaseEvent{}, fmt.Errorf("canary rule %s not found", code)
	}
	if !hasApprovedRuleGate(s.ruleApprovals, code) {
		return RuleReleaseEvent{}, fmt.Errorf("rule %s requires approved gate before publish", code)
	}
	now := time.Now().UTC().Format(time.RFC3339)
	for idx := range s.rules {
		if s.rules[idx].Code == code && fallback(s.rules[idx].Stage, "active") == "active" {
			s.rules[idx].Stage = "archived"
			s.rules[idx].Enabled = false
		}
	}
	s.rules[canaryIndex].Stage = "active"
	s.rules[canaryIndex].Version = nextPublishedVersion(s.rules[canaryIndex].Version)
	event := newRuleReleaseEvent(code, s.rules[canaryIndex].Version, "PUBLISH", actor, note, now)
	s.ruleEvents = append([]RuleReleaseEvent{event}, s.ruleEvents...)
	return event, nil
}

func (s *Store) RollbackRule(code, actor, note string) (RuleReleaseEvent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	activeIndex := s.ruleIndexLocked(code, "active")
	if activeIndex < 0 {
		return RuleReleaseEvent{}, fmt.Errorf("active rule %s not found", code)
	}
	now := time.Now().UTC().Format(time.RFC3339)
	s.rules[activeIndex].Enabled = false
	s.rules[activeIndex].Stage = "canary"
	event := newRuleReleaseEvent(code, s.rules[activeIndex].Version, "ROLLBACK", actor, note, now)
	s.ruleEvents = append([]RuleReleaseEvent{event}, s.ruleEvents...)
	return event, nil
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
	s.completeReviewTaskByMessageLocked(messageID, annotation.Reviewer, annotation.CreatedAt)
	return annotation, nil
}

func (s *Store) AssignReviewTask(id, assignee string) (ReviewTask, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for idx := range s.reviewTasks {
		if s.reviewTasks[idx].ID == id {
			if s.reviewTasks[idx].Status != "COMPLETED" {
				s.reviewTasks[idx].Status = "ASSIGNED"
				s.reviewTasks[idx].Assignee = fallback(assignee, "qa-operator")
			}
			return s.reviewTasks[idx], nil
		}
	}
	return ReviewTask{}, fmt.Errorf("review task %s not found", id)
}

func (s *Store) CompleteReviewTask(id string) (ReviewTask, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for idx := range s.reviewTasks {
		if s.reviewTasks[idx].ID == id {
			s.reviewTasks[idx].Status = "COMPLETED"
			s.reviewTasks[idx].CompletedAt = time.Now().UTC().Format(time.RFC3339)
			if s.reviewTasks[idx].Assignee == "" {
				s.reviewTasks[idx].Assignee = "qa-operator"
			}
			return s.reviewTasks[idx], nil
		}
	}
	return ReviewTask{}, fmt.Errorf("review task %s not found", id)
}

func (s *Store) ExportTrainingSamples(maxScore int) ([]TrainingSample, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return trainingSamples(s.annotations, s.messages, s.conversations, normalizeMaxScore(maxScore)), nil
}

func (s *Store) SaveChannelOpsReport(report ChannelOpsReport) (ChannelOpsReport, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	normalizeChannelOpsReport(&report)
	s.channelReports = append([]ChannelOpsReport{report}, s.channelReports...)
	if len(s.channelReports) > 365 {
		s.channelReports = s.channelReports[:365]
	}
	return report, nil
}

func (s *Store) ListChannelOpsReports(limit int) ([]ChannelOpsReport, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	limit = normalizeReportLimit(limit)
	items := append([]ChannelOpsReport(nil), s.channelReports...)
	if len(items) > limit {
		items = items[:limit]
	}
	for idx := range items {
		items[idx].Content = ""
	}
	return items, nil
}

func (s *Store) ChannelOpsReport(id string) (ChannelOpsReport, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, item := range s.channelReports {
		if item.ID == id {
			return item, nil
		}
	}
	return ChannelOpsReport{}, fmt.Errorf("channel ops report %s not found", id)
}

func (s *Store) PruneChannelOpsReports(retain int) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	retain = normalizeReportRetention(retain)
	if len(s.channelReports) <= retain {
		return 0, nil
	}
	pruned := len(s.channelReports) - retain
	s.channelReports = append([]ChannelOpsReport(nil), s.channelReports[:retain]...)
	return pruned, nil
}

func (s *Store) RecordChannelOpsReportEvent(event ChannelOpsReportEvent) (ChannelOpsReportEvent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	normalizeChannelOpsReportEvent(&event)
	s.reportEvents = append([]ChannelOpsReportEvent{event}, s.reportEvents...)
	if len(s.reportEvents) > 100 {
		s.reportEvents = s.reportEvents[:100]
	}
	return event, nil
}

func (s *Store) ListChannelOpsReportEvents(limit int) ([]ChannelOpsReportEvent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	limit = normalizeReportEventLimit(limit)
	items := append([]ChannelOpsReportEvent(nil), s.reportEvents...)
	if len(items) > limit {
		items = items[:limit]
	}
	return items, nil
}

func (s *Store) Dashboard() (Dashboard, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.expireNotificationPolicyChangesLocked(time.Now().UTC())
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
	openReviews := 0
	for _, task := range s.reviewTasks {
		if task.Status != "COMPLETED" {
			openReviews++
		}
	}
	transfers := withTransferSLAs(s.tickets, s.channelPolicies, time.Now().UTC())
	quality := qualitySummary(s.messages, s.gaps, transfers, s.annotations)
	channelAlerts := channelAlerts(s.channelFailures)
	alertPolicies := channelAlertPolicies(s.alertPolicies, channelAlerts)
	return Dashboard{
		Metrics: []Metric{
			{Label: "Active sessions", Value: fmt.Sprintf("%d", len(s.conversations)), Note: "in-memory V1 runtime"},
			{Label: "Knowledge items", Value: fmt.Sprintf("%d", len(s.knowledge)), Note: "seeded trusted articles"},
			{Label: "Open gaps", Value: fmt.Sprintf("%d", openGaps), Note: "created by no-evidence fallback"},
			{Label: "Channels", Value: fmt.Sprintf("%d", activeChannelCount(s.conversations)), Note: "conversation sources with policies"},
			{Label: "Open transfers", Value: fmt.Sprintf("%d", openTransfers), Note: "waiting for human agents"},
			{Label: "SLA escalations", Value: fmt.Sprintf("%d", escalatedTransferCount(transfers)), Note: "open tickets past response SLA"},
			{Label: "Enabled rules", Value: fmt.Sprintf("%d", enabledRules(s.rules)), Note: "guardrail and transfer policies"},
			{Label: "Channel failures", Value: fmt.Sprintf("%d", channelAlertCount(channelAlerts)), Note: "rejected inbound requests"},
			{Label: "Active alerts", Value: fmt.Sprintf("%d", activeAlertPolicies(alertPolicies)), Note: "notification policies triggered"},
			{Label: "Review tasks", Value: fmt.Sprintf("%d", openReviews), Note: "assistant replies awaiting QA"},
		},
		Conversations:    append([]Conversation(nil), s.conversations...),
		KnowledgeGaps:    append([]KnowledgeGap(nil), s.gaps...),
		Rules:            append([]Rule(nil), s.rules...),
		Transfers:        transfers,
		ChannelPolicies:  append([]ChannelPolicy(nil), s.channelPolicies...),
		Integrations:     append([]ChannelIntegration(nil), s.integrations...),
		ChannelAlerts:    channelAlerts,
		ChannelTrends:    channelFailureTrends(s.channelFailures, time.Now().UTC()),
		ChannelAudits:    append([]ChannelInboundAudit(nil), s.inboundAudits...),
		AuditEvents:      append([]ChannelInboundAuditQualityEvent(nil), s.auditEvents...),
		AlertPolicies:    alertPolicies,
		Notifications:    append([]ChannelNotification(nil), s.notifications...),
		ChannelRunbooks:  channelRunbooks(channelAlerts, alertPolicies, s.notifications, s.inboundAudits, s.runbookChecks),
		PolicyEvents:     append([]NotificationPolicyEvent(nil), s.policyEvents...),
		PolicyChanges:    append([]NotificationPolicyChange(nil), s.policyChanges...),
		Quality:          quality,
		Annotations:      append([]Annotation(nil), s.annotations...),
		ReviewTasks:      append([]ReviewTask(nil), s.reviewTasks...),
		RuleApprovals:    append([]RuleApproval(nil), s.ruleApprovals...),
		RuleEvents:       append([]RuleReleaseEvent(nil), s.ruleEvents...),
		RuleObservations: ruleReleaseObservations(s.ruleEvents, s.rules, transfers, s.annotations, time.Now().UTC()),
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

func agentReply(generator ReplyGenerator, conversationID, content string, evidence []KnowledgeArticle, history []Message, ruleResult RuleTestResult, now string) (Message, *KnowledgeGap) {
	if ruleResult.Action == "recommend_human_transfer" {
		return Message{
			ID: fmt.Sprintf("msg_%d_agent", time.Now().UnixNano()), ConversationID: conversationID, Role: "assistant",
			Content: "我已经为你转接人工客服，并保留当前对话上下文。人工客服接手前，我不会编造处理结果。",
			Engine:  "rule", Safe: true, FallbackReason: fallback(ruleResult.RuleCode, "TRANSFER_THRESHOLD"), CreatedAt: now,
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

func shouldCreateTransfer(reason string) bool {
	return reason == "TRANSFER_THRESHOLD" || reason == "CANCEL_RISK_TRANSFER"
}

func enabledRules(rules []Rule) int {
	total := 0
	for _, rule := range rules {
		if rule.Enabled && fallback(rule.Stage, "active") == "active" {
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
	if shouldCancelTransfer(content) && ruleEnabled(rules, "CANCEL_RISK_TRANSFER") {
		result.Matched = true
		result.RuleCode = "CANCEL_RISK_TRANSFER"
		result.Action = "recommend_human_transfer"
		result.RiskLevel = "MEDIUM"
		result.Fallback = true
		result.Reason = "灰度规则认为取消、退订或退款争议需要人工确认边界。"
		result.Recommended = "进入人工复核或继续观察灰度命中率后再发布。"
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

func shouldCancelTransfer(content string) bool {
	text := strings.ToLower(content)
	return strings.Contains(text, "取消订单") || strings.Contains(text, "退订") || strings.Contains(text, "退款争议") || strings.Contains(text, "取消服务")
}

func ruleEnabled(rules []Rule, code string) bool {
	for _, rule := range rules {
		if rule.Code == code && rule.Enabled {
			return true
		}
	}
	return false
}

func rulesByStage(rules []Rule, stage string) []Rule {
	stage = fallback(stage, "active")
	filtered := make([]Rule, 0, len(rules))
	for _, rule := range rules {
		if fallback(rule.Stage, "active") == stage {
			filtered = append(filtered, rule)
		}
	}
	return filtered
}

func rulesForComparison(rules []Rule) []Rule {
	filtered := make([]Rule, 0, len(rules))
	for _, rule := range rules {
		stage := fallback(rule.Stage, "active")
		if stage == "active" || stage == "canary" {
			filtered = append(filtered, rule)
		}
	}
	return filtered
}

func compareRuleResults(content string, current, canary RuleTestResult) RuleComparison {
	changed := current.RuleCode != canary.RuleCode || current.Action != canary.Action || current.RiskLevel != canary.RiskLevel || current.Fallback != canary.Fallback
	recommendation := "灰度规则与当前规则一致，可以继续观察命中率。"
	if changed {
		recommendation = "灰度规则改变了处置结果，发布前需要复核样本和人工队列压力。"
	}
	return RuleComparison{
		Input:          content,
		Current:        current,
		Canary:         canary,
		Changed:        changed,
		Recommendation: recommendation,
	}
}

func newRuleReleaseEvent(code, version, action, actor, note, createdAt string) RuleReleaseEvent {
	return RuleReleaseEvent{
		ID:        fmt.Sprintf("rule_event_%d", time.Now().UnixNano()),
		RuleCode:  code,
		Version:   fallback(version, "2026-06-active"),
		Action:    action,
		Actor:     fallback(actor, "operator"),
		Note:      fallback(note, "规则发布动作已记录"),
		CreatedAt: createdAt,
	}
}

func newNotificationPolicyEvent(channel, action, actor, before, after, note, createdAt string) NotificationPolicyEvent {
	return NotificationPolicyEvent{
		ID:        fmt.Sprintf("notification_policy_event_%d", time.Now().UnixNano()),
		Channel:   channel,
		Action:    fallback(action, "UPDATE"),
		Actor:     fallback(actor, "ops-a"),
		Before:    before,
		After:     after,
		Note:      fallback(note, "通知策略配置已更新"),
		CreatedAt: createdAt,
	}
}

func newNotificationPolicyChange(current ChannelAlertPolicy, targetURL, secretRef string, maxAttempts, backoffSeconds int, requestedBy, note, createdAt string) NotificationPolicyChange {
	normalizeChannelAlertPolicy(&current)
	createdTime, err := time.Parse(time.RFC3339, createdAt)
	if err != nil {
		createdTime = time.Now().UTC()
		createdAt = createdTime.Format(time.RFC3339)
	}
	next := NotificationPolicyChange{
		Channel:               current.Channel,
		TargetURL:             strings.TrimSpace(targetURL),
		SecretRef:             strings.TrimSpace(secretRef),
		MaxAttempts:           normalizeMaxAttempts(maxAttempts),
		BackoffSeconds:        normalizeBackoffSeconds(backoffSeconds),
		CurrentTargetURL:      current.TargetURL,
		CurrentSecretRef:      current.SecretRef,
		CurrentMaxAttempts:    current.MaxAttempts,
		CurrentBackoffSeconds: current.BackoffSeconds,
	}
	next.Diff = notificationPolicyDiff(next)
	return NotificationPolicyChange{
		ID:                    fmt.Sprintf("notification_policy_change_%d", time.Now().UnixNano()),
		Channel:               next.Channel,
		TargetURL:             next.TargetURL,
		SecretRef:             next.SecretRef,
		MaxAttempts:           next.MaxAttempts,
		BackoffSeconds:        next.BackoffSeconds,
		CurrentTargetURL:      next.CurrentTargetURL,
		CurrentSecretRef:      next.CurrentSecretRef,
		CurrentMaxAttempts:    next.CurrentMaxAttempts,
		CurrentBackoffSeconds: next.CurrentBackoffSeconds,
		Diff:                  next.Diff,
		ConfirmationText:      notificationPolicyConfirmationText("APPROVE", next.Channel),
		RequestedBy:           fallback(requestedBy, "ops-a"),
		Status:                "PENDING",
		Note:                  fallback(note, "通知策略变更待审批"),
		CreatedAt:             createdAt,
		ExpiresAt:             createdTime.Add(notificationPolicyChangeTTL()).UTC().Format(time.RFC3339),
	}
}

func notificationPolicySummary(policy ChannelAlertPolicy) string {
	normalizeChannelAlertPolicy(&policy)
	return fmt.Sprintf("targetUrl=%s secretRef=%s maxAttempts=%d backoffSeconds=%d inboundAuditMinSamples=%d inboundAuditMinAcceptanceRate=%d inboundAuditMaxErrorCount=%d", policy.TargetURL, policy.SecretRef, policy.MaxAttempts, policy.BackoffSeconds, policy.InboundAuditMinSamples, policy.InboundAuditMinAcceptanceRate, policy.InboundAuditMaxErrorCount)
}

func notificationPolicyChangeSummary(change NotificationPolicyChange) string {
	return fmt.Sprintf("targetUrl=%s secretRef=%s maxAttempts=%d backoffSeconds=%d", change.TargetURL, change.SecretRef, change.MaxAttempts, change.BackoffSeconds)
}

func notificationPolicyDiff(change NotificationPolicyChange) []NotificationPolicyDiff {
	diffs := make([]NotificationPolicyDiff, 0, 4)
	appendDiff := func(field, before, after string) {
		if before != after {
			diffs = append(diffs, NotificationPolicyDiff{Field: field, Before: before, After: after})
		}
	}
	appendDiff("targetUrl", change.CurrentTargetURL, change.TargetURL)
	appendDiff("secretRef", change.CurrentSecretRef, change.SecretRef)
	appendDiff("maxAttempts", fmt.Sprintf("%d", normalizeMaxAttempts(change.CurrentMaxAttempts)), fmt.Sprintf("%d", normalizeMaxAttempts(change.MaxAttempts)))
	appendDiff("backoffSeconds", fmt.Sprintf("%d", normalizeBackoffSeconds(change.CurrentBackoffSeconds)), fmt.Sprintf("%d", normalizeBackoffSeconds(change.BackoffSeconds)))
	return diffs
}

func latestApprovedNotificationPolicyChange(changes []NotificationPolicyChange, channel string) (NotificationPolicyChange, bool) {
	for _, change := range changes {
		if change.Status == "APPROVED" && strings.EqualFold(change.Channel, channel) && strings.TrimSpace(change.CurrentTargetURL) != "" && strings.TrimSpace(change.CurrentSecretRef) != "" {
			change.Diff = notificationPolicyDiff(change)
			change.ConfirmationText = notificationPolicyConfirmationText("APPROVE", change.Channel)
			return change, true
		}
	}
	return NotificationPolicyChange{}, false
}

func requireNotificationPolicyApprover(actor string) error {
	switch strings.ToLower(strings.TrimSpace(actor)) {
	case "ops-lead", "security-owner", "platform-owner":
		return nil
	default:
		return fmt.Errorf("notification policy actor %s is not authorized", fallback(actor, "unknown"))
	}
}

func requireNotificationPolicyConfirmation(action, channel, confirmation string) error {
	expected := notificationPolicyConfirmationText(action, channel)
	if strings.TrimSpace(confirmation) != expected {
		return fmt.Errorf("confirmation %q is required", expected)
	}
	return nil
}

func notificationPolicyConfirmationText(action, channel string) string {
	return fmt.Sprintf("%s %s", strings.ToUpper(strings.TrimSpace(action)), strings.TrimSpace(channel))
}

func notificationPolicyChangeTTL() time.Duration {
	return 24 * time.Hour
}

func (s *Store) expireNotificationPolicyChangesLocked(now time.Time) {
	for idx := range s.policyChanges {
		if s.policyChanges[idx].Status != "PENDING" || !notificationPolicyChangeExpired(s.policyChanges[idx], now) {
			continue
		}
		s.policyChanges[idx].Status = "EXPIRED"
		s.policyChanges[idx].ApprovedBy = "system"
		s.policyChanges[idx].ApprovedAt = now.UTC().Format(time.RFC3339)
		before := notificationPolicyFor(s.alertPolicies, s.policyChanges[idx].Channel)
		event := newNotificationPolicyEvent(s.policyChanges[idx].Channel, "EXPIRE", "system", notificationPolicySummary(before), notificationPolicyChangeSummary(s.policyChanges[idx]), "通知策略变更审批超时自动过期", now.UTC().Format(time.RFC3339))
		s.policyEvents = append([]NotificationPolicyEvent{event}, s.policyEvents...)
	}
	if len(s.policyEvents) > 30 {
		s.policyEvents = s.policyEvents[:30]
	}
}

func notificationPolicyChangeExpired(change NotificationPolicyChange, now time.Time) bool {
	if strings.TrimSpace(change.ExpiresAt) == "" {
		return false
	}
	expiresAt, err := time.Parse(time.RFC3339, change.ExpiresAt)
	if err != nil {
		return false
	}
	return !expiresAt.After(now.UTC())
}

func requiresNotificationPolicyApproval(policy ChannelAlertPolicy) bool {
	return policy.Severity == "HIGH" || policy.Severity == "CRITICAL"
}

func notificationPolicyFor(policies []ChannelAlertPolicy, channel string) ChannelAlertPolicy {
	for _, policy := range policies {
		if strings.EqualFold(policy.Channel, channel) {
			normalizeChannelAlertPolicy(&policy)
			return policy
		}
	}
	policy := newDefaultChannelAlertPolicy(channel, "MEDIUM", 5, 60, "ops-webhook")
	normalizeChannelAlertPolicy(&policy)
	return policy
}

func newRuleApproval(code, approver, riskLevel, note string, sampleIDs []string, createdAt string) RuleApproval {
	riskLevel = strings.ToUpper(fallback(riskLevel, "LOW"))
	sampleIDs = normalizeSampleIDs(sampleIDs)
	sampleCount := len(sampleIDs)
	status := "REJECTED"
	if sampleCount >= 3 && riskLevel != "HIGH" && riskLevel != "CRITICAL" {
		status = "APPROVED"
	}
	return RuleApproval{
		ID:          fmt.Sprintf("rule_approval_%d", time.Now().UnixNano()),
		RuleCode:    code,
		Approver:    fallback(approver, "qa-lead"),
		RiskLevel:   riskLevel,
		SampleIDs:   sampleIDs,
		SampleCount: sampleCount,
		Status:      status,
		Note:        fallback(note, "规则发布审批"),
		CreatedAt:   createdAt,
	}
}

func hasApprovedRuleGate(approvals []RuleApproval, code string) bool {
	for _, approval := range approvals {
		if approval.RuleCode == code && approval.Status == "APPROVED" && approval.SampleCount >= 3 {
			return true
		}
	}
	return false
}

func ruleReleaseObservations(events []RuleReleaseEvent, rules []Rule, tickets []TransferTicket, annotations []Annotation, now time.Time) []RuleReleaseObservation {
	observations := make([]RuleReleaseObservation, 0)
	seen := map[string]bool{}
	for _, event := range events {
		if event.Action != "PUBLISH" || seen[event.RuleCode] {
			continue
		}
		seen[event.RuleCode] = true
		releasedAt, ok := parseRFC3339(event.CreatedAt)
		if !ok {
			releasedAt = now
		}
		rule := ruleByCodeAndVersion(rules, event.RuleCode, event.Version)
		transferCount := countTransfersSince(tickets, releasedAt)
		lowScoreCount, averageReview := annotationStatsSince(annotations, releasedAt)
		riskLevel, rollbackRecommended, recommendation := releaseRecommendation(rule.HitCount, transferCount, lowScoreCount, averageReview)
		observations = append(observations, RuleReleaseObservation{
			RuleCode:            event.RuleCode,
			Version:             event.Version,
			ReleasedAt:          event.CreatedAt,
			Window:              observationWindow(releasedAt, now),
			RuleHits:            rule.HitCount,
			TransferTickets:     transferCount,
			LowScoreSamples:     lowScoreCount,
			AverageReview:       averageReview,
			RiskLevel:           riskLevel,
			Recommendation:      recommendation,
			RollbackRecommended: rollbackRecommended,
		})
	}
	return observations
}

func ruleByCodeAndVersion(rules []Rule, code, version string) Rule {
	for _, rule := range rules {
		if rule.Code == code && rule.Version == version {
			return rule
		}
	}
	for _, rule := range rules {
		if rule.Code == code {
			return rule
		}
	}
	return Rule{Code: code, Version: version}
}

func countTransfersSince(tickets []TransferTicket, since time.Time) int {
	count := 0
	for _, ticket := range tickets {
		createdAt, ok := parseRFC3339(ticket.CreatedAt)
		if ok && !createdAt.Before(since) {
			count++
		}
	}
	return count
}

func annotationStatsSince(annotations []Annotation, since time.Time) (int, int) {
	lowScoreCount := 0
	totalScore := 0
	reviewed := 0
	for _, annotation := range annotations {
		createdAt, ok := parseRFC3339(annotation.CreatedAt)
		if !ok || createdAt.Before(since) {
			continue
		}
		reviewed++
		totalScore += annotation.Score
		if needsReview(annotation, 80) {
			lowScoreCount++
		}
	}
	if reviewed == 0 {
		return lowScoreCount, 0
	}
	return lowScoreCount, totalScore / reviewed
}

func releaseRecommendation(ruleHits, transferCount, lowScoreCount, averageReview int) (string, bool, string) {
	if lowScoreCount >= 3 || transferCount >= 5 || (averageReview > 0 && averageReview < 60) {
		return "HIGH", true, "建议暂停放量并回滚，发布后低分样本或人工压力已经超过课堂阈值。"
	}
	if lowScoreCount > 0 || transferCount >= 2 || (averageReview > 0 && averageReview < 75) {
		return "MEDIUM", false, "建议继续小流量观察，优先复盘低分样本和转人工原因。"
	}
	if ruleHits == 0 {
		return "LOW", false, "已发布但暂无命中，保持观察并等待真实流量。"
	}
	return "LOW", false, "发布后指标稳定，可以继续观察或逐步放量。"
}

func observationWindow(since, now time.Time) string {
	if now.Before(since) {
		return "0m"
	}
	duration := now.Sub(since)
	if duration < time.Hour {
		return fmt.Sprintf("%dm", int(duration.Minutes()))
	}
	if duration < 24*time.Hour {
		return fmt.Sprintf("%dh", int(duration.Hours()))
	}
	return fmt.Sprintf("%dd", int(duration.Hours()/24))
}

func parseRFC3339(value string) (time.Time, bool) {
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, false
	}
	return parsed, true
}

func normalizeSampleIDs(sampleIDs []string) []string {
	result := make([]string, 0, len(sampleIDs))
	seen := map[string]bool{}
	for _, id := range sampleIDs {
		id = strings.TrimSpace(id)
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		result = append(result, id)
	}
	return result
}

func nextPublishedVersion(version string) string {
	version = fallback(version, "2026-06-canary")
	if strings.Contains(version, "active") {
		return version
	}
	return strings.Replace(version, "canary", "active", 1)
}

func (s *Store) ruleIndexLocked(code, stage string) int {
	for idx := range s.rules {
		if s.rules[idx].Code == code && fallback(s.rules[idx].Stage, "active") == stage {
			return idx
		}
	}
	return -1
}

func (s *Store) recordRuleHitLocked(code, hitAt string) {
	if strings.TrimSpace(code) == "" {
		return
	}
	for idx := range s.rules {
		if s.rules[idx].Code == code && fallback(s.rules[idx].Stage, "active") == "active" {
			s.rules[idx].HitCount++
			s.rules[idx].LastHitAt = hitAt
			return
		}
	}
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
		{Channel: "WeCom", DisplayName: "企业微信客服", Tone: "专业、内部协同、保留上下文", SLAMinutes: 12, RiskBoost: "HIGH", EscalationNote: "企业微信通常连接私域客户和内部协同，投诉和合同问题优先接管。", Enabled: true},
		{Channel: "App", DisplayName: "App 客服", Tone: "直接、产品化、引导自助", SLAMinutes: 20, RiskBoost: "NORMAL", EscalationNote: "App 内问题优先引导订单和售后入口。", Enabled: true},
		{Channel: "Marketplace", DisplayName: "平台店铺客服", Tone: "谨慎、合规、避免承诺", SLAMinutes: 10, RiskBoost: "HIGH", EscalationNote: "平台投诉可能影响店铺指标，优先升级。", Enabled: true},
		{Channel: "Douyin", DisplayName: "抖音客服", Tone: "短句、快速确认、避免营销承诺", SLAMinutes: 8, RiskBoost: "HIGH", EscalationNote: "直播和短视频渠道情绪扩散快，投诉和售后争议优先升级。", Enabled: true},
		{Channel: "Xiaohongshu", DisplayName: "小红书客服", Tone: "温和、解释充分、注意口碑风险", SLAMinutes: 10, RiskBoost: "HIGH", EscalationNote: "种草和口碑场景要关注公开评价风险，争议内容优先接管。", Enabled: true},
	}
}

func defaultChannelIntegrations(now string) []ChannelIntegration {
	return []ChannelIntegration{
		{Channel: "Web", DisplayName: "Web 客服", Enabled: true, SecretSource: "env", SecretRef: "ANJING_CHANNEL_WEB_SECRET", NextSecretRef: "ANJING_CHANNEL_WEB_NEXT_SECRET", SignatureWindowSeconds: 300, ReplayProtection: true, AllowedOrigins: []string{"https://console.example.com"}, RateLimitPerMinute: 120, RotationHint: "按演示环境手动轮换 env secret", UpdatedAt: now},
		{Channel: "WeChat", DisplayName: "微信客服", Enabled: true, SecretSource: "env", SecretRef: "ANJING_CHANNEL_WECHAT_SECRET", NextSecretRef: "ANJING_CHANNEL_WECHAT_NEXT_SECRET", SignatureWindowSeconds: 300, ReplayProtection: true, AllowedOrigins: []string{"https://wechat.example.com"}, RateLimitPerMinute: 60, RotationHint: "生产接入时对齐微信回调 message id", UpdatedAt: now},
		{Channel: "WeCom", DisplayName: "企业微信客服", Enabled: true, SecretSource: "env", SecretRef: "ANJING_CHANNEL_WECOM_SECRET", NextSecretRef: "ANJING_CHANNEL_WECOM_NEXT_SECRET", SignatureWindowSeconds: 300, ReplayProtection: true, AllowedOrigins: []string{"https://qyapi.weixin.qq.com"}, RateLimitPerMinute: 50, RotationHint: "企业微信回调按 corpId + msgId 做对账并轮换 token", UpdatedAt: now},
		{Channel: "App", DisplayName: "App 客服", Enabled: true, SecretSource: "env", SecretRef: "ANJING_CHANNEL_APP_SECRET", NextSecretRef: "ANJING_CHANNEL_APP_NEXT_SECRET", SignatureWindowSeconds: 300, ReplayProtection: true, AllowedOrigins: []string{"app://agent-customer-service"}, RateLimitPerMinute: 90, RotationHint: "App 版本发布时同步轮换 secret", UpdatedAt: now},
		{Channel: "Marketplace", DisplayName: "平台店铺客服", Enabled: true, SecretSource: "env", SecretRef: "ANJING_CHANNEL_MARKETPLACE_SECRET", NextSecretRef: "ANJING_CHANNEL_MARKETPLACE_NEXT_SECRET", SignatureWindowSeconds: 300, ReplayProtection: true, AllowedOrigins: []string{"https://marketplace.example.com"}, RateLimitPerMinute: 45, RotationHint: "按平台回调密钥周期轮换", UpdatedAt: now},
		{Channel: "Douyin", DisplayName: "抖音客服", Enabled: true, SecretSource: "env", SecretRef: "ANJING_CHANNEL_DOUYIN_SECRET", NextSecretRef: "ANJING_CHANNEL_DOUYIN_NEXT_SECRET", SignatureWindowSeconds: 300, ReplayProtection: true, AllowedOrigins: []string{"https://open.douyin.com"}, RateLimitPerMinute: 80, RotationHint: "按抖音开放平台回调 messageId 做幂等并轮换 secret", UpdatedAt: now},
		{Channel: "Xiaohongshu", DisplayName: "小红书客服", Enabled: true, SecretSource: "env", SecretRef: "ANJING_CHANNEL_XIAOHONGSHU_SECRET", NextSecretRef: "ANJING_CHANNEL_XIAOHONGSHU_NEXT_SECRET", SignatureWindowSeconds: 300, ReplayProtection: true, AllowedOrigins: []string{"https://open.xiaohongshu.com"}, RateLimitPerMinute: 70, RotationHint: "按小红书开放平台 messageId 对账，口碑场景保留审计", UpdatedAt: now},
	}
}

func defaultChannelAlertPolicies() []ChannelAlertPolicy {
	return []ChannelAlertPolicy{
		newDefaultChannelAlertPolicy("Web", "MEDIUM", 8, 60, "ops-webhook"),
		newDefaultChannelAlertPolicy("WeChat", "HIGH", 5, 60, "wechat-oncall"),
		newDefaultChannelAlertPolicy("WeCom", "HIGH", 4, 60, "wecom-oncall"),
		newDefaultChannelAlertPolicy("App", "MEDIUM", 6, 60, "app-oncall"),
		newDefaultChannelAlertPolicy("Marketplace", "HIGH", 3, 60, "marketplace-oncall"),
		newDefaultChannelAlertPolicy("Douyin", "HIGH", 4, 60, "douyin-oncall"),
		newDefaultChannelAlertPolicy("Xiaohongshu", "HIGH", 4, 60, "xiaohongshu-oncall"),
	}
}

func newDefaultChannelAlertPolicy(channel, severity string, threshold, windowMinutes int, target string) ChannelAlertPolicy {
	policy := ChannelAlertPolicy{
		Channel:                       channel,
		Severity:                      severity,
		Threshold:                     threshold,
		WindowMinutes:                 windowMinutes,
		NotifyTarget:                  target,
		TargetURL:                     notificationTargetURL(target),
		SecretRef:                     notificationSecretRef(target),
		MaxAttempts:                   3,
		BackoffSeconds:                60,
		InboundAuditMinSamples:        3,
		InboundAuditMinAcceptanceRate: 80,
		InboundAuditMaxErrorCount:     2,
		Enabled:                       true,
	}
	switch channel {
	case "Web", "App":
		policy.InboundAuditMinSamples = 5
	case "WeChat", "WeCom", "Marketplace", "Douyin", "Xiaohongshu":
		policy.InboundAuditMinSamples = 4
		policy.InboundAuditMinAcceptanceRate = 85
	}
	return policy
}

func channelIntegrationFor(integrations []ChannelIntegration, channel string) ChannelIntegration {
	normalized := strings.ToLower(strings.TrimSpace(channel))
	for _, integration := range integrations {
		if strings.ToLower(integration.Channel) == normalized {
			return integration
		}
	}
	for _, integration := range defaultChannelIntegrations(time.Now().UTC().Format(time.RFC3339)) {
		if strings.ToLower(integration.Channel) == normalized {
			return integration
		}
	}
	return defaultChannelIntegrations(time.Now().UTC().Format(time.RFC3339))[0]
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

func newReviewTask(messageID, conversationID, channel, priority, reason, createdAt string) ReviewTask {
	return ReviewTask{
		ID:             fmt.Sprintf("review_%s", strings.TrimPrefix(messageID, "msg_")),
		MessageID:      messageID,
		ConversationID: conversationID,
		Channel:        fallback(channel, "Web"),
		Status:         "OPEN",
		Priority:       fallback(priority, "NORMAL"),
		Reason:         fallback(reason, "Agent 回复抽检"),
		CreatedAt:      createdAt,
	}
}

func reviewPriority(message Message) string {
	if message.FallbackReason != "" || !message.Safe {
		return "HIGH"
	}
	if message.Engine == "llm+rag" {
		return "HIGH"
	}
	return "NORMAL"
}

func reviewReason(message Message) string {
	switch message.FallbackReason {
	case "NO_EVIDENCE":
		return "无证据兜底需要确认是否应补知识"
	case "TRANSFER_THRESHOLD":
		return "转人工边界需要确认话术与升级理由"
	}
	if message.Engine == "llm+rag" {
		return "模型生成回复需要抽检事实一致性"
	}
	return "RAG 回复抽检证据引用"
}

func (s *Store) completeReviewTaskByMessageLocked(messageID, reviewer, completedAt string) {
	for idx := range s.reviewTasks {
		if s.reviewTasks[idx].MessageID == messageID && s.reviewTasks[idx].Status != "COMPLETED" {
			s.reviewTasks[idx].Status = "COMPLETED"
			s.reviewTasks[idx].Assignee = fallback(reviewer, s.reviewTasks[idx].Assignee)
			s.reviewTasks[idx].CompletedAt = completedAt
			return
		}
	}
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

func normalizeReportLimit(limit int) int {
	if limit <= 0 {
		return 10
	}
	if limit > 30 {
		return 30
	}
	return limit
}

func normalizeReportRetention(retain int) int {
	if retain <= 0 {
		return 30
	}
	if retain > 365 {
		return 365
	}
	return retain
}

func normalizeReportEventLimit(limit int) int {
	if limit <= 0 {
		return 10
	}
	if limit > 50 {
		return 50
	}
	return limit
}

func normalizeChannelInboundAuditLimit(limit int) int {
	if limit <= 0 {
		return 20
	}
	if limit > 100 {
		return 100
	}
	return limit
}

func normalizeChannelAuditQualityEventLimit(limit int) int {
	if limit <= 0 {
		return 20
	}
	if limit > 100 {
		return 100
	}
	return limit
}

func normalizeChannelOpsReport(report *ChannelOpsReport) {
	now := time.Now().UTC()
	if strings.TrimSpace(report.ID) == "" {
		report.ID = fmt.Sprintf("channel_ops_%d", now.UnixNano())
	}
	report.Format = strings.ToLower(strings.TrimSpace(report.Format))
	if report.Format == "md" {
		report.Format = "markdown"
	}
	if report.Format == "" {
		report.Format = "markdown"
	}
	if strings.TrimSpace(report.ContentType) == "" {
		if report.Format == "csv" {
			report.ContentType = "text/csv; charset=utf-8"
		} else {
			report.ContentType = "text/markdown; charset=utf-8"
		}
	}
	if strings.TrimSpace(report.GeneratedAt) == "" {
		report.GeneratedAt = now.Format(time.RFC3339)
	}
}

func normalizeChannelOpsReportEvent(event *ChannelOpsReportEvent) {
	now := time.Now().UTC()
	if strings.TrimSpace(event.ID) == "" {
		event.ID = fmt.Sprintf("channel_ops_event_%d", now.UnixNano())
	}
	event.Action = fallback(strings.ToUpper(strings.TrimSpace(event.Action)), "COMPENSATE")
	event.Actor = fallback(strings.TrimSpace(event.Actor), "operator")
	event.Status = fallback(strings.ToUpper(strings.TrimSpace(event.Status)), "SUCCESS")
	event.Format = fallback(strings.ToLower(strings.TrimSpace(event.Format)), "markdown")
	if strings.TrimSpace(event.CreatedAt) == "" {
		event.CreatedAt = now.Format(time.RFC3339)
	}
}

func normalizeChannelRunbookCheck(check *ChannelRunbookCheck) {
	now := time.Now().UTC()
	if strings.TrimSpace(check.ID) == "" {
		check.ID = fmt.Sprintf("channel_runbook_check_%d", now.UnixNano())
	}
	check.Channel = fallback(strings.TrimSpace(check.Channel), "Unknown")
	check.RunbookStatus = fallback(strings.ToUpper(strings.TrimSpace(check.RunbookStatus)), "DISPATCH")
	check.Step = fallback(strings.TrimSpace(check.Step), "Runbook step")
	if check.StepIndex < 0 {
		check.StepIndex = 0
	}
	check.Actor = fallback(strings.TrimSpace(check.Actor), "operator")
	check.ActionRef = strings.TrimSpace(check.ActionRef)
	check.ReportID = strings.TrimSpace(check.ReportID)
	check.Note = strings.TrimSpace(check.Note)
	if strings.TrimSpace(check.CompletedAt) == "" {
		check.CompletedAt = now.Format(time.RFC3339)
	}
}

func sameChannelRunbookCheck(left ChannelRunbookCheck, right ChannelRunbookCheck) bool {
	return strings.EqualFold(left.Channel, right.Channel) &&
		strings.EqualFold(left.RunbookStatus, right.RunbookStatus) &&
		left.StepIndex == right.StepIndex &&
		strings.EqualFold(left.ActionRef, right.ActionRef)
}

func normalizeChannelInboundAudit(audit *ChannelInboundAudit) {
	now := time.Now().UTC()
	if strings.TrimSpace(audit.ID) == "" {
		audit.ID = fmt.Sprintf("channel_inbound_audit_%d", now.UnixNano())
	}
	audit.Channel = fallback(strings.TrimSpace(audit.Channel), "Unknown")
	audit.Status = fallback(strings.ToUpper(strings.TrimSpace(audit.Status)), "REJECTED")
	audit.Code = fallback(strings.TrimSpace(audit.Code), "channel_inbound")
	if strings.TrimSpace(audit.CreatedAt) == "" {
		audit.CreatedAt = now.Format(time.RFC3339)
	}
}

func normalizeChannelInboundAuditQualityEvent(event *ChannelInboundAuditQualityEvent) {
	now := time.Now().UTC()
	if strings.TrimSpace(event.ID) == "" {
		event.ID = fmt.Sprintf("channel_audit_quality_event_%d", now.UnixNano())
	}
	event.Channel = fallback(strings.TrimSpace(event.Channel), "Unknown")
	event.Severity = fallback(strings.ToUpper(strings.TrimSpace(event.Severity)), "MEDIUM")
	event.Status = fallback(strings.ToUpper(strings.TrimSpace(event.Status)), "OPEN")
	event.FailureCode = fallback(strings.TrimSpace(event.FailureCode), "inbound_acceptance_low")
	if strings.TrimSpace(event.CreatedAt) == "" {
		event.CreatedAt = now.Format(time.RFC3339)
	}
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

func channelAlerts(events []ChannelFailureEvent) []ChannelAlert {
	type key struct {
		channel string
		code    string
	}
	alertsByKey := make(map[key]ChannelAlert)
	order := make([]key, 0)
	for _, event := range events {
		k := key{channel: fallback(event.Channel, "Unknown"), code: fallback(event.Code, "channel_failure")}
		alert, exists := alertsByKey[k]
		if !exists {
			order = append(order, k)
			alert = ChannelAlert{Channel: k.channel, Code: k.code}
		}
		alert.Count++
		if alert.LastSeenAt == "" || event.CreatedAt > alert.LastSeenAt {
			alert.LastSeenAt = event.CreatedAt
			alert.LastReason = event.Reason
			alert.LastOrigin = event.Origin
		}
		alertsByKey[k] = alert
	}
	alerts := make([]ChannelAlert, 0, len(order))
	for _, k := range order {
		alerts = append(alerts, alertsByKey[k])
	}
	return alerts
}

func channelFailureTrends(events []ChannelFailureEvent, now time.Time) []ChannelFailureTrend {
	type key struct {
		channel string
		bucket  string
	}
	cutoff := now.Add(-4 * time.Hour)
	counts := make(map[key]int)
	order := make([]key, 0)
	for _, event := range events {
		createdAt, err := time.Parse(time.RFC3339, event.CreatedAt)
		if err != nil || createdAt.Before(cutoff) {
			continue
		}
		bucket := createdAt.UTC().Truncate(time.Hour).Format(time.RFC3339)
		k := key{channel: fallback(event.Channel, "Unknown"), bucket: bucket}
		if _, exists := counts[k]; !exists {
			order = append(order, k)
		}
		counts[k]++
	}
	trends := make([]ChannelFailureTrend, 0, len(order))
	for _, k := range order {
		trends = append(trends, ChannelFailureTrend{Channel: k.channel, BucketStart: k.bucket, Count: counts[k]})
	}
	return trends
}

func channelAlertPolicies(policies []ChannelAlertPolicy, alerts []ChannelAlert) []ChannelAlertPolicy {
	result := make([]ChannelAlertPolicy, 0, len(policies))
	for _, policy := range policies {
		item := policy
		normalizeChannelAlertPolicy(&item)
		for _, alert := range alerts {
			if alert.Channel == item.Channel {
				item.CurrentCount += alert.Count
				if alert.LastSeenAt > item.LastTriggeredAt {
					item.LastTriggeredAt = alert.LastSeenAt
				}
			}
		}
		item.Active = item.Enabled && item.CurrentCount >= item.Threshold
		if !item.Active {
			item.LastTriggeredAt = ""
		}
		result = append(result, item)
	}
	return result
}

func appendTriggeredNotifications(existing []ChannelNotification, policies []ChannelAlertPolicy, createdAt string) []ChannelNotification {
	result := append([]ChannelNotification(nil), existing...)
	for _, policy := range policies {
		if !policy.Active || hasOpenChannelNotification(result, policy.Channel) {
			continue
		}
		result = append([]ChannelNotification{newChannelNotification(policy, createdAt)}, result...)
	}
	if len(result) > 50 {
		return result[:50]
	}
	return result
}

func hasOpenChannelNotification(items []ChannelNotification, channel string) bool {
	for _, item := range items {
		if item.Channel == channel && (item.Status == "OPEN" || item.Status == "RETRYING") {
			return true
		}
	}
	return false
}

func newChannelNotification(policy ChannelAlertPolicy, createdAt string) ChannelNotification {
	normalizeChannelAlertPolicy(&policy)
	return ChannelNotification{
		ID:             fmt.Sprintf("channel_notice_%d", time.Now().UnixNano()),
		Channel:        policy.Channel,
		Severity:       policy.Severity,
		Target:         policy.NotifyTarget,
		TargetURL:      policy.TargetURL,
		SecretRef:      policy.SecretRef,
		Status:         "OPEN",
		Reason:         fmt.Sprintf("%s failures reached %d/%d in %dm", policy.Channel, policy.CurrentCount, policy.Threshold, policy.WindowMinutes),
		Count:          policy.CurrentCount,
		MaxAttempts:    policy.MaxAttempts,
		BackoffSeconds: policy.BackoffSeconds,
		CreatedAt:      fallback(createdAt, time.Now().UTC().Format(time.RFC3339)),
	}
}

func normalizeChannelAlertPolicy(policy *ChannelAlertPolicy) {
	if policy == nil {
		return
	}
	if strings.TrimSpace(policy.NotifyTarget) == "" {
		policy.NotifyTarget = "ops-webhook"
	}
	policy.TargetURL = fallback(policy.TargetURL, notificationTargetURL(policy.NotifyTarget))
	policy.SecretRef = fallback(policy.SecretRef, notificationSecretRef(policy.NotifyTarget))
	policy.MaxAttempts = normalizeMaxAttempts(policy.MaxAttempts)
	policy.BackoffSeconds = normalizeBackoffSeconds(policy.BackoffSeconds)
	policy.InboundAuditMinSamples = normalizeInboundAuditMinSamples(policy.InboundAuditMinSamples)
	policy.InboundAuditMinAcceptanceRate = normalizeInboundAuditMinAcceptanceRate(policy.InboundAuditMinAcceptanceRate)
	policy.InboundAuditMaxErrorCount = normalizeInboundAuditMaxErrorCount(policy.InboundAuditMaxErrorCount)
}

func applyChannelAlertPolicyInboundAuditThresholds(policy *ChannelAlertPolicy, minSamples int, minAcceptanceRate int, maxErrorCount int) {
	policy.InboundAuditMinSamples = normalizeInboundAuditMinSamples(minSamples)
	policy.InboundAuditMinAcceptanceRate = normalizeInboundAuditMinAcceptanceRate(minAcceptanceRate)
	policy.InboundAuditMaxErrorCount = normalizeInboundAuditMaxErrorCount(maxErrorCount)
}

func channelAlertPolicyDeliveryChanged(policy ChannelAlertPolicy, targetURL string, secretRef string, maxAttempts int, backoffSeconds int) bool {
	normalizeChannelAlertPolicy(&policy)
	targetURL = fallback(strings.TrimSpace(targetURL), notificationTargetURL(policy.NotifyTarget))
	secretRef = fallback(strings.TrimSpace(secretRef), notificationSecretRef(policy.NotifyTarget))
	return policy.TargetURL != targetURL ||
		policy.SecretRef != secretRef ||
		policy.MaxAttempts != normalizeMaxAttempts(maxAttempts) ||
		policy.BackoffSeconds != normalizeBackoffSeconds(backoffSeconds)
}

func channelAlertPolicyInboundAuditThresholdChanged(policy ChannelAlertPolicy, minSamples int, minAcceptanceRate int, maxErrorCount int) bool {
	normalizeChannelAlertPolicy(&policy)
	return policy.InboundAuditMinSamples != normalizeInboundAuditMinSamples(minSamples) ||
		policy.InboundAuditMinAcceptanceRate != normalizeInboundAuditMinAcceptanceRate(minAcceptanceRate) ||
		policy.InboundAuditMaxErrorCount != normalizeInboundAuditMaxErrorCount(maxErrorCount)
}

func normalizeInboundAuditThresholdInputs(policy ChannelAlertPolicy, minSamples int, minAcceptanceRate int, maxErrorCount int) (int, int, int) {
	normalizeChannelAlertPolicy(&policy)
	if minSamples <= 0 {
		minSamples = policy.InboundAuditMinSamples
	}
	if minAcceptanceRate <= 0 {
		minAcceptanceRate = policy.InboundAuditMinAcceptanceRate
	}
	if maxErrorCount <= 0 {
		maxErrorCount = policy.InboundAuditMaxErrorCount
	}
	return normalizeInboundAuditMinSamples(minSamples), normalizeInboundAuditMinAcceptanceRate(minAcceptanceRate), normalizeInboundAuditMaxErrorCount(maxErrorCount)
}

func normalizeInboundAuditMinSamples(value int) int {
	if value <= 0 {
		return 3
	}
	if value > 100 {
		return 100
	}
	return value
}

func normalizeInboundAuditMinAcceptanceRate(value int) int {
	if value <= 0 {
		return 80
	}
	if value > 100 {
		return 100
	}
	return value
}

func normalizeInboundAuditMaxErrorCount(value int) int {
	if value <= 0 {
		return 2
	}
	if value > 100 {
		return 100
	}
	return value
}

func normalizeMaxAttempts(value int) int {
	if value <= 0 {
		return 3
	}
	if value > 10 {
		return 10
	}
	return value
}

func normalizeBackoffSeconds(value int) int {
	if value <= 0 {
		return 60
	}
	if value > 3600 {
		return 3600
	}
	return value
}

func dispatchChannelNotification(notification ChannelNotification, outcome string, delivery NotificationDeliveryClient, now time.Time) ChannelNotification {
	if notification.MaxAttempts <= 0 {
		notification.MaxAttempts = 3
	}
	if notification.BackoffSeconds <= 0 {
		notification.BackoffSeconds = 60
	}
	if strings.TrimSpace(notification.TargetURL) == "" {
		notification.TargetURL = notificationTargetURL(notification.Target)
	}
	if strings.TrimSpace(notification.SecretRef) == "" {
		notification.SecretRef = notificationSecretRef(notification.Target)
	}
	if notification.Status == "ACKED" || notification.Status == "SENT" || notification.Status == "DEAD_LETTER" {
		return notification
	}
	notification.Attempts++
	notification.LastDispatchAt = now.Format(time.RFC3339)
	notification.Signature = signChannelNotification(notification)
	payload := notificationPayload(notification)
	result, err := delivery.DeliverChannelNotification(context.Background(), NotificationDeliveryRequest{
		Notification:  notification,
		Outcome:       outcome,
		SignedPayload: payload,
	})
	if err != nil {
		result = NotificationDeliveryResult{Error: err.Error(), ReceiptStatus: "500 DELIVERY_ERROR", ReceiptBody: err.Error()}
	}
	notification.DeliveryAudit = append(notification.DeliveryAudit, newNotificationDeliveryAudit(notification, payload, result, now))
	if result.Accepted {
		notification.Status = "SENT"
		notification.LastError = ""
		notification.DeadLetterReason = ""
		notification.NextRetryAt = ""
		notification.ReceiptStatus = fallback(result.ReceiptStatus, "202 ACCEPTED")
		notification.ReceiptBody = fallback(result.ReceiptBody, "notification accepted")
		return notification
	}
	notification.LastError = fallback(result.Error, fallback(outcome, "webhook_timeout"))
	notification.ReceiptStatus = fallback(result.ReceiptStatus, "504 TIMEOUT")
	notification.ReceiptBody = fallback(result.ReceiptBody, "delivery client did not return success")
	if notification.Attempts >= notification.MaxAttempts {
		notification.Status = "DEAD_LETTER"
		notification.DeadLetterReason = fmt.Sprintf("failed after %d signed webhook attempts", notification.Attempts)
		notification.NextRetryAt = ""
		return notification
	}
	notification.Status = "RETRYING"
	notification.NextRetryAt = now.Add(time.Duration(notification.BackoffSeconds) * time.Second).Format(time.RFC3339)
	notification.BackoffSeconds = notification.BackoffSeconds * 2
	return notification
}

func signChannelNotification(notification ChannelNotification) string {
	mac := hmac.New(sha256.New, []byte(notificationSecret(notification.SecretRef)))
	_, _ = mac.Write([]byte(notificationPayload(notification)))
	return hex.EncodeToString(mac.Sum(nil))
}

func notificationPayload(notification ChannelNotification) string {
	return fmt.Sprintf("%s|%s|%s|%s|%d|%d", notification.ID, notification.Channel, notification.Target, notification.TargetURL, notification.Count, notification.Attempts)
}

func newNotificationDeliveryAudit(notification ChannelNotification, payload string, result NotificationDeliveryResult, createdAt time.Time) NotificationDeliveryAudit {
	return NotificationDeliveryAudit{
		Attempt:          notification.Attempts,
		TargetURL:        notification.TargetURL,
		SecretRef:        notification.SecretRef,
		SignaturePreview: signaturePreview(notification.Signature),
		PayloadHash:      sha256Hex(payload),
		RequestSummary:   fmt.Sprintf("POST %s channel=%s count=%d", notification.TargetURL, notification.Channel, notification.Count),
		ResponseSummary:  redactedResponseSummary(result),
		CreatedAt:        createdAt.Format(time.RFC3339),
	}
}

func signaturePreview(signature string) string {
	signature = strings.TrimSpace(signature)
	if len(signature) <= 12 {
		return signature
	}
	return signature[:12] + "..."
}

func sha256Hex(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func redactedResponseSummary(result NotificationDeliveryResult) string {
	status := fallback(result.ReceiptStatus, "NO_STATUS")
	body := strings.TrimSpace(result.ReceiptBody)
	if len(body) > 80 {
		body = body[:80] + "..."
	}
	if body == "" {
		body = strings.TrimSpace(result.Error)
	}
	if body == "" {
		return status
	}
	return status + " · " + body
}

func notificationSecret(secretRef string) string {
	if strings.TrimSpace(secretRef) == "" {
		return "notification-demo-secret"
	}
	if value := os.Getenv(secretRef); strings.TrimSpace(value) != "" {
		return value
	}
	return "notification-demo-secret"
}

func notificationSecretRef(target string) string {
	normalized := strings.ToUpper(strings.NewReplacer("-", "_", ".", "_", "/", "_").Replace(strings.TrimSpace(target)))
	if normalized == "" {
		normalized = "OPS_WEBHOOK"
	}
	return "ANJING_NOTIFICATION_" + normalized + "_SECRET"
}

type demoNotificationDeliveryClient struct{}

func (demoNotificationDeliveryClient) DeliverChannelNotification(_ context.Context, req NotificationDeliveryRequest) (NotificationDeliveryResult, error) {
	if strings.EqualFold(req.Outcome, "SUCCESS") {
		return NotificationDeliveryResult{Accepted: true, ReceiptStatus: "202 ACCEPTED", ReceiptBody: "notification accepted by demo webhook"}, nil
	}
	return NotificationDeliveryResult{Accepted: false, ReceiptStatus: "504 TIMEOUT", ReceiptBody: "demo webhook did not return success", Error: fallback(req.Outcome, "webhook_timeout")}, nil
}

func notificationTargetURL(target string) string {
	normalized := strings.ToLower(strings.TrimSpace(target))
	if normalized == "" {
		normalized = "ops-webhook"
	}
	if strings.HasPrefix(normalized, "http://") || strings.HasPrefix(normalized, "https://") {
		return strings.TrimSpace(target)
	}
	return fmt.Sprintf("https://hooks.example.com/anjing/%s", normalized)
}

func activeAlertPolicies(policies []ChannelAlertPolicy) int {
	total := 0
	for _, policy := range policies {
		if policy.Active {
			total++
		}
	}
	return total
}

func channelAlertCount(alerts []ChannelAlert) int {
	total := 0
	for _, alert := range alerts {
		total += alert.Count
	}
	return total
}

func channelRunbooks(alerts []ChannelAlert, policies []ChannelAlertPolicy, notifications []ChannelNotification, audits []ChannelInboundAudit, checks []ChannelRunbookCheck) []ChannelRunbook {
	runbooks := make([]ChannelRunbook, 0)
	for _, policy := range policies {
		if !policy.Active {
			continue
		}
		alert := latestChannelAlert(alerts, policy.Channel)
		notification := latestChannelNotification(notifications, policy.Channel)
		status := "MONITOR"
		nextAction := "继续观察渠道失败趋势"
		escalation := "失败数持续增长时升级给渠道 owner"
		steps := []string{
			"确认失败码、来源和最近触发时间",
			"核对渠道协议矩阵中的签名头、时间窗和 replay key",
			"检查通知目标、secret ref 和最近一次投递审计",
		}
		if notification.ID != "" {
			switch notification.Status {
			case "OPEN":
				status = "DISPATCH"
				nextAction = "发送出站通知并记录投递审计"
				steps = append(steps, "点击发送通知，确认目标 webhook 可达")
			case "RETRYING":
				status = "RETRY"
				nextAction = "等待退避重试或手动再次投递"
				steps = append(steps, "查看 nextRetryAt、lastError 和 response summary")
			case "DEAD_LETTER":
				status = "ESCALATE"
				nextAction = "升级到渠道 owner 并确认死信"
				escalation = "通知连续失败已进入死信，需人工处理 webhook 或 secret"
				steps = append(steps, "带上 deadLetterReason、payload hash 和签名预览升级处理")
			case "SENT":
				status = "ACK"
				nextAction = "等待运营确认告警已处理"
				steps = append(steps, "确认外部回执后点击确认告警")
			case "ACKED":
				status = "DONE"
				nextAction = "处置已确认，保留审计记录"
			}
		}
		runbooks = append(runbooks, ChannelRunbook{
			Channel:           policy.Channel,
			Severity:          policy.Severity,
			Status:            status,
			FailureCode:       alert.Code,
			Owner:             policy.NotifyTarget,
			NextAction:        nextAction,
			Escalation:        escalation,
			NotificationID:    notification.ID,
			NotificationState: notification.Status,
			Steps:             steps,
		})
	}
	runbooks = append(runbooks, channelInboundAuditRunbooks(audits, policies)...)
	return attachChannelRunbookChecks(runbooks, checks)
}

func attachChannelRunbookChecks(runbooks []ChannelRunbook, checks []ChannelRunbookCheck) []ChannelRunbook {
	for idx := range runbooks {
		matched := make([]ChannelRunbookCheck, 0)
		for _, check := range checks {
			if !strings.EqualFold(check.Channel, runbooks[idx].Channel) || !strings.EqualFold(check.RunbookStatus, runbooks[idx].Status) {
				continue
			}
			if check.StepIndex >= 0 && check.StepIndex < len(runbooks[idx].Steps) {
				matched = append(matched, check)
				continue
			}
			for stepIdx, step := range runbooks[idx].Steps {
				if strings.EqualFold(check.Step, step) {
					check.StepIndex = stepIdx
					matched = append(matched, check)
					break
				}
			}
		}
		runbooks[idx].Checks = matched
	}
	return runbooks
}

func channelInboundAuditRunbooks(audits []ChannelInboundAudit, policies []ChannelAlertPolicy) []ChannelRunbook {
	type auditQuality struct {
		channel       string
		total         int
		accepted      int
		rejected      int
		codeCounts    map[string]int
		codeOrder     []string
		lastCreatedAt string
	}
	qualityByChannel := make(map[string]*auditQuality)
	channelOrder := make([]string, 0)
	for _, audit := range audits {
		channel := fallback(audit.Channel, "Unknown")
		quality, exists := qualityByChannel[channel]
		if !exists {
			quality = &auditQuality{channel: channel, codeCounts: make(map[string]int)}
			qualityByChannel[channel] = quality
			channelOrder = append(channelOrder, channel)
		}
		quality.total++
		if strings.EqualFold(audit.Status, "ACCEPTED") {
			quality.accepted++
		} else {
			quality.rejected++
			code := fallback(audit.Code, "inbound_rejected")
			if _, exists := quality.codeCounts[code]; !exists {
				quality.codeOrder = append(quality.codeOrder, code)
			}
			quality.codeCounts[code]++
		}
		if audit.CreatedAt > quality.lastCreatedAt {
			quality.lastCreatedAt = audit.CreatedAt
		}
	}

	runbooks := make([]ChannelRunbook, 0)
	for _, channel := range channelOrder {
		quality := qualityByChannel[channel]
		policy := notificationPolicyFor(policies, channel)
		if quality.total < policy.InboundAuditMinSamples || quality.rejected == 0 {
			continue
		}
		acceptanceRate := quality.accepted * 100 / quality.total
		topCode := topInboundAuditCode(quality.codeCounts, quality.codeOrder)
		topCodeCount := quality.codeCounts[topCode]
		if acceptanceRate >= policy.InboundAuditMinAcceptanceRate && topCodeCount < policy.InboundAuditMaxErrorCount {
			continue
		}
		severity := policy.Severity
		status := "WATCH"
		if acceptanceRate < policy.InboundAuditMinAcceptanceRate-10 {
			severity = "HIGH"
		}
		if acceptanceRate < policy.InboundAuditMinAcceptanceRate-30 || topCodeCount >= policy.InboundAuditMaxErrorCount*2 {
			status = "ESCALATE"
		}
		failureCode := topCode
		if failureCode == "" {
			failureCode = "inbound_acceptance_low"
		}
		runbooks = append(runbooks, ChannelRunbook{
			Channel:     channel,
			Severity:    severity,
			Status:      status,
			FailureCode: failureCode,
			Owner:       policy.NotifyTarget,
			NextAction:  fmt.Sprintf("最近 %d 次入站验收率为 %d%%，当前阈值为 %d%%/%d 样本/%d 次同错码，复核验签、来源白名单、时间窗和 replay key。", quality.total, acceptanceRate, policy.InboundAuditMinAcceptanceRate, policy.InboundAuditMinSamples, policy.InboundAuditMaxErrorCount),
			Escalation:  fmt.Sprintf("若验收率持续低于 %d%% 或同一错误码达到 %d 次，升级 platform-owner 与渠道 owner。", policy.InboundAuditMinAcceptanceRate, policy.InboundAuditMaxErrorCount),
			Steps: []string{
				"查看渠道验收审计筛选同一错误码的最近样本",
				"核对 ChannelIntegration 的 allowed origins、active/next secret ref 和签名窗口",
				"按 external message id、replay key 和 content hash 对账重复请求",
				"确认渠道平台签名头和 timestamp 语义是否与协议矩阵一致",
			},
		})
	}
	return runbooks
}

func channelInboundAuditQualityEvent(audits []ChannelInboundAudit, policies []ChannelAlertPolicy) (ChannelInboundAuditQualityEvent, bool) {
	runbooks := channelInboundAuditRunbooks(audits, policies)
	if len(runbooks) == 0 {
		return ChannelInboundAuditQualityEvent{}, false
	}
	runbook := runbooks[0]
	channelAudits := make([]ChannelInboundAudit, 0)
	for _, audit := range audits {
		if strings.EqualFold(audit.Channel, runbook.Channel) {
			channelAudits = append(channelAudits, audit)
		}
	}
	policy := notificationPolicyFor(policies, runbook.Channel)
	total := len(channelAudits)
	accepted := 0
	rejected := 0
	for _, audit := range channelAudits {
		if strings.EqualFold(audit.Status, "ACCEPTED") {
			accepted++
		} else {
			rejected++
		}
	}
	acceptanceRate := 100
	if total > 0 {
		acceptanceRate = accepted * 100 / total
	}
	event := ChannelInboundAuditQualityEvent{
		Channel:           runbook.Channel,
		Severity:          runbook.Severity,
		Status:            runbook.Status,
		FailureCode:       runbook.FailureCode,
		Total:             total,
		Accepted:          accepted,
		Rejected:          rejected,
		AcceptanceRate:    acceptanceRate,
		MinSamples:        policy.InboundAuditMinSamples,
		MinAcceptanceRate: policy.InboundAuditMinAcceptanceRate,
		MaxErrorCount:     policy.InboundAuditMaxErrorCount,
		Reason:            runbook.NextAction,
	}
	normalizeChannelInboundAuditQualityEvent(&event)
	return event, true
}

func sameChannelInboundAuditQualityEvent(events []ChannelInboundAuditQualityEvent, next ChannelInboundAuditQualityEvent) bool {
	if len(events) == 0 {
		return false
	}
	latest := events[0]
	return strings.EqualFold(latest.Channel, next.Channel) &&
		latest.FailureCode == next.FailureCode &&
		latest.Total == next.Total &&
		latest.AcceptanceRate == next.AcceptanceRate
}

func topInboundAuditCode(counts map[string]int, order []string) string {
	topCode := ""
	topCount := 0
	for _, code := range order {
		count := counts[code]
		if count > topCount {
			topCode = code
			topCount = count
		}
	}
	return topCode
}

func latestChannelAlert(alerts []ChannelAlert, channel string) ChannelAlert {
	for _, alert := range alerts {
		if strings.EqualFold(alert.Channel, channel) {
			return alert
		}
	}
	return ChannelAlert{Channel: channel, Code: "channel_failure"}
}

func latestChannelNotification(notifications []ChannelNotification, channel string) ChannelNotification {
	for _, notification := range notifications {
		if strings.EqualFold(notification.Channel, channel) {
			return notification
		}
	}
	return ChannelNotification{}
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
