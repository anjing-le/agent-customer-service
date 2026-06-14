package store

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type fakeReplyGenerator struct {
	reply string
	err   error
}

func (f fakeReplyGenerator) GenerateReply(context.Context, ReplyRequest) (ReplyGeneration, error) {
	if f.err != nil {
		return ReplyGeneration{}, f.err
	}
	return ReplyGeneration{Content: f.reply, Model: "fake-model"}, nil
}

type recordingDeliveryClient struct {
	requests []NotificationDeliveryRequest
}

func (r *recordingDeliveryClient) DeliverChannelNotification(_ context.Context, req NotificationDeliveryRequest) (NotificationDeliveryResult, error) {
	r.requests = append(r.requests, req)
	return NotificationDeliveryResult{Accepted: true, ReceiptStatus: "204 NO_CONTENT", ReceiptBody: "custom delivery accepted"}, nil
}

func TestSendMessageUsesKnowledgeEvidence(t *testing.T) {
	st := NewSeedStore()

	result, err := st.SendMessage("conv_demo_refund", "这个商品能不能开发票？")
	if err != nil {
		t.Fatalf("send message: %v", err)
	}

	if result.AgentMessage.FallbackReason != "" {
		t.Fatalf("expected evidence answer, got fallback %q", result.AgentMessage.FallbackReason)
	}
	if len(result.Evidence) != 1 || result.Evidence[0].ID != "kb_invoice" {
		t.Fatalf("expected invoice evidence, got %#v", result.Evidence)
	}
	if result.Gap != nil {
		t.Fatalf("expected no gap, got %#v", result.Gap)
	}

	history, err := st.ListMessages("conv_demo_refund")
	if err != nil {
		t.Fatalf("list messages: %v", err)
	}
	if len(history) < 4 {
		t.Fatalf("expected seeded and new conversation history, got %#v", history)
	}
	if history[len(history)-2].Role != "user" || history[len(history)-1].Role != "assistant" {
		t.Fatalf("expected latest user/assistant pair, got %#v", history[len(history)-2:])
	}
}

func TestSendMessageUsesLLMWhenEvidenceExists(t *testing.T) {
	st := NewSeedStore(WithReplyGenerator(fakeReplyGenerator{reply: "模型基于发票知识生成的客服回复。"}))

	result, err := st.SendMessage("conv_demo_refund", "这个商品能不能开发票？")
	if err != nil {
		t.Fatalf("send message: %v", err)
	}

	if result.AgentMessage.Engine != "llm+rag" {
		t.Fatalf("expected llm+rag engine, got %q", result.AgentMessage.Engine)
	}
	if result.AgentMessage.Content != "模型基于发票知识生成的客服回复。" {
		t.Fatalf("expected model reply, got %q", result.AgentMessage.Content)
	}
	if len(result.AgentMessage.EvidenceIDs) == 0 {
		t.Fatalf("expected evidence ids, got %#v", result.AgentMessage)
	}
	if result.AgentMessage.Trace == nil || !result.AgentMessage.Trace.ModelAttempted || result.AgentMessage.Trace.Model != "fake-model" {
		t.Fatalf("expected model trace, got %#v", result.AgentMessage.Trace)
	}
}

func TestSendMessageFallsBackWhenLLMUnavailable(t *testing.T) {
	st := NewSeedStore(WithReplyGenerator(fakeReplyGenerator{err: errors.New("model timeout")}))

	result, err := st.SendMessage("conv_demo_refund", "这个商品能不能开发票？")
	if err != nil {
		t.Fatalf("send message: %v", err)
	}

	if result.AgentMessage.Engine != "rag+rule" {
		t.Fatalf("expected rag+rule fallback, got %q", result.AgentMessage.Engine)
	}
	if result.AgentMessage.FallbackReason != "" {
		t.Fatalf("expected no safety fallback reason for evidence answer, got %q", result.AgentMessage.FallbackReason)
	}
	if result.AgentMessage.Trace == nil || !result.AgentMessage.Trace.ModelFallback || result.AgentMessage.Trace.ModelFallbackReason == "" {
		t.Fatalf("expected model fallback trace, got %#v", result.AgentMessage.Trace)
	}
}

func TestSendMessageCreatesGapWhenEvidenceMissing(t *testing.T) {
	st := NewSeedStore()

	result, err := st.SendMessage("conv_demo_refund", "完全没有资料的新品保价规则是什么？")
	if err != nil {
		t.Fatalf("send message: %v", err)
	}

	if result.AgentMessage.FallbackReason != "NO_EVIDENCE" {
		t.Fatalf("expected no evidence fallback, got %q", result.AgentMessage.FallbackReason)
	}
	if result.Gap == nil || result.Gap.Status != "OPEN" {
		t.Fatalf("expected open knowledge gap, got %#v", result.Gap)
	}
	if result.Conversation.Status != "KnowledgeGap" {
		t.Fatalf("expected conversation status KnowledgeGap, got %q", result.Conversation.Status)
	}
	if result.AgentMessage.Trace == nil || result.AgentMessage.Trace.Strategy != "no_evidence_fallback" {
		t.Fatalf("expected no evidence trace, got %#v", result.AgentMessage.Trace)
	}
}

func TestCreateArticleFromGapResolvesGapAndEnablesRecall(t *testing.T) {
	st := NewSeedStore()
	result, err := st.SendMessage("conv_demo_refund", "新品保价规则是什么？")
	if err != nil {
		t.Fatalf("send message: %v", err)
	}
	if result.Gap == nil {
		t.Fatal("expected knowledge gap")
	}

	article, err := st.CreateArticleFromGap(result.Gap.ID, "新品保价规则", "售后", "新品暂不支持价保，活动页另有说明时以活动规则为准。", []string{"保价", "新品"})
	if err != nil {
		t.Fatalf("create article from gap: %v", err)
	}
	if article.TrustLevel != "HIGH" {
		t.Fatalf("expected high trust article, got %#v", article)
	}

	recalled, err := st.SearchKnowledge("新品保价")
	if err != nil {
		t.Fatalf("search knowledge: %v", err)
	}
	if len(recalled) == 0 || recalled[0].ID != article.ID {
		t.Fatalf("expected new article recall, got %#v", recalled)
	}

	dashboard, err := st.Dashboard()
	if err != nil {
		t.Fatalf("dashboard: %v", err)
	}
	for _, gap := range dashboard.KnowledgeGaps {
		if gap.ID == result.Gap.ID && gap.Status != "RESOLVED" {
			t.Fatalf("expected resolved gap, got %#v", gap)
		}
	}
}

func TestSendMessageRecommendsHumanTransfer(t *testing.T) {
	st := NewSeedStore()

	result, err := st.SendMessage("conv_demo_transfer", "我已经投诉很多次了，现在必须转人工")
	if err != nil {
		t.Fatalf("send message: %v", err)
	}

	if result.AgentMessage.FallbackReason != "TRANSFER_THRESHOLD" {
		t.Fatalf("expected transfer fallback, got %q", result.AgentMessage.FallbackReason)
	}
	if result.Conversation.Status != "NeedsHuman" {
		t.Fatalf("expected NeedsHuman status, got %q", result.Conversation.Status)
	}
	if result.Gap != nil {
		t.Fatalf("expected no knowledge gap for transfer, got %#v", result.Gap)
	}

	dashboard, err := st.Dashboard()
	if err != nil {
		t.Fatalf("dashboard: %v", err)
	}
	if len(dashboard.Transfers) == 0 || dashboard.Transfers[0].Status != "OPEN" {
		t.Fatalf("expected open transfer ticket, got %#v", dashboard.Transfers)
	}
	if dashboard.Transfers[0].Escalated {
		t.Fatalf("new transfer ticket should not be escalated immediately, got %#v", dashboard.Transfers[0])
	}
	if len(dashboard.Transfers[0].Events) != 1 || dashboard.Transfers[0].Events[0].Type != "CREATED" {
		t.Fatalf("expected created event, got %#v", dashboard.Transfers[0].Events)
	}

	resolved, err := st.ResolveTransferTicket(dashboard.Transfers[0].ID, "agent-a", "已电话回访")
	if err != nil {
		t.Fatalf("resolve transfer ticket: %v", err)
	}
	if resolved.Status != "RESOLVED" || resolved.Assignee != "agent-a" {
		t.Fatalf("expected resolved ticket, got %#v", resolved)
	}
	if len(resolved.Events) != 2 || resolved.Events[1].Type != "RESOLVED" || resolved.Events[1].Note != "已电话回访" {
		t.Fatalf("expected resolved event timeline, got %#v", resolved.Events)
	}
}

func TestTransferTicketSLAEscalatesOpenTicket(t *testing.T) {
	createdAt := time.Now().UTC().Add(-45 * time.Minute).Format(time.RFC3339)
	ticket := withTransferSLA(demoTransferTicket("ticket_sla", "conv_sla", "Web", "必须人工处理", createdAt), defaultChannelPolicies(), time.Now().UTC())

	if !ticket.Escalated {
		t.Fatalf("expected escalated ticket, got %#v", ticket)
	}
	if ticket.Priority != "CRITICAL" || ticket.SLAStatus != "BREACHED" {
		t.Fatalf("expected critical breached ticket, got %#v", ticket)
	}
	if ticket.WaitMinutes < ticket.SLAMinutes {
		t.Fatalf("expected wait minutes past SLA, got %#v", ticket)
	}
}

func TestChannelPolicyControlsTransferSLA(t *testing.T) {
	st := NewSeedStore()

	conv, err := st.CreateConversation("平台客户", "Marketplace")
	if err != nil {
		t.Fatalf("create conversation: %v", err)
	}
	result, err := st.SendMessage(conv.ID, "我已经投诉很多次了，必须马上转人工")
	if err != nil {
		t.Fatalf("send message: %v", err)
	}
	if result.Conversation.Channel != "Marketplace" {
		t.Fatalf("expected channel to remain Marketplace, got %#v", result.Conversation)
	}

	dashboard, err := st.Dashboard()
	if err != nil {
		t.Fatalf("dashboard: %v", err)
	}
	if len(dashboard.ChannelPolicies) < 4 {
		t.Fatalf("expected seeded channel policies, got %#v", dashboard.ChannelPolicies)
	}
	if len(dashboard.Integrations) < 4 {
		t.Fatalf("expected seeded channel integrations, got %#v", dashboard.Integrations)
	}
	if len(dashboard.Transfers) == 0 || dashboard.Transfers[0].Channel != "Marketplace" {
		t.Fatalf("expected marketplace transfer ticket, got %#v", dashboard.Transfers)
	}
	if dashboard.Transfers[0].SLAMinutes != 10 {
		t.Fatalf("expected marketplace SLA, got %#v", dashboard.Transfers[0])
	}
}

func TestSubmitAnnotationUpdatesQualitySummary(t *testing.T) {
	st := NewSeedStore()

	result, err := st.SendMessage("conv_demo_refund", "这个商品能不能开发票？")
	if err != nil {
		t.Fatalf("send message: %v", err)
	}

	annotation, err := st.SubmitAnnotation(result.AgentMessage.ID, "qa-a", "PASS", "证据充分，回复安全", AnnotationDimensions{
		Groundedness: 5,
		Safety:       5,
		Helpfulness:  4,
	}, []string{"human_review", "teaching_sample"})
	if err != nil {
		t.Fatalf("submit annotation: %v", err)
	}
	if annotation.Score != 93 {
		t.Fatalf("expected weighted annotation score, got %#v", annotation)
	}
	if annotation.Reviewer != "qa-a" || annotation.Dimensions.Helpfulness != 4 {
		t.Fatalf("expected reviewer and dimensions, got %#v", annotation)
	}

	dashboard, err := st.Dashboard()
	if err != nil {
		t.Fatalf("dashboard: %v", err)
	}
	if len(dashboard.Annotations) != 1 {
		t.Fatalf("expected one annotation, got %#v", dashboard.Annotations)
	}
	if dashboard.Quality.AnnotationCount != 1 || dashboard.Quality.AverageReview != 93 {
		t.Fatalf("expected annotation quality summary, got %#v", dashboard.Quality)
	}
}

func TestReviewTaskAssignmentAndAnnotationCompletion(t *testing.T) {
	st := NewSeedStore()

	result, err := st.SendMessage("conv_demo_refund", "这个商品能不能开发票？")
	if err != nil {
		t.Fatalf("send message: %v", err)
	}
	dashboard, err := st.Dashboard()
	if err != nil {
		t.Fatalf("dashboard: %v", err)
	}
	var task ReviewTask
	for _, candidate := range dashboard.ReviewTasks {
		if candidate.MessageID == result.AgentMessage.ID {
			task = candidate
			break
		}
	}
	if task.ID == "" || task.Status != "OPEN" {
		t.Fatalf("expected open review task for agent reply, got %#v", dashboard.ReviewTasks)
	}

	assigned, err := st.AssignReviewTask(task.ID, "qa-a")
	if err != nil {
		t.Fatalf("assign review task: %v", err)
	}
	if assigned.Status != "ASSIGNED" || assigned.Assignee != "qa-a" {
		t.Fatalf("expected assigned task, got %#v", assigned)
	}

	if _, err := st.SubmitAnnotation(result.AgentMessage.ID, "qa-a", "PASS", "证据充分", AnnotationDimensions{
		Groundedness: 5,
		Safety:       5,
		Helpfulness:  5,
	}, nil); err != nil {
		t.Fatalf("submit annotation: %v", err)
	}
	dashboard, err = st.Dashboard()
	if err != nil {
		t.Fatalf("dashboard: %v", err)
	}
	for _, candidate := range dashboard.ReviewTasks {
		if candidate.ID == task.ID && (candidate.Status != "COMPLETED" || candidate.CompletedAt == "") {
			t.Fatalf("expected completed review task after annotation, got %#v", candidate)
		}
	}
}

func TestExportTrainingSamplesReturnsLowScoreAnnotations(t *testing.T) {
	st := NewSeedStore()

	result, err := st.SendMessage("conv_demo_refund", "这个商品能不能开发票？")
	if err != nil {
		t.Fatalf("send message: %v", err)
	}
	if _, err := st.SubmitAnnotation(result.AgentMessage.ID, "qa-a", "FAIL", "证据引用不完整，需要复盘", AnnotationDimensions{
		Groundedness: 2,
		Safety:       3,
		Helpfulness:  2,
	}, []string{"low_score", "export_candidate"}); err != nil {
		t.Fatalf("submit annotation: %v", err)
	}

	samples, err := st.ExportTrainingSamples(80)
	if err != nil {
		t.Fatalf("export training samples: %v", err)
	}
	if len(samples) != 1 {
		t.Fatalf("expected one low score sample, got %#v", samples)
	}
	sample := samples[0]
	if sample.MessageID != result.AgentMessage.ID || sample.Prompt == "" || sample.Answer == "" {
		t.Fatalf("expected prompt and answer sample, got %#v", sample)
	}
	if sample.Score > 80 || sample.Verdict != "FAIL" {
		t.Fatalf("expected low score failed sample, got %#v", sample)
	}
}

func TestRuleTestDetectsTransferAndNoEvidenceBoundaries(t *testing.T) {
	st := NewSeedStore()

	transfer, err := st.TestRule("我已经投诉很多次了，现在必须转人工")
	if err != nil {
		t.Fatalf("test rule: %v", err)
	}
	if transfer.RuleCode != "TRANSFER_THRESHOLD" || !transfer.Fallback || transfer.RiskLevel != "HIGH" {
		t.Fatalf("expected transfer rule, got %#v", transfer)
	}

	noEvidence, err := st.TestRule("新品保价规则是什么？")
	if err != nil {
		t.Fatalf("test rule: %v", err)
	}
	if noEvidence.RuleCode != "NO_EVIDENCE_FALLBACK" || !noEvidence.Fallback || noEvidence.RiskLevel != "MEDIUM" {
		t.Fatalf("expected no evidence fallback, got %#v", noEvidence)
	}

	evidenceOK, err := st.TestRule("这个商品能不能开发票？")
	if err != nil {
		t.Fatalf("test rule: %v", err)
	}
	if evidenceOK.Matched || evidenceOK.Fallback || evidenceOK.Action != "allow_answer" {
		t.Fatalf("expected allowed answer, got %#v", evidenceOK)
	}
}

func TestRuleComparisonShowsCanaryDecisionChange(t *testing.T) {
	st := NewSeedStore()

	comparison, err := st.CompareRuleVersions("我想取消订单，7 天无理由退货怎么处理")
	if err != nil {
		t.Fatalf("compare rule versions: %v", err)
	}
	if !comparison.Changed {
		t.Fatalf("expected canary to change decision, got %#v", comparison)
	}
	if comparison.Current.Action != "allow_answer" || comparison.Canary.RuleCode != "CANCEL_RISK_TRANSFER" {
		t.Fatalf("expected current allow and canary transfer, got %#v", comparison)
	}

	if _, err := st.TestRule("我已经投诉很多次了，现在必须转人工"); err != nil {
		t.Fatalf("test rule: %v", err)
	}
	dashboard, err := st.Dashboard()
	if err != nil {
		t.Fatalf("dashboard: %v", err)
	}
	for _, rule := range dashboard.Rules {
		if rule.Code == "TRANSFER_THRESHOLD" && rule.HitCount == 0 {
			t.Fatalf("expected active rule hit count, got %#v", rule)
		}
	}
}

func TestPublishAndRollbackRuleRecordsEvents(t *testing.T) {
	st := NewSeedStore()

	if _, err := st.PublishCanaryRule("CANCEL_RISK_TRANSFER", "qa-lead", "未审批直接发布"); err == nil {
		t.Fatal("expected publish to require approval gate")
	}
	approval, err := st.SubmitRuleApproval("CANCEL_RISK_TRANSFER", "qa-lead", "LOW", "3 条灰度样本通过", []string{"sample_cancel_1", "sample_cancel_2", "sample_cancel_3"})
	if err != nil {
		t.Fatalf("submit rule approval: %v", err)
	}
	if approval.Status != "APPROVED" || approval.SampleCount != 3 || len(approval.SampleIDs) != 3 {
		t.Fatalf("expected approved gate, got %#v", approval)
	}

	published, err := st.PublishCanaryRule("CANCEL_RISK_TRANSFER", "qa-lead", "灰度样本通过，发布规则")
	if err != nil {
		t.Fatalf("publish canary rule: %v", err)
	}
	if published.Action != "PUBLISH" || published.Actor != "qa-lead" {
		t.Fatalf("expected publish event, got %#v", published)
	}

	result, err := st.TestRule("我想取消订单，7 天无理由退货怎么处理")
	if err != nil {
		t.Fatalf("test published rule: %v", err)
	}
	if result.RuleCode != "CANCEL_RISK_TRANSFER" {
		t.Fatalf("expected published rule to become active, got %#v", result)
	}
	for idx := 0; idx < 3; idx++ {
		reply, err := st.SendMessage("conv_demo_refund", fmt.Sprintf("我想取消订单，退款争议需要确认 %d", idx))
		if err != nil {
			t.Fatalf("send post-release message: %v", err)
		}
		if _, err := st.SubmitAnnotation(reply.AgentMessage.ID, "qa-lead", "FAIL", "发布后低分样本", AnnotationDimensions{
			Groundedness: 2,
			Safety:       2,
			Helpfulness:  2,
		}, []string{"rule_release_observation"}); err != nil {
			t.Fatalf("submit post-release annotation: %v", err)
		}
	}
	dashboard, err := st.Dashboard()
	if err != nil {
		t.Fatalf("dashboard: %v", err)
	}
	if len(dashboard.RuleObservations) != 1 {
		t.Fatalf("expected one rule observation, got %#v", dashboard.RuleObservations)
	}
	observation := dashboard.RuleObservations[0]
	if observation.RuleCode != "CANCEL_RISK_TRANSFER" || observation.RiskLevel != "HIGH" || !observation.RollbackRecommended {
		t.Fatalf("expected high-risk rollback observation, got %#v", observation)
	}
	if observation.RuleHits < 4 || observation.LowScoreSamples < 3 || observation.TransferTickets < 3 {
		t.Fatalf("expected post-release samples and transfers, got %#v", observation)
	}

	rolledBack, err := st.RollbackRule("CANCEL_RISK_TRANSFER", "qa-lead", "发布后人工压力过高，回滚")
	if err != nil {
		t.Fatalf("rollback rule: %v", err)
	}
	if rolledBack.Action != "ROLLBACK" {
		t.Fatalf("expected rollback event, got %#v", rolledBack)
	}
	dashboard, err = st.Dashboard()
	if err != nil {
		t.Fatalf("dashboard: %v", err)
	}
	if len(dashboard.RuleEvents) != 2 {
		t.Fatalf("expected two rule events, got %#v", dashboard.RuleEvents)
	}
	if len(dashboard.RuleApprovals) != 1 {
		t.Fatalf("expected one approval record, got %#v", dashboard.RuleApprovals)
	}
}

func TestChannelFailureTrendsAndAlertPolicies(t *testing.T) {
	st := NewSeedStore()
	for idx := 0; idx < 3; idx++ {
		if err := st.RecordChannelFailure(ChannelFailureEvent{
			Channel:                "Marketplace",
			Code:                   "channel_signature_invalid",
			Reason:                 "签名错误",
			ExternalConversationID: "buyer-demo",
			ExternalMessageID:      fmt.Sprintf("event-%d", idx),
			Origin:                 "https://marketplace.example.com",
		}); err != nil {
			t.Fatalf("record channel failure: %v", err)
		}
	}

	dashboard, err := st.Dashboard()
	if err != nil {
		t.Fatalf("dashboard: %v", err)
	}
	if len(dashboard.ChannelTrends) == 0 {
		t.Fatalf("expected channel failure trends, got %#v", dashboard.ChannelTrends)
	}
	var marketplacePolicy ChannelAlertPolicy
	for _, policy := range dashboard.AlertPolicies {
		if policy.Channel == "Marketplace" {
			marketplacePolicy = policy
			break
		}
	}
	if !marketplacePolicy.Active || marketplacePolicy.CurrentCount != 3 || marketplacePolicy.NotifyTarget == "" {
		t.Fatalf("expected active marketplace alert policy, got %#v", marketplacePolicy)
	}
	if len(dashboard.Notifications) != 1 || dashboard.Notifications[0].Status != "OPEN" {
		t.Fatalf("expected one open notification, got %#v", dashboard.Notifications)
	}
	acked, err := st.AcknowledgeChannelNotification(dashboard.Notifications[0].ID, "ops-a", "已通知渠道负责人")
	if err != nil {
		t.Fatalf("ack notification: %v", err)
	}
	if acked.Status != "ACKED" || acked.AckedBy != "ops-a" {
		t.Fatalf("expected acked notification, got %#v", acked)
	}
}

func TestUpdateChannelAlertPolicyControlsNewNotifications(t *testing.T) {
	st := NewSeedStore()
	policy, err := st.UpdateChannelAlertPolicy("Marketplace", "https://ops.example.com/hooks/marketplace", "ANJING_NOTIFICATION_CUSTOM_SECRET", 5, 45, "ops-a", "切换到生产通知目标")
	if err != nil {
		t.Fatalf("update channel alert policy: %v", err)
	}
	if policy.TargetURL == "https://ops.example.com/hooks/marketplace" || policy.SecretRef == "ANJING_NOTIFICATION_CUSTOM_SECRET" {
		t.Fatalf("expected high risk policy change to wait for approval, got %#v", policy)
	}
	dashboard, err := st.Dashboard()
	if err != nil {
		t.Fatalf("dashboard: %v", err)
	}
	if len(dashboard.PolicyChanges) != 1 || dashboard.PolicyChanges[0].Status != "PENDING" {
		t.Fatalf("expected pending policy change, got %#v", dashboard.PolicyChanges)
	}
	approved, err := st.ApproveNotificationPolicyChange(dashboard.PolicyChanges[0].ID, "ops-lead", "审批通过")
	if err != nil {
		t.Fatalf("approve policy change: %v", err)
	}
	if approved.TargetURL != "https://ops.example.com/hooks/marketplace" || approved.SecretRef != "ANJING_NOTIFICATION_CUSTOM_SECRET" {
		t.Fatalf("expected approved delivery config, got %#v", approved)
	}
	for idx := 0; idx < 3; idx++ {
		if err := st.RecordChannelFailure(ChannelFailureEvent{
			Channel:           "Marketplace",
			Code:              "channel_signature_invalid",
			Reason:            "签名错误",
			ExternalMessageID: fmt.Sprintf("configured-%d", idx),
			Origin:            "https://marketplace.example.com",
		}); err != nil {
			t.Fatalf("record channel failure: %v", err)
		}
	}
	dashboard, err = st.Dashboard()
	if err != nil {
		t.Fatalf("dashboard: %v", err)
	}
	if len(dashboard.Notifications) != 1 {
		t.Fatalf("expected one configured notification, got %#v", dashboard.Notifications)
	}
	if len(dashboard.PolicyEvents) < 2 {
		t.Fatalf("expected one policy audit event, got %#v", dashboard.PolicyEvents)
	}
	if dashboard.PolicyEvents[0].Actor != "ops-lead" || !strings.Contains(dashboard.PolicyEvents[0].After, "ANJING_NOTIFICATION_CUSTOM_SECRET") {
		t.Fatalf("expected policy audit event with actor and after summary, got %#v", dashboard.PolicyEvents[0])
	}
	notification := dashboard.Notifications[0]
	if notification.TargetURL != "https://ops.example.com/hooks/marketplace" || notification.SecretRef != "ANJING_NOTIFICATION_CUSTOM_SECRET" {
		t.Fatalf("expected configured notification target, got %#v", notification)
	}
	if notification.MaxAttempts != 5 || notification.BackoffSeconds != 45 {
		t.Fatalf("expected configured retry values, got %#v", notification)
	}
}

func TestRejectNotificationPolicyChangeKeepsCurrentPolicy(t *testing.T) {
	st := NewSeedStore()
	original, err := st.UpdateChannelAlertPolicy("Marketplace", "https://ops.example.com/hooks/reject", "ANJING_NOTIFICATION_REJECT_SECRET", 5, 45, "ops-a", "待拒绝变更")
	if err != nil {
		t.Fatalf("request channel alert policy update: %v", err)
	}
	dashboard, err := st.Dashboard()
	if err != nil {
		t.Fatalf("dashboard: %v", err)
	}
	if len(dashboard.PolicyChanges) != 1 || dashboard.PolicyChanges[0].Status != "PENDING" {
		t.Fatalf("expected pending policy change, got %#v", dashboard.PolicyChanges)
	}
	rejected, err := st.RejectNotificationPolicyChange(dashboard.PolicyChanges[0].ID, "ops-lead", "目标未备案")
	if err != nil {
		t.Fatalf("reject policy change: %v", err)
	}
	if rejected.Status != "REJECTED" || rejected.ApprovedBy != "ops-lead" {
		t.Fatalf("expected rejected policy change, got %#v", rejected)
	}
	dashboard, err = st.Dashboard()
	if err != nil {
		t.Fatalf("dashboard: %v", err)
	}
	var marketplace ChannelAlertPolicy
	for _, policy := range dashboard.AlertPolicies {
		if policy.Channel == "Marketplace" {
			marketplace = policy
			break
		}
	}
	if marketplace.TargetURL != original.TargetURL || marketplace.SecretRef != original.SecretRef {
		t.Fatalf("expected rejected change to keep original policy, got %#v", marketplace)
	}
	if dashboard.PolicyEvents[0].Action != "REJECT" || !strings.Contains(dashboard.PolicyEvents[0].Note, "目标未备案") {
		t.Fatalf("expected reject event, got %#v", dashboard.PolicyEvents[0])
	}
}

func TestCancelNotificationPolicyChangeKeepsCurrentPolicy(t *testing.T) {
	st := NewSeedStore()
	original, err := st.UpdateChannelAlertPolicy("Marketplace", "https://ops.example.com/hooks/cancel", "ANJING_NOTIFICATION_CANCEL_SECRET", 5, 45, "ops-a", "待撤销变更")
	if err != nil {
		t.Fatalf("request channel alert policy update: %v", err)
	}
	dashboard, err := st.Dashboard()
	if err != nil {
		t.Fatalf("dashboard: %v", err)
	}
	if len(dashboard.PolicyChanges) != 1 || dashboard.PolicyChanges[0].Status != "PENDING" || dashboard.PolicyChanges[0].ExpiresAt == "" {
		t.Fatalf("expected pending policy change with expiry, got %#v", dashboard.PolicyChanges)
	}
	canceled, err := st.CancelNotificationPolicyChange(dashboard.PolicyChanges[0].ID, "ops-a", "申请人撤销")
	if err != nil {
		t.Fatalf("cancel policy change: %v", err)
	}
	if canceled.Status != "CANCELED" || canceled.ApprovedBy != "ops-a" {
		t.Fatalf("expected canceled policy change, got %#v", canceled)
	}
	dashboard, err = st.Dashboard()
	if err != nil {
		t.Fatalf("dashboard: %v", err)
	}
	var marketplace ChannelAlertPolicy
	for _, policy := range dashboard.AlertPolicies {
		if policy.Channel == "Marketplace" {
			marketplace = policy
			break
		}
	}
	if marketplace.TargetURL != original.TargetURL || marketplace.SecretRef != original.SecretRef {
		t.Fatalf("expected canceled change to keep original policy, got %#v", marketplace)
	}
	if dashboard.PolicyEvents[0].Action != "CANCEL" || !strings.Contains(dashboard.PolicyEvents[0].Note, "申请人撤销") {
		t.Fatalf("expected cancel event, got %#v", dashboard.PolicyEvents[0])
	}
}

func TestExpiredNotificationPolicyChangeIsAudited(t *testing.T) {
	st := NewSeedStore()
	_, err := st.UpdateChannelAlertPolicy("Marketplace", "https://ops.example.com/hooks/expired", "ANJING_NOTIFICATION_EXPIRED_SECRET", 5, 45, "ops-a", "待过期变更")
	if err != nil {
		t.Fatalf("request channel alert policy update: %v", err)
	}
	dashboard, err := st.Dashboard()
	if err != nil {
		t.Fatalf("dashboard: %v", err)
	}
	if len(dashboard.PolicyChanges) != 1 {
		t.Fatalf("expected pending policy change, got %#v", dashboard.PolicyChanges)
	}
	st.mu.Lock()
	st.policyChanges[0].ExpiresAt = time.Now().UTC().Add(-time.Minute).Format(time.RFC3339)
	st.mu.Unlock()

	dashboard, err = st.Dashboard()
	if err != nil {
		t.Fatalf("dashboard: %v", err)
	}
	if dashboard.PolicyChanges[0].Status != "EXPIRED" || dashboard.PolicyChanges[0].ApprovedBy != "system" {
		t.Fatalf("expected expired policy change, got %#v", dashboard.PolicyChanges[0])
	}
	if dashboard.PolicyEvents[0].Action != "EXPIRE" {
		t.Fatalf("expected expire event, got %#v", dashboard.PolicyEvents[0])
	}
	if _, err := st.ApproveNotificationPolicyChange(dashboard.PolicyChanges[0].ID, "ops-lead", "审批通过"); err == nil {
		t.Fatal("expected expired policy change approval to fail")
	}
}

func TestDispatchChannelNotificationUsesDeliveryClientAndSecretRef(t *testing.T) {
	t.Setenv("ANJING_NOTIFICATION_MARKETPLACE_ONCALL_SECRET", "custom-notification-secret")
	client := &recordingDeliveryClient{}
	st := NewSeedStore(WithNotificationDeliveryClient(client))
	for idx := 0; idx < 3; idx++ {
		if err := st.RecordChannelFailure(ChannelFailureEvent{
			Channel:           "Marketplace",
			Code:              "channel_signature_invalid",
			Reason:            "签名错误",
			ExternalMessageID: fmt.Sprintf("custom-delivery-%d", idx),
			Origin:            "https://marketplace.example.com",
		}); err != nil {
			t.Fatalf("record channel failure: %v", err)
		}
	}
	dashboard, err := st.Dashboard()
	if err != nil {
		t.Fatalf("dashboard: %v", err)
	}
	sent, err := st.DispatchChannelNotification(dashboard.Notifications[0].ID, "SUCCESS")
	if err != nil {
		t.Fatalf("dispatch notification: %v", err)
	}
	if sent.Status != "SENT" || sent.ReceiptStatus != "204 NO_CONTENT" {
		t.Fatalf("expected custom delivery receipt, got %#v", sent)
	}
	if sent.SecretRef != "ANJING_NOTIFICATION_MARKETPLACE_ONCALL_SECRET" {
		t.Fatalf("expected secret ref, got %#v", sent)
	}
	if len(client.requests) != 1 || client.requests[0].Notification.ID != sent.ID {
		t.Fatalf("expected delivery client call, got %#v", client.requests)
	}
	expectedSignature := signChannelNotification(client.requests[0].Notification)
	if sent.Signature != expectedSignature {
		t.Fatalf("expected env-secret signature %s, got %s", expectedSignature, sent.Signature)
	}
	if len(sent.DeliveryAudit) != 1 {
		t.Fatalf("expected one delivery audit record, got %#v", sent.DeliveryAudit)
	}
	audit := sent.DeliveryAudit[0]
	if audit.Attempt != 1 || audit.PayloadHash == "" || audit.SignaturePreview != expectedSignature[:12]+"..." {
		t.Fatalf("expected redacted audit summary, got %#v", audit)
	}
	if strings.Contains(audit.RequestSummary, client.requests[0].SignedPayload) || strings.Contains(audit.ResponseSummary, expectedSignature) {
		t.Fatalf("expected audit summary to avoid raw payload/signature, got %#v", audit)
	}
}

func TestHTTPNotificationDeliveryClientPostsSignedWebhook(t *testing.T) {
	var signatureHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		signatureHeader = r.Header.Get("X-Anjing-Signature")
		if r.Header.Get("X-Anjing-Secret-Ref") != "ANJING_NOTIFICATION_MARKETPLACE_ONCALL_SECRET" {
			t.Fatalf("expected secret ref header, got %s", r.Header.Get("X-Anjing-Secret-Ref"))
		}
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte("accepted by test server"))
	}))
	defer server.Close()

	t.Setenv("ANJING_NOTIFICATION_MARKETPLACE_ONCALL_SECRET", "http-notification-secret")
	notification := ChannelNotification{
		ID:          "notice_http_1",
		Channel:     "Marketplace",
		Severity:    "HIGH",
		Target:      "marketplace-oncall",
		TargetURL:   server.URL,
		SecretRef:   "ANJING_NOTIFICATION_MARKETPLACE_ONCALL_SECRET",
		Count:       3,
		Attempts:    1,
		MaxAttempts: 3,
	}
	notification.Signature = signChannelNotification(notification)
	client := NewHTTPNotificationDeliveryClient(time.Second)
	result, err := client.DeliverChannelNotification(context.Background(), NotificationDeliveryRequest{
		Notification:  notification,
		Outcome:       "SUCCESS",
		SignedPayload: notificationPayload(notification),
	})
	if err != nil {
		t.Fatalf("deliver webhook: %v", err)
	}
	if !result.Accepted || result.ReceiptStatus != "202 Accepted" || result.ReceiptBody != "accepted by test server" {
		t.Fatalf("expected accepted receipt, got %#v", result)
	}
	if signatureHeader != notification.Signature {
		t.Fatalf("expected signature header %s, got %s", notification.Signature, signatureHeader)
	}
}
