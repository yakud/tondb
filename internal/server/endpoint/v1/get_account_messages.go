package v1

import (
	"github.com/labstack/echo/v4"
	"gitlab.flora.loc/mills/tondb/swagger/tonapi"
)

type GetAccountMessages struct {
}

func (t *GetAccountMessages) GetV1AccountMessages(ctx echo.Context, params tonapi.GetV1AccountMessagesParams) error {
	panic("implement me")
}
