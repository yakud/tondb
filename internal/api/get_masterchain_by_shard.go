package api

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"gitlab.flora.loc/mills/tondb/internal/utils"

	"gitlab.flora.loc/mills/tondb/internal/ton/storage"

	"github.com/julienschmidt/httprouter"
)

type MasterchainByShard struct {
	shardsDescrStorage *storage.ShardsDescr
}

func (m *MasterchainByShard) Handler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	seqNoStr := r.URL.Query().Get("seq_no")
	if seqNoStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":true,"message":"seq_no empty"}`))
		return
	}

	seqNo, err := strconv.ParseUint(seqNoStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":true,"message":"seqNo parse error"}`))
		return
	}

	shardHex := r.URL.Query().Get("shard_hex")
	if shardHex == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":true,"message":"shard_hex empty"}`))
		return
	}

	shardDec, err := utils.HexToDec(shardHex)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":true,"message":"shard_hex parse error"}`))
		return
	}

	fmt.Println("shard_hex:", shardHex)
	fmt.Println("shard_dec:", shardDec)

	masterBlock, err := m.shardsDescrStorage.GetMasterSeqByShardSeq(shardDec, seqNo)
	if err != nil {
		log.Println("GetMasterSeqByShardSeq error:", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":true,"message":"shardsDescrStorage.GetMasterSeqByShardSeq error"}`))
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

func NewMasterchainByShard(shardsDescrStorage *storage.ShardsDescr) *MasterchainByShard {
	return &MasterchainByShard{
		shardsDescrStorage: shardsDescrStorage,
	}
}
