package streaming

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/google/uuid"
	"github.com/mailru/easygo/netpoll"
)

type WSServer struct {
	poller     netpoll.Poller
	subscriber Subscriber

	ctx    context.Context
	cancel context.CancelFunc
}

func (s *WSServer) Stop() {
	s.cancel()
}

func (s *WSServer) Handler(w http.ResponseWriter, req *http.Request) {
	conn, _, _, err := ws.UpgradeHTTP(req, w)
	if err != nil {
		// handle error
	}

	//isWriterRun := false
	client := NewClient(conn)
	ctx, cancel := context.WithCancel(s.ctx) // todo: положить в юзера
	client.cancelWriter = cancel

	pollerDesc, err := netpoll.HandleRead(conn)
	if err != nil {
		// TODO: handle errors properly
		log.Println(err)
	}

	//defer s.poller.Stop(pollerDesc)
	//defer pollerDesc.Close()
	//
	//ctx, cancel := context.WithCancel(s.ctx)
	//defer cancel()

	err = s.poller.Start(pollerDesc, func(event netpoll.Event) {
		if event&netpoll.EventReadHup != 0 || event&netpoll.EventWriteHup != 0 {
			// TODO
			s.poller.Stop(pollerDesc)
			conn.Close()
			return
		}

		msg, _, err := wsutil.ReadClientData(conn)
		if err != nil {
			// handle error
		}

		filter := &Filter{}
		if err := json.Unmarshal(msg, filter); err == nil {
			sub, err := s.subscriber.Subscribe(client, *filter)
			if err != nil {
				// TODO handle error
			}

			if client.writer == nil {
				client.writer = NewAsyncWriter()
				go client.writer.Run(ctx, client) // TODO
			}

			if err = wsutil.WriteServerText(conn, []byte(sub.id)); err != nil {
				log.Println(err)
			}
		} else {
			if id, err := uuid.Parse(string(msg)); err == nil {
				if err := s.subscriber.Unsubscribe(SubscriptionID(id.String())); err != nil {
					// TODO: Handle error
				}

				if err = wsutil.WriteServerText(conn, []byte("unsubscribed "+id.String())); err != nil {
					// handle error
				}
			}
		}
	})

	if err != nil {
		// handle error
	}
}

func NewWSServer(poller netpoll.Poller, subscriber Subscriber) *WSServer {
	return &WSServer{
		poller:     poller,
		subscriber: subscriber,
	}
}
