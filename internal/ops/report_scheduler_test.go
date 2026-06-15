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
