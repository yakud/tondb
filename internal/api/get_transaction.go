package api

import (
	"fmt"
	"gitlab.flora.loc/mills/tondb/internal/ton/query/filter"
	"gitlab.flora.loc/mills/tondb/swagger/tonapi"
	"log"
	"net/http"
	"strings"

	"gitlab.flora.loc/mills/tondb/internal/ton/storage"

	"github.com/labstack/echo/v4"
	"gitlab.flora.loc/mills/tondb/internal/ton/query"
)

type GetTransactions struct {
	q                  *query.SearchTransactions
	shardsDescrStorage *storage.ShardsDescr
}

func (m *GetTransactions) GetV1Transaction(ctx echo.Context, params tonapi.GetV1TransactionParams) error {
	trFilter, err := filter.NewTransactionHashByIndex(strings.ToUpper(params.Hash))
	if err != nil {
		return ctx.JSONBlob(http.StatusBadRequest, []byte(`{"error":true,"message":"`+err.Error()+`"}`))
	}

	if trFilter == nil {
		return ctx.JSONBlob(http.StatusBadRequest, []byte(`{"error":true,"message":"empty hash filter"}`))
	}

	blocksTransactions, err := m.q.SearchByFilter(trFilter)
	if err != nil {
		log.Println(fmt.Errorf("query SearchByFilter error: %w", err))
		return ctx.JSONBlob(http.StatusBadRequest, []byte(`{"error":true,"message":"SearchByFilter query error"}`))
	}

	return ctx.JSON(http.StatusOK, blocksTransactions)
}

func NewGetTransactions(q *query.SearchTransactions) *GetTransactions {
	return &GetTransactions{
		q: q,
	}
}
