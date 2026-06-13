package store

import "testing"

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
}
