package api

import (
	"github.com/labstack/echo/v4"
	"gitlab.flora.loc/mills/tondb/internal/ton/query"
	"gitlab.flora.loc/mills/tondb/swagger/tonapi"
	"log"
	"net/http"
)

type GetMessage struct {
	q *query.GetMessage
}

func (m *GetMessage) GetV1MessageGet(ctx echo.Context, params tonapi.GetV1MessageGetParams) error {
    if len(params.TrxHash) != 64 {
		return ctx.JSONBlob(http.StatusBadRequest, []byte(`{"error":true,"message":"trx_hash must contain exactly 64 symbols"}`))
	}

	message, err := m.q.SelectMessage(params.TrxHash, uint64(params.MessageLt))
	if err != nil {
		log.Println(err)
		return ctx.JSONBlob(http.StatusInternalServerError, []byte(`{"error":true,"message":"error retrieving message from DB"}`))
	}

	return ctx.JSON(http.StatusOK, message)
}

func NewGetMessage(q *query.GetMessage) *GetMessage {
	return &GetMessage{
		q: q,
	}
}