package api

import (
	"fmt"
	"log"
	"net/http"

	apiFilter "gitlab.flora.loc/mills/tondb/internal/api/filter"

	"gitlab.flora.loc/mills/tondb/internal/ton/storage"

	"github.com/julienschmidt/httprouter"
	"gitlab.flora.loc/mills/tondb/internal/ton/query"
)

type GetTransactions struct {
	q                  *query.SearchTransactions
	shardsDescrStorage *storage.ShardsDescr
}

func (m *GetTransactions) Handler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	trFilter, err := apiFilter.TransactionHashByIndexFilterFromRequest(r, "hash")
	if err != nil {
		http.Error(w, `{"error":true,"message":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	if trFilter == nil {
		http.Error(w, `{"error":true,"message":"empty hash filter"}`, http.StatusBadRequest)
		return
	}

	blocksTransactions, err := m.q.SearchByFilter(trFilter)
	if err != nil {
		log.Println(fmt.Errorf("query SearchByFilter error: %w", err))
		http.Error(w, `{"error":true,"message":"SearchByFilter query error"}`, http.StatusBadRequest)
		return
	}

	resp, err := json.Marshal(blocksTransactions)
	if err != nil {
		http.Error(w, `{"error":true,"message":"response json marshaling error"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
	w.Write(resp)
}

func NewGetTransactions(q *query.SearchTransactions) *GetTransactions {
	return &GetTransactions{
		q: q,
	}
}
