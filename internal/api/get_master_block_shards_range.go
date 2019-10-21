package api

import (
	"log"
	"net/http"

	apiFilter "gitlab.flora.loc/mills/tondb/internal/api/filter"

	"gitlab.flora.loc/mills/tondb/internal/ton/storage"

	"github.com/julienschmidt/httprouter"
)

type MasterchainBlockShards struct {
	shardsDescrStorage *storage.ShardsDescr
}

func (m *MasterchainBlockShards) Handler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// block_master
	blockMasterFilter, err := apiFilter.BlockFilterFromRequest(r, "block_master", maxBlocksPerRequest)
	if err != nil {
		http.Error(w, `{"error":true,"message":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	} else if blockMasterFilter == nil {
		http.Error(w, `{"error":true,"message":"empty block_master"}`, http.StatusBadRequest)
		return
	}

	respShardsBlocks := make([]storage.ShardBlocksRange, 0)
	for _, masterBlockId := range blockMasterFilter.Blocks() {
		shardsBlocks, err := m.shardsDescrStorage.GetShardsSeqRangeInMasterBlock(masterBlockId.SeqNo)
		if err != nil {
			log.Println("GetShardsSeqRangeInMasterBlock error:", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":true,"message":"shardsDescrStorage.GetShardsSeqRangeInMasterBlock error"}`))
			return
		}

		respShardsBlocks = append(respShardsBlocks, shardsBlocks...)
	}

	resp, err := json.Marshal(respShardsBlocks)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":true,"message":"response json marshaling error"}`))
		return
	}

	w.WriteHeader(200)
	w.Write(resp)
}

func NewMasterBlockShardsRange(shardsDescrStorage *storage.ShardsDescr) *MasterchainBlockShards {
	return &MasterchainBlockShards{
		shardsDescrStorage: shardsDescrStorage,
	}
}
