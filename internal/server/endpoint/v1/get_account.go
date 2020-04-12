package v1

import (
	"github.com/labstack/echo/v4"
	"gitlab.flora.loc/mills/tondb/swagger/tonapi"
)

type GetAccount struct {
}

func (t *GetAccount) GetV1Account(ctx echo.Context, params tonapi.GetV1AccountParams) error {
	panic("implement me")
}
