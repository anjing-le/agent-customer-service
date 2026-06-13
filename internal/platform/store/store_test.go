package store

import (
	"context"
	"errors"
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
	ticket := withTransferSLA(demoTransferTicket("ticket_sla", "conv_sla", "必须人工处理", createdAt), time.Now().UTC())

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
