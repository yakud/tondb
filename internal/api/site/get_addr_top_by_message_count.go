package site

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"gitlab.flora.loc/mills/tondb/internal/ton/view/stats"
)

const defaultTopAddrCount = 50

type AddrTopByMessageCountResponse struct {
	TopIn  []stats.AddrCount `json:"top_in"`
	TopOut []stats.AddrCount `json:"top_out"`
}

type GetAddrTopByMessageCount struct {
	q *stats.AddrMessagesCount
}

func (api *GetAddrTopByMessageCount) Handler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	topIn, topOut, err := api.q.SelectTopMessagesCount(defaultTopAddrCount)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":true,"message":"error retrieve top accounts from DB"}`))
		return
	}

	resp, err := json.Marshal(AddrTopByMessageCountResponse{topIn, topOut})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":true,"message":"error serialize response"}`))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func NewGetAddrTopByMessageCount(q *stats.AddrMessagesCount) *GetAddrTopByMessageCount {
	return &GetAddrTopByMessageCount{
		q: q,
	}
}
