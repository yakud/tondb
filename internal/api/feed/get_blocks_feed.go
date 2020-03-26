package feed

import (
	"gitlab.flora.loc/mills/tondb/swagger/tonapi"
	"net/http"

	"gitlab.flora.loc/mills/tondb/internal/ton/view/feed"

	"github.com/labstack/echo/v4"
)

const (
	defaultBlocksFeedCount = 30
)

type GetBlocksFeed struct {
	f *feed.BlocksFeed
}

func (m *GetBlocksFeed) GetV1BlocksFeed(ctx echo.Context, params tonapi.GetV1BlocksFeedParams) error {
	var err error

	// limit
	var limit uint16
	if params.Limit == nil {
		limit = defaultBlocksFeedCount
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
	var scrollId = &feed.BlocksFeedScrollId{}
	if params.ScrollId != nil && len(*params.ScrollId) > 0 {
		if err := UnpackScrollId(*params.ScrollId, scrollId); err != nil {
			return ctx.JSONBlob(http.StatusBadRequest, []byte(`{"error":true,"message":"error unpacking scroll_id"}`))
		}
	} else {
		scrollId.WorkchainId = workchainId
	}

	blocksFeed, newScrollId, err := m.f.SelectBlocks(scrollId, limit)
	if err != nil {
		return ctx.JSONBlob(http.StatusInternalServerError, []byte(`{"error":true,"message":"error fetching blocks"}`))
	}

	newPackedScrollId, err := PackScrollId(newScrollId)
	if err != nil {
		return ctx.JSONBlob(http.StatusInternalServerError, []byte(`{"error":true,"message":"error packing scroll_id"}`))
	}

	respFeed := tonapi.BlocksFeedResponse{
		Blocks:   &blocksFeed,
		ScrollId: &newPackedScrollId,
	}

	return ctx.JSON(200, respFeed)
}

func NewGetBlocksFeed(f *feed.BlocksFeed) *GetBlocksFeed {
	return &GetBlocksFeed{
		f: f,
	}
}
