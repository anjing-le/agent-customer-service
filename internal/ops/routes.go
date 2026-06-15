package ops

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"net/http"
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
	summary := store.ChannelOpsReportSummary{}
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
	return summary
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

	var b strings.Builder
	b.WriteString("# Agent Customer Service Channel Ops Report\n\n")
	b.WriteString(fmt.Sprintf("- Generated at: %s\n", generatedAt.Format(time.RFC3339)))
	b.WriteString("- Window: last 24 hours\n")
	b.WriteString(fmt.Sprintf("- Channel failures: %d\n", totalFailures))
	b.WriteString(fmt.Sprintf("- Active runbooks: %d\n", len(dashboard.ChannelRunbooks)))
	b.WriteString(fmt.Sprintf("- Notifications: open=%d retrying=%d dead_letter=%d\n\n", openNotifications, retryingNotifications, deadLetters))

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

	b.WriteString("## Runbooks\n\n")
	if len(dashboard.ChannelRunbooks) == 0 {
		b.WriteString("- No active runbooks.\n\n")
	} else {
		for _, runbook := range dashboard.ChannelRunbooks {
			b.WriteString(fmt.Sprintf("### %s · %s\n\n", runbook.Channel, runbook.Status))
			b.WriteString(fmt.Sprintf("- Owner: %s\n", runbook.Owner))
			b.WriteString(fmt.Sprintf("- Failure code: %s\n", runbook.FailureCode))
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
	for _, runbook := range dashboard.ChannelRunbooks {
		if err := writer.Write([]string{"runbook", runbook.Channel, runbook.Status, runbook.FailureCode, "", runbook.Owner, runbook.NextAction, runbook.Escalation}); err != nil {
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
