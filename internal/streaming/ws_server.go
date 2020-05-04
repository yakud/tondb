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
		conn.Write([]byte("error occurred when trying to do http upgrade."))
		return
	}

	// todo clients count by ip ~ 10
	// conn.RemoteAddr().

	// todo: metric clients count

	ctx, cancel := context.WithCancel(s.ctx)
	client := NewClient(conn, cancel)

	pollerDesc, err := netpoll.HandleRead(conn)
	if err != nil {
		log.Println(err)
	}

	err = s.poller.Start(pollerDesc, func(event netpoll.Event) {
		if event&netpoll.EventReadHup != 0 || event&netpoll.EventWriteHup != 0 || client.Cancelled() {
			s.poller.Stop(pollerDesc)
			pollerDesc.Close()
			client.Close()
			return
		}

		msg, _, err := wsutil.ReadClientData(conn)
		if err != nil {
			return
		}

		filter := &Filter{}
		if err := json.Unmarshal(msg, filter); err == nil {
			// todo: subs count by ip 30
			// todo: metric subs count

			sub, err := s.subscriber.Subscribe(client, *filter)
			if err != nil {
				log.Println("An error occurred when trying to subscribe client.")
				log.Print("error: ")
				log.Println(err)
				return
			}

			if client.writer == nil {
				client.writer = NewAsyncWriter()
				go client.writer.Run(ctx, client)
			}

			if err = wsutil.WriteServerText(conn, []byte(sub.id)); err != nil {
				log.Println(err)
				return
			}
		} else {
			if id, err := uuid.Parse(string(msg)); err == nil {
				if err := s.subscriber.Unsubscribe(SubscriptionID(id.String())); err != nil {
					log.Println("An error occurred when trying to unsubscribe client.")
					log.Print("error: ")
					log.Println(err)
					return
				}

				if err = wsutil.WriteServerText(conn, []byte("unsubscribed "+id.String())); err != nil {
					log.Println(err)
					return
				}
			}
		}
	})

	if err != nil {
		log.Println("An error occurred when trying to start poller.")
		log.Print("error: ")
		log.Println(err)
		return
	}
}

func NewWSServer(poller netpoll.Poller, subscriber Subscriber, ctx context.Context, cancel context.CancelFunc) *WSServer {
	return &WSServer{
		poller:     poller,
		subscriber: subscriber,
		ctx:        ctx,
		cancel:     cancel,
	}
}
