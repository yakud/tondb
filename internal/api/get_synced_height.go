package api

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/yakud/ton-blocks-stream-receiver/internal/ton/query"
)

type GetSyncedHeight struct {
	q *query.GetSyncedHeight
}

func (api *GetSyncedHeight) Handler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	lastSyncedBlock, err := api.q.GetSyncedHeight()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":true,"message":"error retrieve synced height from DB"}`))
		return
	}

	resp, err := json.Marshal(lastSyncedBlock)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":true,"message":"error serialize response"}`))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func NewGetSyncedHeight(q *query.GetSyncedHeight) *GetSyncedHeight {
	return &GetSyncedHeight{
		q: q,
	}
}
