package stats

import (
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"gitlab.flora.loc/mills/tondb/internal/ton/query/stats"
	"log"
	"math"
	"net/http"
)

type BlocksMetrics struct {
	q *stats.BlocksMetrics
}

func (api *BlocksMetrics) Handler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var wcId string
	wcIds, ok := r.URL.Query()["workchain_id"]
	if ok {
		if len(wcIds) > 1 {
			http.Error(w, `{"error":true,"message":"only one workchain_id parameter can be set"}`, http.StatusBadRequest)
			return
		}
		wcId = wcIds[0]
	}

	res, err := api.q.GetBlocksMetrics(wcId)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":true,"message":"error retrieving blocks metrics"}`))
		return
	}

	if math.IsNaN(res.AvgBlockTime) {
		res.AvgBlockTime = 0
	}

	resp, err := json.Marshal(res)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":true,"message":"error serializing blocks metrics"}`))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func NewBlocksMetrics(q *stats.BlocksMetrics) *BlocksMetrics {
	return &BlocksMetrics{
		q: q,
	}
}

