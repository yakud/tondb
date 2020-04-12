package api

import (
	"net/http"

	"gitlab.flora.loc/mills/tondb/internal/ton/view/state"

	filter2 "gitlab.flora.loc/mills/tondb/internal/api/filter"

	"github.com/julienschmidt/httprouter"
)

type GetAccount struct {
	s *state.AccountState
}

func (m *GetAccount) Handler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	accountFilter, err := filter2.AccountFilterFromRequest(r, "address")
	if err != nil {
		http.Error(w, `{"error":true,"message":"error make account filter: `+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	//// block
	//blockFilter, err := apiFilter.BlockFilterFromRequest(r, "block", maxBlocksPerRequest)
	//if err != nil {
	//	http.Error(w, `{"error":true,"message":"`+err.Error()+`"}`, http.StatusBadRequest)
	//	return
	//} else if blockFilter != nil {
	//	blocksFilter.Or(blockFilter)
	//}
	//
	//// block_master
	//blockMasterFilter, err := apiFilter.BlockFilterFromRequest(r, "block_master", maxBlocksPerRequest)
	//if err != nil {
	//	http.Error(w, `{"error":true,"message":"`+err.Error()+`"}`, http.StatusBadRequest)
	//	return
	//} else if blockMasterFilter != nil {
	//	for _, masterBlockId := range blockMasterFilter.Blocks() {
	//		shardsBlocks, err := m.shardsDescrStorage.GetShardsSeqRangeInMasterBlock(masterBlockId.SeqNo)
	//		if err != nil {
	//			log.Println("GetShardsSeqRangeInMasterBlock error:", err)
	//			w.WriteHeader(http.StatusInternalServerError)
	//			w.Write([]byte(`{"error":true,"message":"shardsDescrStorage.GetShardsSeqRangeInMasterBlock error"}`))
	//			return
	//		}
	//
	//		for _, shardsBlock := range shardsBlocks {
	//			blocksRange, err := filter.NewBlocksRange(
	//				&ton.BlockId{
	//					WorkchainId: shardsBlock.WorkchainId,
	//					Shard:       shardsBlock.Shard,
	//					SeqNo:       shardsBlock.FromSeq,
	//				},
	//				&ton.BlockId{
	//					WorkchainId: shardsBlock.WorkchainId,
	//					Shard:       shardsBlock.Shard,
	//					SeqNo:       shardsBlock.ToSeq,
	//				},
	//			)
	//			if err != nil {
	//				log.Println("NewBlocksRange error:", err)
	//				w.WriteHeader(http.StatusInternalServerError)
	//				w.Write([]byte(`{"error":true,"message":"NewBlocksRange error"}`))
	//				return
	//			}
	//			blocksFilter.Or(blocksRange)
	//		}
	//	}
	//}
	//if blockFilter == nil && blockRangeFilter == nil && blockMasterFilter == nil {
	//	http.Error(w, `{"error":true,"message":"you should set block or block_from+block_to or block_master filter"}`, http.StatusBadRequest)
	//	return
	//}

	accountState, err := m.s.GetAccountWithStats(accountFilter.Addr())
	if err != nil {
		http.Error(w, `{"error":true,"message":"error fetch account: `+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	resp, err := json.Marshal(accountState)
	if err != nil {
		http.Error(w, `{"error":true,"message":"response json marshaling error"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
	w.Write(resp)
}

func NewGetAccount(s *state.AccountState) *GetAccount {
	return &GetAccount{
		s: s,
	}
}
