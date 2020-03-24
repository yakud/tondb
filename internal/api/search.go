package api

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"gitlab.flora.loc/mills/tondb/internal/ton/search"
	"gitlab.flora.loc/mills/tondb/swagger/tonapi"
	"log"
	"net/http"
)

type Search struct {
	searcher *search.Searcher
}

func (m *Search) GetV1Search(ctx echo.Context, params tonapi.GetV1SearchParams) error {
	if params.Q == "" {
		return ctx.JSONBlob(http.StatusBadRequest, []byte(`{"error":true,"message":"empty search query"}`))
	}

	searchResult, err := m.searcher.Search(params.Q)
	if err != nil {
		log.Println(fmt.Errorf("searcher error: %w", err))
		return ctx.JSONBlob(http.StatusInternalServerError, []byte(`{"error":true,"message":"search error"}`))
	}

	if len(searchResult) == 0 {
		return ctx.JSONBlob(http.StatusNoContent, []byte(`{"result":[]}`))
	}

	return ctx.JSON(http.StatusOK, searchResult)
}

func NewSearch(searcher *search.Searcher) *Search {
	return &Search{
		searcher: searcher,
	}
}
