package main

import (
	"context"

	"github.com/anjing-le/agent-customer-service/internal/consoleweb"
	"github.com/anjing-le/agent-customer-service/internal/customer"
	"github.com/anjing-le/agent-customer-service/internal/knowledge"
	"github.com/anjing-le/agent-customer-service/internal/ops"
	"github.com/anjing-le/agent-customer-service/internal/platform/config"
	"github.com/anjing-le/agent-customer-service/internal/platform/db"
	"github.com/anjing-le/agent-customer-service/internal/platform/service"
	"github.com/anjing-le/agent-customer-service/internal/platform/store"
)

func main() {
	cfg := config.Load("platform-all", "10002")
	var st store.Runtime = store.NewSeedStore()
	if cfg.DatabaseURL != "" {
		pool, err := db.Open(context.Background(), cfg.DatabaseURL)
		if err != nil {
			panic(err)
		}
		defer pool.Close()
		st = store.NewPostgresStore(pool)
	}
	mux := service.NewMux(cfg.ServiceName, st, customer.Register, knowledge.Register, ops.Register)
	consoleweb.Register(mux, cfg.StaticDir)
	if err := service.Listen(cfg.Addr, cfg.ServiceName, mux); err != nil {
		panic(err)
	}
}
