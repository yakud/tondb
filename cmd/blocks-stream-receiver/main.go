package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"gitlab.flora.loc/mills/tondb/internal/ton/view/state"

	"gitlab.flora.loc/mills/tondb/internal/ton/view/index"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gitlab.flora.loc/mills/tondb/internal/ton/writer"

	"gitlab.flora.loc/mills/tondb/internal/ton/storage"

	"gitlab.flora.loc/mills/tondb/internal/ch"
	"gitlab.flora.loc/mills/tondb/internal/tlb_pretty"

	"gitlab.flora.loc/mills/tondb/internal/blocks_receiver"
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
func workerBlocksHandler(buffer *writer.BulkBuffer) error {
	var astPretty *tlb_pretty.AstNode
	var err error

	for blockPretty := range blocksChan {
		astPretty = tlbParser.Parse(blockPretty)
		astPretty, err = treeSimplifier.Simplify(astPretty)
		if err != nil {
			log.Fatal(err, "block size:", len(blockPretty), string(blockPretty))
		}

		if t, err := astPretty.Type(); err == nil && t == "account_state" {
			st, err := astTonConverter.ConvertToState(astPretty)
			if err != nil {
				log.Fatal(err, "state:", string(blockPretty))
			}

			if err := buffer.AddState(st); err != nil {
				log.Fatal(err, "state:", string(blockPretty))
			}

		} else {
			block, err := astTonConverter.ConvertToBlock(astPretty)
			if err != nil {
				log.Fatal(err, "block size:", len(blockPretty), string(blockPretty))
			}

			if err := buffer.AddBlock(block); err != nil {
				log.Fatal(err, "block size:", len(blockPretty), string(blockPretty))
			}

			//fmt.Println(block.Info.ShardWorkchainId, block.Info.SeqNo)
			fmt.Print(".")

			if block.Info.WorkchainId == -1 {
				fmt.Print("(-1;", block.Info.SeqNo, ")")
			}
		}
	}

	return nil
}

func main() {
	log.Println("started v0.0.1")
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
		chAddr = "http://127.0.0.1:8123/default?max_query_size=3145728000"
		//chAddr = "http://default:V9AQZJFNX4ygj2vP@192.168.100.3:8123/ton?max_query_size=3145728000"
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

	accountState := storage.NewAccountState(chConnect)
	//accountState.DropTable()
	if err := accountState.CreateTable(); err != nil {
		log.Fatal("accountState CreateTable", err)
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

	indexTransactionBlock := index.NewIndexTransactionBlock(chConnect)
	//shardsDescrStorage.DropTable()
	if err := indexTransactionBlock.CreateTable(); err != nil {
		log.Fatal("indexTransactionBlock CreateTable", err)
	}

	indexNextBlock := index.NewIndexNextBlock(chConnect)
	//shardsDescrStorage.DropTable()
	if err := indexNextBlock.CreateTable(); err != nil {
		log.Fatal("indexNextBlock CreateTable", err)
	}

	indexReverseBlockSeqNo := index.NewIndexReverseBlockSeqNo(chConnect)
	//shardsDescrStorage.DropTable()
	if err := indexReverseBlockSeqNo.CreateTable(); err != nil {
		log.Fatal("indexReverseBlockSeqNo CreateTable", err)
	}

	indexHash := index.NewIndexHash(chConnect)
	//shardsDescrStorage.DropTable()
	if err := indexHash.CreateTable(); err != nil {
		log.Fatal("indexHash CreateTable", err)
	}

	stateAccountState := state.NewAccountState(chConnect)
	//stateAccountState.DropTable()
	if err := stateAccountState.CreateTable(); err != nil {
		log.Fatal("stateAccountState CreateTable", err)
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
		accountState,
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
			if err := workerBlocksHandler(buffer); err != nil {
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
