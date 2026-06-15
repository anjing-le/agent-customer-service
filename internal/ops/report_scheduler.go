package ops

import (
	"context"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/anjing-le/agent-customer-service/internal/platform/store"
)

type ReportSchedulerConfig struct {
	Enabled    bool
	Format     string
	Interval   time.Duration
	Retain     int
	RunOnStart bool
}

type ReportScheduler struct {
	st     store.Runtime
	cfg    ReportSchedulerConfig
	logger *slog.Logger
	mu     sync.RWMutex
	status ReportSchedulerStatus
}

type ReportSchedulerStatus struct {
	Enabled      bool   `json:"enabled"`
	Format       string `json:"format"`
	IntervalMins int    `json:"intervalMins"`
	Retain       int    `json:"retain"`
	RunOnStart   bool   `json:"runOnStart"`
	LastRunAt    string `json:"lastRunAt,omitempty"`
	NextRunAt    string `json:"nextRunAt,omitempty"`
	LastReportID string `json:"lastReportId,omitempty"`
	LastStatus   string `json:"lastStatus"`
	LastError    string `json:"lastError,omitempty"`
	LastPruned   int    `json:"lastPruned"`
}

type ReportCompensationResult struct {
	Event  store.ChannelOpsReportEvent `json:"event"`
	Report store.ChannelOpsReport      `json:"report,omitempty"`
	Status ReportSchedulerStatus       `json:"status"`
}

func NewReportScheduler(st store.Runtime, cfg ReportSchedulerConfig, logger *slog.Logger) *ReportScheduler {
	cfg = normalizeReportSchedulerConfig(cfg)
	if logger == nil {
		logger = slog.Default()
	}
	return &ReportScheduler{st: st, cfg: cfg, logger: logger, status: initialReportSchedulerStatus(cfg)}
}

func (s *ReportScheduler) Start(ctx context.Context) {
	if s == nil || s.st == nil || !s.cfg.Enabled {
		return
	}
	s.setNextRun(time.Now().UTC().Add(s.cfg.Interval))
	go s.loop(ctx)
}

func (s *ReportScheduler) Status() ReportSchedulerStatus {
	if s == nil {
		return initialReportSchedulerStatus(ReportSchedulerConfig{})
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.status
}

func (s *ReportScheduler) RunOnce(ctx context.Context) (store.ChannelOpsReport, int, error) {
	select {
	case <-ctx.Done():
		return store.ChannelOpsReport{}, 0, ctx.Err()
	default:
	}
	dashboard, err := s.st.Dashboard()
	if err != nil {
		return store.ChannelOpsReport{}, 0, err
	}
	report, err := buildChannelOpsReport(dashboard, s.cfg.Format, time.Now().UTC())
	if err != nil {
		return store.ChannelOpsReport{}, 0, err
	}
	report, err = s.st.SaveChannelOpsReport(report)
	if err != nil {
		return store.ChannelOpsReport{}, 0, err
	}
	pruned, err := s.st.PruneChannelOpsReports(s.cfg.Retain)
	if err != nil {
		return store.ChannelOpsReport{}, 0, err
	}
	return report, pruned, nil
}

func (s *ReportScheduler) Compensate(ctx context.Context, actor, note string) (ReportCompensationResult, error) {
	report, pruned, err := s.RunOnce(ctx)
	if err != nil {
		s.recordRunFailure(err)
		event, recordErr := s.st.RecordChannelOpsReportEvent(store.ChannelOpsReportEvent{
			Action: "COMPENSATE",
			Actor:  actor,
			Status: "FAILED",
			Format: s.cfg.Format,
			Note:   note,
			Error:  err.Error(),
		})
		if recordErr != nil {
			return ReportCompensationResult{Event: event, Status: s.Status()}, recordErr
		}
		return ReportCompensationResult{Event: event, Status: s.Status()}, err
	}
	s.recordRunSuccess(report, pruned)
	event, err := s.st.RecordChannelOpsReportEvent(store.ChannelOpsReportEvent{
		Action:   "COMPENSATE",
		Actor:    actor,
		Status:   "SUCCESS",
		ReportID: report.ID,
		Format:   report.Format,
		Pruned:   pruned,
		Note:     note,
	})
	if err != nil {
		return ReportCompensationResult{Report: report, Status: s.Status()}, err
	}
	s.logger.Info("channel ops report compensated", "id", report.ID, "actor", actor, "pruned", pruned)
	return ReportCompensationResult{Event: event, Report: report, Status: s.Status()}, nil
}

func (s *ReportScheduler) loop(ctx context.Context) {
	if s.cfg.RunOnStart {
		s.runAndLog(ctx)
	}
	ticker := time.NewTicker(s.cfg.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.runAndLog(ctx)
		}
	}
}

func (s *ReportScheduler) runAndLog(ctx context.Context) {
	report, pruned, err := s.RunOnce(ctx)
	if err != nil {
		s.recordRunFailure(err)
		s.logger.Error("channel ops report scheduler failed", "error", err)
		return
	}
	s.recordRunSuccess(report, pruned)
	s.logger.Info("channel ops report generated", "id", report.ID, "format", report.Format, "pruned", pruned)
}

func (s *ReportScheduler) recordRunSuccess(report store.ChannelOpsReport, pruned int) {
	now := time.Now().UTC()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status.LastRunAt = now.Format(time.RFC3339)
	s.status.NextRunAt = now.Add(s.cfg.Interval).Format(time.RFC3339)
	s.status.LastReportID = report.ID
	s.status.LastStatus = "SUCCESS"
	s.status.LastError = ""
	s.status.LastPruned = pruned
}

func (s *ReportScheduler) recordRunFailure(err error) {
	now := time.Now().UTC()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status.LastRunAt = now.Format(time.RFC3339)
	s.status.NextRunAt = now.Add(s.cfg.Interval).Format(time.RFC3339)
	s.status.LastStatus = "FAILED"
	s.status.LastError = err.Error()
}

func (s *ReportScheduler) setNextRun(next time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status.NextRunAt = next.UTC().Format(time.RFC3339)
}

func initialReportSchedulerStatus(cfg ReportSchedulerConfig) ReportSchedulerStatus {
	cfg = normalizeReportSchedulerConfig(cfg)
	status := "DISABLED"
	if cfg.Enabled {
		status = "PENDING"
	}
	return ReportSchedulerStatus{
		Enabled:      cfg.Enabled,
		Format:       cfg.Format,
		IntervalMins: int(cfg.Interval / time.Minute),
		Retain:       cfg.Retain,
		RunOnStart:   cfg.RunOnStart,
		LastStatus:   status,
	}
}

func normalizeReportSchedulerConfig(cfg ReportSchedulerConfig) ReportSchedulerConfig {
	cfg.Format = strings.ToLower(strings.TrimSpace(cfg.Format))
	if cfg.Format == "md" {
		cfg.Format = "markdown"
	}
	if cfg.Format == "" {
		cfg.Format = "markdown"
	}
	if cfg.Interval <= 0 {
		cfg.Interval = 24 * time.Hour
	}
	if cfg.Retain <= 0 {
		cfg.Retain = 30
	}
	if cfg.Retain > 365 {
		cfg.Retain = 365
	}
	return cfg
}
