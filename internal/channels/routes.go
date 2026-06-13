package channels

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/anjing-le/agent-customer-service/internal/platform/httpjson"
	"github.com/anjing-le/agent-customer-service/internal/platform/store"
)

type inboundRequest struct {
	Channel                string `json:"channel"`
	ExternalConversationID string `json:"externalConversationId"`
	ExternalMessageID      string `json:"externalMessageId"`
	Customer               string `json:"customer"`
	Content                string `json:"content"`
	Timestamp              string `json:"timestamp"`
	Signature              string `json:"signature"`
}

type wechatInboundRequest struct {
	OpenID    string `json:"openId"`
	MessageID string `json:"msgId"`
	Nickname  string `json:"nickname"`
	Text      string `json:"text"`
	Timestamp string `json:"timestamp"`
	Signature string `json:"signature"`
}

type appInboundRequest struct {
	DeviceID  string `json:"deviceId"`
	MessageID string `json:"messageId"`
	UserName  string `json:"userName"`
	Body      string `json:"body"`
	SentAt    string `json:"sentAt"`
	Signature string `json:"signature"`
}

type marketplaceInboundRequest struct {
	BuyerID    string `json:"buyerId"`
	EventID    string `json:"eventId"`
	BuyerName  string `json:"buyerName"`
	Message    string `json:"message"`
	OccurredAt string `json:"occurredAt"`
	Signature  string `json:"signature"`
}

type Config struct {
	Secrets         map[string]string
	SignatureWindow time.Duration
	Now             func() time.Time
}

var defaultChannelSecrets = map[string]string{
	"web":         "web-demo-secret",
	"wechat":      "wechat-demo-secret",
	"app":         "app-demo-secret",
	"marketplace": "marketplace-demo-secret",
}

func Register(mux *http.ServeMux, st store.Runtime) {
	RegisterWithConfig(mux, st, DefaultConfig())
}

func RegisterWithConfig(mux *http.ServeMux, st store.Runtime, cfg Config) {
	cfg = cfg.withDefaults()
	mux.HandleFunc("/api/channels/inbound", func(w http.ResponseWriter, r *http.Request) {
		if !httpjson.RequireMethod(w, r, http.MethodPost) {
			return
		}
		var req inboundRequest
		if err := httpjson.Decode(r, &req); err != nil {
			httpjson.BadRequest(w, err.Error())
			return
		}
		handleInbound(w, st, cfg, req)
	})
	mux.HandleFunc("/api/channels/wechat/inbound", func(w http.ResponseWriter, r *http.Request) {
		if !httpjson.RequireMethod(w, r, http.MethodPost) {
			return
		}
		var req wechatInboundRequest
		if err := httpjson.Decode(r, &req); err != nil {
			httpjson.BadRequest(w, err.Error())
			return
		}
		handleInbound(w, st, cfg, req.toInbound())
	})
	mux.HandleFunc("/api/channels/app/inbound", func(w http.ResponseWriter, r *http.Request) {
		if !httpjson.RequireMethod(w, r, http.MethodPost) {
			return
		}
		var req appInboundRequest
		if err := httpjson.Decode(r, &req); err != nil {
			httpjson.BadRequest(w, err.Error())
			return
		}
		handleInbound(w, st, cfg, req.toInbound())
	})
	mux.HandleFunc("/api/channels/marketplace/inbound", func(w http.ResponseWriter, r *http.Request) {
		if !httpjson.RequireMethod(w, r, http.MethodPost) {
			return
		}
		var req marketplaceInboundRequest
		if err := httpjson.Decode(r, &req); err != nil {
			httpjson.BadRequest(w, err.Error())
			return
		}
		handleInbound(w, st, cfg, req.toInbound())
	})
}

func handleInbound(w http.ResponseWriter, st store.Runtime, cfg Config, req inboundRequest) {
	if strings.TrimSpace(req.Channel) == "" {
		httpjson.BadRequest(w, "channel is required")
		return
	}
	if strings.TrimSpace(req.ExternalConversationID) == "" {
		httpjson.BadRequest(w, "externalConversationId is required")
		return
	}
	if strings.TrimSpace(req.Content) == "" {
		httpjson.BadRequest(w, "content is required")
		return
	}
	if strings.TrimSpace(req.Timestamp) == "" {
		httpjson.BadRequest(w, "timestamp is required")
		return
	}
	if strings.TrimSpace(req.Signature) == "" {
		httpjson.BadRequest(w, "signature is required")
		return
	}
	integration, err := st.ChannelIntegration(req.Channel)
	if err != nil {
		httpjson.Fail(w, http.StatusInternalServerError, "store_error", err.Error())
		return
	}
	if !integration.Enabled {
		httpjson.Fail(w, http.StatusForbidden, "channel_disabled", "channel integration is disabled")
		return
	}
	ok, code, message := cfg.validSignature(req, integration)
	if !ok {
		httpjson.Fail(w, http.StatusUnauthorized, code, message)
		return
	}
	if integration.ReplayProtection {
		accepted, err := st.RecordChannelInbound(channelInboundReceipt(req))
		if err != nil {
			httpjson.Fail(w, http.StatusInternalServerError, "store_error", err.Error())
			return
		}
		if !accepted {
			httpjson.Fail(w, http.StatusConflict, "duplicate_inbound", "channel inbound message was already accepted")
			return
		}
	}
	result, err := st.ReceiveChannelMessage(store.ChannelInboundMessage{
		Channel:                req.Channel,
		ExternalConversationID: req.ExternalConversationID,
		Customer:               req.Customer,
		Content:                req.Content,
	})
	if err != nil {
		httpjson.Fail(w, http.StatusInternalServerError, "store_error", err.Error())
		return
	}
	httpjson.OK(w, result)
}

func (req wechatInboundRequest) toInbound() inboundRequest {
	return inboundRequest{
		Channel:                "WeChat",
		ExternalConversationID: req.OpenID,
		ExternalMessageID:      req.MessageID,
		Customer:               req.Nickname,
		Content:                req.Text,
		Timestamp:              req.Timestamp,
		Signature:              req.Signature,
	}
}

func (req appInboundRequest) toInbound() inboundRequest {
	return inboundRequest{
		Channel:                "App",
		ExternalConversationID: req.DeviceID,
		ExternalMessageID:      req.MessageID,
		Customer:               req.UserName,
		Content:                req.Body,
		Timestamp:              req.SentAt,
		Signature:              req.Signature,
	}
}

func (req marketplaceInboundRequest) toInbound() inboundRequest {
	return inboundRequest{
		Channel:                "Marketplace",
		ExternalConversationID: req.BuyerID,
		ExternalMessageID:      req.EventID,
		Customer:               req.BuyerName,
		Content:                req.Message,
		Timestamp:              req.OccurredAt,
		Signature:              req.Signature,
	}
}

func DefaultConfig() Config {
	return Config{
		Secrets:         defaultChannelSecrets,
		SignatureWindow: 5 * time.Minute,
		Now:             time.Now,
	}
}

func (cfg Config) withDefaults() Config {
	if len(cfg.Secrets) == 0 {
		cfg.Secrets = defaultChannelSecrets
	}
	if cfg.SignatureWindow <= 0 {
		cfg.SignatureWindow = 5 * time.Minute
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	return cfg
}

func (cfg Config) validSignature(req inboundRequest, integration store.ChannelIntegration) (bool, string, string) {
	timestamp, err := time.Parse(time.RFC3339, strings.TrimSpace(req.Timestamp))
	if err != nil {
		return false, "invalid_timestamp", "timestamp must be RFC3339"
	}
	delta := cfg.Now().Sub(timestamp)
	if delta < 0 {
		delta = -delta
	}
	if delta > cfg.signatureWindow(integration) {
		return false, "stale_signature", "channel signature timestamp is outside allowed window"
	}
	for _, secret := range cfg.channelSecrets(req.Channel, integration) {
		expected := ChannelSignatureWithSecret(secret, req.Channel, req.ExternalConversationID, req.Timestamp, req.Content)
		if hmac.Equal([]byte(expected), []byte(strings.TrimSpace(req.Signature))) {
			return true, "", ""
		}
	}
	return false, "invalid_signature", "channel signature verification failed"
}

func channelInboundReceipt(req inboundRequest) store.ChannelInboundReceipt {
	return store.ChannelInboundReceipt{
		ReplayKey:              replayKey(req),
		Channel:                strings.TrimSpace(req.Channel),
		ExternalConversationID: strings.TrimSpace(req.ExternalConversationID),
		ExternalMessageID:      strings.TrimSpace(req.ExternalMessageID),
		Timestamp:              strings.TrimSpace(req.Timestamp),
		Signature:              strings.TrimSpace(req.Signature),
		ContentHash:            contentHash(req.Content),
	}
}

func ChannelSignature(channel, externalConversationID, timestamp, content string) string {
	secret := DefaultConfig().channelSecret(channel, store.ChannelIntegration{})
	return ChannelSignatureWithSecret(secret, channel, externalConversationID, timestamp, content)
}

func ChannelSignatureWithSecret(secret, channel, externalConversationID, timestamp, content string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(canonicalPayload(channel, externalConversationID, timestamp, content)))
	return hex.EncodeToString(mac.Sum(nil))
}

func canonicalPayload(channel, externalConversationID, timestamp, content string) string {
	return fmt.Sprintf("%s\n%s\n%s\n%s",
		strings.TrimSpace(channel),
		strings.TrimSpace(externalConversationID),
		strings.TrimSpace(timestamp),
		strings.TrimSpace(content),
	)
}

func replayKey(req inboundRequest) string {
	if messageID := strings.TrimSpace(req.ExternalMessageID); messageID != "" {
		sum := sha256.Sum256([]byte(canonicalMessageID(req.Channel, messageID)))
		return hex.EncodeToString(sum[:])
	}
	sum := sha256.Sum256([]byte(canonicalPayload(req.Channel, req.ExternalConversationID, req.Timestamp, strings.TrimSpace(req.Signature))))
	return hex.EncodeToString(sum[:])
}

func canonicalMessageID(channel, externalMessageID string) string {
	return fmt.Sprintf("%s\n%s", strings.TrimSpace(channel), strings.TrimSpace(externalMessageID))
}

func contentHash(content string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(content)))
	return hex.EncodeToString(sum[:])
}

func channelSecret(channel string) string {
	return DefaultConfig().channelSecret(channel, store.ChannelIntegration{})
}

func (cfg Config) channelSecret(channel string, integration store.ChannelIntegration) string {
	secrets := cfg.channelSecrets(channel, integration)
	if len(secrets) == 0 {
		return "web-demo-secret"
	}
	return secrets[0]
}

func (cfg Config) channelSecrets(channel string, integration store.ChannelIntegration) []string {
	secrets := make([]string, 0, 2)
	if secretRef := strings.TrimSpace(integration.SecretRef); secretRef != "" {
		if value := os.Getenv(secretRef); strings.TrimSpace(value) != "" {
			secrets = append(secrets, value)
		}
	}
	if nextSecretRef := strings.TrimSpace(integration.NextSecretRef); nextSecretRef != "" {
		if value := os.Getenv(nextSecretRef); strings.TrimSpace(value) != "" {
			secrets = append(secrets, value)
		}
	}
	if len(secrets) > 0 {
		return secrets
	}
	secret, ok := cfg.Secrets[strings.ToLower(strings.TrimSpace(channel))]
	if !ok || strings.TrimSpace(secret) == "" {
		return []string{"web-demo-secret"}
	}
	return []string{secret}
}

func (cfg Config) signatureWindow(integration store.ChannelIntegration) time.Duration {
	if integration.SignatureWindowSeconds > 0 {
		return time.Duration(integration.SignatureWindowSeconds) * time.Second
	}
	return cfg.SignatureWindow
}
