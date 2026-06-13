package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/anjing-le/agent-customer-service/internal/platform/store"
)

func TestClientGenerateReplyUsesOpenAICompatibleRequest(t *testing.T) {
	var captured chatRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Fatalf("unexpected authorization header %q", r.Header.Get("Authorization"))
		}
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"基于证据的模型回复"}}]}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key", "test-model")
	reply, err := client.GenerateReply(context.Background(), store.ReplyRequest{
		ConversationID: "conv_test",
		Question:       "能开发票吗？",
		Evidence: []store.KnowledgeArticle{{
			ID: "kb_invoice", Title: "电子发票开具", Content: "订单完成后可申请电子发票。", TrustLevel: "HIGH",
		}},
		History: []store.Message{{Role: "user", Content: "上一轮问题"}},
	})
	if err != nil {
		t.Fatalf("generate reply: %v", err)
	}
	if reply.Content != "基于证据的模型回复" {
		t.Fatalf("unexpected reply %q", reply.Content)
	}
	if reply.Model != "test-model" {
		t.Fatalf("unexpected reply model %q", reply.Model)
	}
	if captured.Model != "test-model" {
		t.Fatalf("unexpected model %q", captured.Model)
	}
	if len(captured.Messages) < 3 {
		t.Fatalf("expected system, history and user messages, got %#v", captured.Messages)
	}
}
