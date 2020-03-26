package timeseries

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"gitlab.flora.loc/mills/tondb/internal/ton/query/timeseries"
)

type BlocksByWorkchain struct {
	q *timeseries.GetBlocksByWorkchain
}

func (api *BlocksByWorkchain) GetV1TimeseriesBlocksByWorkchain(ctx echo.Context) error {
	blocksByWorkchain, err := api.q.GetBlocksByWorkchain()
	if err != nil {
		log.Println(err)
		return ctx.JSONBlob(http.StatusBadRequest, []byte(`{"error":true,"message":"error retrieve timeseries"}`))
	}

	return ctx.JSON(http.StatusOK, blocksByWorkchain)
}

func NewBlocksByWorkchain(q *timeseries.GetBlocksByWorkchain) *BlocksByWorkchain {
	return &BlocksByWorkchain{
		q: q,
	}
}
