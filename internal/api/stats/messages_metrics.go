package stats

import (
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"gitlab.flora.loc/mills/tondb/internal/ton/query/stats"
	"log"
	"math"
	"net/http"
)

type MessagesMetrics struct {
	q *stats.MessagesMetrics
}

func (api *MessagesMetrics) Handler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// I should really move all those repeating parts to some kind of utils
	var wcId string
	wcIds, ok := r.URL.Query()["workchain_id"]
	if ok {
		if len(wcIds) > 1 {
			http.Error(w, `{"error":true,"message":"only one workchain_id parameter can be set"}`, http.StatusBadRequest)
			return
		}
		wcId = wcIds[0]
	}

	res, err := api.q.GetMessagesMetrics(wcId)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":true,"message":"error retrieving messages metrics"}`))
		return
	}

	if math.IsNaN(res.Tps) {
		res.Tps = 0
	}

	if math.IsNaN(res.Mps) {
		res.Mps = 0
	}

	resp, err := json.Marshal(res)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":true,"message":"error serializing messages metrics"}`))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func NewMessagesMetrics(q *stats.MessagesMetrics) *MessagesMetrics {
	return &MessagesMetrics{
		q: q,
	}
}

