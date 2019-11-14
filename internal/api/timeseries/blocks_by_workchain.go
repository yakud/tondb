package timeseries

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"gitlab.flora.loc/mills/tondb/internal/ton/query/timeseries"
)

type BlocksByWorkchain struct {
	q *timeseries.GetBlocksByWorkchain
}

func (api *BlocksByWorkchain) Handler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	blocksByWorkchain, err := api.q.GetBlocksByWorkchain()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":true,"message":"error retrieve timeseries"}`))
		return
	}

	resp, err := json.Marshal(blocksByWorkchain)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":true,"message":"error serialize timeseries"}`))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func NewBlocksByWorkchain(q *timeseries.GetBlocksByWorkchain) *BlocksByWorkchain {
	return &BlocksByWorkchain{
		q: q,
	}
}
