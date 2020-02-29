package main

import (
	"context"
	"log"
	"net/http"
	"time"

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

var (
	blocksRootAliases  = [...]string{"/blocks", "/block", "/b"}
	addressRootAliases = [...]string{"/address", "/account", "/a"}
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

	blocksFeed := feed.NewBlocksFeed(chConnectSqlx)
	if err := blocksFeed.CreateTable(); err != nil {
		log.Fatal(err)
	}

	syncedHeightQuery := query.NewGetSyncedHeight(chConnect)
	blockchainHeightQuery := query.NewGetBlockchainHeight(chConnect)
	searchTransactionsQuery := query.NewSearchTransactions(chConnect)
	getBlockInfoQuery := query.NewGetBlockInfo(chConnect)

	if err := ratelimit.RateLimitLua.Load(redisClient).Err(); err != nil {
		log.Fatal("error load redis lua script:", err)
	}
	rateLimiter := ratelimit.NewRateLimiter(redisClient)
	rateLimitMiddleware := middleware.RateLimit(rateLimiter)

	router.GET("/height/synced", rateLimitMiddleware(api.NewGetSyncedHeight(syncedHeightQuery).Handler))
	router.GET("/height/blockchain", rateLimitMiddleware(api.NewGetBlockchainHeight(blockchainHeightQuery).Handler))
	router.GET("/master/block/shards/range", rateLimitMiddleware(api.NewMasterBlockShardsRange(shardsDescrStorage).Handler))
	router.GET("/workchain/block/master", rateLimitMiddleware(api.NewGetWorkchainBlockMaster(shardsDescrStorage).Handler))
	router.GET("/transaction", rateLimitMiddleware(api.NewGetTransactions(searchTransactionsQuery).Handler))
	router.GET("/block/tlb", rateLimitMiddleware(api.NewGetBlockTlb(blocksFetcher).Handler))

	// Block routes
	getBlockInfo := api.NewGetBlockInfo(getBlockInfoQuery, shardsDescrStorage)
	getBlockTransactions := api.NewGetBlockTransactions(searchTransactionsQuery, shardsDescrStorage)
	getBlocksFeed := apifeed.NewGetBlocksFeed(blocksFeed)
	for _, blockRoot := range blocksRootAliases {
		router.GET(blockRoot+"/info", rateLimitMiddleware(getBlockInfo.Handler))
		router.GET(blockRoot+"/transactions", rateLimitMiddleware(getBlockTransactions.Handler))
		router.GET(blockRoot+"/feed", rateLimitMiddleware(getBlocksFeed.Handler))
	}

	// Address (account) routes
	getAccountHandler := api.NewGetAccount(accountState)
	getAccountTransactions := api.NewGetAccountTransactions(accountTransactions)
	getAccountQR := api.NewGetAccountQR()
	for _, addrRoot := range addressRootAliases {
		router.GET(addrRoot, rateLimitMiddleware(getAccountHandler.Handler))
		router.GET(addrRoot+"/transactions", rateLimitMiddleware(getAccountTransactions.Handler))
		router.GET(addrRoot+"/qr", rateLimitMiddleware(getAccountQR.Handler))
	}

	// Messages feed
	messagesFeedGlobal := feed.NewMessagesFeed(chConnectSqlx)
	if err := messagesFeedGlobal.CreateTable(); err != nil {
		log.Fatal(err)
	}

	router.GET("/messages/feed", rateLimitMiddleware(apifeed.NewGetMessagesFeed(messagesFeedGlobal).Handler))

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

	qGetTopWhales := statsQ.NewGetTopWhales(chConnect)

	ctxBgCache, _ := context.WithCancel(context.Background())
	bgCache := cache.NewBackground()
	blocksCache := cache.NewBackground()

	globalMetrics := statsQ.NewGlobalMetrics(chConnect, bgCache)
	if err := globalMetrics.UpdateQuery(); err != nil {
		log.Fatal(err)
	}

	blocksMetrics := statsQ.NewBlocksMetrics(chConnect, bgCache)
	if err := blocksMetrics.UpdateQuery(); err != nil {
		log.Fatal(err)
	}

	addressesMetrics := statsQ.NewAddressesMetrics(chConnect, bgCache)
	if err := addressesMetrics.UpdateQuery(); err != nil {
		log.Fatal(err)
	}

	messagesMetrics := statsQ.NewMessagesMetrics(chConnect, bgCache)
	if err := messagesMetrics.CreateTable(); err != nil {
		log.Fatal(err)
	}
	if err := messagesMetrics.UpdateQuery(); err != nil {
		log.Fatal(err)
	}

	bgCache.AddQuery(globalMetrics)
	bgCache.AddQuery(addressesMetrics)
	bgCache.AddQuery(messagesMetrics)
	blocksCache.AddQuery(blocksMetrics)

	go func() {
		bgCache.RunTicker(ctxBgCache, 5 * time.Second)
	}()

	go func() {
		blocksCache.RunTicker(ctxBgCache, 1 * time.Second)
	}()

	router.GET("/timeseries/blocks-by-workchain", rateLimitMiddleware(timeseries.NewBlocksByWorkchain(qBlocksByWorkchain).Handler))
	router.GET("/timeseries/messages-by-type", rateLimitMiddleware(timeseries.NewMessagesByType(tsMessagesByType).Handler))
	router.GET("/timeseries/volume-by-grams", rateLimitMiddleware(timeseries.NewVolumeByGrams(tsVolumeByGrams).Handler))
	router.GET("/timeseries/messages-ord-count", rateLimitMiddleware(timeseries.NewMessagesOrdCount(tsMessagesOrdCount).Handler))
	router.GET("/addr/top-by-message-count", rateLimitMiddleware(site.NewGetAddrTopByMessageCount(addrMessagesCount).Handler))
	router.GET("/top/whales", rateLimitMiddleware(site.NewGetTopWhales(qGetTopWhales).Handler))
	router.GET("/stats/global", rateLimitMiddleware(statsApi.NewGlobalMetrics(globalMetrics).Handler))
	router.GET("/stats/blocks", rateLimitMiddleware(statsApi.NewBlocksMetrics(blocksMetrics).Handler))
	router.GET("/stats/addresses", rateLimitMiddleware(statsApi.NewAddressesMetrics(addressesMetrics).Handler))
	router.GET("/stats/messages", rateLimitMiddleware(statsApi.NewMessagesMetrics(messagesMetrics).Handler))

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
