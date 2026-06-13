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
			httpjson.OK(w, st.ListConversations())
		case http.MethodPost:
			var req struct {
				Customer string `json:"customer"`
				Channel  string `json:"channel"`
			}
			if err := httpjson.Decode(r, &req); err != nil {
				httpjson.BadRequest(w, err.Error())
				return
			}
			httpjson.Created(w, st.CreateConversation(req.Customer, req.Channel))
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
		httpjson.OK(w, st.SendMessage(req.ConversationID, req.Content))
	})
}
