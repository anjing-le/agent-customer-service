package channels

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/anjing-le/agent-customer-service/internal/platform/store"
)

func TestInboundRouteAcceptsSignedChannelMessage(t *testing.T) {
	mux := http.NewServeMux()
	Register(mux, store.NewSeedStore())

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
	Register(mux, store.NewSeedStore())

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
