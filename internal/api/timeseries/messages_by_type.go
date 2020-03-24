package timeseries

import (
	"log"
	"net/http"

	"gitlab.flora.loc/mills/tondb/internal/ton/view/timeseries"

	"github.com/labstack/echo/v4"
)

type MessagesByType struct {
	q *timeseries.MessagesByType
}

func (api *MessagesByType) GetV1TimeseriesMessagesByType(ctx echo.Context) error {
	res, err := api.q.GetMessagesByType()
	if err != nil {
		log.Println(err)
		return ctx.JSONBlob(http.StatusBadRequest, []byte(`{"error":true,"message":"error retrieving timeseries"}`))
	}

	return ctx.JSON(http.StatusOK, res)
}

func NewMessagesByType(q *timeseries.MessagesByType) *MessagesByType {
	return &MessagesByType{
		q: q,
	}
}
