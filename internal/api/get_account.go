package api

import (
	"gitlab.flora.loc/mills/tondb/internal/ton"
	"gitlab.flora.loc/mills/tondb/internal/ton/query/filter"
	"gitlab.flora.loc/mills/tondb/swagger/tonapi"
	"strings"

	"gitlab.flora.loc/mills/tondb/internal/ton/view/state"

	"github.com/labstack/echo/v4"
)

type GetAccount struct {
	s *state.AccountState
}

func (s *GetAccount) GetV1Account(ctx echo.Context, params tonapi.GetV1AccountParams) error {
	accAddr, err := ton.ParseAccountAddress(strings.TrimSpace(params.Address))
	if err != nil {
		return err
	}

	accountFilter := filter.NewAccount(accAddr)

	accountState, err := s.s.GetAccountWithStats(accountFilter.Addr())
	if err != nil {
		return err
	}

	return ctx.JSON(200, accountState)
}

func NewGetAccount(s *state.AccountState) *GetAccount {
	return &GetAccount{
		s: s,
	}
}
