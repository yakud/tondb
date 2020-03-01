package api

import (
	"github.com/julienschmidt/httprouter"
	"gitlab.flora.loc/mills/tondb/internal/ton/query"
	"log"
	"net/http"
	"strconv"
)

type GetMessage struct {
	q *query.GetMessage
}

func (m *GetMessage) Handler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	trxHash := p.ByName("trx_hash")
	messageLt, err := strconv.ParseUint(p.ByName("message_lt"), 10, 64)
	if err != nil {
		http.Error(w, `{"error":true,"message":"Can't convert message_lt to uint64'"}`, http.StatusBadRequest)
		return
	}

	message, err := m.q.SelectMessage(trxHash, messageLt)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":true,"message":"error retrieving message from DB"}`))
		return
	}

	respJson, err := json.Marshal(&message)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":true,"message":"error serializing response"}`))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(respJson)
}

func NewGetMessage(q *query.GetMessage) *GetMessage {
	return &GetMessage{
		q: q,
	}
}

