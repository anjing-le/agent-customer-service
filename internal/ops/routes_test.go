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

func TestReviewTaskRoutesAssignAndCompleteTask(t *testing.T) {
	st := store.NewSeedStore()
	dashboard, err := st.Dashboard()
	if err != nil {
		t.Fatalf("dashboard: %v", err)
	}
	if len(dashboard.ReviewTasks) == 0 {
		t.Fatal("expected seeded review task")
	}
	taskID := dashboard.ReviewTasks[0].ID

	mux := http.NewServeMux()
	Register(mux, st)

	assignReq := httptest.NewRequest(http.MethodPost, "/api/ops/review-tasks/assign", strings.NewReader(`{"id":"`+taskID+`","assignee":"qa-a"}`))
	assignRec := httptest.NewRecorder()
	mux.ServeHTTP(assignRec, assignReq)

	if assignRec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", assignRec.Code, assignRec.Body.String())
	}
	for _, expected := range []string{`"id":"` + taskID + `"`, `"status":"ASSIGNED"`, `"assignee":"qa-a"`} {
		if !strings.Contains(assignRec.Body.String(), expected) {
			t.Fatalf("expected %s in response, got %s", expected, assignRec.Body.String())
		}
	}

	completeReq := httptest.NewRequest(http.MethodPost, "/api/ops/review-tasks/complete", strings.NewReader(`{"id":"`+taskID+`"}`))
	completeRec := httptest.NewRecorder()
	mux.ServeHTTP(completeRec, completeReq)

	if completeRec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", completeRec.Code, completeRec.Body.String())
	}
	if !strings.Contains(completeRec.Body.String(), `"status":"COMPLETED"`) {
		t.Fatalf("expected completed task, got %s", completeRec.Body.String())
	}
}

func TestExportTrainingSamplesRouteReturnsLowScoreSamples(t *testing.T) {
	st := store.NewSeedStore()
	result, err := st.SendMessage("conv_demo_refund", "这个商品能不能开发票？")
	if err != nil {
		t.Fatalf("send message: %v", err)
	}
	if _, err := st.SubmitAnnotation(result.AgentMessage.ID, "qa-a", "REVIEW", "帮助性不足", store.AnnotationDimensions{
		Groundedness: 3,
		Safety:       4,
		Helpfulness:  2,
	}, []string{"low_score"}); err != nil {
		t.Fatalf("submit annotation: %v", err)
	}

	mux := http.NewServeMux()
	Register(mux, st)

	req := httptest.NewRequest(http.MethodGet, "/api/ops/training-samples/export?maxScore=80", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	for _, expected := range []string{`"messageId":"` + result.AgentMessage.ID + `"`, `"verdict":"REVIEW"`, `"prompt":"这个商品能不能开发票？"`} {
		if !strings.Contains(rec.Body.String(), expected) {
			t.Fatalf("expected %s in response, got %s", expected, rec.Body.String())
		}
	}
}

func TestExportTrainingSamplesRouteRejectsInvalidMaxScore(t *testing.T) {
	mux := http.NewServeMux()
	Register(mux, store.NewSeedStore())

	req := httptest.NewRequest(http.MethodGet, "/api/ops/training-samples/export?maxScore=bad", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "maxScore must be a number") {
		t.Fatalf("expected validation error, got %s", rec.Body.String())
	}
}
