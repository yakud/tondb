package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/yakud/ton-blocks-stream-receiver/internal/ton/writer"

	"github.com/yakud/ton-blocks-stream-receiver/internal/ton/storage"

	"github.com/yakud/ton-blocks-stream-receiver/internal/ch"
	"github.com/yakud/ton-blocks-stream-receiver/internal/tlb_pretty"

	"github.com/yakud/ton-blocks-stream-receiver/internal/blocks_receiver"
)

var tlbParser = tlb_pretty.NewParser()
var treeSimplifier = tlb_pretty.NewTreeSimplifier()
var astTonConverter = tlb_pretty.NewAstTonConverter()
var blocksChan = make(chan []byte, 100000)

func handler() func(resp []byte) error {
	return func(resp []byte) error {
		blocksChan <- resp
		return nil
	}
}
func worker(buffer *writer.BulkBuffer) error {
	var node *tlb_pretty.AstNode
	var err error

	for resp := range blocksChan {
		node = tlbParser.Parse(resp)
		node, err = treeSimplifier.Simplify(node)
		if err != nil {
			log.Fatal(err, "block size:", len(resp), string(resp))
			continue
		}

		block, err := astTonConverter.ConvertToBlock(node)
		if err != nil {
			log.Fatal(err, "block size:", len(resp), string(resp))
			continue
		}

		if err := buffer.Add(block); err != nil {
			log.Fatal(err, "block size:", len(resp), string(resp))
			continue
		}

		//fmt.Println(block.Info.ShardWorkchainId, block.Info.SeqNo)
		fmt.Print(".")

		if block.Info.WorkchainId == -1 {
			fmt.Print("(-1;", block.Info.SeqNo, ")")
		}
	}

	return nil
}

func main() {
	go func() {
		for {
			<-time.After(time.Second * 5)
			fmt.Print("-")
		}
	}()

	promAddr := os.Getenv("PROM_ADDR")
	if promAddr == "" {
		promAddr = "0.0.0.0:8080"
	}

	serverAddr := os.Getenv("ADDR")
	if serverAddr == "" {
		serverAddr = "0.0.0.0:7315"
	}

	chAddr := os.Getenv("CH_ADDR")
	if chAddr == "" {
		chAddr = "http://default:V9AQZJFNX4ygj2vP@192.168.100.3:8123/ton?max_query_size=3145728000"
	}
	chConnect, err := ch.Connect(&chAddr)
	if err != nil {
		log.Fatal(err)
	}

	blocksStorage := storage.NewBlocks(chConnect)
	//blocksStorage.DropTable()
	if err := blocksStorage.CreateTable(); err != nil {
		log.Fatal("blocksStorage CreateTable", err)
	}

	transactionsStorage := storage.NewTransactions(chConnect)
	//transactionsStorage.DropTable()
	if err := transactionsStorage.CreateTable(); err != nil {
		log.Fatal("transactionsStorage CreateTable", err)
	}

	shardsDescrStorage := storage.NewShardsDescr(chConnect)
	//shardsDescrStorage.DropTable()
	if err := shardsDescrStorage.CreateTable(); err != nil {
		log.Fatal("shardsDescrStorage CreateTable", err)
	}

	writeBulksChan := make(chan *writer.Bulk, 10)
	buffer := writer.NewBulkBuffer(writeBulksChan)

	ctxBuffer, _ := context.WithCancel(context.Background())
	wgBuffer := &sync.WaitGroup{}
	wgBuffer.Add(1)
	go buffer.Timeout(ctxBuffer, wgBuffer)

	// Bulk writer
	bulkWriter := writer.NewBulkWriter(
		blocksStorage,
		transactionsStorage,
		shardsDescrStorage,
		writeBulksChan,
	)

	ctxWriter, _ := context.WithCancel(context.Background())
	wgWriter := &sync.WaitGroup{}
	wgWriter.Add(1)
	go func() {
		if err := bulkWriter.Run(ctxWriter, wgWriter, nil); err != nil {
			log.Fatal(err)
		}
	}()

	promServer := http.Server{
		Addr: promAddr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			switch req.URL.Path {
			case "/metrics":
				promhttp.Handler().ServeHTTP(w, req)
			}
		}),
	}

	// prom metrics server
	go func() {
		if err := promServer.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	workers := 4
	for i := 0; i < workers; i++ {
		go func() {
			if err := worker(buffer); err != nil {
				log.Fatal(err)
			}
		}()
	}

	tcpServer := blocks_receiver.NewTcpReceiver(serverAddr)
	ctx, _ := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	wg.Add(1)
	if err := tcpServer.Run(ctx, wg, handler()); err != nil {
		log.Fatal(err)
	}
}
