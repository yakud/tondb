package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"gitlab.flora.loc/mills/tondb/internal/api/swagger"

	"gitlab.flora.loc/mills/tondb/internal/ton/view/index"

	"gitlab.flora.loc/mills/tondb/internal/ton/search"

	"github.com/jmoiron/sqlx"

	"gitlab.flora.loc/mills/tondb/internal/api"
	apifeed "gitlab.flora.loc/mills/tondb/internal/api/feed"
	"gitlab.flora.loc/mills/tondb/internal/api/middleware"
	"gitlab.flora.loc/mills/tondb/internal/api/ratelimit"
	"gitlab.flora.loc/mills/tondb/internal/api/site"
	statsApi "gitlab.flora.loc/mills/tondb/internal/api/stats"
	"gitlab.flora.loc/mills/tondb/internal/api/timeseries"
	"gitlab.flora.loc/mills/tondb/internal/blocks_fetcher"
	"gitlab.flora.loc/mills/tondb/internal/ch"
	"gitlab.flora.loc/mills/tondb/internal/ton/query"
	"gitlab.flora.loc/mills/tondb/internal/ton/query/cache"
	statsQ "gitlab.flora.loc/mills/tondb/internal/ton/query/stats"
	timeseriesQ "gitlab.flora.loc/mills/tondb/internal/ton/query/timeseries"
	"gitlab.flora.loc/mills/tondb/internal/ton/storage"
	"gitlab.flora.loc/mills/tondb/internal/ton/view/feed"
	"gitlab.flora.loc/mills/tondb/internal/ton/view/state"
	"gitlab.flora.loc/mills/tondb/internal/ton/view/stats"
	timeseriesV "gitlab.flora.loc/mills/tondb/internal/ton/view/timeseries"

	"github.com/go-redis/redis"
	"github.com/julienschmidt/httprouter"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/cors"
)

const ApiV1 = "/v1"

var (
	blocksRootAliases  = [...]string{"/blocks", "/block", "/b"}
	addressRootAliases = [...]string{"/address", "/account", "/a"}
	router             = httprouter.New()
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

	accountTransactions := feed.NewAccountTransactions(chConnect)
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

	router.GET("/docs", swagger.NewGetSwaggerDocs().Handler)
	router.GET("/swagger.json", swagger.NewGetSwaggerJson().Handler)

	routerGetVersioning("/height/synced", rateLimitMiddleware(api.NewGetSyncedHeight(syncedHeightQuery).Handler))
	routerGetVersioning("/height/blockchain", rateLimitMiddleware(api.NewGetBlockchainHeight(blockchainHeightQuery).Handler))
	routerGetVersioning("/master/block/shards/range", rateLimitMiddleware(api.NewMasterBlockShardsRange(shardsDescrStorage).Handler))
	routerGetVersioning("/master/block/shards/actual", rateLimitMiddleware(api.NewMasterchainBlockShardsActual(shardsDescrStorage).Handler))
	routerGetVersioning("/workchain/block/master", rateLimitMiddleware(api.NewGetWorkchainBlockMaster(shardsDescrStorage).Handler))
	routerGetVersioning("/transaction", rateLimitMiddleware(api.NewGetTransactions(searchTransactionsQuery).Handler))
	routerGetVersioning("/block/tlb", rateLimitMiddleware(api.NewGetBlockTlb(blocksFetcher).Handler))
	routerGetVersioning("/search", rateLimitMiddleware(api.NewSearch(searcher).Handler))

	// Block routes
	getBlockInfo := api.NewGetBlockInfo(getBlockInfoQuery, shardsDescrStorage)
	getBlockTransactions := api.NewGetBlockTransactions(searchTransactionsQuery, shardsDescrStorage)
	getBlocksFeed := apifeed.NewGetBlocksFeed(blocksFeed)
	for _, blockRoot := range blocksRootAliases {
		routerGetVersioning(blockRoot+"/info", rateLimitMiddleware(getBlockInfo.Handler))
		routerGetVersioning(blockRoot+"/transactions", rateLimitMiddleware(getBlockTransactions.Handler))
		routerGetVersioning(blockRoot+"/feed", rateLimitMiddleware(getBlocksFeed.Handler))
	}

	// Address (account) routes
	getAccountHandler := api.NewGetAccount(accountState)
	getAccountTransactions := api.NewGetAccountTransactions(accountTransactions)
	getAccountQR := api.NewGetAccountQR()
	for _, addrRoot := range addressRootAliases {
		routerGetVersioning(addrRoot, rateLimitMiddleware(getAccountHandler.Handler))
		routerGetVersioning(addrRoot+"/transactions", rateLimitMiddleware(getAccountTransactions.Handler))
		routerGetVersioning(addrRoot+"/qr", rateLimitMiddleware(getAccountQR.Handler))
	}

	// Messages feed
	messagesFeedGlobal := feed.NewMessagesFeed(chConnectSqlx)
	if err := messagesFeedGlobal.CreateTable(); err != nil {
		log.Fatal(err)
	}

	getMessageQuery := query.NewGetMessage(chConnect)

	routerGetVersioning("/messages/feed", rateLimitMiddleware(apifeed.NewGetMessagesFeed(messagesFeedGlobal).Handler))
	routerGetVersioning("/message/get", rateLimitMiddleware(api.NewGetMessage(getMessageQuery).Handler))

	// Transactions feed
	transactionsFeed := feed.NewTransactionsFeed(chConnectSqlx)
	if err := transactionsFeed.CreateTable(); err != nil {
		log.Fatal(err)
	}

	routerGetVersioning("/transactions/feed", rateLimitMiddleware(apifeed.NewGetTransactionsFeed(transactionsFeed).Handler))

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

	ctxBgCache, _ := context.WithCancel(context.Background())
	metricsCache := cache.NewBackground()
	blocksCache := cache.NewBackground()
	whalesCache := cache.NewBackground()

	messagesMetrics := statsQ.NewMessagesMetrics(chConnect, metricsCache)
	if err := messagesMetrics.CreateTable(); err != nil {
		log.Fatal(err)
	}
	if err := messagesMetrics.UpdateQuery(); err != nil {
		log.Fatal(err)
	}

	globalMetrics := statsQ.NewGlobalMetrics(chConnect, metricsCache)
	if err := globalMetrics.UpdateQuery(); err != nil {
		log.Fatal(err)
	}

	blocksMetrics := statsQ.NewBlocksMetrics(chConnect, metricsCache)
	if err := blocksMetrics.UpdateQuery(); err != nil {
		log.Fatal(err)
	}

	addressesMetrics := statsQ.NewAddressesMetrics(chConnect, metricsCache)
	if err := addressesMetrics.UpdateQuery(); err != nil {
		log.Fatal(err)
	}

	topWhales := statsQ.NewGetTopWhales(chConnect, whalesCache, globalMetrics)
	if err := topWhales.UpdateQuery(); err != nil {
		log.Fatal(err)
	}

	sentAndFees := timeseriesQ.NewSentAndFees(chConnect, whalesCache)
	if err := sentAndFees.UpdateQuery(); err != nil {
		log.Fatal(err)
	}

	metricsCache.AddQuery(globalMetrics)
	metricsCache.AddQuery(addressesMetrics)
	metricsCache.AddQuery(messagesMetrics)
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

	routerGetVersioning("/timeseries/blocks-by-workchain", rateLimitMiddleware(timeseries.NewBlocksByWorkchain(qBlocksByWorkchain).Handler))
	routerGetVersioning("/timeseries/messages-by-type", rateLimitMiddleware(timeseries.NewMessagesByType(tsMessagesByType).Handler))
	routerGetVersioning("/timeseries/volume-by-grams", rateLimitMiddleware(timeseries.NewVolumeByGrams(tsVolumeByGrams).Handler))
	routerGetVersioning("/timeseries/messages-ord-count", rateLimitMiddleware(timeseries.NewMessagesOrdCount(tsMessagesOrdCount).Handler))
	routerGetVersioning("/timeseries/sent-and-fees", rateLimitMiddleware(timeseries.NewSentAndFees(sentAndFees).Handler))
	routerGetVersioning("/addr/top-by-message-count", rateLimitMiddleware(site.NewGetAddrTopByMessageCount(addrMessagesCount).Handler))
	routerGetVersioning("/top/whales", rateLimitMiddleware(site.NewGetTopWhales(topWhales).Handler))
	routerGetVersioning("/stats/global", rateLimitMiddleware(statsApi.NewGlobalMetrics(globalMetrics).Handler))
	routerGetVersioning("/stats/blocks", rateLimitMiddleware(statsApi.NewBlocksMetrics(blocksMetrics).Handler))
	routerGetVersioning("/stats/addresses", rateLimitMiddleware(statsApi.NewAddressesMetrics(addressesMetrics).Handler))
	routerGetVersioning("/stats/messages", rateLimitMiddleware(statsApi.NewMessagesMetrics(messagesMetrics).Handler))

	handler := cors.AllowAll().Handler(router)
	srv := &http.Server{
		Addr:    config.Addr,
		Handler: handler,
	}
	log.Println("Start listening:", config.Addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func routerGetVersioning(path string, handle httprouter.Handle) {
	router.GET(path, handle)
	router.GET(ApiV1+path, handle)
}
