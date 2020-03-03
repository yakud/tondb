package api

import (
	"fmt"
	"net/http"
	"strings"

	"gitlab.flora.loc/mills/tondb/internal/utils"

	"gitlab.flora.loc/mills/tondb/internal/blocks_fetcher"

	apiFilter "gitlab.flora.loc/mills/tondb/internal/api/filter"

	"github.com/julienschmidt/httprouter"
)

type GetBlockTlb struct {
	blocksFetcher *blocks_fetcher.Client
}

func (m *GetBlockTlb) Handler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// block
	blockFilter, err := apiFilter.BlockFilterFromRequest(r, "block", 1)
	if err != nil {
		http.Error(w, `{"error":true,"message":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	if len(blockFilter.Blocks()) != 1 {
		http.Error(w, `{"error":true,"message":"should be exactly one block"}`, http.StatusBadRequest)
		return
	}

	block := blockFilter.Blocks()[0]
	blockTlb, err := m.blocksFetcher.FetchBlockTlb(*block)
	if err != nil {
		http.Error(w, `{"error":true,"message":"tlb block fetch error"}`, http.StatusInternalServerError)
		return
	}
	fileName := fmt.Sprintf(
		"(%d,%s,%d).boc",
		block.WorkchainId,
		strings.ToUpper(utils.DecToHex(block.Shard)),
		block.SeqNo,
	)
	w.Header().Set("Content-Disposition", "attachment; filename=\""+fileName+"\"")
	w.WriteHeader(200)
	w.Write(blockTlb)
}

func NewGetBlockTlb(blocksFetcher *blocks_fetcher.Client) *GetBlockTlb {
	return &GetBlockTlb{
		blocksFetcher: blocksFetcher,
	}
}
