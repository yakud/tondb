package server

import (
	"github.com/labstack/echo/v4"
	v1 "gitlab.flora.loc/mills/tondb/internal/server/endpoint/v1"
	"gitlab.flora.loc/mills/tondb/swagger/tonapi"
)

type TonApi struct {
	v1.GetAccount
	v1.GetAccountMessages
}

// TODO: move endpoints to internal/swagger/endpoint/v1

func (t *TonApi) GetV1AccountQr(ctx echo.Context, params tonapi.GetV1AccountQrParams) error {
	panic("implement me")
}

func (t *TonApi) GetV1AddrTopByMessageCount(ctx echo.Context) error {
	panic("implement me")
}

func (t *TonApi) GetV1BlockInfo(ctx echo.Context, params tonapi.GetV1BlockInfoParams) error {
	panic("implement me")
}

func (t *TonApi) GetV1BlockTlb(ctx echo.Context, params tonapi.GetV1BlockTlbParams) error {
	panic("implement me")
}

func (t *TonApi) GetV1BlockTransactions(ctx echo.Context, params tonapi.GetV1BlockTransactionsParams) error {
	panic("implement me")
}

func (t *TonApi) GetV1BlocksFeed(ctx echo.Context, params tonapi.GetV1BlocksFeedParams) error {
	panic("implement me")
}

func (t *TonApi) GetV1HeightBlockchain(ctx echo.Context) error {
	panic("implement me")
}

func (t *TonApi) GetV1HeightSynced(ctx echo.Context) error {
	panic("implement me")
}

func (t *TonApi) GetV1MasterBlockShardsActual(ctx echo.Context, params tonapi.GetV1MasterBlockShardsActualParams) error {
	panic("implement me")
}

func (t *TonApi) GetV1MasterBlockShardsRange(ctx echo.Context, params tonapi.GetV1MasterBlockShardsRangeParams) error {
	panic("implement me")
}

func (t *TonApi) GetV1MessageGet(ctx echo.Context, params tonapi.GetV1MessageGetParams) error {
	panic("implement me")
}

func (t *TonApi) GetV1MessagesFeed(ctx echo.Context, params tonapi.GetV1MessagesFeedParams) error {
	panic("implement me")
}

func (t *TonApi) GetV1Search(ctx echo.Context, params tonapi.GetV1SearchParams) error {
	panic("implement me")
}

func (t *TonApi) GetV1StatsAddresses(ctx echo.Context, params tonapi.GetV1StatsAddressesParams) error {
	panic("implement me")
}

func (t *TonApi) GetV1StatsBlocks(ctx echo.Context, params tonapi.GetV1StatsBlocksParams) error {
	panic("implement me")
}

func (t *TonApi) GetV1StatsGlobal(ctx echo.Context) error {
	panic("implement me")
}

func (t *TonApi) GetV1StatsMessages(ctx echo.Context, params tonapi.GetV1StatsMessagesParams) error {
	panic("implement me")
}

func (t *TonApi) GetV1StatsTransactions(ctx echo.Context, params tonapi.GetV1StatsTransactionsParams) error {
	panic("implement me")
}

func (t *TonApi) GetV1TimeseriesBlocksByWorkchain(ctx echo.Context) error {
	panic("implement me")
}

func (t *TonApi) GetV1TimeseriesMessagesByType(ctx echo.Context) error {
	panic("implement me")
}

func (t *TonApi) GetV1TimeseriesMessagesOrdCount(ctx echo.Context) error {
	panic("implement me")
}

func (t *TonApi) GetV1TimeseriesSentAndFees(ctx echo.Context) error {
	panic("implement me")
}

func (t *TonApi) GetV1TimeseriesVolumeByGrams(ctx echo.Context) error {
	panic("implement me")
}

func (t *TonApi) GetV1TopWhales(ctx echo.Context, params tonapi.GetV1TopWhalesParams) error {
	panic("implement me")
}

func (t *TonApi) GetV1Transaction(ctx echo.Context, params tonapi.GetV1TransactionParams) error {
	panic("implement me")
}

func (t *TonApi) GetV1TransactionsFeed(ctx echo.Context, params tonapi.GetV1TransactionsFeedParams) error {
	panic("implement me")
}

func (t *TonApi) GetV1WorkchainBlockMaster(ctx echo.Context, params tonapi.GetV1WorkchainBlockMasterParams) error {
	panic("implement me")
}
