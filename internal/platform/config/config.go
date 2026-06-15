package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	ServiceName   string
	Addr          string
	StaticDir     string
	DatabaseURL   string
	MigrationsDir string
	LLM           LLMConfig
	Channels      ChannelConfig
	Notifications NotificationConfig
	Reports       ReportConfig
}

type LLMConfig struct {
	APIURL string
	APIKey string
	Model  string
}

type ChannelConfig struct {
	Secrets                map[string]string
	SignatureWindowSeconds int
}

type NotificationConfig struct {
	DeliveryMode      string
	HTTPTimeoutMillis int
}

type ReportConfig struct {
	SchedulerEnabled      bool
	SchedulerIntervalMins int
	SchedulerFormat       string
	SchedulerRetain       int
	SchedulerRunOnStartup bool
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
		Channels: ChannelConfig{
			Secrets: map[string]string{
				"web":         env("ANJING_CHANNEL_WEB_SECRET", "web-demo-secret"),
				"wechat":      env("ANJING_CHANNEL_WECHAT_SECRET", "wechat-demo-secret"),
				"wecom":       env("ANJING_CHANNEL_WECOM_SECRET", "wecom-demo-secret"),
				"app":         env("ANJING_CHANNEL_APP_SECRET", "app-demo-secret"),
				"marketplace": env("ANJING_CHANNEL_MARKETPLACE_SECRET", "marketplace-demo-secret"),
			},
			SignatureWindowSeconds: envInt("ANJING_CHANNEL_SIGNATURE_WINDOW_SECONDS", 300),
		},
		Notifications: NotificationConfig{
			DeliveryMode:      env("ANJING_NOTIFICATION_DELIVERY_MODE", "demo"),
			HTTPTimeoutMillis: envInt("ANJING_NOTIFICATION_HTTP_TIMEOUT_MS", 2000),
		},
		Reports: ReportConfig{
			SchedulerEnabled:      envBool("ANJING_CHANNEL_REPORT_SCHEDULER_ENABLED", false),
			SchedulerIntervalMins: envInt("ANJING_CHANNEL_REPORT_INTERVAL_MINUTES", 1440),
			SchedulerFormat:       env("ANJING_CHANNEL_REPORT_FORMAT", "markdown"),
			SchedulerRetain:       envInt("ANJING_CHANNEL_REPORT_RETAIN", 30),
			SchedulerRunOnStartup: envBool("ANJING_CHANNEL_REPORT_RUN_ON_STARTUP", false),
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

func envInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func envBool(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}
