package stats

import (
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"gitlab.flora.loc/mills/tondb/internal/ton/query/stats"
	"log"
	"net/http"
)

type AddressesMetrics struct {
	q *stats.AddressesMetrics
}

func (api *AddressesMetrics) Handler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var wcId string
	wcIds, ok := r.URL.Query()["workchain_id"]
	if ok {
		if len(wcIds) > 1 {
			http.Error(w, `{"error":true,"message":"only one workchain_id parameter can be set"}`, http.StatusBadRequest)
			return
		}
		wcId = wcIds[0]
	}

	res, err := api.q.GetAddressesMetrics(wcId)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":true,"message":"error retrieving addresses metrics"}`))
		return
	}

	resp, err := json.Marshal(res)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":true,"message":"error serializing addresses metrics"}`))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func NewAddressesMetrics(q *stats.AddressesMetrics) *AddressesMetrics {
	return &AddressesMetrics{
		q: q,
	}
}

