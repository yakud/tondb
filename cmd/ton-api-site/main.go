package main

import (
	"log"
	"net/http"
	"os"

	"gitlab.flora.loc/mills/tondb/internal/ton/view/stats"

	"gitlab.flora.loc/mills/tondb/internal/api/site"

	"gitlab.flora.loc/mills/tondb/internal/ton/view/feed"

	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"
	"gitlab.flora.loc/mills/tondb/internal/api/timeseries"
	"gitlab.flora.loc/mills/tondb/internal/ch"
	timeseriesQ "gitlab.flora.loc/mills/tondb/internal/ton/query/timeseries"
	timeseriesV "gitlab.flora.loc/mills/tondb/internal/ton/view/timeseries"
)

func main() {
	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = "0.0.0.0:8513"
	}

	chAddr := os.Getenv("CH_ADDR")
	if chAddr == "" {
		chAddr = "http://default:V9AQZJFNX4ygj2vP@192.168.100.3:8123/ton2?max_query_size=3145728000"
	}
	chConnect, err := ch.Connect(&chAddr)
	if err != nil {
		log.Fatal(err)
	}

	vBlocksByWorkchain := timeseriesV.NewBlocksByWorkchain(chConnect)
	if err := vBlocksByWorkchain.CreateTable(); err != nil {
		log.Fatal(err)
	}
	qBlocksByWorkchain := timeseriesQ.NewGetBlocksByWorkchain(chConnect)

	tsMessagesByType := timeseriesV.NewMessagesByType(chConnect)
	if err := tsMessagesByType.CreateTable(); err != nil {
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

	router := httprouter.New()
	router.GET("/timeseries/blocks-by-workchain", timeseries.NewBlocksByWorkchain(qBlocksByWorkchain).Handler)
	router.GET("/timeseries/messages-by-type", timeseries.NewMessagesByType(tsMessagesByType).Handler)
	router.GET("/messages/latest", site.NewGetLatestMessages(messagesFeedGlobal).Handler)
	router.GET("/addr/top-by-message-count", site.NewGetAddrTopByMessageCount(addrMessagesCount).Handler)

	handler := cors.Default().Handler(router)
	srv := &http.Server{
		Addr:    addr,
		Handler: handler,
	}
	log.Println("Start listening", addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
