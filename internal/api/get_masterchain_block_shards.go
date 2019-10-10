package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/yakud/ton-blocks-stream-receiver/internal/ton/storage"

	"github.com/julienschmidt/httprouter"
)

type MasterchainBlockShards struct {
	shardsDescrStorage *storage.ShardsDescr
}

func (m *MasterchainBlockShards) Handler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	seqNoStr := p.ByName("seqNo")
	seqNo, err := strconv.ParseUint(seqNoStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":true,"message":"seqNo parse error"}`))
		return
	}

	shardsBlocks, err := m.shardsDescrStorage.GetShardsSeqRangeInMCBlock(seqNo)
	if err != nil {
		log.Println("GetShardsSeqRangeInMCBlock error:", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":true,"message":"shardsDescrStorage.GetShardsSeqRangeInMCBlock error"}`))
		return
	}

	resp, err := json.Marshal(shardsBlocks)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":true,"message":"response json marshaling error"}`))
		return
	}

	w.WriteHeader(200)
	w.Write(resp)
}

func NewMasterchainBlockShards(shardsDescrStorage *storage.ShardsDescr) *MasterchainBlockShards {
	return &MasterchainBlockShards{
		shardsDescrStorage: shardsDescrStorage,
	}
}
