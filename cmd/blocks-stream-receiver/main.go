package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"sync"

	"github.com/yakud/ton-blocks-stream-receiver/internal/tlb_pretty"

	"github.com/yakud/ton-blocks-stream-receiver/internal/blocks_receiver"
)

var tlbParser = tlb_pretty.NewParser()
var treeSimplifier = tlb_pretty.NewTreeSimplifier()
var astTonConverter = tlb_pretty.NewAstTonConverter()

func handler(resp []byte) error {
	//since := time.Now()
	node := tlbParser.Parse(resp)
	simplifiedNode, err := treeSimplifier.Simplify(node)
	if err != nil {
		return err
	}

	_, errOrig := astTonConverter.ConvertToBlock(simplifiedNode)
	if errOrig != nil {
		dd, err := simplifiedNode.ToJSON()
		if err != nil {
			log.Fatal(err)
		}

		if err := ioutil.WriteFile("/Users/user/go/src/github.com/yakud/ton-blocks-stream-receiver/a.json", dd, 0600); err != nil {
			log.Fatal(err)
		}

		log.Fatal(errOrig)
	}

	//fmt.Println(block.Info.ShardWorkchainId, block.Info.SeqNo)
	fmt.Print(".")

	return nil
}

func main() {
	tcpServer := blocks_receiver.NewTcpReceiver()

	ctx, _ := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	wg.Add(1)
	if err := tcpServer.Run(ctx, wg, handler); err != nil {
		log.Fatal(err)
	}
}
