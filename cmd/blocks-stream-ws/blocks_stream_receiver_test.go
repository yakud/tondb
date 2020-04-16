package main

import (
	"context"
	"log"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"

	"github.com/google/uuid"
)

func TestBlocksStreamReceiver(t *testing.T) {
	var blks uint32 = 0
	var id uuid.UUID
	uuidObtained := false
	conn, _, _, err := ws.DefaultDialer.Dial(context.Background(), "ws://0.0.0.0:1818")
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	// send message
	err = wsutil.WriteClientText(conn, []byte(`{"feed_name": "messages", "account_addr": "-1:3333333333333333333333333333333333333333333333333333333333333333"}`))
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			<-time.After(time.Second * 5)
			log.Println(atomic.LoadUint32(&blks))
		}
	}()

	for {
		// receive message
		msg, err := wsutil.ReadServerText(conn)
		if !uuidObtained {
			id = uuid.MustParse(string(msg))
			uuidObtained = true
			go func() {
				for {
					<-time.After(time.Second * 100)
					log.Println("unsubscribing")
					wsutil.WriteClientText(conn, []byte(id.String()))
					return
				}
			}()
			continue
		}

		if err == nil {
			atomic.AddUint32(&blks, 1)
			log.Println(string(msg))
		}
	}
}
