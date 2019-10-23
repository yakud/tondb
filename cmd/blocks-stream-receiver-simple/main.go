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

	return func(resp []byte) error {
		fmt.Print(".")

		node := tlbParser.Parse(resp)
		simplifiedNode, err := treeSimplifier.Simplify(node)
		if err != nil {
			fmt.Println(string(resp))
			log.Fatal(err)
		}

		block, err := astTonConverter.ConvertToBlock(simplifiedNode)
		if err != nil {
			fmt.Println(string(resp))
			log.Fatal(err)
		}

		fmt.Println(block.Info.WorkchainId, block.Info.Shard, block.Info.SeqNo)
		//fmt.Print(".")

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

	tcpServer := blocks_receiver.NewTcpReceiver()

	ctx, _ := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	wg.Add(1)
	if err := tcpServer.Run(ctx, wg, handler()); err != nil {
		log.Fatal(err)
	}
}