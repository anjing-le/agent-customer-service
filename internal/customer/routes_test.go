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
