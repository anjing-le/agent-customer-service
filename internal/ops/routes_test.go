package ops

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/anjing-le/agent-customer-service/internal/platform/store"
)

func TestSubmitAnnotationRouteReturnsCreatedAnnotation(t *testing.T) {
	st := store.NewSeedStore()
	result, err := st.SendMessage("conv_demo_refund", "这个商品能不能开发票？")
	if err != nil {
		t.Fatalf("send message: %v", err)
	}

	mux := http.NewServeMux()
	Register(mux, st)

	body := strings.NewReader(`{"messageId":"` + result.AgentMessage.ID + `","reviewer":"qa-a","verdict":"PASS","note":"证据充分","dimensions":{"groundedness":5,"safety":5,"helpfulness":4},"tags":["human_review"]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/ops/annotations/submit", body)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	for _, expected := range []string{`"messageId":"` + result.AgentMessage.ID + `"`, `"reviewer":"qa-a"`, `"score":93`} {
		if !strings.Contains(rec.Body.String(), expected) {
			t.Fatalf("expected %s in response, got %s", expected, rec.Body.String())
		}
	}
}

func TestSubmitAnnotationRouteRequiresMessageID(t *testing.T) {
	mux := http.NewServeMux()
	Register(mux, store.NewSeedStore())

	req := httptest.NewRequest(http.MethodPost, "/api/ops/annotations/submit", strings.NewReader(`{"reviewer":"qa-a"}`))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "messageId is required") {
		t.Fatalf("expected validation error, got %s", rec.Body.String())
	}
}
