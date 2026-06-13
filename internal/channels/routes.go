package channels

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"

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

var channelSecrets = map[string]string{
	"web":         "web-demo-secret",
	"wechat":      "wechat-demo-secret",
	"app":         "app-demo-secret",
	"marketplace": "marketplace-demo-secret",
}

func Register(mux *http.ServeMux, st store.Runtime) {
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
		if !validSignature(req) {
			httpjson.Fail(w, http.StatusUnauthorized, "invalid_signature", "channel signature verification failed")
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

func validSignature(req inboundRequest) bool {
	expected := ChannelSignature(req.Channel, req.ExternalConversationID, req.Timestamp, req.Content)
	return hmac.Equal([]byte(expected), []byte(strings.TrimSpace(req.Signature)))
}

func ChannelSignature(channel, externalConversationID, timestamp, content string) string {
	secret := channelSecret(channel)
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
	secret, ok := channelSecrets[strings.ToLower(strings.TrimSpace(channel))]
	if !ok {
		return "web-demo-secret"
	}
	return secret
}
