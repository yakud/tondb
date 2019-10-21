package api

import (
	"log"
	"net/http"

	apiFilter "gitlab.flora.loc/mills/tondb/internal/api/filter"

	"gitlab.flora.loc/mills/tondb/internal/ton/storage"

	"github.com/julienschmidt/httprouter"
)

type GetWorkchainBlockMaster struct {
	shardsDescrStorage *storage.ShardsDescr
}

func (m *GetWorkchainBlockMaster) Handler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	blockMasterFilter, err := apiFilter.BlockFilterFromRequest(r, "block", maxBlocksPerRequest)
	if err != nil {
		http.Error(w, `{"error":true,"message":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	} else if blockMasterFilter == nil || len(blockMasterFilter.Blocks()) == 0 {
		http.Error(w, `{"error":true,"message":"empty block"}`, http.StatusBadRequest)
		return
	}

	block := blockMasterFilter.Blocks()[0]
	masterBlock, err := m.shardsDescrStorage.GetMasterByShardBlock(block)
	if err != nil {
		log.Println("GetMasterByShardBlock error:", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":true,"message":"shardsDescrStorage.GetMasterByShardBlock error"}`))
		return
	}

	resp, err := json.Marshal(masterBlock)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":true,"message":"response json marshaling error"}`))
		return
	}

	w.WriteHeader(200)
	w.Write(resp)
}

func NewGetWorkchainBlockMaster(shardsDescrStorage *storage.ShardsDescr) *GetWorkchainBlockMaster {
	return &GetWorkchainBlockMaster{
		shardsDescrStorage: shardsDescrStorage,
	}
}
