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
	}
}
