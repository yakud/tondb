package main

import (
	"log"
	"net/http"
	"os"

	"gitlab.flora.loc/mills/tondb/internal/api"
	"gitlab.flora.loc/mills/tondb/internal/ch"
	"gitlab.flora.loc/mills/tondb/internal/ton/query"
	"gitlab.flora.loc/mills/tondb/internal/ton/storage"

	"github.com/julienschmidt/httprouter"
)

func main() {
	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = "0.0.0.0:8512"
	}

	chAddr := os.Getenv("CH_ADDR")
	if chAddr == "" {
		chAddr = "http://default:V9AQZJFNX4ygj2vP@192.168.100.3:8123/ton?max_query_size=3145728000"
	}
	chConnect, err := ch.Connect(&chAddr)
	if err != nil {
		log.Fatal(err)
	}

	//blocksStorage := storage.NewBlocks(chConnect)
	//transactionsStorage := storage.NewTransactions(chConnect)
	shardsDescrStorage := storage.NewShardsDescr(chConnect)

	syncedHeightQuery := query.NewGetSyncedHeight(chConnect)
	blockchainHeightQuery := query.NewGetBlockchainHeight(chConnect)
	searchTransactionsQuery := query.NewSearchTransactions(chConnect)
	getBlockInfoQuery := query.NewGetBlockInfo(chConnect)

	router := httprouter.New()

	router.GET("/height/synced", api.BasicAuth(api.NewGetSyncedHeight(syncedHeightQuery).Handler))
	router.GET("/height/blockchain", api.BasicAuth(api.NewGetBlockchainHeight(blockchainHeightQuery).Handler))
	router.GET("/master/block/shards/range", api.BasicAuth(api.NewMasterBlockShardsRange(shardsDescrStorage).Handler))
	router.GET("/workchain/block/master", api.BasicAuth(api.NewGetWorkchainBlockMaster(shardsDescrStorage).Handler))
	router.GET("/block/info", api.BasicAuth(api.NewGetBlockInfo(getBlockInfoQuery, shardsDescrStorage).Handler))
	router.GET("/block/transactions", api.BasicAuth(api.NewGetBlockTransactions(searchTransactionsQuery, shardsDescrStorage).Handler))

	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}
	log.Println("Start listening", addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}