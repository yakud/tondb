package api

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"gitlab.flora.loc/mills/tondb/internal/ton/query"
)

type GetSyncedHeight struct {
	q *query.GetSyncedHeight
}

func (api *GetSyncedHeight) GetV1HeightSynced(ctx echo.Context) error {
	lastSyncedBlock, err := api.q.GetSyncedHeight()
	if err != nil {
		return ctx.JSONBlob(http.StatusBadRequest, []byte(`{"error":true,"message":"error retrieve synced height from DB"}`))
	}

	return ctx.JSON(http.StatusOK, lastSyncedBlock)
}

func NewGetSyncedHeight(q *query.GetSyncedHeight) *GetSyncedHeight {
	return &GetSyncedHeight{
		q: q,
	}
}
