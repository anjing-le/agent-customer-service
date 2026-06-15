package ops

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

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

	req = httptest.NewRequest(http.MethodPost, "/api/ops/channel-alert-policies/approve-change", strings.NewReader(`{"id":"`+dashboard.PolicyChanges[0].ID+`","approver":"ops-a","note":"审批通过","confirmation":"`+dashboard.PolicyChanges[0].ConfirmationText+`"}`))
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected unauthorized approval to fail, got %d: %s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/api/ops/channel-alert-policies/approve-change", strings.NewReader(`{"id":"`+dashboard.PolicyChanges[0].ID+`","approver":"ops-lead","note":"审批通过","confirmation":"`+dashboard.PolicyChanges[0].ConfirmationText+`"}`))
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

	req = httptest.NewRequest(http.MethodPost, "/api/ops/channel-alert-policies/rollback", strings.NewReader(`{"channel":"Marketplace","actor":"ops-lead","note":"回滚通知目标","confirmation":"ROLLBACK Marketplace"}`))
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), `"secretRef":"ANJING_NOTIFICATION_CUSTOM_SECRET"`) {
		t.Fatalf("expected rollback response to restore previous secret ref, got %s", rec.Body.String())
	}
}

func TestRejectChannelAlertPolicyChangeRoute(t *testing.T) {
	st := store.NewSeedStore()
	mux := http.NewServeMux()
	Register(mux, st)

	req := httptest.NewRequest(http.MethodPost, "/api/ops/channel-alert-policies/update", strings.NewReader(`{"channel":"Marketplace","targetUrl":"https://ops.example.com/hooks/reject","secretRef":"ANJING_NOTIFICATION_REJECT_SECRET","maxAttempts":4,"backoffSeconds":30,"actor":"ops-a","note":"待拒绝"}`))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	dashboard, err := st.Dashboard()
	if err != nil {
		t.Fatalf("dashboard: %v", err)
	}
	if len(dashboard.PolicyChanges) != 1 {
		t.Fatalf("expected pending policy change, got %#v", dashboard.PolicyChanges)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/ops/channel-alert-policies/reject-change", strings.NewReader(`{"id":"`+dashboard.PolicyChanges[0].ID+`","reviewer":"ops-lead","note":"目标未备案"}`))
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	for _, expected := range []string{`"status":"REJECTED"`, `"approvedBy":"ops-lead"`, `"note":"待拒绝"`} {
		if !strings.Contains(rec.Body.String(), expected) {
			t.Fatalf("expected %s in response, got %s", expected, rec.Body.String())
		}
	}
}

func TestCancelChannelAlertPolicyChangeRoute(t *testing.T) {
	st := store.NewSeedStore()
	mux := http.NewServeMux()
	Register(mux, st)

	req := httptest.NewRequest(http.MethodPost, "/api/ops/channel-alert-policies/update", strings.NewReader(`{"channel":"Marketplace","targetUrl":"https://ops.example.com/hooks/cancel","secretRef":"ANJING_NOTIFICATION_CANCEL_SECRET","maxAttempts":4,"backoffSeconds":30,"actor":"ops-a","note":"待撤销"}`))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	dashboard, err := st.Dashboard()
	if err != nil {
		t.Fatalf("dashboard: %v", err)
	}
	if len(dashboard.PolicyChanges) != 1 {
		t.Fatalf("expected pending policy change, got %#v", dashboard.PolicyChanges)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/ops/channel-alert-policies/cancel-change", strings.NewReader(`{"id":"`+dashboard.PolicyChanges[0].ID+`","actor":"ops-a","note":"申请人撤销"}`))
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	for _, expected := range []string{`"status":"CANCELED"`, `"approvedBy":"ops-a"`, `"note":"待撤销"`} {
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
	if len(finalDashboard.ChannelRunbooks) != 1 || finalDashboard.ChannelRunbooks[0].Status != "ESCALATE" {
		t.Fatalf("expected escalate runbook, got %#v", finalDashboard.ChannelRunbooks)
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

func TestExportChannelOpsReportRouteReturnsMarkdownAndCSV(t *testing.T) {
	st := store.NewSeedStore()
	for idx := 0; idx < 3; idx++ {
		if err := st.RecordChannelFailure(store.ChannelFailureEvent{
			Channel:           "Marketplace",
			Code:              "channel_signature_invalid",
			Reason:            "签名错误",
			ExternalMessageID: fmt.Sprintf("report-%d", idx),
			Origin:            "https://marketplace.example.com",
		}); err != nil {
			t.Fatalf("record failure: %v", err)
		}
	}
	if err := st.RecordChannelInboundAudit(store.ChannelInboundAudit{Channel: "Marketplace", Status: "ACCEPTED", Code: "accepted"}); err != nil {
		t.Fatalf("record accepted audit: %v", err)
	}
	for idx := 0; idx < 3; idx++ {
		if err := st.RecordChannelInboundAudit(store.ChannelInboundAudit{Channel: "Marketplace", Status: "REJECTED", Code: "invalid_signature"}); err != nil {
			t.Fatalf("record rejected audit: %v", err)
		}
	}
	mux := http.NewServeMux()
	Register(mux, st)

	req := httptest.NewRequest(http.MethodGet, "/api/ops/channel-ops-report/export?format=markdown", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Header().Get("Content-Type"), "text/markdown") {
		t.Fatalf("expected markdown content type, got %s", rec.Header().Get("Content-Type"))
	}
	for _, expected := range []string{"# Agent Customer Service Channel Ops Report", "Marketplace", "channel_signature_invalid", "Inbound audits: total=4 accepted=1 rejected=3 acceptance_rate=25%", "Inbound quality events: total=1 active=1 watch=0 recovered=0", "Handoff priorities:", "Marketplace `INBOUND_AUDIT_ACTIVE`", "`invalid_signature`: 3", "Active channels: Marketplace"} {
		if !strings.Contains(rec.Body.String(), expected) {
			t.Fatalf("expected %s in markdown report, got %s", expected, rec.Body.String())
		}
	}

	req = httptest.NewRequest(http.MethodGet, "/api/ops/channel-ops-report/export?format=csv", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Header().Get("Content-Type"), "text/csv") {
		t.Fatalf("expected csv content type, got %s", rec.Header().Get("Content-Type"))
	}
	if !strings.Contains(rec.Body.String(), "section,channel,status,code,count,owner,next_action,escalation") {
		t.Fatalf("expected csv header, got %s", rec.Body.String())
	}
	for _, expected := range []string{"inbound_audit,,ACCEPTANCE_RATE,accepted,1/4 (25%)", "inbound_audit_error,,REJECTED,invalid_signature,3", "inbound_quality,,SUMMARY,events,total=1 active=1 watch=0 recovered=0", "inbound_quality_channel,Marketplace,ACTIVE", "handoff_priority,Marketplace", "INBOUND_AUDIT_ACTIVE"} {
		if !strings.Contains(rec.Body.String(), expected) {
			t.Fatalf("expected %s in csv report, got %s", expected, rec.Body.String())
		}
	}
}

func TestExportChannelOpsReportRouteRejectsInvalidFormat(t *testing.T) {
	mux := http.NewServeMux()
	Register(mux, store.NewSeedStore())

	req := httptest.NewRequest(http.MethodGet, "/api/ops/channel-ops-report/export?format=pdf", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "format must be markdown or csv") {
		t.Fatalf("expected validation error, got %s", rec.Body.String())
	}
}

func TestChannelOpsReportHistoryRoutesGenerateListAndExport(t *testing.T) {
	st := store.NewSeedStore()
	if err := st.RecordChannelFailure(store.ChannelFailureEvent{
		Channel:           "App",
		Code:              "channel_rate_limited",
		Reason:            "渠道限流",
		ExternalMessageID: "history-1",
		Origin:            "app://agent-customer-service",
	}); err != nil {
		t.Fatalf("record failure: %v", err)
	}
	if err := st.RecordChannelInboundAudit(store.ChannelInboundAudit{Channel: "App", Status: "ACCEPTED", Code: "accepted"}); err != nil {
		t.Fatalf("record accepted audit: %v", err)
	}
	mux := http.NewServeMux()
	Register(mux, st)

	generateReq := httptest.NewRequest(http.MethodPost, "/api/ops/channel-ops-reports/generate", strings.NewReader(`{"format":"markdown"}`))
	generateRec := httptest.NewRecorder()
	mux.ServeHTTP(generateRec, generateReq)

	if generateRec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", generateRec.Code, generateRec.Body.String())
	}
	for _, expected := range []string{`"format":"markdown"`, `"contentType":"text/markdown; charset=utf-8"`, `"failureCount":1`, `"handoffPriorities":[{"rank":1`, `"source":"HIGH_FREQUENCY_FAILURE"`, `"inboundAudit":{"total":1,"accepted":1,"rejected":0,"acceptanceRate":100`, "# Agent Customer Service Channel Ops Report"} {
		if !strings.Contains(generateRec.Body.String(), expected) {
			t.Fatalf("expected %s in generated report, got %s", expected, generateRec.Body.String())
		}
	}
	reportID := generatedReportID(generateRec.Body.String())
	if reportID == "" {
		t.Fatalf("expected generated report id, got %s", generateRec.Body.String())
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/ops/channel-ops-reports?limit=5", nil)
	listRec := httptest.NewRecorder()
	mux.ServeHTTP(listRec, listReq)

	if listRec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", listRec.Code, listRec.Body.String())
	}
	for _, expected := range []string{`"id":"` + reportID + `"`, `"failureCount":1`} {
		if !strings.Contains(listRec.Body.String(), expected) {
			t.Fatalf("expected %s in report list, got %s", expected, listRec.Body.String())
		}
	}
	if strings.Contains(listRec.Body.String(), "# Agent Customer Service Channel Ops Report") {
		t.Fatalf("expected list response to omit report content, got %s", listRec.Body.String())
	}

	exportReq := httptest.NewRequest(http.MethodGet, "/api/ops/channel-ops-reports/export?id="+reportID, nil)
	exportRec := httptest.NewRecorder()
	mux.ServeHTTP(exportRec, exportReq)

	if exportRec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", exportRec.Code, exportRec.Body.String())
	}
	if !strings.Contains(exportRec.Header().Get("Content-Type"), "text/markdown") {
		t.Fatalf("expected markdown content type, got %s", exportRec.Header().Get("Content-Type"))
	}
	for _, expected := range []string{"# Agent Customer Service Channel Ops Report", "App", "channel_rate_limited"} {
		if !strings.Contains(exportRec.Body.String(), expected) {
			t.Fatalf("expected %s in exported history report, got %s", expected, exportRec.Body.String())
		}
	}
}

func TestGenerateChannelOpsReportRouteRejectsInvalidFormat(t *testing.T) {
	mux := http.NewServeMux()
	Register(mux, store.NewSeedStore())

	req := httptest.NewRequest(http.MethodPost, "/api/ops/channel-ops-reports/generate", strings.NewReader(`{"format":"pdf"}`))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "format must be markdown or csv") {
		t.Fatalf("expected validation error, got %s", rec.Body.String())
	}
}

func TestChannelOpsReportSchedulerRouteReturnsStatus(t *testing.T) {
	st := store.NewSeedStore()
	scheduler := NewReportScheduler(st, ReportSchedulerConfig{
		Enabled:    true,
		Format:     "csv",
		Interval:   15 * time.Minute,
		Retain:     7,
		RunOnStart: true,
	}, nil)
	scheduler.runAndLog(context.Background())

	mux := http.NewServeMux()
	RegisterWithReportScheduler(mux, st, scheduler)

	req := httptest.NewRequest(http.MethodGet, "/api/ops/channel-ops-report-scheduler", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	for _, expected := range []string{`"enabled":true`, `"format":"csv"`, `"intervalMins":15`, `"retain":7`, `"runOnStart":true`, `"lastStatus":"SUCCESS"`, `"lastReportId":"`} {
		if !strings.Contains(rec.Body.String(), expected) {
			t.Fatalf("expected %s in scheduler status, got %s", expected, rec.Body.String())
		}
	}
}

func TestChannelOpsReportCompensationRouteGeneratesReportAndEvent(t *testing.T) {
	st := store.NewSeedStore()
	scheduler := NewReportScheduler(st, ReportSchedulerConfig{
		Enabled:  false,
		Format:   "markdown",
		Interval: time.Hour,
		Retain:   10,
	}, nil)
	mux := http.NewServeMux()
	RegisterWithReportScheduler(mux, st, scheduler)

	req := httptest.NewRequest(http.MethodPost, "/api/ops/channel-ops-report-scheduler/compensate", strings.NewReader(`{"actor":"ops-lead","note":"补偿日报"}`))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	for _, expected := range []string{`"action":"COMPENSATE"`, `"actor":"ops-lead"`, `"status":"SUCCESS"`, `"reportId":"`, `"lastStatus":"SUCCESS"`} {
		if !strings.Contains(rec.Body.String(), expected) {
			t.Fatalf("expected %s in compensation response, got %s", expected, rec.Body.String())
		}
	}

	eventsReq := httptest.NewRequest(http.MethodGet, "/api/ops/channel-ops-report-events?limit=3", nil)
	eventsRec := httptest.NewRecorder()
	mux.ServeHTTP(eventsRec, eventsReq)

	if eventsRec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", eventsRec.Code, eventsRec.Body.String())
	}
	if !strings.Contains(eventsRec.Body.String(), `"action":"COMPENSATE"`) || !strings.Contains(eventsRec.Body.String(), `"note":"补偿日报"`) {
		t.Fatalf("expected compensation event list, got %s", eventsRec.Body.String())
	}
}

func TestChannelOpsReportEventRoutesFilterAndExport(t *testing.T) {
	st := store.NewSeedStore()
	if _, err := st.RecordChannelOpsReportEvent(store.ChannelOpsReportEvent{
		Action:   "COMPENSATE",
		Actor:    "ops-lead",
		Status:   "SUCCESS",
		ReportID: "report-ok",
		Format:   "markdown",
		Pruned:   1,
		Note:     "manual success",
	}); err != nil {
		t.Fatalf("record success event: %v", err)
	}
	if _, err := st.RecordChannelOpsReportEvent(store.ChannelOpsReportEvent{
		Action: "COMPENSATE",
		Actor:  "qa-a",
		Status: "FAILED",
		Format: "csv",
		Error:  "boom",
	}); err != nil {
		t.Fatalf("record failed event: %v", err)
	}
	mux := http.NewServeMux()
	Register(mux, st)

	req := httptest.NewRequest(http.MethodGet, "/api/ops/channel-ops-report-events?status=FAILED&actor=qa", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	for _, expected := range []string{`"actor":"qa-a"`, `"status":"FAILED"`, `"error":"boom"`} {
		if !strings.Contains(rec.Body.String(), expected) {
			t.Fatalf("expected %s in filtered events, got %s", expected, rec.Body.String())
		}
	}
	if strings.Contains(rec.Body.String(), "report-ok") {
		t.Fatalf("expected filtered events to omit success event, got %s", rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/ops/channel-ops-report-events/export?status=SUCCESS&actor=ops", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Header().Get("Content-Type"), "text/csv") {
		t.Fatalf("expected csv content type, got %s", rec.Header().Get("Content-Type"))
	}
	for _, expected := range []string{"id,action,actor,status,report_id,format,pruned,note,error,created_at", "ops-lead", "manual success"} {
		if !strings.Contains(rec.Body.String(), expected) {
			t.Fatalf("expected %s in exported events, got %s", expected, rec.Body.String())
		}
	}
	if strings.Contains(rec.Body.String(), "qa-a") {
		t.Fatalf("expected exported events to omit failed event, got %s", rec.Body.String())
	}
}

func TestChannelInboundAuditRoutesFilterAndExport(t *testing.T) {
	st := store.NewSeedStore()
	if err := st.RecordChannelInboundAudit(store.ChannelInboundAudit{
		ID:                     "audit-ok",
		Channel:                "Douyin",
		ExternalConversationID: "dy-open-1",
		ExternalMessageID:      "dy-msg-1",
		Origin:                 "https://open.douyin.com",
		Status:                 "ACCEPTED",
		Code:                   "accepted",
		Reason:                 "channel inbound accepted",
		ReplayKey:              "replay-ok",
		SignaturePreview:       "abcdef123456",
		ContentHash:            "content-ok",
	}); err != nil {
		t.Fatalf("record accepted audit: %v", err)
	}
	if err := st.RecordChannelInboundAudit(store.ChannelInboundAudit{
		ID:                     "audit-bad",
		Channel:                "WeChat",
		ExternalConversationID: "wx-open-1",
		ExternalMessageID:      "wx-msg-1",
		Origin:                 "https://wechat.example.com",
		Status:                 "REJECTED",
		Code:                   "invalid_signature",
		Reason:                 "channel signature verification failed",
		ReplayKey:              "replay-bad",
		SignaturePreview:       "bad-signatur",
		ContentHash:            "content-bad",
	}); err != nil {
		t.Fatalf("record rejected audit: %v", err)
	}
	mux := http.NewServeMux()
	Register(mux, st)

	req := httptest.NewRequest(http.MethodGet, "/api/ops/channel-inbound-audits?channel=WeChat&status=REJECTED&code=signature", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	for _, expected := range []string{`"id":"audit-bad"`, `"status":"REJECTED"`, `"code":"invalid_signature"`} {
		if !strings.Contains(rec.Body.String(), expected) {
			t.Fatalf("expected %s in filtered audits, got %s", expected, rec.Body.String())
		}
	}
	if strings.Contains(rec.Body.String(), "audit-ok") {
		t.Fatalf("expected filtered audits to omit accepted audit, got %s", rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/ops/channel-inbound-audits/export?channel=Douyin&status=ACCEPTED", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Header().Get("Content-Type"), "text/csv") {
		t.Fatalf("expected csv content type, got %s", rec.Header().Get("Content-Type"))
	}
	for _, expected := range []string{"id,channel,status,code,reason,origin,external_conversation_id,external_message_id,replay_key,signature_preview,content_hash,created_at", "audit-ok", "Douyin"} {
		if !strings.Contains(rec.Body.String(), expected) {
			t.Fatalf("expected %s in exported audits, got %s", expected, rec.Body.String())
		}
	}
	if strings.Contains(rec.Body.String(), "audit-bad") {
		t.Fatalf("expected exported audits to omit rejected audit, got %s", rec.Body.String())
	}
}

func TestChannelInboundAuditQualityEventRoutesFilterAndExport(t *testing.T) {
	st := store.NewSeedStore()
	for _, audit := range []store.ChannelInboundAudit{
		{Channel: "WeChat", Status: "ACCEPTED", Code: "accepted"},
		{Channel: "WeChat", Status: "REJECTED", Code: "invalid_signature"},
		{Channel: "WeChat", Status: "REJECTED", Code: "invalid_signature"},
		{Channel: "WeChat", Status: "REJECTED", Code: "invalid_signature"},
	} {
		if err := st.RecordChannelInboundAudit(audit); err != nil {
			t.Fatalf("record audit: %v", err)
		}
	}
	mux := http.NewServeMux()
	Register(mux, st)

	req := httptest.NewRequest(http.MethodGet, "/api/ops/channel-inbound-audit-quality-events?channel=WeChat&status=ESCALATE&code=signature", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	for _, expected := range []string{`"channel":"WeChat"`, `"status":"ESCALATE"`, `"failureCode":"invalid_signature"`, `"acceptanceRate":25`} {
		if !strings.Contains(rec.Body.String(), expected) {
			t.Fatalf("expected %s in filtered audit quality events, got %s", expected, rec.Body.String())
		}
	}

	req = httptest.NewRequest(http.MethodGet, "/api/ops/channel-inbound-audit-quality-events/export?channel=WeChat&status=ESCALATE", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	for _, expected := range []string{"id,channel,severity,status,failure_code,total,accepted,rejected,acceptance_rate", "WeChat", "invalid_signature"} {
		if !strings.Contains(rec.Body.String(), expected) {
			t.Fatalf("expected %s in exported audit quality events, got %s", expected, rec.Body.String())
		}
	}
}

func generatedReportID(body string) string {
	const marker = `"id":"`
	start := strings.Index(body, marker)
	if start < 0 {
		return ""
	}
	start += len(marker)
	end := strings.Index(body[start:], `"`)
	if end < 0 {
		return ""
	}
	return body[start : start+end]
}
