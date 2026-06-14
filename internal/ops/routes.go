package ops

import (
	"net/http"
	"strconv"

	"github.com/anjing-le/agent-customer-service/internal/platform/httpjson"
	"github.com/anjing-le/agent-customer-service/internal/platform/store"
)

func Register(mux *http.ServeMux, st store.Runtime) {
	mux.HandleFunc("/api/ops/dashboard", func(w http.ResponseWriter, r *http.Request) {
		if !httpjson.RequireMethod(w, r, http.MethodGet) {
			return
		}
		dashboard, err := st.Dashboard()
		if err != nil {
			httpjson.Fail(w, http.StatusInternalServerError, "store_error", err.Error())
			return
		}
		httpjson.OK(w, dashboard)
	})

	mux.HandleFunc("/api/ops/rules/test", func(w http.ResponseWriter, r *http.Request) {
		if !httpjson.RequireMethod(w, r, http.MethodPost) {
			return
		}
		var req struct {
			Content string `json:"content"`
		}
		if err := httpjson.Decode(r, &req); err != nil {
			httpjson.BadRequest(w, err.Error())
			return
		}
		result, err := st.TestRule(req.Content)
		if err != nil {
			httpjson.Fail(w, http.StatusInternalServerError, "store_error", err.Error())
			return
		}
		httpjson.OK(w, result)
	})

	mux.HandleFunc("/api/ops/rules/compare", func(w http.ResponseWriter, r *http.Request) {
		if !httpjson.RequireMethod(w, r, http.MethodPost) {
			return
		}
		var req struct {
			Content string `json:"content"`
		}
		if err := httpjson.Decode(r, &req); err != nil {
			httpjson.BadRequest(w, err.Error())
			return
		}
		result, err := st.CompareRuleVersions(req.Content)
		if err != nil {
			httpjson.Fail(w, http.StatusInternalServerError, "store_error", err.Error())
			return
		}
		httpjson.OK(w, result)
	})

	mux.HandleFunc("/api/ops/rules/approve", func(w http.ResponseWriter, r *http.Request) {
		if !httpjson.RequireMethod(w, r, http.MethodPost) {
			return
		}
		var req struct {
			Code      string   `json:"code"`
			Approver  string   `json:"approver"`
			RiskLevel string   `json:"riskLevel"`
			SampleIDs []string `json:"sampleIds"`
			Note      string   `json:"note"`
		}
		if err := httpjson.Decode(r, &req); err != nil {
			httpjson.BadRequest(w, err.Error())
			return
		}
		if req.Code == "" {
			httpjson.BadRequest(w, "code is required")
			return
		}
		approval, err := st.SubmitRuleApproval(req.Code, req.Approver, req.RiskLevel, req.Note, req.SampleIDs)
		if err != nil {
			httpjson.Fail(w, http.StatusInternalServerError, "store_error", err.Error())
			return
		}
		httpjson.OK(w, approval)
	})

	mux.HandleFunc("/api/ops/rules/publish-canary", func(w http.ResponseWriter, r *http.Request) {
		if !httpjson.RequireMethod(w, r, http.MethodPost) {
			return
		}
		var req struct {
			Code  string `json:"code"`
			Actor string `json:"actor"`
			Note  string `json:"note"`
		}
		if err := httpjson.Decode(r, &req); err != nil {
			httpjson.BadRequest(w, err.Error())
			return
		}
		if req.Code == "" {
			httpjson.BadRequest(w, "code is required")
			return
		}
		event, err := st.PublishCanaryRule(req.Code, req.Actor, req.Note)
		if err != nil {
			httpjson.Fail(w, http.StatusInternalServerError, "store_error", err.Error())
			return
		}
		httpjson.OK(w, event)
	})

	mux.HandleFunc("/api/ops/rules/rollback", func(w http.ResponseWriter, r *http.Request) {
		if !httpjson.RequireMethod(w, r, http.MethodPost) {
			return
		}
		var req struct {
			Code  string `json:"code"`
			Actor string `json:"actor"`
			Note  string `json:"note"`
		}
		if err := httpjson.Decode(r, &req); err != nil {
			httpjson.BadRequest(w, err.Error())
			return
		}
		if req.Code == "" {
			httpjson.BadRequest(w, "code is required")
			return
		}
		event, err := st.RollbackRule(req.Code, req.Actor, req.Note)
		if err != nil {
			httpjson.Fail(w, http.StatusInternalServerError, "store_error", err.Error())
			return
		}
		httpjson.OK(w, event)
	})

	mux.HandleFunc("/api/ops/transfers/resolve", func(w http.ResponseWriter, r *http.Request) {
		if !httpjson.RequireMethod(w, r, http.MethodPost) {
			return
		}
		var req struct {
			ID       string `json:"id"`
			Assignee string `json:"assignee"`
			Note     string `json:"note"`
		}
		if err := httpjson.Decode(r, &req); err != nil {
			httpjson.BadRequest(w, err.Error())
			return
		}
		if req.ID == "" {
			httpjson.BadRequest(w, "id is required")
			return
		}
		ticket, err := st.ResolveTransferTicket(req.ID, req.Assignee, req.Note)
		if err != nil {
			httpjson.Fail(w, http.StatusInternalServerError, "store_error", err.Error())
			return
		}
		httpjson.OK(w, ticket)
	})

	mux.HandleFunc("/api/ops/review-tasks/assign", func(w http.ResponseWriter, r *http.Request) {
		if !httpjson.RequireMethod(w, r, http.MethodPost) {
			return
		}
		var req struct {
			ID       string `json:"id"`
			Assignee string `json:"assignee"`
		}
		if err := httpjson.Decode(r, &req); err != nil {
			httpjson.BadRequest(w, err.Error())
			return
		}
		if req.ID == "" {
			httpjson.BadRequest(w, "id is required")
			return
		}
		task, err := st.AssignReviewTask(req.ID, req.Assignee)
		if err != nil {
			httpjson.Fail(w, http.StatusInternalServerError, "store_error", err.Error())
			return
		}
		httpjson.OK(w, task)
	})

	mux.HandleFunc("/api/ops/review-tasks/complete", func(w http.ResponseWriter, r *http.Request) {
		if !httpjson.RequireMethod(w, r, http.MethodPost) {
			return
		}
		var req struct {
			ID string `json:"id"`
		}
		if err := httpjson.Decode(r, &req); err != nil {
			httpjson.BadRequest(w, err.Error())
			return
		}
		if req.ID == "" {
			httpjson.BadRequest(w, "id is required")
			return
		}
		task, err := st.CompleteReviewTask(req.ID)
		if err != nil {
			httpjson.Fail(w, http.StatusInternalServerError, "store_error", err.Error())
			return
		}
		httpjson.OK(w, task)
	})

	mux.HandleFunc("/api/ops/channel-notifications/ack", func(w http.ResponseWriter, r *http.Request) {
		if !httpjson.RequireMethod(w, r, http.MethodPost) {
			return
		}
		var req struct {
			ID    string `json:"id"`
			Actor string `json:"actor"`
			Note  string `json:"note"`
		}
		if err := httpjson.Decode(r, &req); err != nil {
			httpjson.BadRequest(w, err.Error())
			return
		}
		if req.ID == "" {
			httpjson.BadRequest(w, "id is required")
			return
		}
		notification, err := st.AcknowledgeChannelNotification(req.ID, req.Actor, req.Note)
		if err != nil {
			httpjson.Fail(w, http.StatusInternalServerError, "store_error", err.Error())
			return
		}
		httpjson.OK(w, notification)
	})

	mux.HandleFunc("/api/ops/channel-notifications/dispatch", func(w http.ResponseWriter, r *http.Request) {
		if !httpjson.RequireMethod(w, r, http.MethodPost) {
			return
		}
		var req struct {
			ID      string `json:"id"`
			Outcome string `json:"outcome"`
		}
		if err := httpjson.Decode(r, &req); err != nil {
			httpjson.BadRequest(w, err.Error())
			return
		}
		if req.ID == "" {
			httpjson.BadRequest(w, "id is required")
			return
		}
		notification, err := st.DispatchChannelNotification(req.ID, req.Outcome)
		if err != nil {
			httpjson.Fail(w, http.StatusInternalServerError, "store_error", err.Error())
			return
		}
		httpjson.OK(w, notification)
	})

	mux.HandleFunc("/api/ops/channel-alert-policies/update", func(w http.ResponseWriter, r *http.Request) {
		if !httpjson.RequireMethod(w, r, http.MethodPost) {
			return
		}
		var req struct {
			Channel        string `json:"channel"`
			TargetURL      string `json:"targetUrl"`
			SecretRef      string `json:"secretRef"`
			MaxAttempts    int    `json:"maxAttempts"`
			BackoffSeconds int    `json:"backoffSeconds"`
			Actor          string `json:"actor"`
			Note           string `json:"note"`
		}
		if err := httpjson.Decode(r, &req); err != nil {
			httpjson.BadRequest(w, err.Error())
			return
		}
		if req.Channel == "" {
			httpjson.BadRequest(w, "channel is required")
			return
		}
		policy, err := st.UpdateChannelAlertPolicy(req.Channel, req.TargetURL, req.SecretRef, req.MaxAttempts, req.BackoffSeconds, req.Actor, req.Note)
		if err != nil {
			httpjson.Fail(w, http.StatusInternalServerError, "store_error", err.Error())
			return
		}
		httpjson.OK(w, policy)
	})

	mux.HandleFunc("/api/ops/channel-alert-policies/approve-change", func(w http.ResponseWriter, r *http.Request) {
		if !httpjson.RequireMethod(w, r, http.MethodPost) {
			return
		}
		var req struct {
			ID           string `json:"id"`
			Approver     string `json:"approver"`
			Note         string `json:"note"`
			Confirmation string `json:"confirmation"`
		}
		if err := httpjson.Decode(r, &req); err != nil {
			httpjson.BadRequest(w, err.Error())
			return
		}
		if req.ID == "" {
			httpjson.BadRequest(w, "id is required")
			return
		}
		policy, err := st.ApproveNotificationPolicyChange(req.ID, req.Approver, req.Note, req.Confirmation)
		if err != nil {
			httpjson.Fail(w, http.StatusInternalServerError, "store_error", err.Error())
			return
		}
		httpjson.OK(w, policy)
	})

	mux.HandleFunc("/api/ops/channel-alert-policies/reject-change", func(w http.ResponseWriter, r *http.Request) {
		if !httpjson.RequireMethod(w, r, http.MethodPost) {
			return
		}
		var req struct {
			ID       string `json:"id"`
			Reviewer string `json:"reviewer"`
			Note     string `json:"note"`
		}
		if err := httpjson.Decode(r, &req); err != nil {
			httpjson.BadRequest(w, err.Error())
			return
		}
		if req.ID == "" {
			httpjson.BadRequest(w, "id is required")
			return
		}
		change, err := st.RejectNotificationPolicyChange(req.ID, req.Reviewer, req.Note)
		if err != nil {
			httpjson.Fail(w, http.StatusInternalServerError, "store_error", err.Error())
			return
		}
		httpjson.OK(w, change)
	})

	mux.HandleFunc("/api/ops/channel-alert-policies/cancel-change", func(w http.ResponseWriter, r *http.Request) {
		if !httpjson.RequireMethod(w, r, http.MethodPost) {
			return
		}
		var req struct {
			ID    string `json:"id"`
			Actor string `json:"actor"`
			Note  string `json:"note"`
		}
		if err := httpjson.Decode(r, &req); err != nil {
			httpjson.BadRequest(w, err.Error())
			return
		}
		if req.ID == "" {
			httpjson.BadRequest(w, "id is required")
			return
		}
		change, err := st.CancelNotificationPolicyChange(req.ID, req.Actor, req.Note)
		if err != nil {
			httpjson.Fail(w, http.StatusInternalServerError, "store_error", err.Error())
			return
		}
		httpjson.OK(w, change)
	})

	mux.HandleFunc("/api/ops/channel-alert-policies/rollback", func(w http.ResponseWriter, r *http.Request) {
		if !httpjson.RequireMethod(w, r, http.MethodPost) {
			return
		}
		var req struct {
			Channel      string `json:"channel"`
			Actor        string `json:"actor"`
			Note         string `json:"note"`
			Confirmation string `json:"confirmation"`
		}
		if err := httpjson.Decode(r, &req); err != nil {
			httpjson.BadRequest(w, err.Error())
			return
		}
		if req.Channel == "" {
			httpjson.BadRequest(w, "channel is required")
			return
		}
		policy, err := st.RollbackChannelAlertPolicy(req.Channel, req.Actor, req.Note, req.Confirmation)
		if err != nil {
			httpjson.Fail(w, http.StatusInternalServerError, "store_error", err.Error())
			return
		}
		httpjson.OK(w, policy)
	})

	mux.HandleFunc("/api/ops/annotations/submit", func(w http.ResponseWriter, r *http.Request) {
		if !httpjson.RequireMethod(w, r, http.MethodPost) {
			return
		}
		var req struct {
			MessageID  string                     `json:"messageId"`
			Reviewer   string                     `json:"reviewer"`
			Verdict    string                     `json:"verdict"`
			Note       string                     `json:"note"`
			Dimensions store.AnnotationDimensions `json:"dimensions"`
			Tags       []string                   `json:"tags"`
		}
		if err := httpjson.Decode(r, &req); err != nil {
			httpjson.BadRequest(w, err.Error())
			return
		}
		if req.MessageID == "" {
			httpjson.BadRequest(w, "messageId is required")
			return
		}
		annotation, err := st.SubmitAnnotation(req.MessageID, req.Reviewer, req.Verdict, req.Note, req.Dimensions, req.Tags)
		if err != nil {
			httpjson.Fail(w, http.StatusInternalServerError, "store_error", err.Error())
			return
		}
		httpjson.Created(w, annotation)
	})

	mux.HandleFunc("/api/ops/training-samples/export", func(w http.ResponseWriter, r *http.Request) {
		if !httpjson.RequireMethod(w, r, http.MethodGet) {
			return
		}
		maxScore := 80
		if raw := r.URL.Query().Get("maxScore"); raw != "" {
			parsed, err := strconv.Atoi(raw)
			if err != nil {
				httpjson.BadRequest(w, "maxScore must be a number")
				return
			}
			maxScore = parsed
		}
		samples, err := st.ExportTrainingSamples(maxScore)
		if err != nil {
			httpjson.Fail(w, http.StatusInternalServerError, "store_error", err.Error())
			return
		}
		httpjson.OK(w, samples)
	})
}
