package site

import (
	"encoding/json"
	"gitlab.flora.loc/mills/tondb/internal/ton/view/feed"
	httpUtils "gitlab.flora.loc/mills/tondb/internal/utils/http"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"gitlab.flora.loc/mills/tondb/internal/ton/query/stats"
)

type GetTopWhales struct {
	q *stats.GetTopWhales
}

func (api *GetTopWhales) Handler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	wcId, err := httpUtils.GetQueryValueInt32(r.URL, "workchain_id")
	if err != nil {
		wcId = int32(feed.EmptyWorkchainId)
	}

	limit, err := httpUtils.GetQueryValueUint32(r.URL, "limit")
	if err != nil {
		limit = uint32(stats.WhalesDefaultLimit / 2)
	}

	offset, _ := httpUtils.GetQueryValueUint32(r.URL, "offset")

	topWhales, err := api.q.GetTopWhales(wcId, limit, offset)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":true,"message":"error retrieving top whales"}`))
		return
	}

	resp, err := json.Marshal(*topWhales)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":true,"message":"error serializing response"}`))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func NewGetTopWhales(q *stats.GetTopWhales) *GetTopWhales {
	return &GetTopWhales{
		q: q,
	}
}
