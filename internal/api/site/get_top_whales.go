package site

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"gitlab.flora.loc/mills/tondb/internal/ton/query/stats"
)

type GetTopWhales struct {
	q *stats.GetTopWhales
}

func (api *GetTopWhales) Handler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	topWhales, err := api.q.GetTopWhales()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":true,"message":"error retrieve top whales from DB"}`))
		return
	}

	resp, err := json.Marshal(*topWhales)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":true,"message":"error serialize response"}`))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func NewGetTopWhales(q *stats.GetTopWhales) *GetTopWhales {
	return &GetTopWhales{
		q: q,
	}
}
