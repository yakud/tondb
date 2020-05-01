package streaming

import (
	"context"
	"errors"
	"net"
	"sync/atomic"

	"github.com/google/uuid"
)

type (
	ClientID string
	JSON     []byte

	Client struct {
		id        ClientID
		conn      net.Conn
		writeChan chan JSON
		writer    *AsyncWriter

		cancelWriter context.CancelFunc
		cancelled    int32

		subs []*Subscription
	}
)

func (c *Client) WriteAsync(j JSON) error {
	select {
	case c.writeChan <- j:
		return nil

	default:
		return errors.New("slow client")
	}
}

func (c *Client) Close() {
	if c.Cancelled() {
		return
	}

	c.Cancel()

	for _, sub := range c.subs {
		sub.Abandon()
	}

	close(c.writeChan)
	c.conn.Close()
	c.cancelWriter()
}

func (c *Client) AddSubscription(sub *Subscription) {
	c.subs = append(c.subs, sub)
}

func (c *Client) Cancelled() bool {
	if atomic.LoadInt32(&(c.cancelled)) != 0 {
		return true
	}
	return false
}

func (c *Client) Cancel() {
	atomic.StoreInt32(&(c.cancelled), 1)
}

func NewClient(conn net.Conn, cancel context.CancelFunc) *Client {
	return &Client{
		id:           ClientID(uuid.New().String()),
		conn:         conn,
		writeChan:    make(chan JSON, 25),
		subs:         make([]*Subscription, 0, 4),
		cancelWriter: cancel,
	}
}
