package api

import (
	"fmt"
	"gitlab.flora.loc/mills/tondb/swagger/tonapi"
	"net/http"
	"strings"

	"gitlab.flora.loc/mills/tondb/internal/utils"

	"gitlab.flora.loc/mills/tondb/internal/blocks_fetcher"

	apiFilter "gitlab.flora.loc/mills/tondb/internal/api/filter"

	"github.com/labstack/echo/v4"
)

type GetBlockTlb struct {
	blocksFetcher *blocks_fetcher.Client
}

func (m *GetBlockTlb) GetV1BlockTlb(ctx echo.Context, params tonapi.GetV1BlockTlbParams) error {
	// block
	blockFilter, err := apiFilter.BlockFilterFromParam(params.Block, 1)
	if err != nil {
		return ctx.JSONBlob(http.StatusBadRequest, []byte(`{"error":true,"message":"`+err.Error()+`"}`))
	}

	if len(blockFilter.Blocks()) != 1 {
		return ctx.JSONBlob(http.StatusBadRequest, []byte(`{"error":true,"message":"should be exactly one block"}`))
	}

	block := blockFilter.Blocks()[0]
	blockTlb, err := m.blocksFetcher.FetchBlockTlb(*block)
	if err != nil {
		return ctx.JSONBlob(http.StatusInternalServerError, []byte(`{"error":true,"message":"tlb block fetch error"}`))
	}
	fileName := fmt.Sprintf(
		"(%d,%s,%d).boc",
		block.WorkchainId,
		strings.ToUpper(utils.DecToHex(block.Shard)),
		block.SeqNo,
	)

	responseWriter := ctx.Response().Writer

	responseWriter.Header().Set("Content-Disposition", "attachment; filename=\""+fileName+"\"")
	responseWriter.WriteHeader(200)
	_, err = responseWriter.Write(blockTlb)

	return err
}

func NewGetBlockTlb(blocksFetcher *blocks_fetcher.Client) *GetBlockTlb {
	return &GetBlockTlb{
		blocksFetcher: blocksFetcher,
	}
}
