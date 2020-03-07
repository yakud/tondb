package timeseries

import (
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"gitlab.flora.loc/mills/tondb/internal/ton/query/timeseries"
	"log"
	"net/http"
)

type SentAndFees struct {
	q *timeseries.SentAndFees
}

func (api *SentAndFees) Handler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	res, err := api.q.GetSentAndFees()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":true,"message":"error retrieving average sent and fees"}`))
		return
	}

	resp, err := json.Marshal(res)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":true,"message":"error serializing average sent and fees"}`))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func NewSentAndFees(q *timeseries.SentAndFees) *SentAndFees {
	return &SentAndFees{
		q: q,
	}
}
