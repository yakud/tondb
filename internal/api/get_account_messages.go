package api

import (
	"gitlab.flora.loc/mills/tondb/internal/ton"
	"gitlab.flora.loc/mills/tondb/internal/ton/query/filter"
	"gitlab.flora.loc/mills/tondb/swagger/tonapi"
	"strings"

	"gitlab.flora.loc/mills/tondb/internal/ton/view/feed"

	"github.com/labstack/echo/v4"
	feed2 "gitlab.flora.loc/mills/tondb/internal/api/feed"
)

const defaultMessagesCount = 30

type GetAccountMessages struct {
	f *feed.AccountMessages
}

func (s *GetAccountMessages) GetV1AccountMessages(ctx echo.Context, params tonapi.GetV1AccountMessagesParams) error {
	accAddr, err := ton.ParseAccountAddress(strings.TrimSpace(params.Address))
	if err != nil {
		return err
	}

	accountFilter := filter.NewAccount(accAddr)

	// scroll_id
	var scrollId = &feed.AccountMessagesScrollId{}
	if params.ScrollId != nil && len(*params.ScrollId) > 0 {
		if err := feed2.UnpackScrollId(*params.ScrollId, scrollId); err != nil {
			return err
		}
	}

	// limit
	var limit int16
	if params.Limit == nil {
		limit = defaultMessagesCount
	} else {
		limit = int16(*params.Limit)
	}

	accountMessages, newScrollId, err := s.f.GetAccountMessages(accountFilter.Addr(), scrollId, limit, nil)
	if err != nil {
		return err
	}

	newPackedScrollId, err := feed2.PackScrollId(newScrollId)
	if err != nil {
		return err
	}

	messagesResponse := tonapi.AccountMessageResponse{
		Messages: &accountMessages,
		ScrollId: &newPackedScrollId,
	}

	return ctx.JSON(200, messagesResponse)
}

func NewGetAccountMessages(f *feed.AccountMessages) *GetAccountMessages {
	return &GetAccountMessages{
		f: f,
	}
}
