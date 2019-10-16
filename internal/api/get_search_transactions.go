package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/yakud/ton-blocks-stream-receiver/internal/ton/query/filter"

	"github.com/yakud/ton-blocks-stream-receiver/internal/ton"
	"github.com/yakud/ton-blocks-stream-receiver/internal/ton/query"

	"github.com/yakud/ton-blocks-stream-receiver/internal/ton/storage"

	"github.com/julienschmidt/httprouter"
)

type GetSearchTransactionsByMaster struct {
	shardsDescrStorage *storage.ShardsDescr
	searchQuery        *query.SearchTransactions
}

func (m *GetSearchTransactionsByMaster) Handler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// Blocks filter
	seqNoStr := r.URL.Query().Get("seq_no")
	seqNo, err := strconv.ParseUint(seqNoStr, 10, 64)
	if err != nil {
		http.Error(w, `{"error":true,"message":"seq_no parse error"}`, http.StatusBadRequest)
		return
	}

	shardsBlocks, err := m.shardsDescrStorage.GetShardsSeqRangeInMCBlock(seqNo)
	if err != nil {
		log.Println("GetShardsSeqRangeInMCBlock error:", err)
		http.Error(w, `{"error":true,"message":"shardsDescrStorage.GetShardsSeqRangeInMCBlock error"}`, http.StatusInternalServerError)
		return
	}

	if len(shardsBlocks) == 0 {
		http.Error(w, `{"error":true,"message":"empty shards"}`, http.StatusInternalServerError)
		return
	}

	filters := make([]filter.Builder, 0)
	blocksIdFilter := make([]*ton.BlockId, 0)
	for _, shard := range shardsBlocks {
		for i := 0; i <= int(shard.ToSeq-shard.FromSeq); i++ {
			blocksIdFilter = append(blocksIdFilter, &ton.BlockId{
				WorkchainId: shard.WorkchainId,
				Shard:       shard.Shard,
				SeqNo:       shard.FromSeq + uint64(i),
			})
		}
	}
	filters = append(filters, filter.NewBlocks(blocksIdFilter...))

	// Addr filter
	addr := r.URL.Query().Get("addr")
	if addr != "" {
		filters = append(filters, filter.NewSrcDestAddr(addr))
	}

	// Search transactions
	transactions, err := m.searchQuery.SearchByFilter(filters...)
	if err != nil {
		log.Println("SearchTransactions.SearchByBlock error:", err)
		http.Error(w, `{"error":true,"message":"Search transactions DB error"}`, http.StatusInternalServerError)
		return
	}

	// Searialize and write response
	resp, err := json.Marshal(transactions)
	if err != nil {
		http.Error(w, `{"error":true,"message":"response json marshaling error"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
	w.Write(resp)
}

func NewGetSearchTransactionsByMaster(
	shardsDescrStorage *storage.ShardsDescr,
	searchQuery *query.SearchTransactions,
) *GetSearchTransactionsByMaster {
	return &GetSearchTransactionsByMaster{
		shardsDescrStorage: shardsDescrStorage,
		searchQuery:        searchQuery,
	}
}
