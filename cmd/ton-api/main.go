package main

import (
	"log"
	"net/http"
	"os"

	"gitlab.flora.loc/mills/tondb/internal/blocks_fetcher"

	"gitlab.flora.loc/mills/tondb/internal/ton/view/stats"

	"gitlab.flora.loc/mills/tondb/internal/api/site"
	"gitlab.flora.loc/mills/tondb/internal/api/timeseries"

	"gitlab.flora.loc/mills/tondb/internal/ton/view/feed"
	"gitlab.flora.loc/mills/tondb/internal/ton/view/state"

	"github.com/rs/cors"

	"gitlab.flora.loc/mills/tondb/internal/api"
	"gitlab.flora.loc/mills/tondb/internal/ch"
	"gitlab.flora.loc/mills/tondb/internal/ton/query"
	"gitlab.flora.loc/mills/tondb/internal/ton/storage"

	"github.com/julienschmidt/httprouter"

	statsQ "gitlab.flora.loc/mills/tondb/internal/ton/query/stats"
	timeseriesQ "gitlab.flora.loc/mills/tondb/internal/ton/query/timeseries"
	timeseriesV "gitlab.flora.loc/mills/tondb/internal/ton/view/timeseries"
)

var (
	blocksRootAliases  = [...]string{"/blocks", "/block", "/b"}
	addressRootAliases = [...]string{"/address", "/account", "/a"}
)

func main() {
	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = "0.0.0.0:8512"
	}

	chAddr := os.Getenv("CH_ADDR")
	if chAddr == "" {
		chAddr = "http://default:V9AQZJFNX4ygj2vP@192.168.100.3:8123/ton2?max_query_size=3145728000"
	}
	blocksFetcherAddr := os.Getenv("TLB_BLOCKS_FETCHER_ADDR")
	if blocksFetcherAddr == "" {
		blocksFetcherAddr = "127.0.0.1:13699"
	}
	chConnect, err := ch.Connect(&chAddr)
	if err != nil {
		log.Fatal(err)
	}

	blocksFetcher, err := blocks_fetcher.NewClient(blocksFetcherAddr)
	if err != nil {
		log.Fatal(err)
	}

	router := httprouter.New()

	// Core API
	//blocksStorage := storage.NewBlocks(chConnect)
	//transactionsStorage := storage.NewTransactions(chConnect)
	shardsDescrStorage := storage.NewShardsDescr(chConnect)
	if err := shardsDescrStorage.CreateTable(); err != nil {
		log.Fatal(err)
	}

	accountState := state.NewAccountState(chConnect)
	if err := accountState.CreateTable(); err != nil {
		log.Fatal(err)
	}

	accountTransactions := feed.NewAccountTransactions(chConnect)
	if err := accountTransactions.CreateTable(); err != nil {
		log.Fatal(err)
	}

	blocksFeed := feed.NewBlocksFeed(chConnect)
	if err := blocksFeed.CreateTable(); err != nil {
		log.Fatal(err)
	}

	syncedHeightQuery := query.NewGetSyncedHeight(chConnect)
	blockchainHeightQuery := query.NewGetBlockchainHeight(chConnect)
	searchTransactionsQuery := query.NewSearchTransactions(chConnect)
	getBlockInfoQuery := query.NewGetBlockInfo(chConnect)

	router.GET("/height/synced", api.BasicAuth(api.NewGetSyncedHeight(syncedHeightQuery).Handler))
	router.GET("/height/blockchain", api.BasicAuth(api.NewGetBlockchainHeight(blockchainHeightQuery).Handler))
	router.GET("/master/block/shards/range", api.BasicAuth(api.NewMasterBlockShardsRange(shardsDescrStorage).Handler))
	router.GET("/workchain/block/master", api.BasicAuth(api.NewGetWorkchainBlockMaster(shardsDescrStorage).Handler))
	router.GET("/transaction", api.BasicAuth(api.NewGetTransactions(searchTransactionsQuery).Handler))
	router.GET("/block/tlb", api.BasicAuth(api.NewGetBlockTlb(blocksFetcher).Handler))

	// Block routes
	for _, blockRoot := range blocksRootAliases {
		router.GET(blockRoot+"/info", api.BasicAuth(api.NewGetBlockInfo(getBlockInfoQuery, shardsDescrStorage).Handler))
		router.GET(blockRoot+"/transactions", api.BasicAuth(api.NewGetBlockTransactions(searchTransactionsQuery, shardsDescrStorage).Handler))
		router.GET(blockRoot+"/feed", api.BasicAuth(api.NewGetBlocksFeed(blocksFeed).Handler))
	}

	// Address (account) routes
	for _, addrRoot := range addressRootAliases {
		router.GET(addrRoot, api.BasicAuth(api.NewGetAccount(accountState).Handler))
		router.GET(addrRoot+"/transactions", api.BasicAuth(api.NewGetAccountTransactions(accountTransactions).Handler))
	}

	// Main API
	vBlocksByWorkchain := timeseriesV.NewBlocksByWorkchain(chConnect)
	if err := vBlocksByWorkchain.CreateTable(); err != nil {
		log.Fatal(err)
	}
	qBlocksByWorkchain := timeseriesQ.NewGetBlocksByWorkchain(chConnect)

	tsMessagesByType := timeseriesV.NewMessagesByType(chConnect)
	if err := tsMessagesByType.CreateTable(); err != nil {
		log.Fatal(err)
	}

	tsVolumeByGrams := timeseriesV.NewVolumeByGrams(chConnect)
	if err := tsVolumeByGrams.CreateTable(); err != nil {
		log.Fatal(err)
	}

	tsMessagesOrdCount := timeseriesV.NewMessagesOrdCount(chConnect)
	if err := tsMessagesOrdCount.CreateTable(); err != nil {
		log.Fatal(err)
	}

	messagesFeedGlobal := feed.NewMessagesFeedGlobal(chConnect)
	if err := messagesFeedGlobal.CreateTable(); err != nil {
		log.Fatal(err)
	}

	addrMessagesCount := stats.NewAddrMessagesCount(chConnect)
	if err := addrMessagesCount.CreateTable(); err != nil {
		log.Fatal(err)
	}

	qGetTopWhales := statsQ.NewGetTopWhales(chConnect)

	router.GET("/timeseries/blocks-by-workchain", timeseries.NewBlocksByWorkchain(qBlocksByWorkchain).Handler)
	router.GET("/timeseries/messages-by-type", timeseries.NewMessagesByType(tsMessagesByType).Handler)
	router.GET("/timeseries/volume-by-grams", timeseries.NewVolumeByGrams(tsVolumeByGrams).Handler)
	router.GET("/timeseries/messages-ord-count", timeseries.NewMessagesOrdCount(tsMessagesOrdCount).Handler)
	router.GET("/messages/latest", site.NewGetLatestMessages(messagesFeedGlobal).Handler)
	router.GET("/addr/top-by-message-count", site.NewGetAddrTopByMessageCount(addrMessagesCount).Handler)
	router.GET("/top/whales", site.NewGetTopWhales(qGetTopWhales).Handler)

	handler := cors.AllowAll().Handler(router)
	srv := &http.Server{
		Addr:    addr,
		Handler: handler,
	}
	log.Println("Start listening", addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
