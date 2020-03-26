package timeseries

import (
	"log"
	"net/http"

	"gitlab.flora.loc/mills/tondb/internal/ton/view/timeseries"

	"github.com/labstack/echo/v4"
)

type MessagesOrdCount struct {
	q *timeseries.MessagesOrdCount
}

func (api *MessagesOrdCount) GetV1TimeseriesMessagesOrdCount(ctx echo.Context) error {
	res, err := api.q.GetMessagesOrdCount()
	if err != nil {
		log.Println(err)
		return ctx.JSONBlob(http.StatusBadRequest, []byte(`{"error":true,"message":"error retrieving timeseries"}`))
	}

	return ctx.JSON(http.StatusOK, res)
}

func NewMessagesOrdCount(q *timeseries.MessagesOrdCount) *MessagesOrdCount {
	return &MessagesOrdCount{
		q: q,
	}
}
