package ops

import (
	"context"
	"log/slog"
	"strings"
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
}

func NewReportScheduler(st store.Runtime, cfg ReportSchedulerConfig, logger *slog.Logger) *ReportScheduler {
	cfg = normalizeReportSchedulerConfig(cfg)
	if logger == nil {
		logger = slog.Default()
	}
	return &ReportScheduler{st: st, cfg: cfg, logger: logger}
}

func (s *ReportScheduler) Start(ctx context.Context) {
	if s == nil || s.st == nil || !s.cfg.Enabled {
		return
	}
	go s.loop(ctx)
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
		s.logger.Error("channel ops report scheduler failed", "error", err)
		return
	}
	s.logger.Info("channel ops report generated", "id", report.ID, "format", report.Format, "pruned", pruned)
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
