package main

import (
	"context"
	"net/http"
	"time"

	"github.com/anjing-le/agent-customer-service/internal/ops"
	"github.com/anjing-le/agent-customer-service/internal/platform/config"
	"github.com/anjing-le/agent-customer-service/internal/platform/db"
	"github.com/anjing-le/agent-customer-service/internal/platform/service"
	"github.com/anjing-le/agent-customer-service/internal/platform/store"
)

func main() {
	cfg := config.Load("ops-api", "10003")
	var seedOptions []store.Option
	var postgresOptions []store.PostgresOption
	if delivery := config.NotificationDeliveryClient(cfg.Notifications); delivery != nil {
		seedOptions = append(seedOptions, store.WithNotificationDeliveryClient(delivery))
		postgresOptions = append(postgresOptions, store.WithPostgresNotificationDeliveryClient(delivery))
	}
	var st store.Runtime = store.NewSeedStore(seedOptions...)
	if cfg.DatabaseURL != "" {
		pool, err := db.Open(context.Background(), cfg.DatabaseURL)
		if err != nil {
			panic(err)
		}
		defer pool.Close()
		st = store.NewPostgresStore(pool, postgresOptions...)
	}
	reportScheduler := ops.NewReportScheduler(st, ops.ReportSchedulerConfig{
		Enabled:    cfg.Reports.SchedulerEnabled,
		Format:     cfg.Reports.SchedulerFormat,
		Interval:   time.Duration(cfg.Reports.SchedulerIntervalMins) * time.Minute,
		Retain:     cfg.Reports.SchedulerRetain,
		RunOnStart: cfg.Reports.SchedulerRunOnStartup,
	}, nil)
	reportScheduler.Start(context.Background())
	opsRegister := func(mux *http.ServeMux, st store.Runtime) {
		ops.RegisterWithReportScheduler(mux, st, reportScheduler)
	}
	mux := service.NewMux(cfg.ServiceName, st, opsRegister)
	if err := service.Listen(cfg.Addr, cfg.ServiceName, mux); err != nil {
		panic(err)
	}
}
