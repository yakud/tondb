package site

import (
	"gitlab.flora.loc/mills/tondb/internal/ton/view/feed"
	"gitlab.flora.loc/mills/tondb/swagger/tonapi"
	"net/http"

	"github.com/labstack/echo/v4"
	"gitlab.flora.loc/mills/tondb/internal/ton/query/stats"
)

type GetTopWhales struct {
	q *stats.GetTopWhales
}

func (api *GetTopWhales) GetV1TopWhales(ctx echo.Context, params tonapi.GetV1TopWhalesParams) error {
	var wcId int32
	if params.WorkchainId == nil {
		wcId = int32(feed.EmptyWorkchainId)
	} else {
		wcId = *params.WorkchainId
	}

	var limit uint32
	if params.Limit == nil {
		limit = uint32(stats.WhalesDefaultPageLimit)
	} else {
		limit = uint32(*params.Limit)
	}

	var offset uint32
	if params.Offset != nil {
		offset = uint32(*params.Offset)
	}

	topWhales, err := api.q.GetTopWhales(wcId, limit, offset)
	if err != nil {
		return ctx.JSONBlob(http.StatusBadRequest, []byte(`{"error":true,"message":"error retrieving top whales"}`))
	}

	return ctx.JSON(http.StatusOK, topWhales)
}

func NewGetTopWhales(q *stats.GetTopWhales) *GetTopWhales {
	return &GetTopWhales{
		q: q,
	}
}
