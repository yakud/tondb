package streaming

import (
	"context"

	"github.com/gobwas/ws/wsutil"
)

type AsyncWriter struct {
	client      *Client
	jsonBuffer  []JSON
}

func (w *AsyncWriter) Run(ctx context.Context) error {
	for {
		select {
		case json, ok := <-w.client.writeChan:
			if !ok {
				return nil
			}

			w.jsonBuffer = append(w.jsonBuffer, json)

		default:
			// channel is empty, flush buffer to user
			// TODO: we send here separate messages for every entry in buffer, maybe it's better to join all or some
			//  entries with some delimiter and then send them as one message
			for _, json := range w.jsonBuffer {
				if err := wsutil.WriteServerText(w.client.conn, json); err != nil {
					// TODO: handle error properly, maybe we need to return it?
				}
			}

			w.jsonBuffer = make([]JSON, 0, 25)

		case <-ctx.Done():
			return nil
		}
	}
}

func NewAsyncWriter(client *Client) *AsyncWriter {
	return &AsyncWriter{
		client:     client,
		jsonBuffer: make([]JSON, 0, 25),
	}
}
