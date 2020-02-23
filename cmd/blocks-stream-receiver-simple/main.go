package main

import (
	"context"
	"fmt"
	"log"
	"sync"

	"gitlab.flora.loc/mills/tondb/internal/tlb_pretty"

	"gitlab.flora.loc/mills/tondb/internal/blocks_receiver"
)

var tlbParser = tlb_pretty.NewParser()
var treeSimplifier = tlb_pretty.NewTreeSimplifier()
var astTonConverter = tlb_pretty.NewAstTonConverter()

func handler() func(resp []byte) error {

	return func(blockPretty []byte) error {
		//fmt.Println(string(blockPretty))

		astPretty := tlbParser.Parse(blockPretty)
		astPretty, err := treeSimplifier.Simplify(astPretty)
		if err != nil {
			log.Fatal(err, "block size:", len(blockPretty), string(blockPretty))
		}

		if t, err := astPretty.Type(); err == nil && t == "account_state" {
			//st, err := astTonConverter.ConvertToState(astPretty)
			//if err != nil {
			//	log.Fatal(err, "state:", string(blockPretty))
			//}

		} else {
			block, err := astTonConverter.ConvertToBlock(astPretty)
			if err != nil {
				log.Fatal(err, "block size:", len(blockPretty), string(blockPretty))
			}

			//fmt.Println(block.Info.ShardWorkchainId, block.Info.SeqNo)
			fmt.Print(".")

			if block.Info.WorkchainId == -1 {
				fmt.Print("(-1;", block.Info.SeqNo, ")")
			}
		}

		return nil
	}
}

func main() {
	//go func() {
	//	for {
	//		<-time.After(time.Second * 5)
	//		fmt.Print("-")
	//	}
	//}()

	tcpServer := blocks_receiver.NewTcpReceiver("0.0.0.0:8189")

	ctx, _ := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	wg.Add(1)
	if err := tcpServer.Run(ctx, wg, handler()); err != nil {
		log.Fatal(err)
	}
}
