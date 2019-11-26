package timeseries

import (
	"encoding/json"
	"log"
	"net/http"

	"gitlab.flora.loc/mills/tondb/internal/ton/view/timeseries"

	"github.com/julienschmidt/httprouter"
)

type MessagesByType struct {
	q *timeseries.MessagesByType
}

func (api *MessagesByType) Handler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	res, err := api.q.GetMessagesByType()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":true,"message":"error retrieve timeseries"}`))
		return
	}

	resp, err := json.Marshal(res)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":true,"message":"error serialize timeseries"}`))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func NewMessagesByType(q *timeseries.MessagesByType) *MessagesByType {
	return &MessagesByType{
		q: q,
	}
}
