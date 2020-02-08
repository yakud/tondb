package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"gitlab.flora.loc/mills/tondb/internal/tlb_pretty"

	"gitlab.flora.loc/mills/tondb/internal/blocks_receiver"
)

var tlbParser = tlb_pretty.NewParser()
var treeSimplifier = tlb_pretty.NewTreeSimplifier()
var astTonConverter = tlb_pretty.NewAstTonConverter()
var mu = sync.Mutex{}
var out *string
var singleFile, consoleMode, silentMode *bool
var prevBlocksCount uint64
var blocksCounter uint64

func handler() func(resp []byte) error {

	return func(blockPretty []byte) error {

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

			fmt.Print(".")
			atomic.AddUint64(&blocksCounter, 1)

			if *silentMode {
				return nil
			}

			if *consoleMode {
				fmt.Println(string(blockPretty))
				return nil
			}

			var f *os.File

			if *singleFile {
				// I'm not sure whether we receive blocks concurrently or not so I will put on lock just in case
				mu.Lock()
				defer mu.Unlock()

				f, err = os.OpenFile(*out + "/blocks.pretty", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			} else {
				f, err = os.OpenFile(fmt.Sprintf(*out + "/blocks%d_%d_%d.pretty", block.Info.WorkchainId, block.Info.SeqNo, block.Info.Shard), os.O_CREATE|os.O_WRONLY, 0644)
			}

			if err != nil {
				log.Fatal(err)
			}

			if _, err = f.Write(blockPretty); err != nil {
				log.Fatal(err)
			}

			if err = f.Sync(); err != nil {
				log.Fatal(err)
			}

			if err = f.Close(); err != nil {
				log.Fatal(err)
			}
		}

		return nil
	}
}

func main() {
	go func() {
		for {
			<-time.After(time.Second * 5)
			currCnt := atomic.LoadUint64(&blocksCounter)
			fmt.Println(fmt.Sprintf("\nTotal blocks handled: %d", currCnt))
			fmt.Println(fmt.Sprintf("Handling about %d blocks/sec.", (currCnt - prevBlocksCount) / 5))
			prevBlocksCount = currCnt
		}
	}()

	out = flag.String("out", "", "Output directory where pretty blocks will be stored.")
	singleFile = flag.Bool("single_file", false, "If true, all blocks will be written to a single file or to different files otherwise.")
	consoleMode = flag.Bool("console_mode", false, "If true, all blocks will be written to STDOUT instead of files.")
	silentMode = flag.Bool("silent_mode", false, "If true, all blocks will be just silently simplified and converted without writing them anywhere.")

	flag.Parse()

	tcpServer := blocks_receiver.NewTcpReceiver("0.0.0.0:8189")

	ctx, _ := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	wg.Add(1)
	if err := tcpServer.Run(ctx, wg, handler()); err != nil {
		log.Fatal(err)
	}
}
