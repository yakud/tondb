package streaming_new

import "context"

type AsyncWriter struct {
}

func (w *AsyncWriter) Run(ctx context.Context, client *Client) error {
	for {
		select {
		case j, ok := <-client.writeChan:
			if !ok {
				return nil
			}

			client.conn.Write(j)

		case <-ctx.Done():
			return nil
		}
	}
}
