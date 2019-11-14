package main

import (
	"log"
	"net/http"
	"os"

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
		chAddr = "http://default:V9AQZJFNX4ygj2vP@192.168.100.3:8123/ton?max_query_size=3145728000"
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

	router := httprouter.New()
	router.GET("/timeseries/blocks-by-workchain", timeseries.NewBlocksByWorkchain(qBlocksByWorkchain).Handler)

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
