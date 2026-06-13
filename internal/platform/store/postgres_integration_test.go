package store

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/anjing-le/agent-customer-service/internal/platform/db"
)

func TestPostgresStoreRuntime(t *testing.T) {
	databaseURL := os.Getenv("ANJING_INTEGRATION_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set ANJING_INTEGRATION_DATABASE_URL to run PostgreSQL integration test")
	}

	ctx := context.Background()
	pool, err := db.Open(ctx, databaseURL)
	if err != nil {
		t.Fatalf("open postgres: %v", err)
	}
	defer pool.Close()

	if _, err := db.RunMigrations(ctx, pool, "../../../infra/postgres/migrations"); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	st := NewPostgresStore(pool)
	replayKey := fmt.Sprintf("integration-replay-key-%d", time.Now().UnixNano())
	accepted, err := st.RecordChannelInbound(ChannelInboundReceipt{
		ReplayKey:              replayKey,
		Channel:                "WeChat",
		ExternalConversationID: "wx-integration",
		ExternalMessageID:      "wx-message-" + replayKey,
		Timestamp:              "2026-06-14T02:30:00Z",
		Signature:              "integration-signature",
		ContentHash:            "integration-content-hash",
	})
	if err != nil {
		t.Fatalf("record channel inbound: %v", err)
	}
	if !accepted {
		t.Fatalf("expected first channel inbound receipt to be accepted")
	}
	accepted, err = st.RecordChannelInbound(ChannelInboundReceipt{
		ReplayKey:              replayKey,
		Channel:                "WeChat",
		ExternalConversationID: "wx-integration",
		ExternalMessageID:      "wx-message-" + replayKey,
		Timestamp:              "2026-06-14T02:30:00Z",
		Signature:              "integration-signature",
		ContentHash:            "integration-content-hash",
	})
	if err != nil {
		t.Fatalf("record duplicate channel inbound: %v", err)
	}
	if accepted {
		t.Fatalf("expected duplicate channel inbound receipt to be rejected")
	}

	knowledge, err := st.SearchKnowledge("发票")
	if err != nil {
		t.Fatalf("search knowledge: %v", err)
	}
	if len(knowledge) == 0 || knowledge[0].ID != "kb_invoice" {
		t.Fatalf("expected invoice knowledge, got %#v", knowledge)
	}

	result, err := st.SendMessage("conv_integration", "完全没有资料的新品保价规则是什么？")
	if err != nil {
		t.Fatalf("send message: %v", err)
	}
	if result.Gap == nil || result.AgentMessage.FallbackReason != "NO_EVIDENCE" {
		t.Fatalf("expected no evidence gap, got %#v", result)
	}

	transfer, err := st.SendMessage("conv_integration_transfer", "我已经投诉很多次了，必须转人工")
	if err != nil {
		t.Fatalf("send transfer message: %v", err)
	}
	if transfer.AgentMessage.FallbackReason != "TRANSFER_THRESHOLD" {
		t.Fatalf("expected transfer fallback, got %#v", transfer)
	}

	dashboard, err := st.Dashboard()
	if err != nil {
		t.Fatalf("dashboard: %v", err)
	}
	if len(dashboard.KnowledgeGaps) == 0 {
		t.Fatalf("expected persisted knowledge gap, got %#v", dashboard)
	}
	if len(dashboard.Transfers) == 0 {
		t.Fatalf("expected persisted transfer ticket, got %#v", dashboard)
	}
	if len(dashboard.ChannelPolicies) == 0 {
		t.Fatalf("expected channel policies, got %#v", dashboard)
	}
	if len(dashboard.Integrations) == 0 {
		t.Fatalf("expected channel integrations, got %#v", dashboard)
	}
}
