package api

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"gitlab.flora.loc/mills/tondb/internal/ton/query"
)

type GetBlockchainHeight struct {
	q *query.GetBlockchainHeight
}

func (api *GetBlockchainHeight) GetV1HeightBlockchain(ctx echo.Context) error {
	lastSyncedBlock, err := api.q.GetBlockchainHeight()
	if err != nil {
		log.Println(err)
		return ctx.JSONBlob(http.StatusBadRequest, []byte(`{"error":true,"message":"error retrieving blockchain height from DB"}`))
	}

	return ctx.JSON(200, lastSyncedBlock)
}

func NewGetBlockchainHeight(q *query.GetBlockchainHeight) *GetBlockchainHeight {
	return &GetBlockchainHeight{
		q: q,
	}
}
