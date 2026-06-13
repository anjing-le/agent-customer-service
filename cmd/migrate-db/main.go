package main

import (
	"context"
	"fmt"

	"github.com/anjing-le/agent-customer-service/internal/platform/config"
	"github.com/anjing-le/agent-customer-service/internal/platform/db"
)

func main() {
	cfg := config.Load("migrate-db", "0")
	pool, err := db.Open(context.Background(), cfg.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer pool.Close()

	applied, err := db.RunMigrations(context.Background(), pool, cfg.MigrationsDir)
	if err != nil {
		panic(err)
	}
	fmt.Printf("applied migrations: %d\n", len(applied))
	for _, name := range applied {
		fmt.Printf("- %s\n", name)
	}
}
