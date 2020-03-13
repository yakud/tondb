package feed

import (
	"encoding/json"
	"log"
	"net/http"

	httputils "gitlab.flora.loc/mills/tondb/internal/utils/http"

	"gitlab.flora.loc/mills/tondb/internal/ton/view/feed"

	"github.com/julienschmidt/httprouter"
)

const defaultLatestTransactionsCount = 50

type GetTransactionsFeedResponse struct {
	Transactions []*feed.TransactionInFeed `json:"transactions"`
	ScrollId     string                    `json:"scroll_id"`
}

type GetTransactionsFeed struct {
	q *feed.TransactionsFeed
}

func (api *GetTransactionsFeed) Handler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var err error

	// limit
	limit, err := httputils.GetQueryValueUint16(r.URL, "limit")
	if err != nil {
		limit = defaultLatestMessagesCount
	}

	// workchain_id
	workchainId, err := httputils.GetQueryValueInt32(r.URL, "workchain_id")
	if err != nil {
		workchainId = feed.EmptyWorkchainId
	}

	// scroll_id
	var scrollId = &feed.TransactionsFeedScrollId{}
	packedScrollId, err := httputils.GetQueryValueString(r.URL, "scroll_id")
	if err == nil && len(packedScrollId) > 0 {
		if err := UnpackScrollId(packedScrollId, scrollId); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":true,"message":"error unpack scroll_id"}`))
			return
		}
	} else {
		scrollId.WorkchainId = workchainId
	}

	transactionsFeed, newScrollId, err := api.q.SelectTransactions(scrollId, limit)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":true,"message":"error retrieve transactions from DB"}`))
		return
	}
	newPackedScrollId, err := PackScrollId(newScrollId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":true,"message":"error pack scroll_id"}`))
		return
	}

	resp := GetTransactionsFeedResponse{
		Transactions: transactionsFeed,
		ScrollId:     newPackedScrollId,
	}

	respJson, err := json.Marshal(&resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":true,"message":"error serialize response"}`))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(respJson)
}

func NewGetTransactionsFeed(q *feed.TransactionsFeed) *GetTransactionsFeed {
	return &GetTransactionsFeed{
		q: q,
	}
}
