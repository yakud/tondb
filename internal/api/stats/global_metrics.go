package stats

import (
	"github.com/labstack/echo/v4"
	"gitlab.flora.loc/mills/tondb/internal/ton/query/stats"
	"log"
	"net/http"
)

type GlobalMetrics struct {
	q *stats.GlobalMetrics
}

func (api *GlobalMetrics) GetV1StatsGlobal(ctx echo.Context) error {
	res, err := api.q.GetGlobalMetrics()
	if err != nil {
		log.Println(err)
		return ctx.JSONBlob(http.StatusBadRequest, []byte(`{"error":true,"message":"error retrieving global metrics"}`))
	}

	return ctx.JSON(http.StatusOK, res)
}

func NewGlobalMetrics(q *stats.GlobalMetrics) *GlobalMetrics {
	return &GlobalMetrics{
		q: q,
	}
}
