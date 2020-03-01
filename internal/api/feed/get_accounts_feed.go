package feed

import (
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"gitlab.flora.loc/mills/tondb/internal/ton/view/feed"
	httputils "gitlab.flora.loc/mills/tondb/internal/utils/http"
	"log"
	"net/http"
)

const defaultLatestAccountsCount = 50

type GetAccountsFeed struct {
	q *feed.AccountsFeed
}

func (api *GetAccountsFeed) Handler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var err error

	// limit
	limit, err := httputils.GetQueryValueUint16(r.URL, "limit")
	if err != nil {
		limit = defaultLatestAccountsCount
	}

	// offset
	offset, err := httputils.GetQueryValueUint32(r.URL, "offset")
	if err != nil {
		offset = 0
	}

	// workchain_id
	workchainId, err := httputils.GetQueryValueInt32(r.URL, "workchain_id")
	if err != nil {
		workchainId = feed.EmptyWorkchainId
	}

	feed, err := api.q.SelectAccounts(workchainId, limit, offset)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":true,"message":"error retrieving accounts from DB"}`))
		return
	}

	respJson, err := json.Marshal(&feed)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":true,"message":"error serializing response"}`))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(respJson)
}

func NewGetAccountsFeed(q *feed.AccountsFeed) *GetAccountsFeed {
	return &GetAccountsFeed{
		q: q,
	}
}
