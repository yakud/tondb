package site

import (
	"gitlab.flora.loc/mills/tondb/swagger/tonapi"

	"github.com/labstack/echo/v4"
	"gitlab.flora.loc/mills/tondb/internal/ton/view/stats"
)

const defaultTopAddrCount = 50

type GetAddrTopByMessageCount struct {
	q *stats.AddrMessagesCount
}

func (api *GetAddrTopByMessageCount) GetV1AddrTopByMessageCount(ctx echo.Context) error {
	topIn, topOut, err := api.q.SelectTopMessagesCount(defaultTopAddrCount)
	if err != nil {
		return err
	}

	addr := tonapi.AddrTopByMessageCountResponse{
		TopIn:  &topIn,
		TopOut: &topOut,
	}

	return ctx.JSON(200, addr)
}

func NewGetAddrTopByMessageCount(q *stats.AddrMessagesCount) *GetAddrTopByMessageCount {
	return &GetAddrTopByMessageCount{
		q: q,
	}
}
