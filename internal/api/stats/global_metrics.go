package stats

import (
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"gitlab.flora.loc/mills/tondb/internal/ton/query/stats"
	"log"
	"net/http"
)

type GlobalMetrics struct {
	q *stats.GlobalMetrics
}

func (api *GlobalMetrics) Handler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	res, err := api.q.GetGlobalMetrics()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":true,"message":"error retrieving global metrics"}`))
		return
	}

	resp, err := json.Marshal(res)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":true,"message":"error serializing global metrics"}`))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func NewGlobalMetrics(q *stats.GlobalMetrics) *GlobalMetrics {
	return &GlobalMetrics{
		q: q,
	}
}
