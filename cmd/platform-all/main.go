package main

import (
	"context"

	"github.com/anjing-le/agent-customer-service/internal/channels"
	"github.com/anjing-le/agent-customer-service/internal/consoleweb"
	"github.com/anjing-le/agent-customer-service/internal/customer"
	"github.com/anjing-le/agent-customer-service/internal/knowledge"
	"github.com/anjing-le/agent-customer-service/internal/ops"
	"github.com/anjing-le/agent-customer-service/internal/platform/config"
	"github.com/anjing-le/agent-customer-service/internal/platform/db"
	"github.com/anjing-le/agent-customer-service/internal/platform/llm"
	"github.com/anjing-le/agent-customer-service/internal/platform/service"
	"github.com/anjing-le/agent-customer-service/internal/platform/store"
)

func main() {
	cfg := config.Load("platform-all", "10002")
	generator := llm.NewClient(cfg.LLM.APIURL, cfg.LLM.APIKey, cfg.LLM.Model)
	var seedOptions []store.Option
	var postgresOptions []store.PostgresOption
	if generator.Enabled() {
		seedOptions = append(seedOptions, store.WithReplyGenerator(generator))
		postgresOptions = append(postgresOptions, store.WithPostgresReplyGenerator(generator))
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
	mux := service.NewMux(cfg.ServiceName, st, customer.Register, channels.Register, knowledge.Register, ops.Register)
	consoleweb.Register(mux, cfg.StaticDir)
	if err := service.Listen(cfg.Addr, cfg.ServiceName, mux); err != nil {
		panic(err)
	}
}
