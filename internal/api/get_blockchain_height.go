package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/yakud/ton-blocks-stream-receiver/internal/ton/query"
)

type GetBlockchainHeight struct {
	q *query.GetBlockchainHeight
}

func (api *GetBlockchainHeight) Handler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	lastSyncedBlock, err := api.q.GetBlockchainHeight()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":true,"message":"error retrieve blockchain height from DB"}`))
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

func NewGetBlockchainHeight(q *query.GetBlockchainHeight) *GetBlockchainHeight {
	return &GetBlockchainHeight{
		q: q,
	}
}
