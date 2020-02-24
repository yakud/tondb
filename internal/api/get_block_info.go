package api

import (
	"fmt"
	"log"
	"net/http"
	"time"

	apiFilter "gitlab.flora.loc/mills/tondb/internal/api/filter"
	"gitlab.flora.loc/mills/tondb/internal/ton"
	"gitlab.flora.loc/mills/tondb/internal/ton/query"
	"gitlab.flora.loc/mills/tondb/internal/ton/query/filter"
	"gitlab.flora.loc/mills/tondb/internal/ton/storage"

	"github.com/julienschmidt/httprouter"
)

const maxBlocksPerRequest = 1000

type GetBlockInfo struct {
	q                  *query.GetBlockInfo
	shardsDescrStorage *storage.ShardsDescr
}

func (m *GetBlockInfo) Handler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	blocksFilter := filter.NewOr()

	// block
	blockFilter, err := apiFilter.BlockFilterFromRequest(r, "block", maxBlocksPerRequest)
	if err != nil {
		http.Error(w, `{"error":true,"message":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	} else if blockFilter != nil {
		blocksFilter.Or(blockFilter)
	}

	// block_from + block_to
	blockRangeFilter, err := apiFilter.BlockRangeFilterFromRequest(r, "block_from", "block_to", maxBlocksPerRequest)
	if err != nil {
		http.Error(w, `{"error":true,"message":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	} else if blockRangeFilter != nil {
		blocksFilter.Or(blockRangeFilter)
	}

	// block_master
	blockMasterFilter, err := apiFilter.BlockFilterFromRequest(r, "block_master", maxBlocksPerRequest)
	if err != nil {
		http.Error(w, `{"error":true,"message":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	} else if blockMasterFilter != nil {
		for _, masterBlockId := range blockMasterFilter.Blocks() {
			shardsBlocks, err := m.shardsDescrStorage.GetShardsSeqRangeInMasterBlock(masterBlockId.SeqNo)
			if err != nil {
				log.Println("GetShardsSeqRangeInMasterBlock error:", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error":true,"message":"shardsDescrStorage.GetShardsSeqRangeInMasterBlock error"}`))
				return
			}

			for _, shardsBlock := range shardsBlocks {
				blocksRange, err := filter.NewBlocksRange(
					&ton.BlockId{
						WorkchainId: shardsBlock.WorkchainId,
						Shard:       shardsBlock.Shard,
						SeqNo:       shardsBlock.FromSeq,
					},
					&ton.BlockId{
						WorkchainId: shardsBlock.WorkchainId,
						Shard:       shardsBlock.Shard,
						SeqNo:       shardsBlock.ToSeq,
					},
				)
				if err != nil {
					log.Println("NewBlocksRange error:", err)
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(`{"error":true,"message":"NewBlocksRange error"}`))
					return
				}
				blocksFilter.Or(blocksRange)
			}
		}
	}
	if blockFilter == nil && blockRangeFilter == nil && blockMasterFilter == nil {
		http.Error(w, `{"error":true,"message":"you should set block or block_from+block_to or block_master filter"}`, http.StatusBadRequest)
		return
	}

	// Make query
	t := time.Now()
	blockInfo, err := m.q.GetBlockInfo(blocksFilter)
	if err != nil {
		log.Println(fmt.Errorf("query GetBlockInfo error: %w", err))
		http.Error(w, `{"error":true,"message":"GetBlockInfo query error"}`, http.StatusBadRequest)
		return
	}
	log.Println("block info query for: ", time.Now().Sub(t))

	resp, err := json.Marshal(blockInfo)
	if err != nil {
		http.Error(w, `{"error":true,"message":"response json marshaling error"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
	w.Write(resp)
}

func NewGetBlockInfo(q *query.GetBlockInfo, shardsDescrStorage *storage.ShardsDescr) *GetBlockInfo {
	return &GetBlockInfo{
		q:                  q,
		shardsDescrStorage: shardsDescrStorage,
	}
}
