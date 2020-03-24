package api

import (
	"fmt"
	"gitlab.flora.loc/mills/tondb/swagger/tonapi"
	"log"
	"net/http"

	"gitlab.flora.loc/mills/tondb/internal/ton"

	"gitlab.flora.loc/mills/tondb/internal/ton/storage"

	apiFilter "gitlab.flora.loc/mills/tondb/internal/api/filter"
	"gitlab.flora.loc/mills/tondb/internal/ton/query"
	"gitlab.flora.loc/mills/tondb/internal/ton/query/filter"

	"github.com/labstack/echo/v4"
)

type GetBlockTransactions struct {
	q                  *query.SearchTransactions
	shardsDescrStorage *storage.ShardsDescr
}

func (m *GetBlockTransactions) GetV1BlockTransactions(ctx echo.Context, params tonapi.GetV1BlockTransactionsParams) error {
	blocksFilter := filter.NewOr()

	// block
	blockFilter, err := apiFilter.BlockFilterFromParam(params.Block, maxBlocksPerRequest)
	if err != nil {
		return ctx.JSONBlob(http.StatusBadRequest, []byte(`{"error":true,"message":"`+err.Error()+`"}`))
	} else if blockFilter != nil {
		blocksFilter.Or(blockFilter)
	}

	// block_from + block_to
	blockRangeFilter, err := apiFilter.BlockRangeFilterFromParams(params.BlockFrom, params.BlockTo, maxBlocksPerRequest)
	if err != nil {
		return ctx.JSONBlob(http.StatusBadRequest, []byte(`{"error":true,"message":"`+err.Error()+`"}`))
	} else if blockRangeFilter != nil {
		blocksFilter.Or(blockRangeFilter)
	}

	// block_master
	blockMasterFilter, err := apiFilter.BlockFilterFromParam(params.BlockMaster, maxBlocksPerRequest)
	if err != nil {
		return ctx.JSONBlob(http.StatusBadRequest, []byte(`{"error":true,"message":"`+err.Error()+`"}`))
	} else if blockMasterFilter != nil {
		for _, masterBlockId := range blockMasterFilter.Blocks() {
			if masterBlockId.WorkchainId != -1 {
				return ctx.JSONBlob(http.StatusBadRequest, []byte(`{"error":true,"message":"block_master should have workchain_id:-1"}`))
			}

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

	getTransactionsFilter := filter.NewAnd(blocksFilter)

	// dir
	if params.Dir != nil {
		getTransactionsFilter.And(filter.NewArrayHas("Messages.Direction", *params.Dir))
	}

	// addr + lt
	if addrAndLtFilter, err := apiFilter.AddrAndLtFromParams(params.Addr, params.Lt); err != nil {
		return ctx.JSONBlob(http.StatusBadRequest, []byte(`{"error":true,"message":"`+err.Error()+`"}`))
	} else if addrAndLtFilter != nil {
		getTransactionsFilter.And(addrAndLtFilter)
	} else {
		// addr
		if messageAddrFilter, err := apiFilter.MessageAddrFromParam(params.Addr); err != nil {
			return ctx.JSONBlob(http.StatusBadRequest, []byte(`{"error":true,"message":"`+err.Error()+`"}`))
		} else if messageAddrFilter != nil {
			getTransactionsFilter.And(messageAddrFilter)
		}
	}

	// message_type
	if messageTypeFilter, err := apiFilter.MessageTypeFromParam(params.MessageType); err != nil {
		return ctx.JSONBlob(http.StatusBadRequest, []byte(`{"error":true,"message":"`+err.Error()+`"}`))
	} else if messageTypeFilter != nil {
		getTransactionsFilter.And(messageTypeFilter)
	}

	// type
	if typeFilter, err := apiFilter.TypeFromParams(params.Type); err != nil {
		return ctx.JSONBlob(http.StatusBadRequest, []byte(`{"error":true,"message":"`+err.Error()+`"}`))
	} else if typeFilter != nil {
		getTransactionsFilter.And(typeFilter)
	}

	// hash
	if hashFilter, err := apiFilter.HashFromParams(params.Hash); err != nil {
		return ctx.JSONBlob(http.StatusBadRequest, []byte(`{"error":true,"message":"`+err.Error()+`"}`))
	} else if hashFilter != nil {
		getTransactionsFilter.And(hashFilter)
	}

	// Make query
	blocksTransactions, err := m.q.SearchByFilter(getTransactionsFilter)
	if err != nil {
		log.Println(fmt.Errorf("query SearchByFilter error: %w", err))
		return ctx.JSONBlob(http.StatusBadRequest, []byte(`{"error":true,"message":"SearchByFilter query error"}`))
	}

	return ctx.JSON(200, blocksTransactions)
}

func NewGetBlockTransactions(q *query.SearchTransactions, shardsDescrStorage *storage.ShardsDescr) *GetBlockTransactions {
	return &GetBlockTransactions{
		q:                  q,
		shardsDescrStorage: shardsDescrStorage,
	}
}
