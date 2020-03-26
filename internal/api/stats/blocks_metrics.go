package stats

import (
	"github.com/labstack/echo/v4"
	"gitlab.flora.loc/mills/tondb/internal/ton/query/stats"
	"gitlab.flora.loc/mills/tondb/swagger/tonapi"
	"log"
	"math"
	"net/http"
)

type BlocksMetrics struct {
	q *stats.BlocksMetrics
}

func (api *BlocksMetrics) GetV1StatsBlocks(ctx echo.Context, params tonapi.GetV1StatsBlocksParams) error {
	var wcId string
	if params.WorkchainId != nil {
		wcId = * params.WorkchainId
	}

	res, err := api.q.GetBlocksMetrics(wcId)
	if err != nil {
		log.Println(err)
		return ctx.JSONBlob(http.StatusBadRequest, []byte(`{"error":true,"message":"error retrieving blocks metrics"}`))
	}

	if math.IsNaN(res.AvgBlockTime) {
		res.AvgBlockTime = 0
	}

	return ctx.JSON(http.StatusOK, res)
}

func NewBlocksMetrics(q *stats.BlocksMetrics) *BlocksMetrics {
	return &BlocksMetrics{
		q: q,
	}
}

