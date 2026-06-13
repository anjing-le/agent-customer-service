package main

import (
	"github.com/anjing-le/agent-customer-service/internal/consoleweb"
	"github.com/anjing-le/agent-customer-service/internal/platform/config"
	"github.com/anjing-le/agent-customer-service/internal/platform/service"
	"github.com/anjing-le/agent-customer-service/internal/platform/store"
)

func main() {
	cfg := config.Load("console-web", "20002")
	st := store.NewSeedStore()
	mux := service.NewMux(cfg.ServiceName, st)
	consoleweb.Register(mux, cfg.StaticDir)
	if err := service.Listen(cfg.Addr, cfg.ServiceName, mux); err != nil {
		panic(err)
	}
}
