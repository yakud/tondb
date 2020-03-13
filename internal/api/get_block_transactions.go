package api

import (
	"fmt"
	"log"
	"net/http"

	"gitlab.flora.loc/mills/tondb/internal/ton"

	"gitlab.flora.loc/mills/tondb/internal/ton/storage"

	apiFilter "gitlab.flora.loc/mills/tondb/internal/api/filter"
	"gitlab.flora.loc/mills/tondb/internal/ton/query"
	"gitlab.flora.loc/mills/tondb/internal/ton/query/filter"

	"github.com/julienschmidt/httprouter"
)

type GetBlockTransactions struct {
	q                  *query.SearchTransactions
	shardsDescrStorage *storage.ShardsDescr
}

func (m *GetBlockTransactions) Handler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
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
			if masterBlockId.WorkchainId != -1 {
				http.Error(w, `{"error":true,"message":"block_master should have workchain_id:-1"}`, http.StatusBadRequest)
				return
			}

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

	getTransactionsFilter := filter.NewAnd(blocksFilter)

	// dir
	if messageDirectionFilter, err := apiFilter.MessageDirectionFromRequest(r); err != nil {
		http.Error(w, `{"error":true,"message":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	} else if messageDirectionFilter != nil {
		getTransactionsFilter.And(messageDirectionFilter)
	}

	// addr + lt
	if addrAndLtFilter, err := apiFilter.AddrAndLtFromRequest(r); err != nil {
		http.Error(w, `{"error":true,"message":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	} else if addrAndLtFilter != nil {
		getTransactionsFilter.And(addrAndLtFilter)
	} else {
		// addr
		if messageAddrFilter, err := apiFilter.MessageAddrFromRequest(r); err != nil {
			http.Error(w, `{"error":true,"message":"`+err.Error()+`"}`, http.StatusBadRequest)
			return
		} else if messageAddrFilter != nil {
			getTransactionsFilter.And(messageAddrFilter)
		}
	}

	// message_type
	if messageTypeFilter, err := apiFilter.MessageTypeFromRequest(r); err != nil {
		http.Error(w, `{"error":true,"message":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	} else if messageTypeFilter != nil {
		getTransactionsFilter.And(messageTypeFilter)
	}

	// type
	if typeFilter, err := apiFilter.TypeFromRequest(r); err != nil {
		http.Error(w, `{"error":true,"message":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	} else if typeFilter != nil {
		getTransactionsFilter.And(typeFilter)
	}

	// hash
	if hashFilter, err := apiFilter.HashFromRequest(r); err != nil {
		http.Error(w, `{"error":true,"message":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	} else if hashFilter != nil {
		getTransactionsFilter.And(hashFilter)
	}

	// Make query
	blocksTransactions, err := m.q.SearchByFilter(getTransactionsFilter)
	if err != nil {
		log.Println(fmt.Errorf("query SearchByFilter error: %w", err))
		http.Error(w, `{"error":true,"message":"SearchByFilter query error"}`, http.StatusBadRequest)
		return
	}

	resp, err := json.Marshal(blocksTransactions)
	if err != nil {
		http.Error(w, `{"error":true,"message":"response json marshaling error"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
	w.Write(resp)
}

func NewGetBlockTransactions(q *query.SearchTransactions, shardsDescrStorage *storage.ShardsDescr) *GetBlockTransactions {
	return &GetBlockTransactions{
		q:                  q,
		shardsDescrStorage: shardsDescrStorage,
	}
}
