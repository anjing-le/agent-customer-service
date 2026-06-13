package config

import (
	"fmt"
	"os"
)

type Config struct {
	ServiceName   string
	Addr          string
	StaticDir     string
	DatabaseURL   string
	MigrationsDir string
	LLM           LLMConfig
}

type LLMConfig struct {
	APIURL string
	APIKey string
	Model  string
}

func Load(serviceName, defaultPort string) Config {
	return Config{
		ServiceName:   serviceName,
		Addr:          env("ANJING_ADDR", fmt.Sprintf(":%s", defaultPort)),
		StaticDir:     env("ANJING_CONSOLE_DIST", "apps/console/dist"),
		DatabaseURL:   env("ANJING_DATABASE_URL", ""),
		MigrationsDir: env("ANJING_MIGRATIONS_DIR", "infra/postgres/migrations"),
		LLM: LLMConfig{
			APIURL: env("ANJING_LLM_API_URL", ""),
			APIKey: env("ANJING_LLM_API_KEY", ""),
			Model:  env("ANJING_LLM_MODEL", "gpt-4o-mini"),
		},
	}
}

func env(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
