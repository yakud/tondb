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
		select {
		case json, ok := <-client.writeChan:
			if !ok {
				return nil
			}

			jsonBuffer = append(jsonBuffer, json)

		default:
			// channel is empty, flush buffer to user
			// TODO: we send here separate messages for every entry in buffer, maybe it's better to join all or some
			//  entries with some delimiter and then send them as one message
			for _, json := range jsonBuffer {
				if err := wsutil.WriteServerText(client.conn, json); err != nil {
					// TODO: handle error properly, maybe we need to return it?
				}
			}

			jsonBuffer = make([]JSON, 0, 25)

		case <-ctx.Done():
			return nil
		}
	}
}

func NewAsyncWriter() *AsyncWriter {
	return &AsyncWriter{}
}
