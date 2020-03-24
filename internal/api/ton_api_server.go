package api

import (
	"gitlab.flora.loc/mills/tondb/internal/api/feed"
	"gitlab.flora.loc/mills/tondb/internal/api/site"
	"gitlab.flora.loc/mills/tondb/internal/api/stats"
	"gitlab.flora.loc/mills/tondb/internal/api/timeseries"
)

type TonApiServer struct {
	*GetAccount
	*GetAccountMessages
	*site.GetAddrTopByMessageCount
	*GetBlockInfo
	*GetBlockTlb
	*GetBlockTransactions
	*feed.GetBlocksFeed
	*GetBlockchainHeight
	*GetSyncedHeight
	*MasterchainBlockShardsActual
	*MasterchainBlockShards
	*GetMessage
	*feed.GetMessagesFeed
	*Search
	*stats.AddressesMetrics
	*stats.BlocksMetrics
	*stats.GlobalMetrics
	*stats.MessagesMetrics
	*stats.TrxMetrics
	*timeseries.BlocksByWorkchain
	*timeseries.MessagesByType
	*timeseries.MessagesOrdCount
	*timeseries.SentAndFees
	*timeseries.VolumeByGrams
	*site.GetTopWhales
	*GetTransactions
	*feed.GetTransactionsFeed
	*GetWorkchainBlockMaster
}

func NewTonApiServer(getAccount *GetAccount, getAccountMessages *GetAccountMessages,
	addrTopByMessageCount *site.GetAddrTopByMessageCount, getBlockInfo *GetBlockInfo, getBlockTlb *GetBlockTlb,
	getBlockTransactions *GetBlockTransactions, blocksFeed *feed.GetBlocksFeed, blockchainHeight *GetBlockchainHeight,
	syncedHeight *GetSyncedHeight, masterchainBlockShardsActual *MasterchainBlockShardsActual,
	masterchainBlockShards *MasterchainBlockShards, getMessage *GetMessage, getMessagesFeed *feed.GetMessagesFeed,
	search *Search, addressesMetrics *stats.AddressesMetrics, blocksMetrics *stats.BlocksMetrics,
	globalMetrics *stats.GlobalMetrics, messagesMetrics *stats.MessagesMetrics, trxMetrics *stats.TrxMetrics,
	blocksByWorkchain *timeseries.BlocksByWorkchain, messagesByType *timeseries.MessagesByType,
	messagesOrdCount *timeseries.MessagesOrdCount, sentAndFees *timeseries.SentAndFees, volumeByGrams *timeseries.VolumeByGrams,
	getTopWhales *site.GetTopWhales, getTransactions *GetTransactions, getTransactionsFeed *feed.GetTransactionsFeed,
	getWorkchainBlockMaster *GetWorkchainBlockMaster) *TonApiServer {
	return &TonApiServer{
		GetAccount:                   getAccount,
		GetAccountMessages:           getAccountMessages,
		GetAddrTopByMessageCount:     addrTopByMessageCount,
		GetBlockInfo:                 getBlockInfo,
		GetBlockTlb:                  getBlockTlb,
		GetBlockTransactions:         getBlockTransactions,
		GetBlocksFeed:                blocksFeed,
		GetBlockchainHeight:          blockchainHeight,
		GetSyncedHeight:              syncedHeight,
		MasterchainBlockShardsActual: masterchainBlockShardsActual,
		MasterchainBlockShards:       masterchainBlockShards,
		GetMessage:                   getMessage,
		GetMessagesFeed:              getMessagesFeed,
		Search:                       search,
		AddressesMetrics:             addressesMetrics,
		BlocksMetrics:                blocksMetrics,
		GlobalMetrics:                globalMetrics,
		MessagesMetrics:              messagesMetrics,
		TrxMetrics:                   trxMetrics,
		BlocksByWorkchain:            blocksByWorkchain,
		MessagesByType:               messagesByType,
		MessagesOrdCount:             messagesOrdCount,
		SentAndFees:                  sentAndFees,
		VolumeByGrams:                volumeByGrams,
		GetTopWhales:                 getTopWhales,
		GetTransactions:              getTransactions,
		GetTransactionsFeed:          getTransactionsFeed,
		GetWorkchainBlockMaster:      getWorkchainBlockMaster,
	}
}
