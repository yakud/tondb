package api

import (
	"gitlab.flora.loc/mills/tondb/internal/ton"
	"gitlab.flora.loc/mills/tondb/internal/ton/query/filter"
	"gitlab.flora.loc/mills/tondb/swagger/tonapi"
	"log"
	"net/http"

	"gitlab.flora.loc/mills/tondb/internal/ton/storage"

	"github.com/labstack/echo/v4"
)

type GetWorkchainBlockMaster struct {
	shardsDescrStorage *storage.ShardsDescr
}

func (m *GetWorkchainBlockMaster) GetV1WorkchainBlockMaster(ctx echo.Context, params tonapi.GetV1WorkchainBlockMasterParams) error {
	parsedBlock, err := ton.ParseBlockId(params.Block)
	if err != nil {
		return ctx.JSONBlob(http.StatusBadRequest, []byte(`{"error":true,"message":"`+err.Error()+`"}`))
	}

	blockMasterFilter := filter.NewBlocks(parsedBlock)
	if blockMasterFilter == nil || len(blockMasterFilter.Blocks()) == 0 {
		return ctx.JSONBlob(http.StatusBadRequest, []byte(`{"error":true,"message":"empty block"}`))
	}

	block := blockMasterFilter.Blocks()[0]
	masterBlock, err := m.shardsDescrStorage.GetMasterByShardBlock(block)
	if err != nil {
		log.Println("GetMasterByShardBlock error:", err)
		return ctx.JSONBlob(http.StatusInternalServerError, []byte(`{"error":true,"message":"shardsDescrStorage.GetMasterByShardBlock error"}`))
	}

	return ctx.JSON(http.StatusOK, masterBlock)
}

func NewGetWorkchainBlockMaster(shardsDescrStorage *storage.ShardsDescr) *GetWorkchainBlockMaster {
	return &GetWorkchainBlockMaster{
		shardsDescrStorage: shardsDescrStorage,
	}
}
