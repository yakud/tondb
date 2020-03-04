package api

import (
	"fmt"
	"log"
	"net/http"

	httputils "gitlab.flora.loc/mills/tondb/internal/utils/http"

	"gitlab.flora.loc/mills/tondb/internal/ton/search"

	"github.com/julienschmidt/httprouter"
)

type Search struct {
	searcher *search.Searcher
}

func (m *Search) Handler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	searchQuery, err := httputils.GetQueryValueString(r.URL, "q")
	if err != nil || searchQuery == "" {
		http.Error(w, `{"error":true,"message":"empty search query"}`, http.StatusBadRequest)
		return
	}

	searchResult, err := m.searcher.Search(searchQuery)
	if err != nil {
		log.Println(fmt.Errorf("searcher error: %w", err))
		http.Error(w, `{"error":true,"message":"search error"}`, http.StatusInternalServerError)
		return
	}

	if len(searchResult) == 0 {
		w.WriteHeader(http.StatusNoContent)
		w.Write([]byte(`{"result":[]}`))
		return
	}

	resp, err := json.Marshal(searchResult)
	if err != nil {
		http.Error(w, `{"error":true,"message":"response json marshaling error"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
	w.Write(resp)
}

func NewSearch(searcher *search.Searcher) *Search {
	return &Search{
		searcher: searcher,
	}
}
