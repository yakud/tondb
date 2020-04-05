package main

import (
	"context"
	"log"
	"testing"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

func TestBlocksStreamReceiver(t *testing.T) {
	conn, _, _, err := ws.DefaultDialer.Dial(context.Background(), "ws://0.0.0.0:1818")
	if err != nil {
		log.Fatal(err)
	}

	// send message
	err = wsutil.WriteClientMessage(conn, ws.OpText, []byte(`{"feed_name": "blocks", "workchain_id": 0}`))
	if err != nil {
		log.Fatal(err)
	}

	for {
		// receive message
		msg, _, err := wsutil.ReadServerData(conn)
		if err != nil {
			log.Println(err)
		}
		log.Println(string(msg))
	}
}
