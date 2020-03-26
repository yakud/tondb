package stats

import (
	"github.com/labstack/echo/v4"
	"gitlab.flora.loc/mills/tondb/internal/ton/query/stats"
	"gitlab.flora.loc/mills/tondb/swagger/tonapi"
	"log"
	"net/http"
)

type TrxMetrics struct {
	q *stats.TrxMetrics
}

func (api *TrxMetrics) GetV1StatsTransactions(ctx echo.Context, params tonapi.GetV1StatsTransactionsParams) error {
	var wcId string
	if params.WorkchainId != nil {
		wcId = * params.WorkchainId
	}

	res, err := api.q.GetTrxMetrics(wcId)
	if err != nil {
		log.Println(err)
		return ctx.JSONBlob(http.StatusBadRequest, []byte(`{"error":true,"message":"error retrieving blocks metrics"}`))
	}

	return ctx.JSON(http.StatusOK, res)
}

func NewTrxMetrics(q *stats.TrxMetrics) *TrxMetrics {
	return &TrxMetrics{
		q: q,
	}
}

