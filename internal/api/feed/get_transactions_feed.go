package feed

import (
	"gitlab.flora.loc/mills/tondb/swagger/tonapi"
	"log"
	"net/http"

	"gitlab.flora.loc/mills/tondb/internal/ton/view/feed"

	"github.com/labstack/echo/v4"
)

const defaultLatestTransactionsCount = 50

type GetTransactionsFeed struct {
	q *feed.TransactionsFeed
}

func (api *GetTransactionsFeed) GetV1TransactionsFeed(ctx echo.Context, params tonapi.GetV1TransactionsFeedParams) error {
	// limit
	var limit uint16
	if params.Limit == nil {
		limit = defaultLatestMessagesCount
	} else {
		limit = uint16(*params.Limit)
	}

	// workchain_id
	var workchainId int32
	if params.WorkchainId == nil {
		workchainId = feed.EmptyWorkchainId
	} else {
		workchainId = *params.WorkchainId
	}

	// scroll_id
	var scrollId = &feed.TransactionsFeedScrollId{}
	if params.ScrollId != nil && len(*params.ScrollId) > 0 {
		if err := UnpackScrollId(*params.ScrollId, scrollId); err != nil {
			return ctx.JSONBlob(http.StatusBadRequest, []byte(`{"error":true,"message":"error unpacking scroll_id"}`))
		}
	} else {
		scrollId.WorkchainId = workchainId
	}

	transactionsFeed, newScrollId, err := api.q.SelectTransactions(scrollId, limit)
	if err != nil {
		log.Println(err)
		return ctx.JSONBlob(http.StatusInternalServerError, []byte(`{"error":true,"message":"error retrieving transactions from DB"}`))
	}
	newPackedScrollId, err := PackScrollId(newScrollId)
	if err != nil {
		return ctx.JSONBlob(http.StatusInternalServerError, []byte(`{"error":true,"message":"error packing scroll_id"}`))
	}

	resp := tonapi.TransactionsFeedResponse{
		Transactions: transactionsFeed,
		ScrollId:     newPackedScrollId,
	}

	return ctx.JSON(http.StatusOK, resp)
}

func NewGetTransactionsFeed(q *feed.TransactionsFeed) *GetTransactionsFeed {
	return &GetTransactionsFeed{
		q: q,
	}
}
