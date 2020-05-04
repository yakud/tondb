package streaming

import (
	"gitlab.flora.loc/mills/tondb/internal/ton"
)

type StreamReceiver struct {
	converter  FeedConverter
	subscriber Subscriber
}

func (r *StreamReceiver) HandleBlock(block *ton.Block) error {
	publisher := NewJSONPublisher()

	index, err := r.makeFeedIndex(block)
	if err != nil {
		return err
	}

	return r.subscriber.IterateSubscriptions(func(subscriptions *Subscriptions) error {
		switch subscriptions.filter.FeedName {
		case FeedNameBlocks:
			block, err := index.FetchBlock(subscriptions.filter)
			if err != nil {
				return err
			}

			if block == nil {
				return nil
			}

			// Async Send block to all clients
			for _, sub := range subscriptions.subs {
				if sub.IsAbandoned() {
					continue
				}

				if err := publisher.PublishBlock(sub, block); err != nil {
					sub.client.Close()
				}
			}

		case FeedNameTransactions:
			trxs, err := index.FetchTransactions(subscriptions.filter)
			if err != nil {
				return err
			}

			if len(trxs) == 0 {
				return nil
			}

			// Async Send msgs to all clients
			for _, sub := range subscriptions.subs {
				if sub.IsAbandoned() {
					continue
				}

				if err := publisher.PublishTransactions(sub, trxs); err != nil {
					sub.client.Close()
				}
			}

		case FeedNameMessages:
			msgs, err := index.FetchMessage(subscriptions.filter)
			if err != nil {
				return err
			}

			if len(msgs) == 0 {
				return nil
			}

			// Async Send msgs to all clients
			for _, sub := range subscriptions.subs {
				if sub.IsAbandoned() {
					continue
				}

				if err := publisher.PublishMessages(sub, msgs); err != nil {
					sub.client.Close()
				}
			}
		}

		return nil
	})
}

func (r *StreamReceiver) makeFeedIndex(block *ton.Block) (*Index, error) {
	index := NewIndex()

	blockInFeed, err := r.converter.ConvertBlock(block)
	if err != nil {
		return nil, err
	}

	if err := index.IndexBlock(blockInFeed); err != nil {
		return nil, err
	}

	for _, trx := range block.Transactions {
		trxInFeed, err := r.converter.ConvertTransaction(trx)
		if err != nil {
			return nil, err
		}

		if err := index.IndexTransaction(trxInFeed); err != nil {
			return nil, err
		}

		if trx.InMsg != nil {
			inMsgInFeed, err := r.converter.ConvertMessage(block, trx, trx.InMsg, string(MessageDirectionIn), trx.Lt)
			if err != nil {
				return nil, err
			}

			if err := index.IndexMessage(inMsgInFeed); err != nil {
				return nil, err
			}
		}

		for _, msg := range trx.OutMsgs {
			msgInFeed, err := r.converter.ConvertMessage(block, trx, msg, string(MessageDirectionOut), trx.Lt)
			if err != nil {
				return nil, err
			}

			if err := index.IndexMessage(msgInFeed); err != nil {
				return nil, err
			}
		}
	}

	return index, nil
}

func NewStreamReceiver(subscriber Subscriber) *StreamReceiver {
	return &StreamReceiver{
		converter:  &FeedConverterImpl{},
		subscriber: subscriber,
	}
}
