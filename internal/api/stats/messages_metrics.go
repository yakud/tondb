package stats

import (
	"github.com/labstack/echo/v4"
	"gitlab.flora.loc/mills/tondb/internal/ton/query/stats"
	"gitlab.flora.loc/mills/tondb/swagger/tonapi"
	"log"
	"math"
	"net/http"
)

type MessagesMetrics struct {
	q *stats.MessagesMetrics
}

func (api *MessagesMetrics) GetV1StatsMessages(ctx echo.Context, params tonapi.GetV1StatsMessagesParams) error {
	var wcId string
	if params.WorkchainId != nil {
		wcId = * params.WorkchainId
	}

	res, err := api.q.GetMessagesMetrics(wcId)
	if err != nil {
		log.Println(err)
		return ctx.JSONBlob(http.StatusBadRequest, []byte(`{"error":true,"message":"error retrieving messages metrics"}`))
	}

	if math.IsNaN(res.Tps) {
		res.Tps = 0
	}

	if math.IsNaN(res.Mps) {
		res.Mps = 0
	}

	return ctx.JSON(http.StatusOK, res)
}

func NewMessagesMetrics(q *stats.MessagesMetrics) *MessagesMetrics {
	return &MessagesMetrics{
		q: q,
	}
}

