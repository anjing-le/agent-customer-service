package ops

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/anjing-le/agent-customer-service/internal/platform/httpjson"
	"github.com/anjing-le/agent-customer-service/internal/platform/store"
)

func Register(mux *http.ServeMux, st store.Runtime) {
	registerRoutes(mux, st, nil)
}

func RegisterWithReportScheduler(mux *http.ServeMux, st store.Runtime, scheduler *ReportScheduler) {
	registerRoutes(mux, st, scheduler)
}

func registerRoutes(mux *http.ServeMux, st store.Runtime, scheduler *ReportScheduler) {
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

	mux.HandleFunc("/api/ops/channel-ops-report-scheduler", func(w http.ResponseWriter, r *http.Request) {
		if !httpjson.RequireMethod(w, r, http.MethodGet) {
			return
		}
		if scheduler == nil {
			httpjson.OK(w, initialReportSchedulerStatus(ReportSchedulerConfig{}))
			return
		}
		httpjson.OK(w, scheduler.Status())
	})

	mux.HandleFunc("/api/ops/channel-ops-report-events", func(w http.ResponseWriter, r *http.Request) {
		if !httpjson.RequireMethod(w, r, http.MethodGet) {
			return
		}
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		events, err := st.ListChannelOpsReportEvents(limit)
		if err != nil {
			httpjson.Fail(w, http.StatusInternalServerError, "store_error", err.Error())
			return
		}
		httpjson.OK(w, filterChannelOpsReportEvents(events, r.URL.Query().Get("status"), r.URL.Query().Get("actor")))
	})

	mux.HandleFunc("/api/ops/channel-ops-report-events/export", func(w http.ResponseWriter, r *http.Request) {
		if !httpjson.RequireMethod(w, r, http.MethodGet) {
			return
		}
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if limit <= 0 {
			limit = 50
		}
		events, err := st.ListChannelOpsReportEvents(limit)
		if err != nil {
			httpjson.Fail(w, http.StatusInternalServerError, "store_error", err.Error())
			return
		}
		report, err := renderChannelOpsReportEventsCSV(filterChannelOpsReportEvents(events, r.URL.Query().Get("status"), r.URL.Query().Get("actor")))
		if err != nil {
			httpjson.Fail(w, http.StatusInternalServerError, "report_error", err.Error())
			return
		}
		w.Header().Set("Content-Type", "text/csv; charset=utf-8")
		w.Header().Set("Content-Disposition", `attachment; filename="agent-customer-service-channel-ops-events.csv"`)
		_, _ = w.Write(report)
	})

	mux.HandleFunc("/api/ops/channel-inbound-audits", func(w http.ResponseWriter, r *http.Request) {
		if !httpjson.RequireMethod(w, r, http.MethodGet) {
			return
		}
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		audits, err := st.ListChannelInboundAudits(limit)
		if err != nil {
			httpjson.Fail(w, http.StatusInternalServerError, "store_error", err.Error())
			return
		}
		httpjson.OK(w, filterChannelInboundAudits(audits, r.URL.Query().Get("channel"), r.URL.Query().Get("status"), r.URL.Query().Get("code")))
	})

	mux.HandleFunc("/api/ops/channel-inbound-audit-quality-events", func(w http.ResponseWriter, r *http.Request) {
		if !httpjson.RequireMethod(w, r, http.MethodGet) {
			return
		}
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		events, err := st.ListChannelInboundAuditQualityEvents(limit)
		if err != nil {
			httpjson.Fail(w, http.StatusInternalServerError, "store_error", err.Error())
			return
		}
		httpjson.OK(w, filterChannelInboundAuditQualityEvents(events, r.URL.Query().Get("channel"), r.URL.Query().Get("status"), r.URL.Query().Get("code")))
	})

	mux.HandleFunc("/api/ops/channel-inbound-audit-quality-events/export", func(w http.ResponseWriter, r *http.Request) {
		if !httpjson.RequireMethod(w, r, http.MethodGet) {
			return
		}
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if limit <= 0 {
			limit = 100
		}
		events, err := st.ListChannelInboundAuditQualityEvents(limit)
		if err != nil {
			httpjson.Fail(w, http.StatusInternalServerError, "store_error", err.Error())
			return
		}
		report, err := renderChannelInboundAuditQualityEventsCSV(filterChannelInboundAuditQualityEvents(events, r.URL.Query().Get("channel"), r.URL.Query().Get("status"), r.URL.Query().Get("code")))
		if err != nil {
			httpjson.Fail(w, http.StatusInternalServerError, "report_error", err.Error())
			return
		}
		w.Header().Set("Content-Type", "text/csv; charset=utf-8")
		w.Header().Set("Content-Disposition", `attachment; filename="agent-customer-service-channel-audit-quality-events.csv"`)
		_, _ = w.Write(report)
	})

	mux.HandleFunc("/api/ops/channel-inbound-audits/export", func(w http.ResponseWriter, r *http.Request) {
		if !httpjson.RequireMethod(w, r, http.MethodGet) {
			return
		}
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if limit <= 0 {
			limit = 100
		}
		audits, err := st.ListChannelInboundAudits(limit)
		if err != nil {
			httpjson.Fail(w, http.StatusInternalServerError, "store_error", err.Error())
			return
		}
		report, err := renderChannelInboundAuditsCSV(filterChannelInboundAudits(audits, r.URL.Query().Get("channel"), r.URL.Query().Get("status"), r.URL.Query().Get("code")))
		if err != nil {
			httpjson.Fail(w, http.StatusInternalServerError, "report_error", err.Error())
			return
		}
		w.Header().Set("Content-Type", "text/csv; charset=utf-8")
		w.Header().Set("Content-Disposition", `attachment; filename="agent-customer-service-channel-inbound-audits.csv"`)
		_, _ = w.Write(report)
	})

	mux.HandleFunc("/api/ops/channel-ops-report-scheduler/compensate", func(w http.ResponseWriter, r *http.Request) {
		if !httpjson.RequireMethod(w, r, http.MethodPost) {
			return
		}
		var req struct {
			Actor string `json:"actor"`
			Note  string `json:"note"`
		}
		if err := httpjson.Decode(r, &req); err != nil {
			httpjson.BadRequest(w, err.Error())
			return
		}
		activeScheduler := scheduler
		if activeScheduler == nil {
			activeScheduler = NewReportScheduler(st, ReportSchedulerConfig{}, nil)
		}
		result, err := activeScheduler.Compensate(context.Background(), req.Actor, req.Note)
		if err != nil && result.Event.ID == "" {
			httpjson.Fail(w, http.StatusInternalServerError, "store_error", err.Error())
			return
		}
		httpjson.OK(w, result)
	})

	mux.HandleFunc("/api/ops/channel-ops-report/export", func(w http.ResponseWriter, r *http.Request) {
		if !httpjson.RequireMethod(w, r, http.MethodGet) {
			return
		}
		dashboard, err := st.Dashboard()
		if err != nil {
			httpjson.Fail(w, http.StatusInternalServerError, "store_error", err.Error())
			return
		}
		format := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("format")))
		if format == "" {
			format = "markdown"
		}
		switch format {
		case "markdown", "md":
			report := renderChannelOpsReportMarkdown(dashboard, time.Now().UTC())
			w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
			w.Header().Set("Content-Disposition", `attachment; filename="agent-customer-service-channel-ops.md"`)
			_, _ = w.Write([]byte(report))
		case "csv":
			report, err := renderChannelOpsReportCSV(dashboard)
			if err != nil {
				httpjson.Fail(w, http.StatusInternalServerError, "report_error", err.Error())
				return
			}
			w.Header().Set("Content-Type", "text/csv; charset=utf-8")
			w.Header().Set("Content-Disposition", `attachment; filename="agent-customer-service-channel-ops.csv"`)
			_, _ = w.Write(report)
		default:
			httpjson.BadRequest(w, "format must be markdown or csv")
		}
	})

	mux.HandleFunc("/api/ops/channel-ops-reports", func(w http.ResponseWriter, r *http.Request) {
		if !httpjson.RequireMethod(w, r, http.MethodGet) {
			return
		}
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		reports, err := st.ListChannelOpsReports(limit)
		if err != nil {
			httpjson.Fail(w, http.StatusInternalServerError, "store_error", err.Error())
			return
		}
		httpjson.OK(w, reports)
	})

	mux.HandleFunc("/api/ops/channel-ops-reports/generate", func(w http.ResponseWriter, r *http.Request) {
		if !httpjson.RequireMethod(w, r, http.MethodPost) {
			return
		}
		var req struct {
			Format string `json:"format"`
		}
		if err := httpjson.Decode(r, &req); err != nil {
			httpjson.BadRequest(w, err.Error())
			return
		}
		dashboard, err := st.Dashboard()
		if err != nil {
			httpjson.Fail(w, http.StatusInternalServerError, "store_error", err.Error())
			return
		}
		report, err := buildChannelOpsReport(dashboard, req.Format, time.Now().UTC())
		if err != nil {
			httpjson.BadRequest(w, err.Error())
			return
		}
		report, err = st.SaveChannelOpsReport(report)
		if err != nil {
			httpjson.Fail(w, http.StatusInternalServerError, "store_error", err.Error())
			return
		}
		httpjson.Created(w, report)
	})

	mux.HandleFunc("/api/ops/channel-ops-reports/export", func(w http.ResponseWriter, r *http.Request) {
		if !httpjson.RequireMethod(w, r, http.MethodGet) {
			return
		}
		id := strings.TrimSpace(r.URL.Query().Get("id"))
		if id == "" {
			httpjson.BadRequest(w, "id is required")
			return
		}
		report, err := st.ChannelOpsReport(id)
		if err != nil {
			httpjson.Fail(w, http.StatusNotFound, "not_found", err.Error())
			return
		}
		writeChannelOpsReport(w, report)
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

	mux.HandleFunc("/api/ops/channel-runbook-checks", func(w http.ResponseWriter, r *http.Request) {
		if !httpjson.RequireMethod(w, r, http.MethodGet) {
			return
		}
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		checks, err := st.ListChannelRunbookChecks(limit)
		if err != nil {
			httpjson.Fail(w, http.StatusInternalServerError, "store_error", err.Error())
			return
		}
		httpjson.OK(w, filterChannelRunbookChecks(
			checks,
			r.URL.Query().Get("channel"),
			r.URL.Query().Get("status"),
			r.URL.Query().Get("checkStatus"),
			r.URL.Query().Get("actor"),
			r.URL.Query().Get("assignee"),
			r.URL.Query().Get("actionRef"),
			r.URL.Query().Get("overdue"),
		))
	})

	mux.HandleFunc("/api/ops/channel-runbook-checks/export", func(w http.ResponseWriter, r *http.Request) {
		if !httpjson.RequireMethod(w, r, http.MethodGet) {
			return
		}
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if limit <= 0 {
			limit = 100
		}
		checks, err := st.ListChannelRunbookChecks(limit)
		if err != nil {
			httpjson.Fail(w, http.StatusInternalServerError, "store_error", err.Error())
			return
		}
		report, err := renderChannelRunbookChecksCSV(filterChannelRunbookChecks(
			checks,
			r.URL.Query().Get("channel"),
			r.URL.Query().Get("status"),
			r.URL.Query().Get("checkStatus"),
			r.URL.Query().Get("actor"),
			r.URL.Query().Get("assignee"),
			r.URL.Query().Get("actionRef"),
			r.URL.Query().Get("overdue"),
		))
		if err != nil {
			httpjson.Fail(w, http.StatusInternalServerError, "report_error", err.Error())
			return
		}
		w.Header().Set("Content-Type", "text/csv; charset=utf-8")
		w.Header().Set("Content-Disposition", `attachment; filename="agent-customer-service-channel-runbook-checks.csv"`)
		_, _ = w.Write(report)
	})

	mux.HandleFunc("/api/ops/channel-runbook-checks/assign", func(w http.ResponseWriter, r *http.Request) {
		handleAssignChannelRunbookChecks(w, r, st)
	})

	mux.HandleFunc("/api/ops/channel-runbook-checks/complete", func(w http.ResponseWriter, r *http.Request) {
		handleChannelRunbookCheckStatus(w, r, st, "DONE", "Runbook check completed")
	})

	mux.HandleFunc("/api/ops/channel-runbook-checks/block", func(w http.ResponseWriter, r *http.Request) {
		handleChannelRunbookCheckStatus(w, r, st, "BLOCKED", "Runbook check blocked")
	})

	mux.HandleFunc("/api/ops/channel-runbook-checks/recover", func(w http.ResponseWriter, r *http.Request) {
		handleChannelRunbookCheckStatus(w, r, st, "DONE", "Runbook check recovered")
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
			Channel                       string `json:"channel"`
			TargetURL                     string `json:"targetUrl"`
			SecretRef                     string `json:"secretRef"`
			MaxAttempts                   int    `json:"maxAttempts"`
			BackoffSeconds                int    `json:"backoffSeconds"`
			InboundAuditMinSamples        int    `json:"inboundAuditMinSamples"`
			InboundAuditMinAcceptanceRate int    `json:"inboundAuditMinAcceptanceRate"`
			InboundAuditMaxErrorCount     int    `json:"inboundAuditMaxErrorCount"`
			Actor                         string `json:"actor"`
			Note                          string `json:"note"`
		}
		if err := httpjson.Decode(r, &req); err != nil {
			httpjson.BadRequest(w, err.Error())
			return
		}
		if req.Channel == "" {
			httpjson.BadRequest(w, "channel is required")
			return
		}
		policy, err := st.UpdateChannelAlertPolicy(req.Channel, req.TargetURL, req.SecretRef, req.MaxAttempts, req.BackoffSeconds, req.InboundAuditMinSamples, req.InboundAuditMinAcceptanceRate, req.InboundAuditMaxErrorCount, req.Actor, req.Note)
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

func handleChannelRunbookCheckStatus(w http.ResponseWriter, r *http.Request, st store.Runtime, checkStatus string, defaultNote string) {
	if !httpjson.RequireMethod(w, r, http.MethodPost) {
		return
	}
	var req struct {
		Channel       string `json:"channel"`
		RunbookStatus string `json:"runbookStatus"`
		Step          string `json:"step"`
		StepIndex     int    `json:"stepIndex"`
		ActionRef     string `json:"actionRef"`
		ReportID      string `json:"reportId"`
		Assignee      string `json:"assignee"`
		DueAt         string `json:"dueAt"`
		Actor         string `json:"actor"`
		Note          string `json:"note"`
	}
	if err := httpjson.Decode(r, &req); err != nil {
		httpjson.BadRequest(w, err.Error())
		return
	}
	if strings.TrimSpace(req.Channel) == "" || strings.TrimSpace(req.Step) == "" {
		httpjson.BadRequest(w, "channel and step are required")
		return
	}
	check, err := st.CompleteChannelRunbookCheck(store.ChannelRunbookCheck{
		Channel:       req.Channel,
		RunbookStatus: req.RunbookStatus,
		CheckStatus:   checkStatus,
		Step:          req.Step,
		StepIndex:     req.StepIndex,
		ActionRef:     req.ActionRef,
		ReportID:      req.ReportID,
		Assignee:      req.Assignee,
		DueAt:         req.DueAt,
		Actor:         req.Actor,
		Note:          fallbackString(req.Note, defaultNote),
	})
	if err != nil {
		httpjson.Fail(w, http.StatusInternalServerError, "store_error", err.Error())
		return
	}
	httpjson.OK(w, check)
}

func handleAssignChannelRunbookChecks(w http.ResponseWriter, r *http.Request, st store.Runtime) {
	if !httpjson.RequireMethod(w, r, http.MethodPost) {
		return
	}
	var req struct {
		Channel       string `json:"channel"`
		RunbookStatus string `json:"runbookStatus"`
		Assignee      string `json:"assignee"`
		DueAt         string `json:"dueAt"`
		Actor         string `json:"actor"`
		Note          string `json:"note"`
		ReportID      string `json:"reportId"`
		ActionRef     string `json:"actionRef"`
		StepIndexes   []int  `json:"stepIndexes"`
	}
	if err := httpjson.Decode(r, &req); err != nil {
		httpjson.BadRequest(w, err.Error())
		return
	}
	if strings.TrimSpace(req.Channel) == "" || strings.TrimSpace(req.RunbookStatus) == "" || strings.TrimSpace(req.Assignee) == "" {
		httpjson.BadRequest(w, "channel, runbookStatus and assignee are required")
		return
	}
	dashboard, err := st.Dashboard()
	if err != nil {
		httpjson.Fail(w, http.StatusInternalServerError, "store_error", err.Error())
		return
	}
	runbook, ok := channelRunbookForAssignment(dashboard.ChannelRunbooks, req.Channel, req.RunbookStatus)
	if !ok {
		httpjson.Fail(w, http.StatusNotFound, "not_found", "channel runbook not found")
		return
	}
	selected := runbookAssignmentStepIndexes(runbook, req.StepIndexes)
	existingByStep := map[int]store.ChannelRunbookCheck{}
	for _, check := range runbook.Checks {
		existingByStep[check.StepIndex] = check
	}
	assigned := make([]store.ChannelRunbookCheck, 0, len(selected))
	skipped := 0
	for _, stepIndex := range selected {
		if stepIndex < 0 || stepIndex >= len(runbook.Steps) {
			skipped++
			continue
		}
		existing := existingByStep[stepIndex]
		if strings.EqualFold(existing.CheckStatus, "DONE") {
			skipped++
			continue
		}
		checkStatus := fallbackString(existing.CheckStatus, "TODO")
		actionRef := fallbackString(existing.ActionRef, fallbackString(req.ActionRef, runbook.Channel+":"+runbook.Status))
		dueAt := fallbackString(req.DueAt, existing.DueAt)
		reportID := fallbackString(req.ReportID, existing.ReportID)
		check, err := st.CompleteChannelRunbookCheck(store.ChannelRunbookCheck{
			Channel:       runbook.Channel,
			RunbookStatus: runbook.Status,
			CheckStatus:   checkStatus,
			Step:          runbook.Steps[stepIndex],
			StepIndex:     stepIndex,
			ActionRef:     actionRef,
			ReportID:      reportID,
			Assignee:      req.Assignee,
			DueAt:         dueAt,
			Actor:         req.Actor,
			Note:          fallbackString(req.Note, "Runbook check assigned"),
		})
		if err != nil {
			httpjson.Fail(w, http.StatusInternalServerError, "store_error", err.Error())
			return
		}
		assigned = append(assigned, check)
	}
	httpjson.OK(w, struct {
		Assigned int                         `json:"assigned"`
		Skipped  int                         `json:"skipped"`
		Checks   []store.ChannelRunbookCheck `json:"checks"`
	}{
		Assigned: len(assigned),
		Skipped:  skipped,
		Checks:   assigned,
	})
}

func channelRunbookForAssignment(runbooks []store.ChannelRunbook, channel string, status string) (store.ChannelRunbook, bool) {
	for _, runbook := range runbooks {
		if strings.EqualFold(runbook.Channel, channel) && strings.EqualFold(runbook.Status, status) {
			return runbook, true
		}
	}
	return store.ChannelRunbook{}, false
}

func runbookAssignmentStepIndexes(runbook store.ChannelRunbook, requested []int) []int {
	if len(requested) > 0 {
		return append([]int(nil), requested...)
	}
	indexes := make([]int, 0, len(runbook.Steps))
	for idx := range runbook.Steps {
		indexes = append(indexes, idx)
	}
	return indexes
}

func fallbackString(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func filterChannelOpsReportEvents(events []store.ChannelOpsReportEvent, status string, actor string) []store.ChannelOpsReportEvent {
	status = strings.ToUpper(strings.TrimSpace(status))
	actor = strings.ToLower(strings.TrimSpace(actor))
	if status == "ALL" {
		status = ""
	}
	filtered := make([]store.ChannelOpsReportEvent, 0, len(events))
	for _, event := range events {
		if status != "" && !strings.EqualFold(event.Status, status) {
			continue
		}
		if actor != "" && !strings.Contains(strings.ToLower(event.Actor), actor) {
			continue
		}
		filtered = append(filtered, event)
	}
	return filtered
}

func renderChannelOpsReportEventsCSV(events []store.ChannelOpsReportEvent) ([]byte, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	if err := writer.Write([]string{"id", "action", "actor", "status", "report_id", "format", "pruned", "note", "error", "created_at"}); err != nil {
		return nil, err
	}
	for _, event := range events {
		if err := writer.Write([]string{
			event.ID,
			event.Action,
			event.Actor,
			event.Status,
			event.ReportID,
			event.Format,
			fmt.Sprintf("%d", event.Pruned),
			event.Note,
			event.Error,
			event.CreatedAt,
		}); err != nil {
			return nil, err
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func filterChannelRunbookChecks(checks []store.ChannelRunbookCheck, channel string, status string, checkStatus string, actor string, assignee string, actionRef string, overdue string) []store.ChannelRunbookCheck {
	channel = strings.ToLower(strings.TrimSpace(channel))
	status = strings.ToUpper(strings.TrimSpace(status))
	checkStatus = strings.ToUpper(strings.TrimSpace(checkStatus))
	actor = strings.ToLower(strings.TrimSpace(actor))
	assignee = strings.ToLower(strings.TrimSpace(assignee))
	actionRef = strings.ToLower(strings.TrimSpace(actionRef))
	overdueValue, filterOverdue := optionalBoolQuery(overdue)
	if channel == "all" {
		channel = ""
	}
	if status == "ALL" {
		status = ""
	}
	if checkStatus == "ALL" {
		checkStatus = ""
	}
	filtered := make([]store.ChannelRunbookCheck, 0, len(checks))
	for _, check := range checks {
		if channel != "" && !strings.EqualFold(check.Channel, channel) {
			continue
		}
		if status != "" && !strings.EqualFold(check.RunbookStatus, status) {
			continue
		}
		if checkStatus != "" && !strings.EqualFold(check.CheckStatus, checkStatus) {
			continue
		}
		if actor != "" && !strings.Contains(strings.ToLower(check.Actor), actor) {
			continue
		}
		if assignee != "" && !strings.Contains(strings.ToLower(check.Assignee), assignee) {
			continue
		}
		if actionRef != "" && !strings.Contains(strings.ToLower(check.ActionRef), actionRef) {
			continue
		}
		if filterOverdue && store.ChannelRunbookCheckOverdue(check, time.Now().UTC()) != overdueValue {
			continue
		}
		filtered = append(filtered, check)
	}
	return filtered
}

func optionalBoolQuery(value string) (bool, bool) {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" || value == "all" {
		return false, false
	}
	if parsed, err := strconv.ParseBool(value); err == nil {
		return parsed, true
	}
	switch value {
	case "yes", "y", "overdue":
		return true, true
	case "no", "n":
		return false, true
	default:
		return false, false
	}
}

func renderChannelRunbookChecksCSV(checks []store.ChannelRunbookCheck) ([]byte, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	if err := writer.Write([]string{"id", "channel", "runbook_status", "check_status", "step_index", "step", "action_ref", "report_id", "assignee", "due_at", "overdue", "actor", "note", "completed_at"}); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	for _, check := range checks {
		if err := writer.Write([]string{
			check.ID,
			check.Channel,
			check.RunbookStatus,
			check.CheckStatus,
			fmt.Sprintf("%d", check.StepIndex),
			check.Step,
			check.ActionRef,
			check.ReportID,
			check.Assignee,
			check.DueAt,
			fmt.Sprintf("%t", store.ChannelRunbookCheckOverdue(check, now)),
			check.Actor,
			check.Note,
			check.CompletedAt,
		}); err != nil {
			return nil, err
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func filterChannelInboundAuditQualityEvents(events []store.ChannelInboundAuditQualityEvent, channel string, status string, code string) []store.ChannelInboundAuditQualityEvent {
	channel = strings.ToLower(strings.TrimSpace(channel))
	status = strings.ToUpper(strings.TrimSpace(status))
	code = strings.ToLower(strings.TrimSpace(code))
	if channel == "all" {
		channel = ""
	}
	if status == "ALL" {
		status = ""
	}
	if code == "all" {
		code = ""
	}
	filtered := make([]store.ChannelInboundAuditQualityEvent, 0, len(events))
	for _, event := range events {
		if channel != "" && !strings.EqualFold(event.Channel, channel) {
			continue
		}
		if status != "" && !strings.EqualFold(event.Status, status) {
			continue
		}
		if code != "" && !strings.Contains(strings.ToLower(event.FailureCode), code) {
			continue
		}
		filtered = append(filtered, event)
	}
	return filtered
}

func renderChannelInboundAuditQualityEventsCSV(events []store.ChannelInboundAuditQualityEvent) ([]byte, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	if err := writer.Write([]string{"id", "channel", "severity", "status", "failure_code", "total", "accepted", "rejected", "acceptance_rate", "min_samples", "min_acceptance_rate", "max_error_count", "reason", "created_at"}); err != nil {
		return nil, err
	}
	for _, event := range events {
		if err := writer.Write([]string{
			event.ID,
			event.Channel,
			event.Severity,
			event.Status,
			event.FailureCode,
			fmt.Sprintf("%d", event.Total),
			fmt.Sprintf("%d", event.Accepted),
			fmt.Sprintf("%d", event.Rejected),
			fmt.Sprintf("%d", event.AcceptanceRate),
			fmt.Sprintf("%d", event.MinSamples),
			fmt.Sprintf("%d", event.MinAcceptanceRate),
			fmt.Sprintf("%d", event.MaxErrorCount),
			event.Reason,
			event.CreatedAt,
		}); err != nil {
			return nil, err
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func filterChannelInboundAudits(audits []store.ChannelInboundAudit, channel string, status string, code string) []store.ChannelInboundAudit {
	channel = strings.ToLower(strings.TrimSpace(channel))
	status = strings.ToUpper(strings.TrimSpace(status))
	code = strings.ToLower(strings.TrimSpace(code))
	if channel == "all" {
		channel = ""
	}
	if status == "ALL" {
		status = ""
	}
	if code == "all" {
		code = ""
	}
	filtered := make([]store.ChannelInboundAudit, 0, len(audits))
	for _, audit := range audits {
		if channel != "" && !strings.EqualFold(audit.Channel, channel) {
			continue
		}
		if status != "" && !strings.EqualFold(audit.Status, status) {
			continue
		}
		if code != "" && !strings.Contains(strings.ToLower(audit.Code), code) {
			continue
		}
		filtered = append(filtered, audit)
	}
	return filtered
}

func renderChannelInboundAuditsCSV(audits []store.ChannelInboundAudit) ([]byte, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	if err := writer.Write([]string{"id", "channel", "status", "code", "reason", "origin", "external_conversation_id", "external_message_id", "replay_key", "signature_preview", "content_hash", "created_at"}); err != nil {
		return nil, err
	}
	for _, audit := range audits {
		if err := writer.Write([]string{
			audit.ID,
			audit.Channel,
			audit.Status,
			audit.Code,
			audit.Reason,
			audit.Origin,
			audit.ExternalConversationID,
			audit.ExternalMessageID,
			audit.ReplayKey,
			audit.SignaturePreview,
			audit.ContentHash,
			audit.CreatedAt,
		}); err != nil {
			return nil, err
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func buildChannelOpsReport(dashboard store.Dashboard, format string, generatedAt time.Time) (store.ChannelOpsReport, error) {
	format = strings.ToLower(strings.TrimSpace(format))
	if format == "" {
		format = "markdown"
	}
	if format == "md" {
		format = "markdown"
	}
	report := store.ChannelOpsReport{
		ID:          fmt.Sprintf("channel_ops_%d", generatedAt.UnixNano()),
		Format:      format,
		Summary:     channelOpsReportSummary(dashboard),
		GeneratedAt: generatedAt.UTC().Format(time.RFC3339),
	}
	switch format {
	case "markdown":
		report.ContentType = "text/markdown; charset=utf-8"
		report.Content = renderChannelOpsReportMarkdown(dashboard, generatedAt)
	case "csv":
		content, err := renderChannelOpsReportCSV(dashboard)
		if err != nil {
			return store.ChannelOpsReport{}, err
		}
		report.ContentType = "text/csv; charset=utf-8"
		report.Content = string(content)
	default:
		return store.ChannelOpsReport{}, fmt.Errorf("format must be markdown or csv")
	}
	return report, nil
}

func writeChannelOpsReport(w http.ResponseWriter, report store.ChannelOpsReport) {
	contentType := strings.TrimSpace(report.ContentType)
	if contentType == "" {
		contentType = "text/markdown; charset=utf-8"
		if report.Format == "csv" {
			contentType = "text/csv; charset=utf-8"
		}
	}
	extension := "md"
	if report.Format == "csv" {
		extension = "csv"
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="agent-customer-service-channel-ops-%s.%s"`, report.ID, extension))
	_, _ = w.Write([]byte(report.Content))
}

func channelOpsReportSummary(dashboard store.Dashboard) store.ChannelOpsReportSummary {
	summary := store.ChannelOpsReportSummary{
		Channels: []string{},
		InboundAuditQuality: store.ChannelInboundAuditQualitySummary{
			ActiveChannels:    []string{},
			WatchChannels:     []string{},
			RecoveredChannels: []string{},
		},
	}
	seenChannels := map[string]bool{}
	for _, alert := range dashboard.ChannelAlerts {
		summary.FailureCount += alert.Count
		if alert.Channel != "" && !seenChannels[alert.Channel] {
			seenChannels[alert.Channel] = true
			summary.Channels = append(summary.Channels, alert.Channel)
		}
	}
	summary.ActiveRunbooks = len(dashboard.ChannelRunbooks)
	for _, runbook := range dashboard.ChannelRunbooks {
		if runbook.Channel != "" && !seenChannels[runbook.Channel] {
			seenChannels[runbook.Channel] = true
			summary.Channels = append(summary.Channels, runbook.Channel)
		}
	}
	for _, notification := range dashboard.Notifications {
		switch notification.Status {
		case "OPEN":
			summary.OpenNotifications++
		case "RETRYING":
			summary.Retrying++
		case "DEAD_LETTER":
			summary.DeadLetters++
		}
		if notification.Channel != "" && !seenChannels[notification.Channel] {
			seenChannels[notification.Channel] = true
			summary.Channels = append(summary.Channels, notification.Channel)
		}
	}
	summary.InboundAudit = channelInboundAuditSummary(dashboard.ChannelAudits)
	summary.InboundAuditQuality = channelInboundAuditQualitySummary(dashboard.ChannelAudits, dashboard.AlertPolicies, dashboard.AuditEvents)
	summary.RunbookSummary = channelOpsRunbookSummary(dashboard.ChannelRunbooks)
	summary.RunbookLoads = dashboard.RunbookLoads
	summary.HandoffPriorities = channelOpsHandoffPriorities(dashboard)
	for _, channel := range append(append(summary.InboundAuditQuality.ActiveChannels, summary.InboundAuditQuality.WatchChannels...), summary.InboundAuditQuality.RecoveredChannels...) {
		if channel != "" && !seenChannels[channel] {
			seenChannels[channel] = true
			summary.Channels = append(summary.Channels, channel)
		}
	}
	return summary
}

func channelOpsRunbookSummary(runbooks []store.ChannelRunbook) store.ChannelRunbookSummary {
	var summary store.ChannelRunbookSummary
	for _, runbook := range runbooks {
		summary.Total += runbook.CheckSummary.Total
		summary.Done += runbook.CheckSummary.Done
		summary.Blocked += runbook.CheckSummary.Blocked
		summary.Overdue += runbook.CheckSummary.Overdue
		summary.Todo += runbook.CheckSummary.Todo
	}
	return summary
}

type channelOpsHandoffCandidate struct {
	item  store.ChannelOpsHandoffPriority
	score int
}

func channelOpsHandoffPriorities(dashboard store.Dashboard) []store.ChannelOpsHandoffPriority {
	candidates := make([]channelOpsHandoffCandidate, 0)
	inboundQuality := channelInboundAuditQualitySummary(dashboard.ChannelAudits, dashboard.AlertPolicies, dashboard.AuditEvents)
	eventByChannel := latestChannelInboundAuditQualityEventByChannel(dashboard.AuditEvents)
	for _, runbook := range dashboard.ChannelRunbooks {
		if runbook.CheckSummary.Overdue <= 0 {
			continue
		}
		notification := channelNotificationForHandoff(dashboard.Notifications, runbook.Channel, "OPEN", "RETRYING", "DEAD_LETTER")
		candidates = append(candidates, channelOpsHandoffCandidate{
			score: 450000 + runbook.CheckSummary.Overdue,
			item: store.ChannelOpsHandoffPriority{
				Channel:           runbook.Channel,
				Severity:          fallbackReportValue(runbook.Severity),
				Source:            "RUNBOOK_OVERDUE",
				Reason:            fmt.Sprintf("%d overdue runbook checks; blocked=%d owner=%s", runbook.CheckSummary.Overdue, runbook.CheckSummary.Blocked, fallbackReportValue(runbook.Owner)),
				RecommendedAction: "review overdue runbook checks, assign an owner, and recover or escalate blocked steps",
				Count:             runbook.CheckSummary.Overdue,
				ActionType:        "REVIEW_RUNBOOK",
				ActionRef:         handoffRunbookRef(runbook),
				ActionLabel:       handoffRunbookLabel(runbook.Channel, runbook),
				NotificationID:    notification.ID,
				RunbookStatus:     runbook.Status,
			},
		})
	}
	for _, channel := range inboundQuality.ActiveChannels {
		event := eventByChannel[channel]
		snapshot := channelAuditQuality(dashboard.ChannelAudits, channel)
		runbook := channelRunbookForHandoff(dashboard.ChannelRunbooks, channel, "ESCALATE")
		notification := channelNotificationForHandoff(dashboard.Notifications, channel, "OPEN", "RETRYING", "DEAD_LETTER")
		reason := fmt.Sprintf("inbound acceptance is %d%% with %d samples", snapshot.AcceptanceRate, snapshot.Total)
		if event.FailureCode != "" {
			reason = fmt.Sprintf("%s; top failure `%s`", reason, event.FailureCode)
		}
		candidates = append(candidates, channelOpsHandoffCandidate{
			score: 400000 + snapshot.TopErrorCount,
			item: store.ChannelOpsHandoffPriority{
				Channel:           channel,
				Severity:          fallbackReportValue(event.Severity),
				Source:            "INBOUND_AUDIT_ACTIVE",
				Reason:            reason,
				RecommendedAction: "pause blind channel rollout, inspect signature/origin/replay configuration, then rerun inbound examples",
				Count:             snapshot.TopErrorCount,
				ActionType:        "REVIEW_RUNBOOK",
				ActionRef:         handoffRunbookRef(runbook),
				ActionLabel:       handoffRunbookLabel(channel, runbook),
				NotificationID:    notification.ID,
				RunbookStatus:     runbook.Status,
			},
		})
	}
	for _, notification := range dashboard.Notifications {
		switch notification.Status {
		case "DEAD_LETTER":
			runbook := channelRunbookForHandoff(dashboard.ChannelRunbooks, notification.Channel, "ESCALATE")
			candidates = append(candidates, channelOpsHandoffCandidate{
				score: 300000 + notification.Attempts,
				item: store.ChannelOpsHandoffPriority{
					Channel:           notification.Channel,
					Severity:          fallbackReportValue(notification.Severity),
					Source:            "DEAD_LETTER",
					Reason:            fallbackReportValue(notification.DeadLetterReason),
					RecommendedAction: "acknowledge the dead letter, inspect delivery audit, and repair the webhook target before replay",
					Count:             notification.Attempts,
					ActionType:        "ACK_NOTIFICATION",
					ActionRef:         notification.ID,
					ActionLabel:       "Acknowledge dead-letter notification",
					NotificationID:    notification.ID,
					RunbookStatus:     runbook.Status,
				},
			})
		case "RETRYING":
			runbook := channelRunbookForHandoff(dashboard.ChannelRunbooks, notification.Channel, "DISPATCH")
			candidates = append(candidates, channelOpsHandoffCandidate{
				score: 100000 + notification.Attempts,
				item: store.ChannelOpsHandoffPriority{
					Channel:           notification.Channel,
					Severity:          fallbackReportValue(notification.Severity),
					Source:            "RETRYING_DELIVERY",
					Reason:            fallbackReportValue(notification.LastError),
					RecommendedAction: "check target health before retry budget is exhausted",
					Count:             notification.Attempts,
					ActionType:        "ACK_NOTIFICATION",
					ActionRef:         notification.ID,
					ActionLabel:       "Acknowledge or dispatch retrying notification",
					NotificationID:    notification.ID,
					RunbookStatus:     runbook.Status,
				},
			})
		}
	}
	alerts := append([]store.ChannelAlert(nil), dashboard.ChannelAlerts...)
	sort.Slice(alerts, func(i, j int) bool {
		if alerts[i].Count == alerts[j].Count {
			if alerts[i].Channel == alerts[j].Channel {
				return alerts[i].Code < alerts[j].Code
			}
			return alerts[i].Channel < alerts[j].Channel
		}
		return alerts[i].Count > alerts[j].Count
	})
	for _, alert := range alerts {
		runbook := channelRunbookForHandoff(dashboard.ChannelRunbooks, alert.Channel, "DISPATCH")
		notification := channelNotificationForHandoff(dashboard.Notifications, alert.Channel, "OPEN", "RETRYING", "DEAD_LETTER")
		candidates = append(candidates, channelOpsHandoffCandidate{
			score: 200000 + alert.Count,
			item: store.ChannelOpsHandoffPriority{
				Channel:           alert.Channel,
				Severity:          "WATCH",
				Source:            "HIGH_FREQUENCY_FAILURE",
				Reason:            fmt.Sprintf("%d `%s` failures; last reason %s", alert.Count, alert.Code, fallbackReportValue(alert.LastReason)),
				RecommendedAction: "compare the latest failure origin with channel policy and runbook checks",
				Count:             alert.Count,
				ActionType:        "REVIEW_RUNBOOK",
				ActionRef:         handoffRunbookRef(runbook),
				ActionLabel:       handoffRunbookLabel(alert.Channel, runbook),
				NotificationID:    notification.ID,
				RunbookStatus:     runbook.Status,
			},
		})
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].score == candidates[j].score {
			if candidates[i].item.Channel == candidates[j].item.Channel {
				return candidates[i].item.Source < candidates[j].item.Source
			}
			return candidates[i].item.Channel < candidates[j].item.Channel
		}
		return candidates[i].score > candidates[j].score
	})
	priorities := make([]store.ChannelOpsHandoffPriority, 0, len(candidates))
	seen := map[string]bool{}
	for _, candidate := range candidates {
		key := candidate.item.Source + ":" + candidate.item.Channel + ":" + candidate.item.Reason
		if seen[key] {
			continue
		}
		seen[key] = true
		candidate.item.Rank = len(priorities) + 1
		priorities = append(priorities, candidate.item)
		if len(priorities) == 6 {
			break
		}
	}
	return priorities
}

func channelRunbookForHandoff(runbooks []store.ChannelRunbook, channel string, preferredStatus string) store.ChannelRunbook {
	var fallback store.ChannelRunbook
	for _, runbook := range runbooks {
		if !strings.EqualFold(runbook.Channel, channel) {
			continue
		}
		if fallback.Channel == "" {
			fallback = runbook
		}
		if preferredStatus != "" && strings.EqualFold(runbook.Status, preferredStatus) {
			return runbook
		}
	}
	return fallback
}

func channelNotificationForHandoff(notifications []store.ChannelNotification, channel string, statuses ...string) store.ChannelNotification {
	for _, status := range statuses {
		for _, notification := range notifications {
			if strings.EqualFold(notification.Channel, channel) && strings.EqualFold(notification.Status, status) {
				return notification
			}
		}
	}
	return store.ChannelNotification{}
}

func handoffRunbookRef(runbook store.ChannelRunbook) string {
	if runbook.Channel == "" {
		return ""
	}
	if runbook.Status == "" {
		return runbook.Channel
	}
	return runbook.Channel + ":" + runbook.Status
}

func handoffRunbookLabel(channel string, runbook store.ChannelRunbook) string {
	if runbook.Status == "" {
		return "Review channel runbook"
	}
	return fmt.Sprintf("Review %s %s runbook", channel, runbook.Status)
}

func latestChannelInboundAuditQualityEventByChannel(events []store.ChannelInboundAuditQualityEvent) map[string]store.ChannelInboundAuditQualityEvent {
	latest := map[string]store.ChannelInboundAuditQualityEvent{}
	for _, event := range events {
		if strings.TrimSpace(event.Channel) == "" {
			continue
		}
		if _, exists := latest[event.Channel]; exists {
			continue
		}
		latest[event.Channel] = event
	}
	return latest
}

func channelInboundAuditSummary(audits []store.ChannelInboundAudit) store.ChannelInboundAuditSummary {
	summary := store.ChannelInboundAuditSummary{Total: len(audits)}
	errorCounts := map[string]int{}
	for _, audit := range audits {
		switch strings.ToUpper(strings.TrimSpace(audit.Status)) {
		case "ACCEPTED":
			summary.Accepted++
		case "REJECTED":
			summary.Rejected++
			errorCounts[fallbackReportValue(audit.Code)]++
		}
	}
	if summary.Total > 0 {
		summary.AcceptanceRate = int(float64(summary.Accepted) / float64(summary.Total) * 100)
	}
	codes := make([]store.ChannelInboundAuditCodeCount, 0, len(errorCounts))
	for code, count := range errorCounts {
		codes = append(codes, store.ChannelInboundAuditCodeCount{Code: code, Count: count})
	}
	sort.Slice(codes, func(i, j int) bool {
		if codes[i].Count == codes[j].Count {
			return codes[i].Code < codes[j].Code
		}
		return codes[i].Count > codes[j].Count
	})
	if len(codes) > 3 {
		codes = codes[:3]
	}
	summary.TopErrorCodes = codes
	return summary
}

func channelInboundAuditQualitySummary(audits []store.ChannelInboundAudit, policies []store.ChannelAlertPolicy, events []store.ChannelInboundAuditQualityEvent) store.ChannelInboundAuditQualitySummary {
	summary := store.ChannelInboundAuditQualitySummary{
		EventCount:        len(events),
		ActiveChannels:    []string{},
		WatchChannels:     []string{},
		RecoveredChannels: []string{},
	}
	latestByChannel := map[string]store.ChannelInboundAuditQualityEvent{}
	order := make([]string, 0)
	for _, event := range events {
		if strings.TrimSpace(event.Channel) == "" {
			continue
		}
		if _, exists := latestByChannel[event.Channel]; !exists {
			order = append(order, event.Channel)
			latestByChannel[event.Channel] = event
		}
	}
	for _, channel := range order {
		event := latestByChannel[channel]
		current := channelAuditQuality(audits, channel)
		policy := channelAlertPolicyForReport(policies, channel)
		minSamples := fallbackInt(policy.InboundAuditMinSamples, event.MinSamples)
		minAcceptanceRate := fallbackInt(policy.InboundAuditMinAcceptanceRate, event.MinAcceptanceRate)
		maxErrorCount := fallbackInt(policy.InboundAuditMaxErrorCount, event.MaxErrorCount)
		recovered := current.Total >= minSamples && current.AcceptanceRate >= minAcceptanceRate && current.TopErrorCount < maxErrorCount
		if recovered {
			summary.Recovered++
			summary.RecoveredChannels = append(summary.RecoveredChannels, channel)
			continue
		}
		if strings.EqualFold(event.Status, "ESCALATE") {
			summary.Active++
			summary.ActiveChannels = append(summary.ActiveChannels, channel)
			continue
		}
		summary.Watch++
		summary.WatchChannels = append(summary.WatchChannels, channel)
	}
	return summary
}

type channelAuditQualitySnapshot struct {
	Total          int
	Accepted       int
	AcceptanceRate int
	TopErrorCount  int
}

func channelAuditQuality(audits []store.ChannelInboundAudit, channel string) channelAuditQualitySnapshot {
	snapshot := channelAuditQualitySnapshot{}
	errorCounts := map[string]int{}
	for _, audit := range audits {
		if !strings.EqualFold(audit.Channel, channel) {
			continue
		}
		snapshot.Total++
		if strings.EqualFold(audit.Status, "ACCEPTED") {
			snapshot.Accepted++
			continue
		}
		errorCounts[audit.Code]++
		if errorCounts[audit.Code] > snapshot.TopErrorCount {
			snapshot.TopErrorCount = errorCounts[audit.Code]
		}
	}
	if snapshot.Total > 0 {
		snapshot.AcceptanceRate = snapshot.Accepted * 100 / snapshot.Total
	}
	return snapshot
}

func channelAlertPolicyForReport(policies []store.ChannelAlertPolicy, channel string) store.ChannelAlertPolicy {
	for _, policy := range policies {
		if strings.EqualFold(policy.Channel, channel) {
			return policy
		}
	}
	return store.ChannelAlertPolicy{}
}

func fallbackInt(value int, fallback int) int {
	if value > 0 {
		return value
	}
	return fallback
}

func renderChannelOpsReportMarkdown(dashboard store.Dashboard, generatedAt time.Time) string {
	totalFailures := 0
	for _, alert := range dashboard.ChannelAlerts {
		totalFailures += alert.Count
	}
	openNotifications := 0
	retryingNotifications := 0
	deadLetters := 0
	for _, notification := range dashboard.Notifications {
		switch notification.Status {
		case "OPEN":
			openNotifications++
		case "RETRYING":
			retryingNotifications++
		case "DEAD_LETTER":
			deadLetters++
		}
	}
	inboundAudit := channelInboundAuditSummary(dashboard.ChannelAudits)
	inboundQuality := channelInboundAuditQualitySummary(dashboard.ChannelAudits, dashboard.AlertPolicies, dashboard.AuditEvents)
	runbookSummary := channelOpsRunbookSummary(dashboard.ChannelRunbooks)
	handoffPriorities := channelOpsHandoffPriorities(dashboard)
	runbookLoads := dashboard.RunbookLoads

	var b strings.Builder
	b.WriteString("# Agent Customer Service Channel Ops Report\n\n")
	b.WriteString(fmt.Sprintf("- Generated at: %s\n", generatedAt.Format(time.RFC3339)))
	b.WriteString("- Window: last 24 hours\n")
	b.WriteString(fmt.Sprintf("- Channel failures: %d\n", totalFailures))
	b.WriteString(fmt.Sprintf("- Active runbooks: %d\n", len(dashboard.ChannelRunbooks)))
	b.WriteString(fmt.Sprintf("- Runbook progress: done=%d blocked=%d overdue=%d todo=%d total=%d\n", runbookSummary.Done, runbookSummary.Blocked, runbookSummary.Overdue, runbookSummary.Todo, runbookSummary.Total))
	b.WriteString(fmt.Sprintf("- Runbook assignees: %d\n", len(runbookLoads)))
	b.WriteString(fmt.Sprintf("- Inbound audits: total=%d accepted=%d rejected=%d acceptance_rate=%d%%\n", inboundAudit.Total, inboundAudit.Accepted, inboundAudit.Rejected, inboundAudit.AcceptanceRate))
	b.WriteString(fmt.Sprintf("- Inbound quality events: total=%d active=%d watch=%d recovered=%d\n", inboundQuality.EventCount, inboundQuality.Active, inboundQuality.Watch, inboundQuality.Recovered))
	b.WriteString(fmt.Sprintf("- Handoff priorities: %d\n", len(handoffPriorities)))
	b.WriteString(fmt.Sprintf("- Notifications: open=%d retrying=%d dead_letter=%d\n\n", openNotifications, retryingNotifications, deadLetters))

	b.WriteString("## Handoff Priorities\n\n")
	if len(handoffPriorities) == 0 {
		b.WriteString("- No urgent handoff items.\n\n")
	} else {
		for _, item := range handoffPriorities {
			b.WriteString(fmt.Sprintf("%d. %s `%s` [%s]: %s. Next: %s. Action: %s (%s)\n", item.Rank, item.Channel, item.Source, item.Severity, item.Reason, item.RecommendedAction, item.ActionLabel, fallbackReportValue(item.ActionRef)))
			if item.NotificationID != "" || item.RunbookStatus != "" {
				b.WriteString(fmt.Sprintf("   - Links: notification=%s runbook=%s\n", fallbackReportValue(item.NotificationID), fallbackReportValue(item.RunbookStatus)))
			}
		}
		b.WriteString("\n")
	}

	b.WriteString("## Runbook Assignee Load\n\n")
	if len(runbookLoads) == 0 {
		b.WriteString("- No assigned runbook checks.\n\n")
	} else {
		for _, load := range runbookLoads {
			b.WriteString(fmt.Sprintf("- %s: todo=%d blocked=%d overdue=%d done=%d total=%d channels=%s next_due=%s\n", load.Assignee, load.Todo, load.Blocked, load.Overdue, load.Done, load.Total, channelListOrNone(load.Channels), fallbackReportValue(load.NextDueAt)))
		}
		b.WriteString("\n")
	}

	b.WriteString("## Alerts\n\n")
	if len(dashboard.ChannelAlerts) == 0 {
		b.WriteString("- No channel failures.\n\n")
	} else {
		for _, alert := range dashboard.ChannelAlerts {
			b.WriteString(fmt.Sprintf("- %s `%s`: %d failures, last seen %s, origin %s\n", alert.Channel, alert.Code, alert.Count, alert.LastSeenAt, fallbackReportValue(alert.LastOrigin)))
		}
		b.WriteString("\n")
	}

	b.WriteString("## Trends\n\n")
	if len(dashboard.ChannelTrends) == 0 {
		b.WriteString("- No hourly trend buckets.\n\n")
	} else {
		for _, trend := range dashboard.ChannelTrends {
			b.WriteString(fmt.Sprintf("- %s %s: %d\n", trend.Channel, trend.BucketStart, trend.Count))
		}
		b.WriteString("\n")
	}

	b.WriteString("## Inbound Acceptance\n\n")
	if inboundAudit.Total == 0 {
		b.WriteString("- No inbound audit records.\n\n")
	} else {
		b.WriteString(fmt.Sprintf("- Accepted: %d / %d (%d%%)\n", inboundAudit.Accepted, inboundAudit.Total, inboundAudit.AcceptanceRate))
		b.WriteString(fmt.Sprintf("- Rejected: %d\n", inboundAudit.Rejected))
		if len(inboundAudit.TopErrorCodes) > 0 {
			b.WriteString("- Top error codes:\n")
			for _, item := range inboundAudit.TopErrorCodes {
				b.WriteString(fmt.Sprintf("  - `%s`: %d\n", item.Code, item.Count))
			}
		}
		b.WriteString("\n")
	}

	b.WriteString("## Inbound Quality Events\n\n")
	if inboundQuality.EventCount == 0 {
		b.WriteString("- No inbound quality threshold events.\n\n")
	} else {
		b.WriteString(fmt.Sprintf("- Active channels: %s\n", channelListOrNone(inboundQuality.ActiveChannels)))
		b.WriteString(fmt.Sprintf("- Watch channels: %s\n", channelListOrNone(inboundQuality.WatchChannels)))
		b.WriteString(fmt.Sprintf("- Recovered channels: %s\n\n", channelListOrNone(inboundQuality.RecoveredChannels)))
	}

	b.WriteString("## Runbooks\n\n")
	if len(dashboard.ChannelRunbooks) == 0 {
		b.WriteString("- No active runbooks.\n\n")
	} else {
		for _, runbook := range dashboard.ChannelRunbooks {
			b.WriteString(fmt.Sprintf("### %s · %s\n\n", runbook.Channel, runbook.Status))
			b.WriteString(fmt.Sprintf("- Owner: %s\n", runbook.Owner))
			b.WriteString(fmt.Sprintf("- Failure code: %s\n", runbook.FailureCode))
			b.WriteString(fmt.Sprintf("- Progress: done=%d blocked=%d overdue=%d todo=%d total=%d\n", runbook.CheckSummary.Done, runbook.CheckSummary.Blocked, runbook.CheckSummary.Overdue, runbook.CheckSummary.Todo, runbook.CheckSummary.Total))
			b.WriteString(fmt.Sprintf("- Next action: %s\n", runbook.NextAction))
			b.WriteString(fmt.Sprintf("- Escalation: %s\n", runbook.Escalation))
			for _, step := range runbook.Steps {
				b.WriteString(fmt.Sprintf("- [ ] %s\n", step))
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("## Notifications\n\n")
	if len(dashboard.Notifications) == 0 {
		b.WriteString("- No notification events.\n")
	} else {
		for _, notification := range dashboard.Notifications {
			b.WriteString(fmt.Sprintf("- %s %s: status=%s attempts=%d/%d target=%s ackedBy=%s\n", notification.Channel, notification.ID, notification.Status, notification.Attempts, notification.MaxAttempts, notification.Target, fallbackReportValue(notification.AckedBy)))
		}
	}
	return b.String()
}

func renderChannelOpsReportCSV(dashboard store.Dashboard) ([]byte, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	if err := writer.Write([]string{"section", "channel", "status", "code", "count", "owner", "next_action", "escalation"}); err != nil {
		return nil, err
	}
	for _, alert := range dashboard.ChannelAlerts {
		if err := writer.Write([]string{"alert", alert.Channel, "", alert.Code, fmt.Sprintf("%d", alert.Count), "", alert.LastReason, alert.LastOrigin}); err != nil {
			return nil, err
		}
	}
	for _, trend := range dashboard.ChannelTrends {
		if err := writer.Write([]string{"trend", trend.Channel, trend.BucketStart, "", fmt.Sprintf("%d", trend.Count), "", "", ""}); err != nil {
			return nil, err
		}
	}
	inboundAudit := channelInboundAuditSummary(dashboard.ChannelAudits)
	if err := writer.Write([]string{"inbound_audit", "", "ACCEPTANCE_RATE", "accepted", fmt.Sprintf("%d/%d (%d%%)", inboundAudit.Accepted, inboundAudit.Total, inboundAudit.AcceptanceRate), "", "", ""}); err != nil {
		return nil, err
	}
	if err := writer.Write([]string{"inbound_audit", "", "REJECTED", "rejected", fmt.Sprintf("%d", inboundAudit.Rejected), "", "", ""}); err != nil {
		return nil, err
	}
	for _, item := range inboundAudit.TopErrorCodes {
		if err := writer.Write([]string{"inbound_audit_error", "", "REJECTED", item.Code, fmt.Sprintf("%d", item.Count), "", "", ""}); err != nil {
			return nil, err
		}
	}
	inboundQuality := channelInboundAuditQualitySummary(dashboard.ChannelAudits, dashboard.AlertPolicies, dashboard.AuditEvents)
	runbookSummary := channelOpsRunbookSummary(dashboard.ChannelRunbooks)
	if err := writer.Write([]string{"runbook_summary", "", "SUMMARY", "progress", fmt.Sprintf("done=%d blocked=%d overdue=%d todo=%d total=%d", runbookSummary.Done, runbookSummary.Blocked, runbookSummary.Overdue, runbookSummary.Todo, runbookSummary.Total), "", "", ""}); err != nil {
		return nil, err
	}
	for _, load := range dashboard.RunbookLoads {
		detail := fmt.Sprintf("todo=%d blocked=%d overdue=%d done=%d total=%d channels=%s next_due=%s", load.Todo, load.Blocked, load.Overdue, load.Done, load.Total, strings.Join(load.Channels, "|"), fallbackReportValue(load.NextDueAt))
		if err := writer.Write([]string{"runbook_assignee", "", "LOAD", load.Assignee, detail, load.Assignee, "review assigned runbook checks", ""}); err != nil {
			return nil, err
		}
	}
	if err := writer.Write([]string{"inbound_quality", "", "SUMMARY", "events", fmt.Sprintf("total=%d active=%d watch=%d recovered=%d", inboundQuality.EventCount, inboundQuality.Active, inboundQuality.Watch, inboundQuality.Recovered), "", "", ""}); err != nil {
		return nil, err
	}
	for _, channel := range inboundQuality.ActiveChannels {
		if err := writer.Write([]string{"inbound_quality_channel", channel, "ACTIVE", "", "", "", "still below threshold", ""}); err != nil {
			return nil, err
		}
	}
	for _, channel := range inboundQuality.WatchChannels {
		if err := writer.Write([]string{"inbound_quality_channel", channel, "WATCH", "", "", "", "watch threshold trend", ""}); err != nil {
			return nil, err
		}
	}
	for _, channel := range inboundQuality.RecoveredChannels {
		if err := writer.Write([]string{"inbound_quality_channel", channel, "RECOVERED", "", "", "", "current audit quality recovered", ""}); err != nil {
			return nil, err
		}
	}
	for _, item := range channelOpsHandoffPriorities(dashboard) {
		linkSummary := fmt.Sprintf("action=%s ref=%s notification=%s runbook=%s reason=%s", fallbackReportValue(item.ActionType), fallbackReportValue(item.ActionRef), fallbackReportValue(item.NotificationID), fallbackReportValue(item.RunbookStatus), item.Reason)
		if err := writer.Write([]string{"handoff_priority", item.Channel, item.Severity, item.Source, fmt.Sprintf("%d", item.Count), item.ActionLabel, item.RecommendedAction, linkSummary}); err != nil {
			return nil, err
		}
	}
	for _, runbook := range dashboard.ChannelRunbooks {
		progress := fmt.Sprintf("done=%d blocked=%d overdue=%d todo=%d total=%d", runbook.CheckSummary.Done, runbook.CheckSummary.Blocked, runbook.CheckSummary.Overdue, runbook.CheckSummary.Todo, runbook.CheckSummary.Total)
		if err := writer.Write([]string{"runbook", runbook.Channel, runbook.Status, runbook.FailureCode, progress, runbook.Owner, runbook.NextAction, runbook.Escalation}); err != nil {
			return nil, err
		}
	}
	for _, notification := range dashboard.Notifications {
		if err := writer.Write([]string{"notification", notification.Channel, notification.Status, "", fmt.Sprintf("%d/%d", notification.Attempts, notification.MaxAttempts), notification.Target, notification.DeadLetterReason, notification.AckedBy}); err != nil {
			return nil, err
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func fallbackReportValue(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "-"
	}
	return value
}

func channelListOrNone(channels []string) string {
	if len(channels) == 0 {
		return "-"
	}
	return strings.Join(channels, ", ")
}
