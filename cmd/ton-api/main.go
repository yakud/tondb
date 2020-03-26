package main

import (
	"context"
	"gitlab.flora.loc/mills/tondb/internal/api/middleware"
	"gitlab.flora.loc/mills/tondb/internal/api/site"
	"gitlab.flora.loc/mills/tondb/internal/api/swagger"
	"gitlab.flora.loc/mills/tondb/internal/blocks_fetcher"
	"gitlab.flora.loc/mills/tondb/internal/ton/query"
	"gitlab.flora.loc/mills/tondb/internal/ton/query/cache"
	"gitlab.flora.loc/mills/tondb/internal/ton/search"
	"gitlab.flora.loc/mills/tondb/internal/ton/view/index"
	"gitlab.flora.loc/mills/tondb/internal/ton/view/stats"
	"time"

	"gitlab.flora.loc/mills/tondb/swagger/tonapi"
	"log"

	"github.com/jmoiron/sqlx"

	"gitlab.flora.loc/mills/tondb/internal/api"
	apifeed "gitlab.flora.loc/mills/tondb/internal/api/feed"
	"gitlab.flora.loc/mills/tondb/internal/api/ratelimit"
	statsApi "gitlab.flora.loc/mills/tondb/internal/api/stats"
	"gitlab.flora.loc/mills/tondb/internal/api/timeseries"
	"gitlab.flora.loc/mills/tondb/internal/ch"
	statsQ "gitlab.flora.loc/mills/tondb/internal/ton/query/stats"
	timeseriesQ "gitlab.flora.loc/mills/tondb/internal/ton/query/timeseries"
	"gitlab.flora.loc/mills/tondb/internal/ton/storage"
	"gitlab.flora.loc/mills/tondb/internal/ton/view/feed"
	"gitlab.flora.loc/mills/tondb/internal/ton/view/state"
	timeseriesV "gitlab.flora.loc/mills/tondb/internal/ton/view/timeseries"

	"github.com/go-redis/redis"
	"github.com/kelseyhightower/envconfig"
	"github.com/labstack/echo/v4"
)

const ApiV1 = "/v1"

var (
	blocksRootAliases  = [...]string{"/blocks", "/block", "/b"}
	addressRootAliases = [...]string{"/address", "/account", "/a"}
	router             = echo.New()
)

func main() {
	config := api.Config{}
	if err := envconfig.Process("", &config); err != nil {
		log.Fatal(err.Error())
	}
	chConnect, err := ch.Connect(&config.ChAddr)
	if err != nil {
		log.Fatal(err)
	}
	chConnectSqlx := sqlx.NewDb(chConnect, "clickhouse")

	blocksFetcher, err := blocks_fetcher.NewClient(config.TlbBlocksFetcherAddr)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("connect redis:", config.RedisAddr)
	redisClient := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,
		Password: config.RedisPassword,
		DB:       config.RedisDB,
	})
	if pong, err := redisClient.Ping().Result(); pong != "PONG" || err != nil {
		log.Fatalf("error redis connect: %s %v", config.RedisAddr, err)
	}

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

	accountTransactions := feed.NewAccountMessages(chConnect)
	if err := accountTransactions.CreateTable(); err != nil {
		log.Fatal(err)
	}

	blocksFeed := feed.NewBlocksFeed(chConnectSqlx)
	if err := blocksFeed.CreateTable(); err != nil {
		log.Fatal(err)
	}

	syncedHeightQuery := query.NewGetSyncedHeight(chConnect)
	blockchainHeightQuery := query.NewGetBlockchainHeight(chConnect)
	searchTransactionsQuery := query.NewSearchTransactions(chConnect)
	getBlockInfoQuery := query.NewGetBlockInfo(chConnect)
	indexReverseBlockSeqNo := index.NewIndexReverseBlockSeqNo(chConnect)
	indexHash := index.NewIndexHash(chConnect)

	searcher := search.NewSearcher(
		accountState,
		getBlockInfoQuery,
		indexReverseBlockSeqNo,
		indexHash,
	)

	if err := ratelimit.RateLimitLua.Load(redisClient).Err(); err != nil {
		log.Fatal("error load redis lua script:", err)
	}
	rateLimiter := ratelimit.NewRateLimiter(redisClient)
	rateLimitMiddleware := middleware.RateLimit(rateLimiter)

	getBlockInfo := api.NewGetBlockInfo(getBlockInfoQuery, shardsDescrStorage)
	getBlockTransactions := api.NewGetBlockTransactions(searchTransactionsQuery, shardsDescrStorage)
	getBlocksFeed := apifeed.NewGetBlocksFeed(blocksFeed)

	getAccountHandler := api.NewGetAccount(accountState)
	getAccountMessages := api.NewGetAccountMessages(accountTransactions)

	// Messages feed
	messagesFeedGlobal := feed.NewMessagesFeed(chConnectSqlx)
	if err := messagesFeedGlobal.CreateTable(); err != nil {
		log.Fatal(err)
	}

	getMessageQuery := query.NewGetMessage(chConnect)

	// Transactions feed
	transactionsFeed := feed.NewTransactionsFeed(chConnectSqlx)
	if err := transactionsFeed.CreateTable(); err != nil {
		log.Fatal(err)
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

	addrMessagesCount := stats.NewAddrMessagesCount(chConnect)
	if err := addrMessagesCount.CreateTable(); err != nil {
		log.Fatal(err)
	}

	accountMessagesCount := stats.NewAccountMessagesCount(chConnect)
	if err := accountMessagesCount.CreateTable(); err != nil {
		log.Fatal(err)
	}

	ctxBgCache, _ := context.WithCancel(context.Background())
	metricsCache := cache.NewBackground()
	blocksCache := cache.NewBackground()
	whalesCache := cache.NewBackground()

	messagesMetrics := statsQ.NewMessagesMetrics(chConnect, metricsCache)
	if err := messagesMetrics.CreateTable(); err != nil {
		log.Fatal(err)
	}
	if err := messagesMetrics.UpdateQuery(); err != nil {
		log.Fatal("messagesMetrics error:", err)
	}

	globalMetrics := statsQ.NewGlobalMetrics(chConnect, metricsCache)
	if err := globalMetrics.UpdateQuery(); err != nil {
		log.Fatal("globalMetrics error:", err)
	}

	blocksMetrics := statsQ.NewBlocksMetrics(chConnect, metricsCache)
	if err := blocksMetrics.UpdateQuery(); err != nil {
		log.Fatal("blocksMetrics error:", err)
	}

	addressesMetrics := statsQ.NewAddressesMetrics(chConnect, metricsCache)
	if err := addressesMetrics.UpdateQuery(); err != nil {
		log.Fatal("addressesMetrics error:", err)
	}

	trxMetrics := statsQ.NewTrxMetrics(chConnect, metricsCache)
	if err := trxMetrics.UpdateQuery(); err != nil {
		log.Fatal("trxMetrics error:", err)
	}

	topWhales := statsQ.NewGetTopWhales(chConnect, whalesCache, addressesMetrics)
	if err := topWhales.UpdateQuery(); err != nil {
		log.Fatal("topWhales error:", err)
	}

	sentAndFees := timeseriesQ.NewSentAndFees(chConnect, whalesCache)
	if err := sentAndFees.UpdateQuery(); err != nil {
		log.Fatal("sentAndFees error:", err)
	}

	metricsCache.AddQuery(globalMetrics)
	metricsCache.AddQuery(addressesMetrics)
	metricsCache.AddQuery(messagesMetrics)
	metricsCache.AddQuery(trxMetrics)
	blocksCache.AddQuery(blocksMetrics)
	whalesCache.AddQuery(topWhales)
	whalesCache.AddQuery(sentAndFees)

	go func() {
		metricsCache.RunTicker(ctxBgCache, 5*time.Second)
	}()

	go func() {
		whalesCache.RunTicker(ctxBgCache, 10*time.Second)
	}()

	go func() {
		blocksCache.RunTicker(ctxBgCache, 1*time.Second)
	}()

	tonApiServer := api.NewTonApiServer(getAccountHandler, getAccountMessages, site.NewGetAddrTopByMessageCount(addrMessagesCount),
		getBlockInfo, api.NewGetBlockTlb(blocksFetcher), getBlockTransactions, getBlocksFeed,
		api.NewGetBlockchainHeight(blockchainHeightQuery), api.NewGetSyncedHeight(syncedHeightQuery),
		api.NewMasterchainBlockShardsActual(shardsDescrStorage), api.NewMasterBlockShardsRange(shardsDescrStorage),
		api.NewGetMessage(getMessageQuery), apifeed.NewGetMessagesFeed(messagesFeedGlobal), api.NewSearch(searcher),
		statsApi.NewAddressesMetrics(addressesMetrics), statsApi.NewBlocksMetrics(blocksMetrics),
		statsApi.NewGlobalMetrics(globalMetrics), statsApi.NewMessagesMetrics(messagesMetrics),
		statsApi.NewTrxMetrics(trxMetrics), timeseries.NewBlocksByWorkchain(qBlocksByWorkchain),
		timeseries.NewMessagesByType(tsMessagesByType), timeseries.NewMessagesOrdCount(tsMessagesOrdCount),
		timeseries.NewSentAndFees(sentAndFees), timeseries.NewVolumeByGrams(tsVolumeByGrams), site.NewGetTopWhales(topWhales),
		api.NewGetTransactions(searchTransactionsQuery), apifeed.NewGetTransactionsFeed(transactionsFeed),
		api.NewGetWorkchainBlockMaster(shardsDescrStorage))

	router.Pre(rateLimitMiddleware)
	registerHandlers(tonApiServer)

	router.GET("/docs", swagger.NewGetSwaggerDocs().Handler)
	router.GET("/swagger.json", swagger.NewGetSwaggerJson().Handler)

	if err := router.Start(":8512"); err != nil {
		log.Fatal(err)
	}
}

func routerGetVersioning(path string, handle echo.HandlerFunc) {
	router.GET(path, handle)
	router.GET(ApiV1+path, handle)
}

func registerHandlers(server *api.TonApiServer) {

	wrapper := tonapi.ServerInterfaceWrapper{
		Handler: server,
	}

	for _, addrRoot := range addressRootAliases {
		routerGetVersioning(addrRoot, wrapper.GetV1Account)
		routerGetVersioning(addrRoot+"/messages", wrapper.GetV1AccountMessages)
		routerGetVersioning(addrRoot+"/transactions", wrapper.GetV1AccountMessages)
		routerGetVersioning(addrRoot+"/qr", wrapper.GetV1AccountQr)
	}

	for _, blockRoot := range blocksRootAliases {
		routerGetVersioning(blockRoot+"/info", wrapper.GetV1BlockInfo)
		routerGetVersioning(blockRoot+"/tlb", wrapper.GetV1BlockTlb)
		routerGetVersioning(blockRoot+"/transactions", wrapper.GetV1BlockTransactions) // TODO: Remove this in favor of /messages. This is deprecated.
		routerGetVersioning(blockRoot+"/feed", wrapper.GetV1BlocksFeed)
	}

	routerGetVersioning("/addr/top-by-message-count", wrapper.GetV1AddrTopByMessageCount)
	routerGetVersioning("/height/blockchain", wrapper.GetV1HeightBlockchain)
	routerGetVersioning("/height/synced", wrapper.GetV1HeightSynced)
	routerGetVersioning("/master/block/shards/actual", wrapper.GetV1MasterBlockShardsActual)
	routerGetVersioning("/master/block/shards/range", wrapper.GetV1MasterBlockShardsRange)
	routerGetVersioning("/message/get", wrapper.GetV1MessageGet)
	routerGetVersioning("/messages/feed", wrapper.GetV1MessagesFeed)
	routerGetVersioning("/search", wrapper.GetV1Search)
	routerGetVersioning("/stats/addresses", wrapper.GetV1StatsAddresses)
	routerGetVersioning("/stats/blocks", wrapper.GetV1StatsBlocks)
	routerGetVersioning("/stats/global", wrapper.GetV1StatsGlobal)
	routerGetVersioning("/stats/messages", wrapper.GetV1StatsMessages)
	routerGetVersioning("/stats/transactions", wrapper.GetV1StatsTransactions)
	routerGetVersioning("/timeseries/blocks-by-workchain", wrapper.GetV1TimeseriesBlocksByWorkchain)
	routerGetVersioning("/timeseries/messages-by-type", wrapper.GetV1TimeseriesMessagesByType)
	routerGetVersioning("/timeseries/messages-ord-count", wrapper.GetV1TimeseriesMessagesOrdCount)
	routerGetVersioning("/timeseries/sent-and-fees", wrapper.GetV1TimeseriesSentAndFees)
	routerGetVersioning("/timeseries/volume-by-grams", wrapper.GetV1TimeseriesVolumeByGrams)
	routerGetVersioning("/top/whales", wrapper.GetV1TopWhales)
	routerGetVersioning("/transaction", wrapper.GetV1Transaction)
	routerGetVersioning("/transactions/feed", wrapper.GetV1TransactionsFeed)
	routerGetVersioning("/workchain/block/master", wrapper.GetV1WorkchainBlockMaster)

}
