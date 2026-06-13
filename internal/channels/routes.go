package channels

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/anjing-le/agent-customer-service/internal/platform/httpjson"
	"github.com/anjing-le/agent-customer-service/internal/platform/store"
)

type inboundRequest struct {
	Channel                string `json:"channel"`
	ExternalConversationID string `json:"externalConversationId"`
	Customer               string `json:"customer"`
	Content                string `json:"content"`
	Timestamp              string `json:"timestamp"`
	Signature              string `json:"signature"`
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
		ok, code, message := cfg.validSignature(req)
		if !ok {
			httpjson.Fail(w, http.StatusUnauthorized, code, message)
			return
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
	})
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

func (cfg Config) validSignature(req inboundRequest) (bool, string, string) {
	timestamp, err := time.Parse(time.RFC3339, strings.TrimSpace(req.Timestamp))
	if err != nil {
		return false, "invalid_timestamp", "timestamp must be RFC3339"
	}
	delta := cfg.Now().Sub(timestamp)
	if delta < 0 {
		delta = -delta
	}
	if delta > cfg.SignatureWindow {
		return false, "stale_signature", "channel signature timestamp is outside allowed window"
	}
	expected := ChannelSignatureWithSecret(cfg.channelSecret(req.Channel), req.Channel, req.ExternalConversationID, req.Timestamp, req.Content)
	if !hmac.Equal([]byte(expected), []byte(strings.TrimSpace(req.Signature))) {
		return false, "invalid_signature", "channel signature verification failed"
	}
	return true, "", ""
}

func ChannelSignature(channel, externalConversationID, timestamp, content string) string {
	secret := DefaultConfig().channelSecret(channel)
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

func channelSecret(channel string) string {
	return DefaultConfig().channelSecret(channel)
}

func (cfg Config) channelSecret(channel string) string {
	secret, ok := cfg.Secrets[strings.ToLower(strings.TrimSpace(channel))]
	if !ok {
		return "web-demo-secret"
	}
	return secret
}
