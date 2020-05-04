package streaming

import (
	"context"
	"encoding/json"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/google/uuid"
	"github.com/mailru/easygo/netpoll"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	ClientsLimitByIp = 10
	SubsLimitByIp    = 30
)

var (
	clientsCountMetrics = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "clients_count",
		Help: "Number of clients",
	}, []string{})

	subsCountMetrics = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "subs_count",
		Help: "Number of subscriptions",
	}, []string{})

	subsCountLabeledMetrics = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "subs_count_labeled",
		Help: "Labeled subscriptions counter",
	}, []string{"feed_name", "has_workchain", "has_shard", "has_acc_addr", "has_message_direction", "has_custom_filters"})
)

type WSServer struct {
	poller     netpoll.Poller
	subscriber Subscriber

	clientsCountByIp *RateLimiter
	subsCountByIp    *RateLimiter

	ctx    context.Context
	cancel context.CancelFunc
}

func (s *WSServer) Stop() {
	s.cancel()
}

func (s *WSServer) Handler(w http.ResponseWriter, req *http.Request) {
	conn, _, _, err := ws.UpgradeHTTP(req, w)
	if err != nil {
		http.Error(w, "error occurred when trying to do http upgrade.", http.StatusInternalServerError)
		return
	}

	clientIp := conn.RemoteAddr().String()
	if err := s.clientsCountByIp.CheckLimitAndInc(clientIp); err != nil {
		http.Error(w, err.Error(), http.StatusTooManyRequests)
	}

	clientsCountMetrics.WithLabelValues().Inc()

	ctx, cancel := context.WithCancel(s.ctx)
	onClientCancelCallback := func() {
		s.clientsCountByIp.Dec(clientIp)
		clientsCountMetrics.WithLabelValues().Dec()
	}
	client := NewClient(conn, cancel, onClientCancelCallback)

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

			if err:= s.subsCountByIp.CheckLimitAndInc(clientIp); err != nil {
				http.Error(w, err.Error(), http.StatusTooManyRequests)
			}

			sub, err := s.subscriber.Subscribe(client, *filter)
			if err != nil {
				s.subsCountByIp.Dec(clientIp)
				log.Println("An error occurred when trying to subscribe client.")
				log.Print("error: ")
				log.Println(err)
				return
			}

			hasWorkchain := "0"
			if sub.filter.WorkchainId != nil {
				hasWorkchain = "1"
			}
			hasShard := "0"
			if sub.filter.Shard != nil {
				hasShard = "1"
			}
			hasAccAddr := "0"
			if sub.filter.AccountAddr != nil {
				hasAccAddr = "1"
			}
			hasMsgDir := "0"
			if sub.filter.MessageDirection != nil {
				hasMsgDir = "1"
			}
			hasCustFilters := "0"
			if len(sub.filter.CustomFilters) > 0 {
				hasCustFilters = "1"
			}
			subsCountMetrics.WithLabelValues().Inc()
			subsCountLabeledMetrics.WithLabelValues(string(sub.filter.FeedName), hasWorkchain, hasShard, hasAccAddr, hasMsgDir, hasCustFilters).Inc()

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

				s.subsCountByIp.Dec(clientIp)
				subsCountMetrics.WithLabelValues().Dec()

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

		clientsCountByIp: NewRateLimiter(ClientsLimitByIp),
		subsCountByIp:    NewRateLimiter(SubsLimitByIp),

		ctx:    ctx,
		cancel: cancel,
	}
}
