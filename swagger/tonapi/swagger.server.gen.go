// Package tonapi provides primitives to interact the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen DO NOT EDIT.
package tonapi

import (
	"fmt"
	"github.com/deepmap/oapi-codegen/pkg/runtime"
	"github.com/labstack/echo/v4"
	"net/http"
)

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// Get account// (GET /v1/account)
	GetV1Account(ctx echo.Context, params GetV1AccountParams) error
	// Get account messages// (GET /v1/account/messages)
	GetV1AccountMessages(ctx echo.Context, params GetV1AccountMessagesParams) error
	// Get account QR code// (GET /v1/account/qr)
	GetV1AccountQr(ctx echo.Context, params GetV1AccountQrParams) error
	// Top account by message in/out// (GET /v1/addr/top-by-message-count)
	GetV1AddrTopByMessageCount(ctx echo.Context) error
	// Get block info// (GET /v1/block/info)
	GetV1BlockInfo(ctx echo.Context, params GetV1BlockInfoParams) error
	// Get block in TL-B format// (GET /v1/block/tlb)
	GetV1BlockTlb(ctx echo.Context, params GetV1BlockTlbParams) error
	// Get block transactions// (GET /v1/block/transactions)
	GetV1BlockTransactions(ctx echo.Context, params GetV1BlockTransactionsParams) error
	// Get blocks feed// (GET /v1/blocks/feed)
	GetV1BlocksFeed(ctx echo.Context, params GetV1BlocksFeedParams) error
	// Get blockchain height// (GET /v1/height/blockchain)
	GetV1HeightBlockchain(ctx echo.Context) error
	// Get synced height// (GET /v1/height/synced)
	GetV1HeightSynced(ctx echo.Context) error
	// Get master block actual shards// (GET /v1/master/block/shards/actual)
	GetV1MasterBlockShardsActual(ctx echo.Context, params GetV1MasterBlockShardsActualParams) error
	// Get master block shards range// (GET /v1/master/block/shards/range)
	GetV1MasterBlockShardsRange(ctx echo.Context, params GetV1MasterBlockShardsRangeParams) error
	// Get message// (GET /v1/message/get)
	GetV1MessageGet(ctx echo.Context, params GetV1MessageGetParams) error
	// Messages feed// (GET /v1/messages/feed)
	GetV1MessagesFeed(ctx echo.Context, params GetV1MessagesFeedParams) error
	// Search// (GET /v1/search)
	GetV1Search(ctx echo.Context, params GetV1SearchParams) error
	// Get addresses metrics// (GET /v1/stats/addresses)
	GetV1StatsAddresses(ctx echo.Context, params GetV1StatsAddressesParams) error
	// Get blocks metrics// (GET /v1/stats/blocks)
	GetV1StatsBlocks(ctx echo.Context, params GetV1StatsBlocksParams) error
	// Get some global blockchain metrics// (GET /v1/stats/global)
	GetV1StatsGlobal(ctx echo.Context) error
	// Get messages metrics// (GET /v1/stats/messages)
	GetV1StatsMessages(ctx echo.Context, params GetV1StatsMessagesParams) error
	// Get transactions metrics// (GET /v1/stats/transactions)
	GetV1StatsTransactions(ctx echo.Context, params GetV1StatsTransactionsParams) error
	// Blocks by workchain// (GET /v1/timeseries/blocks-by-workchain)
	GetV1TimeseriesBlocksByWorkchain(ctx echo.Context) error
	// Messages by type// (GET /v1/timeseries/messages-by-type)
	GetV1TimeseriesMessagesByType(ctx echo.Context) error
	// Messages ord count// (GET /v1/timeseries/messages-ord-count)
	GetV1TimeseriesMessagesOrdCount(ctx echo.Context) error
	// Average sent and fees in nanograms for all blocks of every day for last 30 days// (GET /v1/timeseries/sent-and-fees)
	GetV1TimeseriesSentAndFees(ctx echo.Context) error
	// Volume by grams// (GET /v1/timeseries/volume-by-grams)
	GetV1TimeseriesVolumeByGrams(ctx echo.Context) error
	// Top whales accounts by gram amount held// (GET /v1/top/whales)
	GetV1TopWhales(ctx echo.Context, params GetV1TopWhalesParams) error
	// Get transaction// (GET /v1/transaction)
	GetV1Transaction(ctx echo.Context, params GetV1TransactionParams) error
	// Transactions feed// (GET /v1/transactions/feed)
	GetV1TransactionsFeed(ctx echo.Context, params GetV1TransactionsFeedParams) error
	// Get master block by workchain block// (GET /v1/workchain/block/master)
	GetV1WorkchainBlockMaster(ctx echo.Context, params GetV1WorkchainBlockMasterParams) error
}

// ServerInterfaceWrapper converts echo contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler ServerInterface
}

// GetV1Account converts echo context to params.
func (w *ServerInterfaceWrapper) GetV1Account(ctx echo.Context) error {
	var err error

	// Parameter object where we will unmarshal all parameters from the context
	var params GetV1AccountParams
	// ------------- Required query parameter "address" -------------
	if paramValue := ctx.QueryParam("address"); paramValue != "" {

	} else {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Query argument address is required, but not found"))
	}

	err = runtime.BindQueryParameter("form", true, true, "address", ctx.QueryParams(), &params.Address)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter address: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetV1Account(ctx, params)
	return err
}

// GetV1AccountMessages converts echo context to params.
func (w *ServerInterfaceWrapper) GetV1AccountMessages(ctx echo.Context) error {
	var err error

	// Parameter object where we will unmarshal all parameters from the context
	var params GetV1AccountMessagesParams
	// ------------- Required query parameter "address" -------------
	if paramValue := ctx.QueryParam("address"); paramValue != "" {

	} else {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Query argument address is required, but not found"))
	}

	err = runtime.BindQueryParameter("form", true, true, "address", ctx.QueryParams(), &params.Address)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter address: %s", err))
	}

	// ------------- Optional query parameter "scroll_id" -------------
	if paramValue := ctx.QueryParam("scroll_id"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "scroll_id", ctx.QueryParams(), &params.ScrollId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter scroll_id: %s", err))
	}

	// ------------- Optional query parameter "limit" -------------
	if paramValue := ctx.QueryParam("limit"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "limit", ctx.QueryParams(), &params.Limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter limit: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetV1AccountMessages(ctx, params)
	return err
}

// GetV1AccountQr converts echo context to params.
func (w *ServerInterfaceWrapper) GetV1AccountQr(ctx echo.Context) error {
	var err error

	// Parameter object where we will unmarshal all parameters from the context
	var params GetV1AccountQrParams
	// ------------- Required query parameter "address" -------------
	if paramValue := ctx.QueryParam("address"); paramValue != "" {

	} else {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Query argument address is required, but not found"))
	}

	err = runtime.BindQueryParameter("form", true, true, "address", ctx.QueryParams(), &params.Address)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter address: %s", err))
	}

	// ------------- Optional query parameter "format" -------------
	if paramValue := ctx.QueryParam("format"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "format", ctx.QueryParams(), &params.Format)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter format: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetV1AccountQr(ctx, params)
	return err
}

// GetV1AddrTopByMessageCount converts echo context to params.
func (w *ServerInterfaceWrapper) GetV1AddrTopByMessageCount(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetV1AddrTopByMessageCount(ctx)
	return err
}

// GetV1BlockInfo converts echo context to params.
func (w *ServerInterfaceWrapper) GetV1BlockInfo(ctx echo.Context) error {
	var err error

	// Parameter object where we will unmarshal all parameters from the context
	var params GetV1BlockInfoParams
	// ------------- Optional query parameter "block" -------------
	if paramValue := ctx.QueryParam("block"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "block", ctx.QueryParams(), &params.Block)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter block: %s", err))
	}

	// ------------- Optional query parameter "block_master" -------------
	if paramValue := ctx.QueryParam("block_master"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "block_master", ctx.QueryParams(), &params.BlockMaster)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter block_master: %s", err))
	}

	// ------------- Optional query parameter "block_from" -------------
	if paramValue := ctx.QueryParam("block_from"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "block_from", ctx.QueryParams(), &params.BlockFrom)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter block_from: %s", err))
	}

	// ------------- Optional query parameter "block_to" -------------
	if paramValue := ctx.QueryParam("block_to"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "block_to", ctx.QueryParams(), &params.BlockTo)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter block_to: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetV1BlockInfo(ctx, params)
	return err
}

// GetV1BlockTlb converts echo context to params.
func (w *ServerInterfaceWrapper) GetV1BlockTlb(ctx echo.Context) error {
	var err error

	// Parameter object where we will unmarshal all parameters from the context
	var params GetV1BlockTlbParams
	// ------------- Optional query parameter "block" -------------
	if paramValue := ctx.QueryParam("block"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "block", ctx.QueryParams(), &params.Block)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter block: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetV1BlockTlb(ctx, params)
	return err
}

// GetV1BlockTransactions converts echo context to params.
func (w *ServerInterfaceWrapper) GetV1BlockTransactions(ctx echo.Context) error {
	var err error

	// Parameter object where we will unmarshal all parameters from the context
	var params GetV1BlockTransactionsParams
	// ------------- Optional query parameter "block" -------------
	if paramValue := ctx.QueryParam("block"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "block", ctx.QueryParams(), &params.Block)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter block: %s", err))
	}

	// ------------- Optional query parameter "block_master" -------------
	if paramValue := ctx.QueryParam("block_master"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "block_master", ctx.QueryParams(), &params.BlockMaster)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter block_master: %s", err))
	}

	// ------------- Optional query parameter "block_from" -------------
	if paramValue := ctx.QueryParam("block_from"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "block_from", ctx.QueryParams(), &params.BlockFrom)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter block_from: %s", err))
	}

	// ------------- Optional query parameter "block_to" -------------
	if paramValue := ctx.QueryParam("block_to"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "block_to", ctx.QueryParams(), &params.BlockTo)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter block_to: %s", err))
	}

	// ------------- Optional query parameter "dir" -------------
	if paramValue := ctx.QueryParam("dir"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "dir", ctx.QueryParams(), &params.Dir)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter dir: %s", err))
	}

	// ------------- Optional query parameter "addr" -------------
	if paramValue := ctx.QueryParam("addr"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "addr", ctx.QueryParams(), &params.Addr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter addr: %s", err))
	}

	// ------------- Optional query parameter "message_type" -------------
	if paramValue := ctx.QueryParam("message_type"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "message_type", ctx.QueryParams(), &params.MessageType)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter message_type: %s", err))
	}

	// ------------- Optional query parameter "type" -------------
	if paramValue := ctx.QueryParam("type"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "type", ctx.QueryParams(), &params.Type)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter type: %s", err))
	}

	// ------------- Optional query parameter "hash" -------------
	if paramValue := ctx.QueryParam("hash"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "hash", ctx.QueryParams(), &params.Hash)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter hash: %s", err))
	}

	// ------------- Optional query parameter "lt" -------------
	if paramValue := ctx.QueryParam("lt"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "lt", ctx.QueryParams(), &params.Lt)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter lt: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetV1BlockTransactions(ctx, params)
	return err
}

// GetV1BlocksFeed converts echo context to params.
func (w *ServerInterfaceWrapper) GetV1BlocksFeed(ctx echo.Context) error {
	var err error

	// Parameter object where we will unmarshal all parameters from the context
	var params GetV1BlocksFeedParams
	// ------------- Optional query parameter "scroll_id" -------------
	if paramValue := ctx.QueryParam("scroll_id"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "scroll_id", ctx.QueryParams(), &params.ScrollId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter scroll_id: %s", err))
	}

	// ------------- Optional query parameter "limit" -------------
	if paramValue := ctx.QueryParam("limit"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "limit", ctx.QueryParams(), &params.Limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter limit: %s", err))
	}

	// ------------- Optional query parameter "workchain_id" -------------
	if paramValue := ctx.QueryParam("workchain_id"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "workchain_id", ctx.QueryParams(), &params.WorkchainId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter workchain_id: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetV1BlocksFeed(ctx, params)
	return err
}

// GetV1HeightBlockchain converts echo context to params.
func (w *ServerInterfaceWrapper) GetV1HeightBlockchain(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetV1HeightBlockchain(ctx)
	return err
}

// GetV1HeightSynced converts echo context to params.
func (w *ServerInterfaceWrapper) GetV1HeightSynced(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetV1HeightSynced(ctx)
	return err
}

// GetV1MasterBlockShardsActual converts echo context to params.
func (w *ServerInterfaceWrapper) GetV1MasterBlockShardsActual(ctx echo.Context) error {
	var err error

	// Parameter object where we will unmarshal all parameters from the context
	var params GetV1MasterBlockShardsActualParams
	// ------------- Required query parameter "block_master" -------------
	if paramValue := ctx.QueryParam("block_master"); paramValue != "" {

	} else {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Query argument block_master is required, but not found"))
	}

	err = runtime.BindQueryParameter("form", true, true, "block_master", ctx.QueryParams(), &params.BlockMaster)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter block_master: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetV1MasterBlockShardsActual(ctx, params)
	return err
}

// GetV1MasterBlockShardsRange converts echo context to params.
func (w *ServerInterfaceWrapper) GetV1MasterBlockShardsRange(ctx echo.Context) error {
	var err error

	// Parameter object where we will unmarshal all parameters from the context
	var params GetV1MasterBlockShardsRangeParams
	// ------------- Required query parameter "block_master" -------------
	if paramValue := ctx.QueryParam("block_master"); paramValue != "" {

	} else {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Query argument block_master is required, but not found"))
	}

	err = runtime.BindQueryParameter("form", true, true, "block_master", ctx.QueryParams(), &params.BlockMaster)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter block_master: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetV1MasterBlockShardsRange(ctx, params)
	return err
}

// GetV1MessageGet converts echo context to params.
func (w *ServerInterfaceWrapper) GetV1MessageGet(ctx echo.Context) error {
	var err error

	// Parameter object where we will unmarshal all parameters from the context
	var params GetV1MessageGetParams
	// ------------- Required query parameter "trx_hash" -------------
	if paramValue := ctx.QueryParam("trx_hash"); paramValue != "" {

	} else {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Query argument trx_hash is required, but not found"))
	}

	err = runtime.BindQueryParameter("form", true, true, "trx_hash", ctx.QueryParams(), &params.TrxHash)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter trx_hash: %s", err))
	}

	// ------------- Required query parameter "message_lt" -------------
	if paramValue := ctx.QueryParam("message_lt"); paramValue != "" {

	} else {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Query argument message_lt is required, but not found"))
	}

	err = runtime.BindQueryParameter("form", true, true, "message_lt", ctx.QueryParams(), &params.MessageLt)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter message_lt: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetV1MessageGet(ctx, params)
	return err
}

// GetV1MessagesFeed converts echo context to params.
func (w *ServerInterfaceWrapper) GetV1MessagesFeed(ctx echo.Context) error {
	var err error

	// Parameter object where we will unmarshal all parameters from the context
	var params GetV1MessagesFeedParams
	// ------------- Optional query parameter "scroll_id" -------------
	if paramValue := ctx.QueryParam("scroll_id"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "scroll_id", ctx.QueryParams(), &params.ScrollId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter scroll_id: %s", err))
	}

	// ------------- Optional query parameter "limit" -------------
	if paramValue := ctx.QueryParam("limit"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "limit", ctx.QueryParams(), &params.Limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter limit: %s", err))
	}

	// ------------- Optional query parameter "workchain_id" -------------
	if paramValue := ctx.QueryParam("workchain_id"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "workchain_id", ctx.QueryParams(), &params.WorkchainId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter workchain_id: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetV1MessagesFeed(ctx, params)
	return err
}

// GetV1Search converts echo context to params.
func (w *ServerInterfaceWrapper) GetV1Search(ctx echo.Context) error {
	var err error

	// Parameter object where we will unmarshal all parameters from the context
	var params GetV1SearchParams
	// ------------- Required query parameter "q" -------------
	if paramValue := ctx.QueryParam("q"); paramValue != "" {

	} else {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Query argument q is required, but not found"))
	}

	err = runtime.BindQueryParameter("form", true, true, "q", ctx.QueryParams(), &params.Q)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter q: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetV1Search(ctx, params)
	return err
}

// GetV1StatsAddresses converts echo context to params.
func (w *ServerInterfaceWrapper) GetV1StatsAddresses(ctx echo.Context) error {
	var err error

	// Parameter object where we will unmarshal all parameters from the context
	var params GetV1StatsAddressesParams
	// ------------- Optional query parameter "workchain_id" -------------
	if paramValue := ctx.QueryParam("workchain_id"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "workchain_id", ctx.QueryParams(), &params.WorkchainId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter workchain_id: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetV1StatsAddresses(ctx, params)
	return err
}

// GetV1StatsBlocks converts echo context to params.
func (w *ServerInterfaceWrapper) GetV1StatsBlocks(ctx echo.Context) error {
	var err error

	// Parameter object where we will unmarshal all parameters from the context
	var params GetV1StatsBlocksParams
	// ------------- Optional query parameter "workchain_id" -------------
	if paramValue := ctx.QueryParam("workchain_id"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "workchain_id", ctx.QueryParams(), &params.WorkchainId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter workchain_id: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetV1StatsBlocks(ctx, params)
	return err
}

// GetV1StatsGlobal converts echo context to params.
func (w *ServerInterfaceWrapper) GetV1StatsGlobal(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetV1StatsGlobal(ctx)
	return err
}

// GetV1StatsMessages converts echo context to params.
func (w *ServerInterfaceWrapper) GetV1StatsMessages(ctx echo.Context) error {
	var err error

	// Parameter object where we will unmarshal all parameters from the context
	var params GetV1StatsMessagesParams
	// ------------- Optional query parameter "workchain_id" -------------
	if paramValue := ctx.QueryParam("workchain_id"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "workchain_id", ctx.QueryParams(), &params.WorkchainId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter workchain_id: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetV1StatsMessages(ctx, params)
	return err
}

// GetV1StatsTransactions converts echo context to params.
func (w *ServerInterfaceWrapper) GetV1StatsTransactions(ctx echo.Context) error {
	var err error

	// Parameter object where we will unmarshal all parameters from the context
	var params GetV1StatsTransactionsParams
	// ------------- Optional query parameter "workchain_id" -------------
	if paramValue := ctx.QueryParam("workchain_id"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "workchain_id", ctx.QueryParams(), &params.WorkchainId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter workchain_id: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetV1StatsTransactions(ctx, params)
	return err
}

// GetV1TimeseriesBlocksByWorkchain converts echo context to params.
func (w *ServerInterfaceWrapper) GetV1TimeseriesBlocksByWorkchain(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetV1TimeseriesBlocksByWorkchain(ctx)
	return err
}

// GetV1TimeseriesMessagesByType converts echo context to params.
func (w *ServerInterfaceWrapper) GetV1TimeseriesMessagesByType(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetV1TimeseriesMessagesByType(ctx)
	return err
}

// GetV1TimeseriesMessagesOrdCount converts echo context to params.
func (w *ServerInterfaceWrapper) GetV1TimeseriesMessagesOrdCount(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetV1TimeseriesMessagesOrdCount(ctx)
	return err
}

// GetV1TimeseriesSentAndFees converts echo context to params.
func (w *ServerInterfaceWrapper) GetV1TimeseriesSentAndFees(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetV1TimeseriesSentAndFees(ctx)
	return err
}

// GetV1TimeseriesVolumeByGrams converts echo context to params.
func (w *ServerInterfaceWrapper) GetV1TimeseriesVolumeByGrams(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetV1TimeseriesVolumeByGrams(ctx)
	return err
}

// GetV1TopWhales converts echo context to params.
func (w *ServerInterfaceWrapper) GetV1TopWhales(ctx echo.Context) error {
	var err error

	// Parameter object where we will unmarshal all parameters from the context
	var params GetV1TopWhalesParams
	// ------------- Optional query parameter "workchain_id" -------------
	if paramValue := ctx.QueryParam("workchain_id"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "workchain_id", ctx.QueryParams(), &params.WorkchainId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter workchain_id: %s", err))
	}

	// ------------- Optional query parameter "limit" -------------
	if paramValue := ctx.QueryParam("limit"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "limit", ctx.QueryParams(), &params.Limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter limit: %s", err))
	}

	// ------------- Optional query parameter "offset" -------------
	if paramValue := ctx.QueryParam("offset"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "offset", ctx.QueryParams(), &params.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter offset: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetV1TopWhales(ctx, params)
	return err
}

// GetV1Transaction converts echo context to params.
func (w *ServerInterfaceWrapper) GetV1Transaction(ctx echo.Context) error {
	var err error

	// Parameter object where we will unmarshal all parameters from the context
	var params GetV1TransactionParams
	// ------------- Optional query parameter "hash" -------------
	if paramValue := ctx.QueryParam("hash"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "hash", ctx.QueryParams(), &params.Hash)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter hash: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetV1Transaction(ctx, params)
	return err
}

// GetV1TransactionsFeed converts echo context to params.
func (w *ServerInterfaceWrapper) GetV1TransactionsFeed(ctx echo.Context) error {
	var err error

	// Parameter object where we will unmarshal all parameters from the context
	var params GetV1TransactionsFeedParams
	// ------------- Optional query parameter "scroll_id" -------------
	if paramValue := ctx.QueryParam("scroll_id"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "scroll_id", ctx.QueryParams(), &params.ScrollId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter scroll_id: %s", err))
	}

	// ------------- Optional query parameter "limit" -------------
	if paramValue := ctx.QueryParam("limit"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "limit", ctx.QueryParams(), &params.Limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter limit: %s", err))
	}

	// ------------- Optional query parameter "workchain_id" -------------
	if paramValue := ctx.QueryParam("workchain_id"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "workchain_id", ctx.QueryParams(), &params.WorkchainId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter workchain_id: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetV1TransactionsFeed(ctx, params)
	return err
}

// GetV1WorkchainBlockMaster converts echo context to params.
func (w *ServerInterfaceWrapper) GetV1WorkchainBlockMaster(ctx echo.Context) error {
	var err error

	// Parameter object where we will unmarshal all parameters from the context
	var params GetV1WorkchainBlockMasterParams
	// ------------- Required query parameter "block" -------------
	if paramValue := ctx.QueryParam("block"); paramValue != "" {

	} else {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Query argument block is required, but not found"))
	}

	err = runtime.BindQueryParameter("form", true, true, "block", ctx.QueryParams(), &params.Block)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter block: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetV1WorkchainBlockMaster(ctx, params)
	return err
}

// RegisterHandlers adds each server route to the EchoRouter.
func RegisterHandlers(router interface {
	CONNECT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	DELETE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	GET(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	HEAD(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	OPTIONS(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PATCH(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	POST(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PUT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	TRACE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
}, si ServerInterface) {

	wrapper := ServerInterfaceWrapper{
		Handler: si,
	}

	router.GET("/v1/account", wrapper.GetV1Account)
	router.GET("/v1/account/messages", wrapper.GetV1AccountMessages)
	router.GET("/v1/account/qr", wrapper.GetV1AccountQr)
	router.GET("/v1/addr/top-by-message-count", wrapper.GetV1AddrTopByMessageCount)
	router.GET("/v1/block/info", wrapper.GetV1BlockInfo)
	router.GET("/v1/block/tlb", wrapper.GetV1BlockTlb)
	router.GET("/v1/block/transactions", wrapper.GetV1BlockTransactions)
	router.GET("/v1/blocks/feed", wrapper.GetV1BlocksFeed)
	router.GET("/v1/height/blockchain", wrapper.GetV1HeightBlockchain)
	router.GET("/v1/height/synced", wrapper.GetV1HeightSynced)
	router.GET("/v1/master/block/shards/actual", wrapper.GetV1MasterBlockShardsActual)
	router.GET("/v1/master/block/shards/range", wrapper.GetV1MasterBlockShardsRange)
	router.GET("/v1/message/get", wrapper.GetV1MessageGet)
	router.GET("/v1/messages/feed", wrapper.GetV1MessagesFeed)
	router.GET("/v1/search", wrapper.GetV1Search)
	router.GET("/v1/stats/addresses", wrapper.GetV1StatsAddresses)
	router.GET("/v1/stats/blocks", wrapper.GetV1StatsBlocks)
	router.GET("/v1/stats/global", wrapper.GetV1StatsGlobal)
	router.GET("/v1/stats/messages", wrapper.GetV1StatsMessages)
	router.GET("/v1/stats/transactions", wrapper.GetV1StatsTransactions)
	router.GET("/v1/timeseries/blocks-by-workchain", wrapper.GetV1TimeseriesBlocksByWorkchain)
	router.GET("/v1/timeseries/messages-by-type", wrapper.GetV1TimeseriesMessagesByType)
	router.GET("/v1/timeseries/messages-ord-count", wrapper.GetV1TimeseriesMessagesOrdCount)
	router.GET("/v1/timeseries/sent-and-fees", wrapper.GetV1TimeseriesSentAndFees)
	router.GET("/v1/timeseries/volume-by-grams", wrapper.GetV1TimeseriesVolumeByGrams)
	router.GET("/v1/top/whales", wrapper.GetV1TopWhales)
	router.GET("/v1/transaction", wrapper.GetV1Transaction)
	router.GET("/v1/transactions/feed", wrapper.GetV1TransactionsFeed)
	router.GET("/v1/workchain/block/master", wrapper.GetV1WorkchainBlockMaster)

}