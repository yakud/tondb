package main

import (
	"context"
	"encoding/json"
	"fmt"
	"gitlab.flora.loc/mills/tondb/internal/streaming"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"gitlab.flora.loc/mills/tondb/internal/tlb_pretty"

	"gitlab.flora.loc/mills/tondb/internal/blocks_receiver"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"

	"github.com/google/uuid"
	"github.com/panjf2000/ants/v2"
)

var tlbParser = tlb_pretty.NewParser()
var treeSimplifier = tlb_pretty.NewTreeSimplifier()
var astTonConverter = tlb_pretty.NewAstTonConverter()
var blocksChan = make(chan []byte, 100000)
var subManager *streaming.SubManager

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
			log.Fatal(err, "block size:", len(blockPretty), string(blockPretty))
		}

		if t, err := astPretty.Type(); err == nil && t == "account_state" {
			// ignore state

		} else {
			block, err := astTonConverter.ConvertToBlock(astPretty)
			if err != nil {
				log.Fatal(err, "block size:", len(blockPretty), string(blockPretty))
			}

			if err := subManager.HandleBlock(block); err != nil {
				log.Fatal(err, "block size:", len(blockPretty), string(blockPretty))
			}

			//fmt.Println(block.Info.ShardWorkchainId, block.Info.SeqNo)
			fmt.Print(".")

			//if block.Info.WorkchainId == -1 {
			//	fmt.Print("(-1;", block.Info.SeqNo, ")")
			//}
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

	serverAddr := os.Getenv("ADDR")
	if serverAddr == "" {
		serverAddr = "0.0.0.0:7315"
	}

	pool, err := ants.NewPool(256)
	if err != nil {
		log.Fatal(err)
	}

	defer pool.Release()

	subManager = streaming.NewSubManager(pool)

	wsServer := http.Server{
		Addr: "0.0.0.0:1818",
		Handler: http.HandlerFunc(wsHandler(pool)),
	}

	go func() {
		if err := wsServer.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	workers := 4
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

func wsHandler(pool *ants.Pool) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		conn, _, _, err := ws.UpgradeHTTP(req, w)
		if err != nil {
			// handle error
		}

		if err = pool.Submit(manageWsConn(conn)); err != nil {
			// handle error
		}
	}
}

func manageWsConn(conn net.Conn) func() {
	return func() {
		defer conn.Close()

		connSubHandlers := streaming.NewConnSubHandlers(subManager)

		for {
			msg, _, err := wsutil.ReadClientData(conn)
			if err != nil {
				// handle error
			}

			params := &streaming.Params{}
			if err := json.Unmarshal(msg, params); err == nil {
				id := uuid.New().String()
				connSubHandlers.AddHandler(conn, *params, id)

				if err = wsutil.WriteServerText(conn, []byte(id)); err != nil {
					log.Println(err)
				}
			} else {
				if id, err := uuid.Parse(string(msg)); err == nil {
					connSubHandlers.RemoveHandler(id.String())

					if err = wsutil.WriteServerText(conn, []byte("unsubscribed " + id.String())); err != nil {
						// handle error
					}
				} else {
					continue
				}
			}
		}
	}
}

