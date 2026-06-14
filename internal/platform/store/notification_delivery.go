package store

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type HTTPNotificationDeliveryClient struct {
	client *http.Client
}

func NewHTTPNotificationDeliveryClient(timeout time.Duration) *HTTPNotificationDeliveryClient {
	if timeout <= 0 {
		timeout = 2 * time.Second
	}
	return &HTTPNotificationDeliveryClient{client: &http.Client{Timeout: timeout}}
}

func (c *HTTPNotificationDeliveryClient) DeliverChannelNotification(ctx context.Context, req NotificationDeliveryRequest) (NotificationDeliveryResult, error) {
	body, err := json.Marshal(map[string]any{
		"id":            req.Notification.ID,
		"channel":       req.Notification.Channel,
		"severity":      req.Notification.Severity,
		"target":        req.Notification.Target,
		"reason":        req.Notification.Reason,
		"count":         req.Notification.Count,
		"attempt":       req.Notification.Attempts,
		"signedPayload": req.SignedPayload,
		"outcome":       req.Outcome,
	})
	if err != nil {
		return NotificationDeliveryResult{}, fmt.Errorf("marshal notification webhook body: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, req.Notification.TargetURL, bytes.NewReader(body))
	if err != nil {
		return NotificationDeliveryResult{}, fmt.Errorf("create notification webhook request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Anjing-Notification-ID", req.Notification.ID)
	httpReq.Header.Set("X-Anjing-Signature", req.Notification.Signature)
	httpReq.Header.Set("X-Anjing-Secret-Ref", req.Notification.SecretRef)

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return NotificationDeliveryResult{}, fmt.Errorf("deliver notification webhook: %w", err)
	}
	defer resp.Body.Close()
	responseBody, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
	statusText := fmt.Sprintf("%d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	result := NotificationDeliveryResult{
		Accepted:      resp.StatusCode >= 200 && resp.StatusCode < 300,
		ReceiptStatus: statusText,
		ReceiptBody:   string(responseBody),
	}
	if !result.Accepted {
		result.Error = statusText
	}
	return result, nil
}
