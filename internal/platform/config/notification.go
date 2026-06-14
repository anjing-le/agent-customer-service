package config

import (
	"strings"
	"time"

	"github.com/anjing-le/agent-customer-service/internal/platform/store"
)

func NotificationDeliveryClient(cfg NotificationConfig) store.NotificationDeliveryClient {
	if !strings.EqualFold(strings.TrimSpace(cfg.DeliveryMode), "http") {
		return nil
	}
	return store.NewHTTPNotificationDeliveryClient(time.Duration(cfg.HTTPTimeoutMillis) * time.Millisecond)
}
