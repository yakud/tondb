package streaming

import (
	"context"

	"github.com/gobwas/ws/wsutil"
)

type AsyncWriter struct {
}

func (w *AsyncWriter) Run(ctx context.Context, client *Client) error {
	var jsonBuffer = make([]JSON, 0, 25)

	for {
		if client.Cancelled() {
			return nil
		}

		select {
		case json, ok := <-client.writeChan:
			if !ok {
				return nil
			}

			jsonBuffer = append(jsonBuffer, json)

		case <-ctx.Done():
			return nil

		default:
			// channel is empty, flush buffer to user
			for _, json := range jsonBuffer {
				// todo: make it in one bulk write
				if err := wsutil.WriteServerText(client.conn, json); err != nil {
					client.Close()
					return err
				}
			}

			jsonBuffer = jsonBuffer[:0]
		}
	}
}

func NewAsyncWriter() *AsyncWriter {
	return &AsyncWriter{}
}