package api

import (
	"fmt"
	"gitlab.flora.loc/mills/tondb/swagger/tonapi"
	"log"
	"net/http"

	apiFilter "gitlab.flora.loc/mills/tondb/internal/api/filter"
	"gitlab.flora.loc/mills/tondb/internal/ton"
	"gitlab.flora.loc/mills/tondb/internal/ton/query"
	"gitlab.flora.loc/mills/tondb/internal/ton/query/filter"
	"gitlab.flora.loc/mills/tondb/internal/ton/storage"

	"github.com/labstack/echo/v4"
)

const maxBlocksPerRequest = 1000

type GetBlockInfo struct {
	q                  *query.GetBlockInfo
	shardsDescrStorage *storage.ShardsDescr
}

func (m *GetBlockInfo) GetV1BlockInfo(ctx echo.Context, params tonapi.GetV1BlockInfoParams) error {
	blocksFilter := filter.NewOr()

	// block
	blockFilter, err := apiFilter.BlockFilterFromParam(params.Block, maxBlocksPerRequest)
	if err != nil {
		return err
	} else if blockFilter != nil {
		blocksFilter.Or(blockFilter)
	}

	// block_from + block_to
	blockRangeFilter, err := apiFilter.BlockRangeFilterFromParams(params.BlockFrom, params.BlockTo, maxBlocksPerRequest)
	if err != nil {
		return err
	} else if blockRangeFilter != nil {
		blocksFilter.Or(blockRangeFilter)
	}

	// block_master
	blockMasterFilter, err := apiFilter.BlockFilterFromParam(params.BlockMaster, maxBlocksPerRequest)
	if err != nil {
		return err
	} else if blockMasterFilter != nil {
		for _, masterBlockId := range blockMasterFilter.Blocks() {
			shardsBlocks, err := m.shardsDescrStorage.GetShardsSeqRangeInMasterBlock(masterBlockId.SeqNo)
			if err != nil {
				log.Println("GetShardsSeqRangeInMasterBlock error:", err)
				return ctx.JSONBlob(http.StatusInternalServerError, []byte(`{"error":true,"message":"shardsDescrStorage.GetShardsSeqRangeInMasterBlock error"}`))
			}

			for _, shardsBlock := range shardsBlocks {
				blocksRange, err := filter.NewBlocksRange(
					&ton.BlockId{
						WorkchainId: shardsBlock.WorkchainId,
						Shard:       uint64(shardsBlock.Shard),
						SeqNo:       uint64(shardsBlock.FromSeq),
					},
					&ton.BlockId{
						WorkchainId: shardsBlock.WorkchainId,
						Shard:       uint64(shardsBlock.Shard),
						SeqNo:       uint64(shardsBlock.ToSeq),
					},
				)
				if err != nil {
					log.Println("NewBlocksRange error:", err)
					return ctx.JSONBlob(http.StatusInternalServerError, []byte(`{"error":true,"message":"NewBlocksRange error"}`))
				}
				blocksFilter.Or(blocksRange)
			}
		}
	}
	if blockFilter == nil && blockRangeFilter == nil && blockMasterFilter == nil {
		return ctx.JSONBlob(http.StatusBadRequest, []byte(`{"error":true,"message":"you should set block or block_from+block_to or block_master filter"}`))
	}

	// Make query
	blockInfo, err := m.q.GetBlockInfo(blocksFilter)
	if err != nil {
		log.Println(fmt.Errorf("query GetBlockInfo error: %w", err))
		return ctx.JSONBlob(http.StatusBadRequest, []byte(`{"error":true,"message":"GetBlockInfo query error"}`))
	}

	if len(blockInfo) == 0 {
		return ctx.String(http.StatusNotFound, `Block not found`)
	}

	return ctx.JSON(200, blockInfo)
}

func NewGetBlockInfo(q *query.GetBlockInfo, shardsDescrStorage *storage.ShardsDescr) *GetBlockInfo {
	return &GetBlockInfo{
		q:                  q,
		shardsDescrStorage: shardsDescrStorage,
	}
}
