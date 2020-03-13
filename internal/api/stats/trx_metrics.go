package stats

import (
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"gitlab.flora.loc/mills/tondb/internal/ton/query/stats"
	httputils "gitlab.flora.loc/mills/tondb/internal/utils/http"
	"log"
	"net/http"
)

type TrxMetrics struct {
	q *stats.TrxMetrics
}

func (api *TrxMetrics) Handler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	wcId, err := httputils.GetQueryValueString(r.URL, "workchain_id")
	if err != nil {
		wcId = ""
	}

	res, err := api.q.GetTrxMetrics(wcId)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":true,"message":"error retrieving blocks metrics"}`))
		return
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

func NewTrxMetrics(q *stats.TrxMetrics) *TrxMetrics {
	return &TrxMetrics{
		q: q,
	}
}

