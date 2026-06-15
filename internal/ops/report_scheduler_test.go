package ops

import (
	"context"
	"testing"
	"time"

	"github.com/anjing-le/agent-customer-service/internal/platform/store"
)

func TestReportSchedulerRunOnceGeneratesAndPrunesReports(t *testing.T) {
	st := store.NewSeedStore()
	scheduler := NewReportScheduler(st, ReportSchedulerConfig{
		Enabled:  true,
		Format:   "markdown",
		Interval: time.Hour,
		Retain:   2,
	}, nil)

	for idx := 0; idx < 3; idx++ {
		report, pruned, err := scheduler.RunOnce(context.Background())
		if err != nil {
			t.Fatalf("run scheduler: %v", err)
		}
		if report.Format != "markdown" || report.Content == "" {
			t.Fatalf("expected markdown report content, got %#v", report)
		}
		if idx < 2 && pruned != 0 {
			t.Fatalf("expected no pruning on run %d, got %d", idx, pruned)
		}
		if idx == 2 && pruned != 1 {
			t.Fatalf("expected one pruned report, got %d", pruned)
		}
	}

	reports, err := st.ListChannelOpsReports(10)
	if err != nil {
		t.Fatalf("list reports: %v", err)
	}
	if len(reports) != 2 {
		t.Fatalf("expected two retained reports, got %#v", reports)
	}
}

func TestReportSchedulerRunOnceHonorsCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	scheduler := NewReportScheduler(store.NewSeedStore(), ReportSchedulerConfig{Enabled: true}, nil)
	if _, _, err := scheduler.RunOnce(ctx); err == nil {
		t.Fatal("expected canceled context error")
	}
}

func TestReportSchedulerStatusTracksRunResult(t *testing.T) {
	st := store.NewSeedStore()
	scheduler := NewReportScheduler(st, ReportSchedulerConfig{
		Enabled:    true,
		Format:     "csv",
		Interval:   30 * time.Minute,
		Retain:     5,
		RunOnStart: true,
	}, nil)

	scheduler.runAndLog(context.Background())
	status := scheduler.Status()

	if !status.Enabled || status.Format != "csv" || status.IntervalMins != 30 || status.Retain != 5 || !status.RunOnStart {
		t.Fatalf("expected scheduler config in status, got %#v", status)
	}
	if status.LastStatus != "SUCCESS" || status.LastReportID == "" || status.LastRunAt == "" || status.NextRunAt == "" {
		t.Fatalf("expected successful run status, got %#v", status)
	}
}

func TestReportSchedulerCompensateRecordsAuditEvent(t *testing.T) {
	st := store.NewSeedStore()
	scheduler := NewReportScheduler(st, ReportSchedulerConfig{
		Enabled:  false,
		Format:   "markdown",
		Interval: time.Hour,
		Retain:   5,
	}, nil)

	result, err := scheduler.Compensate(context.Background(), "ops-lead", "补生成日报")
	if err != nil {
		t.Fatalf("compensate report: %v", err)
	}
	if result.Event.Action != "COMPENSATE" || result.Event.Actor != "ops-lead" || result.Event.Status != "SUCCESS" || result.Event.ReportID == "" {
		t.Fatalf("expected compensation audit event, got %#v", result.Event)
	}
	if result.Report.ID != result.Event.ReportID {
		t.Fatalf("expected event to reference report, got %#v", result)
	}
	events, err := st.ListChannelOpsReportEvents(10)
	if err != nil {
		t.Fatalf("list report events: %v", err)
	}
	if len(events) != 1 || events[0].ID != result.Event.ID {
		t.Fatalf("expected stored compensation event, got %#v", events)
	}
}
