package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/yakud/ton-blocks-stream-receiver/internal/ton/writer"

	"github.com/yakud/ton-blocks-stream-receiver/internal/ton/storage"

	"github.com/yakud/ton-blocks-stream-receiver/internal/ch"
	"github.com/yakud/ton-blocks-stream-receiver/internal/tlb_pretty"

	"github.com/yakud/ton-blocks-stream-receiver/internal/blocks_receiver"
)

var tlbParser = tlb_pretty.NewParser()
var treeSimplifier = tlb_pretty.NewTreeSimplifier()
var astTonConverter = tlb_pretty.NewAstTonConverter()

func handler(buffer *writer.BulkBuffer) func(resp []byte) error {

	return func(resp []byte) error {
		node := tlbParser.Parse(resp)
		simplifiedNode, err := treeSimplifier.Simplify(node)
		if err != nil {
			log.Fatal(err)
		}

		block, err := astTonConverter.ConvertToBlock(simplifiedNode)
		if err != nil {
			log.Fatal(err)
		}

		if err := buffer.Add(block); err != nil {
			log.Fatal(err)
		}

		//fmt.Println(block.Info.ShardWorkchainId, block.Info.SeqNo)
		fmt.Print(".")

		if block.Info.ShardWorkchainId == -1 {
			fmt.Print("(-1;", block.Info.SeqNo, ")")
		}

		return nil
	}
}

func main() {
	go func() {
		for {
			<-time.After(time.Second * 5)
			fmt.Print("-")
		}
	}()

	chAddr := os.Getenv("CH_ADDR")
	if chAddr == "" {
		chAddr = "http://default:V9AQZJFNX4ygj2vP@192.168.100.3:8123/ton?max_query_size=3145728000"
	}
	chConnect, err := ch.Connect(&chAddr)
	if err != nil {
		log.Fatal(err)
	}

	blocksStorage := storage.NewBlocks(chConnect)
	blocksStorage.DropTable()
	if err := blocksStorage.CreateTable(); err != nil {
		log.Fatal("blocksStorage CreateTable", err)
	}

	transactionsStorage := storage.NewTransactions(chConnect)
	transactionsStorage.DropTable()
	if err := transactionsStorage.CreateTable(); err != nil {
		log.Fatal("transactionsStorage CreateTable", err)
	}

	shardsDescrStorage := storage.NewShardsDescr(chConnect)
	shardsDescrStorage.DropTable()
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

	tcpServer := blocks_receiver.NewTcpReceiver()

	ctx, _ := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	wg.Add(1)
	if err := tcpServer.Run(ctx, wg, handler(buffer)); err != nil {
		log.Fatal(err)
	}
}
