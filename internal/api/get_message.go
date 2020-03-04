package api

import (
	"github.com/julienschmidt/httprouter"
	"gitlab.flora.loc/mills/tondb/internal/ton/query"
	httputils "gitlab.flora.loc/mills/tondb/internal/utils/http"
	"log"
	"net/http"
)

type GetMessage struct {
	q *query.GetMessage
}

func (m *GetMessage) Handler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	trxHash, err := httputils.GetQueryValueString(r.URL, "trx_hash")
	if err != nil{
		http.Error(w, `{"error":true,"message":` + err.Error() + `}`, http.StatusBadRequest)
		return
	}
    if len(trxHash) != 64 {
		http.Error(w, `{"error":true,"message":"trx_hash must contain exactly 64 symbols"}`, http.StatusBadRequest)
		return
	}

	messageLt, err := httputils.GetQueryValueUint(r.URL, "message_lt", 64)
	if err != nil {
		http.Error(w, `{"error":true,"message":` + err.Error() + `}`, http.StatusBadRequest)
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