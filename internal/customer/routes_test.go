package customer

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/anjing-le/agent-customer-service/internal/platform/store"
)

type failingRuntime struct{}

func (f failingRuntime) ListConversations() ([]store.Conversation, error) {
	return nil, errors.New("database unavailable")
}

func (f failingRuntime) CreateConversation(string, string) (store.Conversation, error) {
	return store.Conversation{}, errors.New("database unavailable")
}

func (f failingRuntime) ListMessages(string) ([]store.Message, error) {
	return nil, errors.New("database unavailable")
}

func (f failingRuntime) SendMessage(string, string) (store.SendMessageResult, error) {
	return store.SendMessageResult{}, errors.New("database unavailable")
}

func (f failingRuntime) ListKnowledge() ([]store.KnowledgeArticle, error) {
	return nil, errors.New("database unavailable")
}

func (f failingRuntime) SearchKnowledge(string) ([]store.KnowledgeArticle, error) {
	return nil, errors.New("database unavailable")
}

func (f failingRuntime) ResolveKnowledgeGap(string) (store.KnowledgeGap, error) {
	return store.KnowledgeGap{}, errors.New("database unavailable")
}

func (f failingRuntime) CreateArticleFromGap(string, string, string, string, []string) (store.KnowledgeArticle, error) {
	return store.KnowledgeArticle{}, errors.New("database unavailable")
}

func (f failingRuntime) TestRule(string) (store.RuleTestResult, error) {
	return store.RuleTestResult{}, errors.New("database unavailable")
}

func (f failingRuntime) ResolveTransferTicket(string, string, string) (store.TransferTicket, error) {
	return store.TransferTicket{}, errors.New("database unavailable")
}

func (f failingRuntime) SubmitAnnotation(string, string, string, string, store.AnnotationDimensions, []string) (store.Annotation, error) {
	return store.Annotation{}, errors.New("database unavailable")
}

func (f failingRuntime) Dashboard() (store.Dashboard, error) {
	return store.Dashboard{}, errors.New("database unavailable")
}

func TestConversationRouteReturnsStoreErrorEnvelope(t *testing.T) {
	mux := http.NewServeMux()
	Register(mux, failingRuntime{})

	req := httptest.NewRequest(http.MethodGet, "/api/customer-service/conversations", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"code":"store_error"`) {
		t.Fatalf("expected store_error envelope, got %s", rec.Body.String())
	}
}

func TestListMessagesRouteReturnsConversationHistory(t *testing.T) {
	mux := http.NewServeMux()
	Register(mux, store.NewSeedStore())

	req := httptest.NewRequest(http.MethodGet, "/api/customer-service/messages?conversationId=conv_demo_refund", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"conversationId":"conv_demo_refund"`) {
		t.Fatalf("expected conversation history, got %s", rec.Body.String())
	}
}

func TestStreamMessagesRouteReturnsSSEEvents(t *testing.T) {
	mux := http.NewServeMux()
	Register(mux, store.NewSeedStore())

	body := strings.NewReader(`{"conversationId":"conv_demo_refund","content":"这个商品能不能开发票？"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/customer-service/messages/stream", body)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if contentType := rec.Header().Get("Content-Type"); !strings.Contains(contentType, "text/event-stream") {
		t.Fatalf("expected event stream content type, got %q", contentType)
	}
	bodyText := rec.Body.String()
	for _, expected := range []string{"event: meta", "event: delta", "event: done", `"engine":"rag+rule"`, `"agentMessage"`} {
		if !strings.Contains(bodyText, expected) {
			t.Fatalf("expected %q in SSE body, got %s", expected, bodyText)
		}
	}
}
