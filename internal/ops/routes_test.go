package ops

import (
	"fmt"
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

func TestCompareRulesRouteReturnsCanaryResult(t *testing.T) {
	mux := http.NewServeMux()
	Register(mux, store.NewSeedStore())

	req := httptest.NewRequest(http.MethodPost, "/api/ops/rules/compare", strings.NewReader(`{"content":"我想取消订单，7 天无理由退货怎么处理"}`))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	for _, expected := range []string{`"changed":true`, `"current"`, `"canary"`, `"ruleCode":"CANCEL_RISK_TRANSFER"`} {
		if !strings.Contains(rec.Body.String(), expected) {
			t.Fatalf("expected %s in response, got %s", expected, rec.Body.String())
		}
	}
}

func TestRuleReleaseRoutesPublishAndRollback(t *testing.T) {
	mux := http.NewServeMux()
	Register(mux, store.NewSeedStore())

	blockedReq := httptest.NewRequest(http.MethodPost, "/api/ops/rules/publish-canary", strings.NewReader(`{"code":"CANCEL_RISK_TRANSFER","actor":"qa-lead","note":"未审批直接发布"}`))
	blockedRec := httptest.NewRecorder()
	mux.ServeHTTP(blockedRec, blockedReq)

	if blockedRec.Code != http.StatusInternalServerError {
		t.Fatalf("expected publish gate error, got %d: %s", blockedRec.Code, blockedRec.Body.String())
	}
	if !strings.Contains(blockedRec.Body.String(), "requires approved gate") {
		t.Fatalf("expected approval gate error, got %s", blockedRec.Body.String())
	}

	approvalReq := httptest.NewRequest(http.MethodPost, "/api/ops/rules/approve", strings.NewReader(`{"code":"CANCEL_RISK_TRANSFER","approver":"qa-lead","riskLevel":"LOW","sampleIds":["sample_cancel_1","sample_cancel_2","sample_cancel_3"],"note":"灰度样本通过"}`))
	approvalRec := httptest.NewRecorder()
	mux.ServeHTTP(approvalRec, approvalReq)

	if approvalRec.Code != http.StatusOK {
		t.Fatalf("expected approval 200, got %d: %s", approvalRec.Code, approvalRec.Body.String())
	}
	if !strings.Contains(approvalRec.Body.String(), `"status":"APPROVED"`) || !strings.Contains(approvalRec.Body.String(), `"sampleCount":3`) {
		t.Fatalf("expected approved gate, got %s", approvalRec.Body.String())
	}

	publishReq := httptest.NewRequest(http.MethodPost, "/api/ops/rules/publish-canary", strings.NewReader(`{"code":"CANCEL_RISK_TRANSFER","actor":"qa-lead","note":"发布灰度规则"}`))
	publishRec := httptest.NewRecorder()
	mux.ServeHTTP(publishRec, publishReq)

	if publishRec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", publishRec.Code, publishRec.Body.String())
	}
	for _, expected := range []string{`"ruleCode":"CANCEL_RISK_TRANSFER"`, `"action":"PUBLISH"`, `"actor":"qa-lead"`} {
		if !strings.Contains(publishRec.Body.String(), expected) {
			t.Fatalf("expected %s in response, got %s", expected, publishRec.Body.String())
		}
	}

	rollbackReq := httptest.NewRequest(http.MethodPost, "/api/ops/rules/rollback", strings.NewReader(`{"code":"CANCEL_RISK_TRANSFER","actor":"qa-lead","note":"回滚灰度规则"}`))
	rollbackRec := httptest.NewRecorder()
	mux.ServeHTTP(rollbackRec, rollbackReq)

	if rollbackRec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rollbackRec.Code, rollbackRec.Body.String())
	}
	if !strings.Contains(rollbackRec.Body.String(), `"action":"ROLLBACK"`) {
		t.Fatalf("expected rollback event, got %s", rollbackRec.Body.String())
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

func TestAcknowledgeChannelNotificationRoute(t *testing.T) {
	st := store.NewSeedStore()
	for idx := 0; idx < 3; idx++ {
		if err := st.RecordChannelFailure(store.ChannelFailureEvent{
			Channel:           "Marketplace",
			Code:              "channel_signature_invalid",
			Reason:            "签名错误",
			ExternalMessageID: fmt.Sprintf("event-%d", idx),
			Origin:            "https://marketplace.example.com",
		}); err != nil {
			t.Fatalf("record failure: %v", err)
		}
	}
	dashboard, err := st.Dashboard()
	if err != nil {
		t.Fatalf("dashboard: %v", err)
	}
	if len(dashboard.Notifications) == 0 {
		t.Fatal("expected notification")
	}

	mux := http.NewServeMux()
	Register(mux, st)

	req := httptest.NewRequest(http.MethodPost, "/api/ops/channel-notifications/ack", strings.NewReader(`{"id":"`+dashboard.Notifications[0].ID+`","actor":"ops-a","note":"已确认"}`))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"status":"ACKED"`) {
		t.Fatalf("expected acked notification, got %s", rec.Body.String())
	}
}

func TestUpdateChannelAlertPolicyRoute(t *testing.T) {
	st := store.NewSeedStore()
	mux := http.NewServeMux()
	Register(mux, st)

	req := httptest.NewRequest(http.MethodPost, "/api/ops/channel-alert-policies/update", strings.NewReader(`{"channel":"Marketplace","targetUrl":"https://ops.example.com/hooks/marketplace","secretRef":"ANJING_NOTIFICATION_CUSTOM_SECRET","maxAttempts":4,"backoffSeconds":30,"actor":"ops-a","note":"切换通知目标"}`))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	dashboard, err := st.Dashboard()
	if err != nil {
		t.Fatalf("dashboard: %v", err)
	}
	if len(dashboard.PolicyChanges) != 1 || dashboard.PolicyChanges[0].Status != "PENDING" {
		t.Fatalf("expected pending policy change, got %#v", dashboard.PolicyChanges)
	}
	if len(dashboard.PolicyEvents) != 1 || dashboard.PolicyEvents[0].Action != "REQUEST_APPROVAL" {
		t.Fatalf("expected policy audit event, got %#v", dashboard.PolicyEvents)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/ops/channel-alert-policies/approve-change", strings.NewReader(`{"id":"`+dashboard.PolicyChanges[0].ID+`","approver":"ops-lead","note":"审批通过"}`))
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	for _, expected := range []string{`"targetUrl":"https://ops.example.com/hooks/marketplace"`, `"secretRef":"ANJING_NOTIFICATION_CUSTOM_SECRET"`, `"maxAttempts":4`, `"backoffSeconds":30`} {
		if !strings.Contains(rec.Body.String(), expected) {
			t.Fatalf("expected %s in response, got %s", expected, rec.Body.String())
		}
	}
}

func TestDispatchChannelNotificationRouteRetriesAndDeadLetters(t *testing.T) {
	st := store.NewSeedStore()
	for idx := 0; idx < 3; idx++ {
		if err := st.RecordChannelFailure(store.ChannelFailureEvent{
			Channel:           "Marketplace",
			Code:              "channel_signature_invalid",
			Reason:            "签名错误",
			ExternalMessageID: fmt.Sprintf("dead-%d", idx),
			Origin:            "https://marketplace.example.com",
		}); err != nil {
			t.Fatalf("record failure: %v", err)
		}
	}
	dashboard, err := st.Dashboard()
	if err != nil {
		t.Fatalf("dashboard: %v", err)
	}
	notificationID := dashboard.Notifications[0].ID

	mux := http.NewServeMux()
	Register(mux, st)

	for idx := 0; idx < 3; idx++ {
		req := httptest.NewRequest(http.MethodPost, "/api/ops/channel-notifications/dispatch", strings.NewReader(`{"id":"`+notificationID+`","outcome":"webhook_timeout"}`))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
		if !strings.Contains(rec.Body.String(), `"signature":"`) {
			t.Fatalf("expected signed dispatch, got %s", rec.Body.String())
		}
		if idx == 0 && (!strings.Contains(rec.Body.String(), `"nextRetryAt":"`) || !strings.Contains(rec.Body.String(), `"backoffSeconds":120`)) {
			t.Fatalf("expected retry backoff in response, got %s", rec.Body.String())
		}
	}
	finalDashboard, err := st.Dashboard()
	if err != nil {
		t.Fatalf("dashboard: %v", err)
	}
	if finalDashboard.Notifications[0].Status != "DEAD_LETTER" || finalDashboard.Notifications[0].Attempts != 3 {
		t.Fatalf("expected dead letter notification, got %#v", finalDashboard.Notifications[0])
	}
}

func TestDispatchChannelNotificationRouteRecordsReceipt(t *testing.T) {
	st := store.NewSeedStore()
	for idx := 0; idx < 3; idx++ {
		if err := st.RecordChannelFailure(store.ChannelFailureEvent{
			Channel:           "Marketplace",
			Code:              "channel_signature_invalid",
			Reason:            "签名错误",
			ExternalMessageID: fmt.Sprintf("sent-%d", idx),
			Origin:            "https://marketplace.example.com",
		}); err != nil {
			t.Fatalf("record failure: %v", err)
		}
	}
	dashboard, err := st.Dashboard()
	if err != nil {
		t.Fatalf("dashboard: %v", err)
	}

	mux := http.NewServeMux()
	Register(mux, st)

	req := httptest.NewRequest(http.MethodPost, "/api/ops/channel-notifications/dispatch", strings.NewReader(`{"id":"`+dashboard.Notifications[0].ID+`","outcome":"SUCCESS"}`))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	for _, expected := range []string{`"status":"SENT"`, `"receiptStatus":"202 ACCEPTED"`, `"targetUrl":"https://hooks.example.com/anjing/marketplace-oncall"`} {
		if !strings.Contains(rec.Body.String(), expected) {
			t.Fatalf("expected %s in response, got %s", expected, rec.Body.String())
		}
	}
	for _, expected := range []string{`"deliveryAudit":[`, `"payloadHash":"`, `"signaturePreview":"`, `"requestSummary":"POST https://hooks.example.com/anjing/marketplace-oncall`} {
		if !strings.Contains(rec.Body.String(), expected) {
			t.Fatalf("expected audit field %s in response, got %s", expected, rec.Body.String())
		}
	}
	if strings.Contains(rec.Body.String(), "signedPayload") {
		t.Fatalf("expected response to avoid raw signed payload, got %s", rec.Body.String())
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
