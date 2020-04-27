package streaming_new

import (
	"fmt"
	"net"
)

type (
	ClientID string
	JSON     []byte

	Client struct {
		id        ClientID
		conn      net.Conn
		writeChan chan JSON // TODO: buffered
	}
)

func (c *Client) WriteAsync(j JSON) error {
	select {
	case c.writeChan <- j:
		return nil

	default:
		return fmt.Errorf("slow client")
	}
}

//func NewClient() *Client {
//	....
//}
