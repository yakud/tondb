package stats

import (
	"github.com/labstack/echo/v4"
	"gitlab.flora.loc/mills/tondb/internal/ton/query/stats"
	"gitlab.flora.loc/mills/tondb/swagger/tonapi"
	"log"
	"net/http"
)

type AddressesMetrics struct {
	q *stats.AddressesMetrics
}

func (api *AddressesMetrics) GetV1StatsAddresses(ctx echo.Context, params tonapi.GetV1StatsAddressesParams) error {
	var wcId string
	if params.WorkchainId != nil {
		wcId = * params.WorkchainId
	}

	res, err := api.q.GetAddressesMetrics(wcId)
	if err != nil {
		log.Println(err)
		return ctx.JSONBlob(http.StatusBadRequest, []byte(`{"error":true,"message":"error retrieving addresses metrics"}`))
	}

	return ctx.JSON(http.StatusOK, res)
}

func NewAddressesMetrics(q *stats.AddressesMetrics) *AddressesMetrics {
	return &AddressesMetrics{
		q: q,
	}
}

