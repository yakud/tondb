package main

import (
	"log"
	"net/http"
	"os"

	"github.com/yakud/ton-blocks-stream-receiver/internal/ton/query"

	"github.com/yakud/ton-blocks-stream-receiver/internal/ch"
	"github.com/yakud/ton-blocks-stream-receiver/internal/ton/storage"

	"github.com/yakud/ton-blocks-stream-receiver/internal/api"

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
	transactionsStorage := storage.NewTransactions(chConnect)
	shardsDescrStorage := storage.NewShardsDescr(chConnect)

	syncedHeightQuery := query.NewGetSyncedHeight(chConnect)
	blockchainHeightQuery := query.NewGetBlockchainHeight(chConnect)

	router := httprouter.New()
	router.GET("/height/synced", api.NewGetSyncedHeight(syncedHeightQuery).Handler)
	router.GET("/height/blockchain", api.NewGetBlockchainHeight(blockchainHeightQuery).Handler)

	router.GET("/masterchain/block/:seqNo/shards", api.NewMasterchainBlockShards(shardsDescrStorage).Handler)
	router.GET("/workchain/block/masterchain", api.NewMasterchainByShard(shardsDescrStorage).Handler)
	router.GET("/search/transactions", api.NewSearchTransactions(shardsDescrStorage, transactionsStorage).Handler)

	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	log.Println("Start listening", addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
