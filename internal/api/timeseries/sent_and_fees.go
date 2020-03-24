package timeseries

import (
	"github.com/labstack/echo/v4"
	"gitlab.flora.loc/mills/tondb/internal/ton/query/timeseries"
	"log"
	"net/http"
)

type SentAndFees struct {
	q *timeseries.SentAndFees
}

func (api *SentAndFees) GetV1TimeseriesSentAndFees(ctx echo.Context) error {
	res, err := api.q.GetSentAndFees()
	if err != nil {
		log.Println(err)
		return ctx.JSONBlob(http.StatusBadRequest, []byte(`{"error":true,"message":"error retrieving average sent and fees"}`))
	}

	return ctx.JSON(http.StatusOK, res)
}

func NewSentAndFees(q *timeseries.SentAndFees) *SentAndFees {
	return &SentAndFees{
		q: q,
	}
}
