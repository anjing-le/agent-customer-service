package customer

import (
	"net/http"

	"github.com/anjing-le/agent-customer-service/internal/platform/httpjson"
	"github.com/anjing-le/agent-customer-service/internal/platform/store"
)

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

	mux.HandleFunc("/api/customer-service/messages", func(w http.ResponseWriter, r *http.Request) {
		if !httpjson.RequireMethod(w, r, http.MethodPost) {
			return
		}
		var req struct {
			ConversationID string `json:"conversationId"`
			Content        string `json:"content"`
		}
		if err := httpjson.Decode(r, &req); err != nil {
			httpjson.BadRequest(w, err.Error())
			return
		}
		if req.Content == "" {
			httpjson.BadRequest(w, "content is required")
			return
		}
		result, err := st.SendMessage(req.ConversationID, req.Content)
		if err != nil {
			httpjson.Fail(w, http.StatusInternalServerError, "store_error", err.Error())
			return
		}
		httpjson.OK(w, result)
	})
}
