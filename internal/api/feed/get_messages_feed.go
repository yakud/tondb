package feed

import (
	"gitlab.flora.loc/mills/tondb/swagger/tonapi"
	"log"
	"net/http"

	"gitlab.flora.loc/mills/tondb/internal/ton/view/feed"

	"github.com/labstack/echo/v4"
)

const defaultLatestMessagesCount = 50

type GetMessagesFeedResponse struct {
	Messages []*feed.MessageInFeed `json:"messages"`
	ScrollId string                `json:"scroll_id"`
}

type GetMessagesFeed struct {
	q *feed.MessagesFeed
}

func (api *GetMessagesFeed) GetV1MessagesFeed(ctx echo.Context, params tonapi.GetV1MessagesFeedParams) error {
	var err error

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
	var scrollId = &feed.MessagesFeedScrollId{}
	if params.ScrollId != nil && len(*params.ScrollId) > 0 {
		if err := UnpackScrollId(*params.ScrollId, scrollId); err != nil {
			return ctx.JSONBlob(http.StatusBadRequest, []byte(`{"error":true,"message":"error unpacking scroll_id"}`))
		}
	} else {
		scrollId.WorkchainId = workchainId
	}

	messagesFeed, newScrollId, err := api.q.SelectMessages(scrollId, limit)
	if err != nil {
		log.Println(err)
		return ctx.JSONBlob(http.StatusInternalServerError, []byte(`{"error":true,"message":"error retrieving messages from DB"}`))
	}
	newPackedScrollId, err := PackScrollId(newScrollId)
	if err != nil {
		return ctx.JSONBlob(http.StatusInternalServerError, []byte(`{"error":true,"message":"error packing scroll_id"}`))
	}

	resp := tonapi.MessageFeedResponse{
		Messages: messagesFeed,
		ScrollId: newPackedScrollId,
	}

	return ctx.JSON(http.StatusOK, resp)
}

func NewGetMessagesFeed(q *feed.MessagesFeed) *GetMessagesFeed {
	return &GetMessagesFeed{
		q: q,
	}
}
