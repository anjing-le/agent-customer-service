package main

import (
	"github.com/anjing-le/agent-customer-service/internal/customer"
	"github.com/anjing-le/agent-customer-service/internal/knowledge"
	"github.com/anjing-le/agent-customer-service/internal/platform/config"
	"github.com/anjing-le/agent-customer-service/internal/platform/service"
	"github.com/anjing-le/agent-customer-service/internal/platform/store"
)

func main() {
	cfg := config.Load("customer-service-api", "10002")
	st := store.NewSeedStore()
	mux := service.NewMux(cfg.ServiceName, st, customer.Register, knowledge.Register)
	if err := service.Listen(cfg.Addr, cfg.ServiceName, mux); err != nil {
		panic(err)
	}
}
