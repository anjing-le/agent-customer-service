package knowledge

import (
	"net/http"

	"github.com/anjing-le/agent-customer-service/internal/platform/httpjson"
	"github.com/anjing-le/agent-customer-service/internal/platform/store"
)

func Register(mux *http.ServeMux, st store.Runtime) {
	mux.HandleFunc("/api/knowledge/articles", func(w http.ResponseWriter, r *http.Request) {
		if !httpjson.RequireMethod(w, r, http.MethodGet) {
			return
		}
		items, err := st.ListKnowledge()
		if err != nil {
			httpjson.Fail(w, http.StatusInternalServerError, "store_error", err.Error())
			return
		}
		httpjson.OK(w, items)
	})

	mux.HandleFunc("/api/knowledge/search", func(w http.ResponseWriter, r *http.Request) {
		if !httpjson.RequireMethod(w, r, http.MethodPost) {
			return
		}
		var req struct {
			Query string `json:"query"`
		}
		if err := httpjson.Decode(r, &req); err != nil {
			httpjson.BadRequest(w, err.Error())
			return
		}
		items, err := st.SearchKnowledge(req.Query)
		if err != nil {
			httpjson.Fail(w, http.StatusInternalServerError, "store_error", err.Error())
			return
		}
		httpjson.OK(w, items)
	})

	mux.HandleFunc("/api/knowledge/gaps/resolve", func(w http.ResponseWriter, r *http.Request) {
		if !httpjson.RequireMethod(w, r, http.MethodPost) {
			return
		}
		var req struct {
			ID string `json:"id"`
		}
		if err := httpjson.Decode(r, &req); err != nil {
			httpjson.BadRequest(w, err.Error())
			return
		}
		if req.ID == "" {
			httpjson.BadRequest(w, "id is required")
			return
		}
		gap, err := st.ResolveKnowledgeGap(req.ID)
		if err != nil {
			httpjson.Fail(w, http.StatusInternalServerError, "store_error", err.Error())
			return
		}
		httpjson.OK(w, gap)
	})

	mux.HandleFunc("/api/knowledge/gaps/create-article", func(w http.ResponseWriter, r *http.Request) {
		if !httpjson.RequireMethod(w, r, http.MethodPost) {
			return
		}
		var req struct {
			GapID    string   `json:"gapId"`
			Title    string   `json:"title"`
			Category string   `json:"category"`
			Content  string   `json:"content"`
			Tags     []string `json:"tags"`
		}
		if err := httpjson.Decode(r, &req); err != nil {
			httpjson.BadRequest(w, err.Error())
			return
		}
		if req.GapID == "" {
			httpjson.BadRequest(w, "gapId is required")
			return
		}
		article, err := st.CreateArticleFromGap(req.GapID, req.Title, req.Category, req.Content, req.Tags)
		if err != nil {
			httpjson.Fail(w, http.StatusInternalServerError, "store_error", err.Error())
			return
		}
		httpjson.Created(w, article)
	})
}
