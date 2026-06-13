package channels

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/anjing-le/agent-customer-service/internal/platform/store"
)

func TestInboundRouteAcceptsSignedChannelMessage(t *testing.T) {
	mux := http.NewServeMux()
	RegisterWithConfig(mux, store.NewSeedStore(), testConfig())

	timestamp := "2026-06-14T02:10:00Z"
	content := "这个商品能不能开发票？"
	signature := ChannelSignature("WeChat", "wx-open-1", timestamp, content)
	body := strings.NewReader(`{"channel":"WeChat","externalConversationId":"wx-open-1","customer":"微信客户","content":"` + content + `","timestamp":"` + timestamp + `","signature":"` + signature + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/channels/inbound", body)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	for _, expected := range []string{`"channel":"WeChat"`, `"conversationId":"conv_wechat_wx_open_1"`, `"agentMessage"`} {
		if !strings.Contains(rec.Body.String(), expected) {
			t.Fatalf("expected %s in response, got %s", expected, rec.Body.String())
		}
	}
}

func TestInboundRouteRejectsInvalidSignature(t *testing.T) {
	mux := http.NewServeMux()
	RegisterWithConfig(mux, store.NewSeedStore(), testConfig())

	body := strings.NewReader(`{"channel":"WeChat","externalConversationId":"wx-open-1","customer":"微信客户","content":"你好","timestamp":"2026-06-14T02:10:00Z","signature":"bad"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/channels/inbound", body)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"code":"invalid_signature"`) {
		t.Fatalf("expected invalid signature error, got %s", rec.Body.String())
	}
}

func TestInboundRouteRejectsStaleTimestamp(t *testing.T) {
	mux := http.NewServeMux()
	RegisterWithConfig(mux, store.NewSeedStore(), testConfig())

	timestamp := "2026-06-14T02:00:00Z"
	content := "你好"
	signature := ChannelSignature("WeChat", "wx-open-1", timestamp, content)
	body := strings.NewReader(`{"channel":"WeChat","externalConversationId":"wx-open-1","customer":"微信客户","content":"` + content + `","timestamp":"` + timestamp + `","signature":"` + signature + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/channels/inbound", body)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"code":"stale_signature"`) {
		t.Fatalf("expected stale signature error, got %s", rec.Body.String())
	}
}

func TestInboundRouteRejectsDuplicateReplay(t *testing.T) {
	mux := http.NewServeMux()
	RegisterWithConfig(mux, store.NewSeedStore(), testConfig())

	timestamp := "2026-06-14T02:10:00Z"
	content := "这个商品能不能开发票？"
	signature := ChannelSignature("WeChat", "wx-open-duplicate", timestamp, content)
	body := `{"channel":"WeChat","externalConversationId":"wx-open-duplicate","customer":"微信客户","content":"` + content + `","timestamp":"` + timestamp + `","signature":"` + signature + `"}`

	firstReq := httptest.NewRequest(http.MethodPost, "/api/channels/inbound", strings.NewReader(body))
	firstRec := httptest.NewRecorder()
	mux.ServeHTTP(firstRec, firstReq)
	if firstRec.Code != http.StatusOK {
		t.Fatalf("expected first request 200, got %d: %s", firstRec.Code, firstRec.Body.String())
	}

	secondReq := httptest.NewRequest(http.MethodPost, "/api/channels/inbound", strings.NewReader(body))
	secondRec := httptest.NewRecorder()
	mux.ServeHTTP(secondRec, secondReq)
	if secondRec.Code != http.StatusConflict {
		t.Fatalf("expected duplicate request 409, got %d: %s", secondRec.Code, secondRec.Body.String())
	}
	if !strings.Contains(secondRec.Body.String(), `"code":"duplicate_inbound"`) {
		t.Fatalf("expected duplicate inbound error, got %s", secondRec.Body.String())
	}
}

func TestInboundRouteSkipsReplayWhenIntegrationDisablesIt(t *testing.T) {
	mux := http.NewServeMux()
	RegisterWithConfig(mux, store.NewSeedStore(store.WithChannelIntegrations([]store.ChannelIntegration{
		testIntegration("WeChat", true, "ANJING_CHANNEL_WECHAT_SECRET", 300, false),
	})), testConfig())

	timestamp := "2026-06-14T02:10:00Z"
	content := "这个商品能不能开发票？"
	signature := ChannelSignature("WeChat", "wx-open-no-replay", timestamp, content)
	body := `{"channel":"WeChat","externalConversationId":"wx-open-no-replay","customer":"微信客户","content":"` + content + `","timestamp":"` + timestamp + `","signature":"` + signature + `"}`

	for idx := 0; idx < 2; idx++ {
		req := httptest.NewRequest(http.MethodPost, "/api/channels/inbound", strings.NewReader(body))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected request %d to bypass replay protection, got %d: %s", idx+1, rec.Code, rec.Body.String())
		}
	}
}

func TestInboundRouteUsesExternalMessageIDForReplay(t *testing.T) {
	mux := http.NewServeMux()
	RegisterWithConfig(mux, store.NewSeedStore(), testConfig())

	content := "这个商品能不能开发票？"
	firstTimestamp := "2026-06-14T02:10:00Z"
	firstSignature := ChannelSignature("WeChat", "wx-open-message-id", firstTimestamp, content)
	firstBody := `{"channel":"WeChat","externalConversationId":"wx-open-message-id","externalMessageId":"wechat-msg-1","customer":"微信客户","content":"` + content + `","timestamp":"` + firstTimestamp + `","signature":"` + firstSignature + `"}`
	firstReq := httptest.NewRequest(http.MethodPost, "/api/channels/inbound", strings.NewReader(firstBody))
	firstRec := httptest.NewRecorder()
	mux.ServeHTTP(firstRec, firstReq)
	if firstRec.Code != http.StatusOK {
		t.Fatalf("expected first message id request 200, got %d: %s", firstRec.Code, firstRec.Body.String())
	}

	secondTimestamp := "2026-06-14T02:10:10Z"
	secondSignature := ChannelSignature("WeChat", "wx-open-message-id", secondTimestamp, content)
	secondBody := `{"channel":"WeChat","externalConversationId":"wx-open-message-id","externalMessageId":"wechat-msg-1","customer":"微信客户","content":"` + content + `","timestamp":"` + secondTimestamp + `","signature":"` + secondSignature + `"}`
	secondReq := httptest.NewRequest(http.MethodPost, "/api/channels/inbound", strings.NewReader(secondBody))
	secondRec := httptest.NewRecorder()
	mux.ServeHTTP(secondRec, secondReq)
	if secondRec.Code != http.StatusConflict {
		t.Fatalf("expected duplicate message id request 409, got %d: %s", secondRec.Code, secondRec.Body.String())
	}
	if !strings.Contains(secondRec.Body.String(), `"code":"duplicate_inbound"`) {
		t.Fatalf("expected duplicate inbound error, got %s", secondRec.Body.String())
	}
}

func TestInboundRouteScopesExternalMessageIDByChannel(t *testing.T) {
	mux := http.NewServeMux()
	RegisterWithConfig(mux, store.NewSeedStore(), testConfig())

	timestamp := "2026-06-14T02:10:00Z"
	content := "这个商品能不能开发票？"
	for _, channel := range []string{"WeChat", "App"} {
		signature := ChannelSignature(channel, channel+"-conversation", timestamp, content)
		body := `{"channel":"` + channel + `","externalConversationId":"` + channel + `-conversation","externalMessageId":"shared-message-id","customer":"渠道客户","content":"` + content + `","timestamp":"` + timestamp + `","signature":"` + signature + `"}`
		req := httptest.NewRequest(http.MethodPost, "/api/channels/inbound", strings.NewReader(body))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected channel %s to accept scoped message id, got %d: %s", channel, rec.Code, rec.Body.String())
		}
	}
}

func TestWeChatAdapterNormalizesInboundMessage(t *testing.T) {
	mux := http.NewServeMux()
	RegisterWithConfig(mux, store.NewSeedStore(), testConfig())

	timestamp := "2026-06-14T02:10:00Z"
	content := "这个商品能不能开发票？"
	signature := ChannelSignature("WeChat", "wx-open-adapter", timestamp, content)
	body := strings.NewReader(`{"openId":"wx-open-adapter","msgId":"wechat-adapter-msg-1","nickname":"微信客户","text":"` + content + `","timestamp":"` + timestamp + `","signature":"` + signature + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/channels/wechat/inbound", body)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	for _, expected := range []string{`"channel":"WeChat"`, `"conversationId":"conv_wechat_wx_open_adapter"`, `"agentMessage"`} {
		if !strings.Contains(rec.Body.String(), expected) {
			t.Fatalf("expected %s in response, got %s", expected, rec.Body.String())
		}
	}
}

func TestAppAdapterNormalizesInboundMessage(t *testing.T) {
	mux := http.NewServeMux()
	RegisterWithConfig(mux, store.NewSeedStore(), testConfig())

	timestamp := "2026-06-14T02:10:00Z"
	content := "这个商品能不能开发票？"
	signature := ChannelSignature("App", "device-1", timestamp, content)
	body := strings.NewReader(`{"deviceId":"device-1","messageId":"app-msg-1","userName":"App 客户","body":"` + content + `","sentAt":"` + timestamp + `","signature":"` + signature + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/channels/app/inbound", body)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"conversationId":"conv_app_device_1"`) {
		t.Fatalf("expected app conversation id, got %s", rec.Body.String())
	}
}

func TestMarketplaceAdapterUsesEventIDForReplay(t *testing.T) {
	mux := http.NewServeMux()
	RegisterWithConfig(mux, store.NewSeedStore(), testConfig())

	content := "这个商品能不能开发票？"
	firstTimestamp := "2026-06-14T02:10:00Z"
	firstSignature := ChannelSignature("Marketplace", "buyer-1", firstTimestamp, content)
	firstBody := `{"buyerId":"buyer-1","eventId":"market-event-1","buyerName":"平台买家","message":"` + content + `","occurredAt":"` + firstTimestamp + `","signature":"` + firstSignature + `"}`
	firstReq := httptest.NewRequest(http.MethodPost, "/api/channels/marketplace/inbound", strings.NewReader(firstBody))
	firstRec := httptest.NewRecorder()
	mux.ServeHTTP(firstRec, firstReq)
	if firstRec.Code != http.StatusOK {
		t.Fatalf("expected first marketplace event 200, got %d: %s", firstRec.Code, firstRec.Body.String())
	}

	secondTimestamp := "2026-06-14T02:10:10Z"
	secondSignature := ChannelSignature("Marketplace", "buyer-1", secondTimestamp, content)
	secondBody := `{"buyerId":"buyer-1","eventId":"market-event-1","buyerName":"平台买家","message":"` + content + `","occurredAt":"` + secondTimestamp + `","signature":"` + secondSignature + `"}`
	secondReq := httptest.NewRequest(http.MethodPost, "/api/channels/marketplace/inbound", strings.NewReader(secondBody))
	secondRec := httptest.NewRecorder()
	mux.ServeHTTP(secondRec, secondReq)
	if secondRec.Code != http.StatusConflict {
		t.Fatalf("expected duplicate marketplace event 409, got %d: %s", secondRec.Code, secondRec.Body.String())
	}
	if !strings.Contains(secondRec.Body.String(), `"code":"duplicate_inbound"`) {
		t.Fatalf("expected duplicate inbound error, got %s", secondRec.Body.String())
	}
}

func TestInboundRouteAcceptsConfiguredChannelSecret(t *testing.T) {
	mux := http.NewServeMux()
	cfg := testConfig()
	cfg.Secrets = map[string]string{"wechat": "custom-wechat-secret"}
	RegisterWithConfig(mux, store.NewSeedStore(), cfg)

	timestamp := "2026-06-14T02:10:00Z"
	content := "这个商品能不能开发票？"
	signature := ChannelSignatureWithSecret("custom-wechat-secret", "WeChat", "wx-open-2", timestamp, content)
	body := strings.NewReader(`{"channel":"WeChat","externalConversationId":"wx-open-2","customer":"微信客户","content":"` + content + `","timestamp":"` + timestamp + `","signature":"` + signature + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/channels/inbound", body)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestInboundRouteUsesIntegrationSecretRef(t *testing.T) {
	t.Setenv("ANJING_TEST_WECHAT_SECRET", "runtime-wechat-secret")
	mux := http.NewServeMux()
	RegisterWithConfig(mux, store.NewSeedStore(store.WithChannelIntegrations([]store.ChannelIntegration{
		testIntegration("WeChat", true, "ANJING_TEST_WECHAT_SECRET", 300, true),
	})), testConfig())

	timestamp := "2026-06-14T02:10:00Z"
	content := "这个商品能不能开发票？"
	signature := ChannelSignatureWithSecret("runtime-wechat-secret", "WeChat", "wx-open-runtime-secret", timestamp, content)
	body := strings.NewReader(`{"channel":"WeChat","externalConversationId":"wx-open-runtime-secret","customer":"微信客户","content":"` + content + `","timestamp":"` + timestamp + `","signature":"` + signature + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/channels/inbound", body)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestInboundRouteAcceptsNextSecretRefDuringRotation(t *testing.T) {
	t.Setenv("ANJING_TEST_ACTIVE_WECHAT_SECRET", "active-wechat-secret")
	t.Setenv("ANJING_TEST_NEXT_WECHAT_SECRET", "next-wechat-secret")
	mux := http.NewServeMux()
	integration := testIntegration("WeChat", true, "ANJING_TEST_ACTIVE_WECHAT_SECRET", 300, true)
	integration.NextSecretRef = "ANJING_TEST_NEXT_WECHAT_SECRET"
	RegisterWithConfig(mux, store.NewSeedStore(store.WithChannelIntegrations([]store.ChannelIntegration{integration})), testConfig())

	timestamp := "2026-06-14T02:10:00Z"
	content := "这个商品能不能开发票？"
	signature := ChannelSignatureWithSecret("next-wechat-secret", "WeChat", "wx-open-next-secret", timestamp, content)
	body := strings.NewReader(`{"channel":"WeChat","externalConversationId":"wx-open-next-secret","externalMessageId":"next-secret-msg-1","customer":"微信客户","content":"` + content + `","timestamp":"` + timestamp + `","signature":"` + signature + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/channels/inbound", body)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected next secret to verify during rotation, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestInboundRouteAcceptsAllowedOrigin(t *testing.T) {
	mux := http.NewServeMux()
	integration := testIntegration("WeChat", true, "ANJING_CHANNEL_WECHAT_SECRET", 300, true)
	integration.AllowedOrigins = []string{"https://wechat.example.com"}
	RegisterWithConfig(mux, store.NewSeedStore(store.WithChannelIntegrations([]store.ChannelIntegration{integration})), testConfig())

	timestamp := "2026-06-14T02:10:00Z"
	content := "这个商品能不能开发票？"
	signature := ChannelSignature("WeChat", "wx-open-origin-ok", timestamp, content)
	body := strings.NewReader(`{"channel":"WeChat","externalConversationId":"wx-open-origin-ok","customer":"微信客户","content":"` + content + `","timestamp":"` + timestamp + `","signature":"` + signature + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/channels/inbound", body)
	req.Header.Set("X-Channel-Origin", "https://wechat.example.com")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected allowed origin 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestInboundRouteRejectsDeniedOrigin(t *testing.T) {
	mux := http.NewServeMux()
	integration := testIntegration("WeChat", true, "ANJING_CHANNEL_WECHAT_SECRET", 300, true)
	integration.AllowedOrigins = []string{"https://wechat.example.com"}
	RegisterWithConfig(mux, store.NewSeedStore(store.WithChannelIntegrations([]store.ChannelIntegration{integration})), testConfig())

	timestamp := "2026-06-14T02:10:00Z"
	content := "这个商品能不能开发票？"
	signature := ChannelSignature("WeChat", "wx-open-origin-denied", timestamp, content)
	body := strings.NewReader(`{"channel":"WeChat","externalConversationId":"wx-open-origin-denied","customer":"微信客户","content":"` + content + `","timestamp":"` + timestamp + `","signature":"` + signature + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/channels/inbound", body)
	req.Header.Set("X-Channel-Origin", "https://evil.example.com")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected denied origin 403, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"code":"channel_origin_denied"`) {
		t.Fatalf("expected origin denied error, got %s", rec.Body.String())
	}
}

func TestInboundRouteRejectsRateLimitedChannel(t *testing.T) {
	mux := http.NewServeMux()
	integration := testIntegration("WeChat", true, "ANJING_CHANNEL_WECHAT_SECRET", 300, false)
	integration.RateLimitPerMinute = 1
	RegisterWithConfig(mux, store.NewSeedStore(store.WithChannelIntegrations([]store.ChannelIntegration{integration})), testConfig())

	timestamp := "2026-06-14T02:10:00Z"
	content := "这个商品能不能开发票？"
	for idx, externalID := range []string{"wx-open-rate-1", "wx-open-rate-2"} {
		signature := ChannelSignature("WeChat", externalID, timestamp, content)
		body := strings.NewReader(`{"channel":"WeChat","externalConversationId":"` + externalID + `","customer":"微信客户","content":"` + content + `","timestamp":"` + timestamp + `","signature":"` + signature + `"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/channels/inbound", body)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if idx == 0 && rec.Code != http.StatusOK {
			t.Fatalf("expected first request 200, got %d: %s", rec.Code, rec.Body.String())
		}
		if idx == 1 {
			if rec.Code != http.StatusTooManyRequests {
				t.Fatalf("expected second request 429, got %d: %s", rec.Code, rec.Body.String())
			}
			if !strings.Contains(rec.Body.String(), `"code":"channel_rate_limited"`) {
				t.Fatalf("expected rate limited error, got %s", rec.Body.String())
			}
		}
	}
}

func TestInboundRouteSkipsRateLimitWhenDisabled(t *testing.T) {
	mux := http.NewServeMux()
	integration := testIntegration("WeChat", true, "ANJING_CHANNEL_WECHAT_SECRET", 300, false)
	integration.RateLimitPerMinute = 0
	RegisterWithConfig(mux, store.NewSeedStore(store.WithChannelIntegrations([]store.ChannelIntegration{integration})), testConfig())

	timestamp := "2026-06-14T02:10:00Z"
	content := "这个商品能不能开发票？"
	for _, externalID := range []string{"wx-open-rate-off-1", "wx-open-rate-off-2"} {
		signature := ChannelSignature("WeChat", externalID, timestamp, content)
		body := strings.NewReader(`{"channel":"WeChat","externalConversationId":"` + externalID + `","customer":"微信客户","content":"` + content + `","timestamp":"` + timestamp + `","signature":"` + signature + `"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/channels/inbound", body)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected disabled rate limit request 200, got %d: %s", rec.Code, rec.Body.String())
		}
	}
}

func TestInboundRouteRejectsDisabledIntegration(t *testing.T) {
	mux := http.NewServeMux()
	RegisterWithConfig(mux, store.NewSeedStore(store.WithChannelIntegrations([]store.ChannelIntegration{
		testIntegration("WeChat", false, "ANJING_CHANNEL_WECHAT_SECRET", 300, true),
	})), testConfig())

	timestamp := "2026-06-14T02:10:00Z"
	content := "你好"
	signature := ChannelSignature("WeChat", "wx-open-disabled", timestamp, content)
	body := strings.NewReader(`{"channel":"WeChat","externalConversationId":"wx-open-disabled","customer":"微信客户","content":"` + content + `","timestamp":"` + timestamp + `","signature":"` + signature + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/channels/inbound", body)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"code":"channel_disabled"`) {
		t.Fatalf("expected disabled channel error, got %s", rec.Body.String())
	}
}

func TestInboundRouteUsesIntegrationSignatureWindow(t *testing.T) {
	mux := http.NewServeMux()
	RegisterWithConfig(mux, store.NewSeedStore(store.WithChannelIntegrations([]store.ChannelIntegration{
		testIntegration("WeChat", true, "ANJING_CHANNEL_WECHAT_SECRET", 20, true),
	})), testConfig())

	timestamp := "2026-06-14T02:10:00Z"
	content := "你好"
	signature := ChannelSignature("WeChat", "wx-open-window", timestamp, content)
	body := strings.NewReader(`{"channel":"WeChat","externalConversationId":"wx-open-window","customer":"微信客户","content":"` + content + `","timestamp":"` + timestamp + `","signature":"` + signature + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/channels/inbound", body)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"code":"stale_signature"`) {
		t.Fatalf("expected stale signature error, got %s", rec.Body.String())
	}
}

func testConfig() Config {
	return Config{
		Secrets:         defaultChannelSecrets,
		SignatureWindow: 5 * time.Minute,
		Now: func() time.Time {
			return time.Date(2026, 6, 14, 2, 10, 30, 0, time.UTC)
		},
	}
}

func testIntegration(channel string, enabled bool, secretRef string, windowSeconds int, replayProtection bool) store.ChannelIntegration {
	return store.ChannelIntegration{
		Channel:                channel,
		DisplayName:            channel,
		Enabled:                enabled,
		SecretSource:           "env",
		SecretRef:              secretRef,
		SignatureWindowSeconds: windowSeconds,
		ReplayProtection:       replayProtection,
		RateLimitPerMinute:     60,
	}
}
