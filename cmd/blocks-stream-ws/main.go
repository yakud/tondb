package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"gitlab.flora.loc/mills/tondb/internal/streaming"

	"gitlab.flora.loc/mills/tondb/internal/tlb_pretty"

	"gitlab.flora.loc/mills/tondb/internal/blocks_receiver"

	"github.com/mailru/easygo/netpoll"
)

var tlbParser = tlb_pretty.NewParser()
var treeSimplifier = tlb_pretty.NewTreeSimplifier()
var astTonConverter = tlb_pretty.NewAstTonConverter()
var blocksChan = make(chan []byte, 100000)
var subscriber = streaming.NewSubscriber()
var streamReceiver = streaming.NewStreamReceiver(subscriber)

func handler() func(resp []byte) error {
	return func(resp []byte) error {
		blocksChan <- resp
		return nil
	}
}
func workerBlocksHandler() error {
	var astPretty *tlb_pretty.AstNode
	var err error

	for blockPretty := range blocksChan {
		astPretty = tlbParser.Parse(blockPretty)
		astPretty, err = treeSimplifier.Simplify(astPretty)
		if err != nil {
			log.Println(err, "block size:", len(blockPretty), string(blockPretty))
			return err
		}

		if t, err := astPretty.Type(); err == nil && t == "account_state" {
			// ignore state
			continue
		}

		block, err := astTonConverter.ConvertToBlock(astPretty)
		if err != nil {
			log.Println(err, "block size:", len(blockPretty), string(blockPretty))
			return err
		}

		if err := streamReceiver.HandleBlock(block); err != nil {
			log.Println(err, "block size:", len(blockPretty), string(blockPretty))
			continue
		}

		fmt.Print(".")
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

	serverAddr := os.Getenv("ADDR")
	if serverAddr == "" {
		serverAddr = "0.0.0.0:7315"
	}

	poller, err := netpoll.New(nil)
	if err != nil {
		log.Fatal(err)
	}

	subscriberCtx, _ := context.WithCancel(context.Background())
	go subscriber.GarbageCollection(subscriberCtx, 5*time.Minute)

	wsServerCtx, wsServerCancel := context.WithCancel(context.Background())
	wsHandler := streaming.NewWSServer(poller, subscriber, wsServerCtx, wsServerCancel)

	wsServer := http.Server{
		Addr:    "0.0.0.0:1818",
		Handler: http.HandlerFunc(wsHandler.Handler),
	}

	go func() {
		if err := wsServer.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	workers := 1
	for i := 0; i < workers; i++ {
		go func() {
			if err := workerBlocksHandler(); err != nil {
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
