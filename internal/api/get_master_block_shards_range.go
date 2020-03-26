package api

import (
	"gitlab.flora.loc/mills/tondb/swagger/tonapi"
	"log"
	"net/http"

	apiFilter "gitlab.flora.loc/mills/tondb/internal/api/filter"

	"gitlab.flora.loc/mills/tondb/internal/ton/storage"

	"github.com/labstack/echo/v4"
)

type MasterchainBlockShards struct {
	shardsDescrStorage *storage.ShardsDescr
}

func (m *MasterchainBlockShards) GetV1MasterBlockShardsRange(ctx echo.Context, params tonapi.GetV1MasterBlockShardsRangeParams) error {
	// block_master
	blockMasterFilter, err := apiFilter.BlockFilterFromParam(&params.BlockMaster, maxBlocksPerRequest)
	if err != nil {
		return ctx.JSONBlob(http.StatusBadRequest, []byte(`{"error":true,"message":"`+err.Error()+`"}`))
	} else if blockMasterFilter == nil {
		return ctx.JSONBlob(http.StatusBadRequest, []byte(`{"error":true,"message":"empty block_master"}`))
	}

	respShardsBlocks := make([]tonapi.ShardBlocksRange, 0)
	for _, masterBlockId := range blockMasterFilter.Blocks() {
		shardsBlocks, err := m.shardsDescrStorage.GetShardsSeqRangeInMasterBlock(masterBlockId.SeqNo)
		if err != nil {
			log.Println("GetShardsSeqRangeInMasterBlock error:", err)
			return ctx.JSONBlob(http.StatusInternalServerError, []byte(`{"error":true,"message":"shardsDescrStorage.GetShardsSeqRangeInMasterBlock error"}`))
		}

		respShardsBlocks = append(respShardsBlocks, shardsBlocks...)
	}

	return ctx.JSON(http.StatusOK, respShardsBlocks)
}

func NewMasterBlockShardsRange(shardsDescrStorage *storage.ShardsDescr) *MasterchainBlockShards {
	return &MasterchainBlockShards{
		shardsDescrStorage: shardsDescrStorage,
	}
}
