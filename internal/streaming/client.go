package streaming

import (
	"context"
	"fmt"
	"net"

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
		isCancel     bool
	}
)

func (c *Client) Close() {
	if !c.isCancel {
		close(c.writeChan)
		c.conn.Close()
		c.cancelWriter()
	}
}

func (c *Client) WriteAsync(j JSON) error {
	select {
	case c.writeChan <- j:
		return nil

	default:
		return fmt.Errorf("slow client")
	}
}

func NewClient(conn net.Conn) *Client {
	return &Client{
		id:        ClientID(uuid.New().String()),
		conn:      conn,
		writeChan: make(chan JSON, 25),
	}
}
