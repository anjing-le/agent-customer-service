package customer

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/anjing-le/agent-customer-service/internal/platform/httpjson"
	"github.com/anjing-le/agent-customer-service/internal/platform/store"
)

type sendMessageRequest struct {
	ConversationID string `json:"conversationId"`
	Content        string `json:"content"`
}

type streamEvent struct {
	Type string `json:"type"`
	Data any    `json:"data,omitempty"`
}

func Register(mux *http.ServeMux, st store.Runtime) {
	mux.HandleFunc("/api/customer-service/conversations", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			items, err := st.ListConversations()
			if err != nil {
				httpjson.Fail(w, http.StatusInternalServerError, "store_error", err.Error())
				return
			}
			httpjson.OK(w, items)
		case http.MethodPost:
			var req struct {
				Customer string `json:"customer"`
				Channel  string `json:"channel"`
			}
			if err := httpjson.Decode(r, &req); err != nil {
				httpjson.BadRequest(w, err.Error())
				return
			}
			conv, err := st.CreateConversation(req.Customer, req.Channel)
			if err != nil {
				httpjson.Fail(w, http.StatusInternalServerError, "store_error", err.Error())
				return
			}
			httpjson.Created(w, conv)
		default:
			httpjson.MethodNotAllowed(w)
		}
	})

	mux.HandleFunc("/api/customer-service/messages/stream", func(w http.ResponseWriter, r *http.Request) {
		if !httpjson.RequireMethod(w, r, http.MethodPost) {
			return
		}
		req, ok := decodeSendMessageRequest(w, r)
		if !ok {
			return
		}
		result, err := st.SendMessage(req.ConversationID, req.Content)
		if err != nil {
			httpjson.Fail(w, http.StatusInternalServerError, "store_error", err.Error())
			return
		}
		writeMessageStream(w, result)
	})

	mux.HandleFunc("/api/customer-service/messages", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			conversationID := r.URL.Query().Get("conversationId")
			if conversationID == "" {
				httpjson.BadRequest(w, "conversationId is required")
				return
			}
			items, err := st.ListMessages(conversationID)
			if err != nil {
				httpjson.Fail(w, http.StatusInternalServerError, "store_error", err.Error())
				return
			}
			httpjson.OK(w, items)
		case http.MethodPost:
			req, ok := decodeSendMessageRequest(w, r)
			if !ok {
				return
			}
			result, err := st.SendMessage(req.ConversationID, req.Content)
			if err != nil {
				httpjson.Fail(w, http.StatusInternalServerError, "store_error", err.Error())
				return
			}
			httpjson.OK(w, result)
		default:
			httpjson.MethodNotAllowed(w)
		}
	})
}

func decodeSendMessageRequest(w http.ResponseWriter, r *http.Request) (sendMessageRequest, bool) {
	var req sendMessageRequest
	if err := httpjson.Decode(r, &req); err != nil {
		httpjson.BadRequest(w, err.Error())
		return req, false
	}
	if strings.TrimSpace(req.Content) == "" {
		httpjson.BadRequest(w, "content is required")
		return req, false
	}
	return req, true
}

func writeMessageStream(w http.ResponseWriter, result store.SendMessageResult) {
	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	writeSSE(w, "meta", streamEvent{Type: "meta", Data: map[string]any{
		"engine":         result.AgentMessage.Engine,
		"fallbackReason": result.AgentMessage.FallbackReason,
		"evidenceCount":  len(result.Evidence),
		"trace":          result.AgentMessage.Trace,
	}})
	for _, chunk := range splitContent(result.AgentMessage.Content, 18) {
		writeSSE(w, "delta", streamEvent{Type: "delta", Data: map[string]string{"content": chunk}})
	}
	writeSSE(w, "done", streamEvent{Type: "done", Data: result})
}

func writeSSE(w http.ResponseWriter, event string, payload streamEvent) {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return
	}
	_, _ = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, encoded)
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
}

func splitContent(content string, size int) []string {
	if content == "" {
		return []string{""}
	}
	if size <= 0 {
		size = utf8.RuneCountInString(content)
	}
	runes := []rune(content)
	chunks := make([]string, 0, (len(runes)+size-1)/size)
	for start := 0; start < len(runes); start += size {
		end := start + size
		if end > len(runes) {
			end = len(runes)
		}
		chunks = append(chunks, string(runes[start:end]))
	}
	return chunks
}
