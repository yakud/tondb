package streaming

import (
	"gitlab.flora.loc/mills/tondb/internal/ton"
	"gitlab.flora.loc/mills/tondb/internal/ton/view/feed"
)

type StreamReceiver struct {
	converter  FeedConverter
	subscriber Subscriber
}

func (r *StreamReceiver) HandleBlock(block *ton.Block) error {
	publisher := NewJSONPublisher()

	index := NewIndex()

	blockInFeed, err := r.converter.ConvertBlock(block)
	if err != nil {
		return err
	}

	err = index.IndexBlock(blockInFeed)
	if err != nil {
		return err
	}

	transactionsInFeed := make([]*feed.TransactionInFeed, 0, len(block.Transactions))
	messagesInFeed := make([]*feed.MessageInFeed, 0, len(block.Transactions)*5)

	for _, trx := range block.Transactions {
		trxInFeed, err := r.converter.ConvertTransaction(trx)
		if err != nil {
			return err
		}
		transactionsInFeed = append(transactionsInFeed, trxInFeed)

		err = index.IndexTransaction(trxInFeed)
		if err != nil {
			return err
		}

		if trx.InMsg != nil {
			inMsgInFeed, err := r.converter.ConvertMessage(block, trx.InMsg, string(MessageDirectionIn), trx.Lt)
			if err != nil {
				return err
			}
			messagesInFeed = append(messagesInFeed, inMsgInFeed)

			err = index.IndexMessage(inMsgInFeed)
			if err != nil {
				return err
			}
		}

		for _, msg := range trx.OutMsgs {
			msgInFeed, err := r.converter.ConvertMessage(block, msg, string(MessageDirectionOut), trx.Lt)
			if err != nil {
				return err
			}
			messagesInFeed = append(messagesInFeed, msgInFeed)

			err = index.IndexMessage(msgInFeed)
			if err != nil {
				return err
			}
		}
	}

	return r.subscriber.IterateSubscriptions(func(subscriptions *Subscriptions) error {
		switch subscriptions.filter.FeedName {
		case FeedNameBlocks:
			block, err := index.FetchBlocks(subscriptions.filter)
			if err != nil {
				return err
			}

			if block == nil {
				return nil
			}
			// Async Send block to all clients
			for _, sub := range subscriptions.subs {
				if sub.GetAbandoned() {
					continue
				}

				err := publisher.PublishBlock(sub, block)
				if err != nil {
					if err = r.subscriber.Unsubscribe(sub.id); err != nil {
						return err
					}
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
				if sub.GetAbandoned() {
					continue
				}

				err := publisher.PublishTransactions(sub, trxs)
				if err != nil {
					if err = r.subscriber.Unsubscribe(sub.id); err != nil {
						return err
					}
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
				if sub.GetAbandoned() {
					continue
				}

				err := publisher.PublishMessages(sub, msgs)
				if err != nil {
					if err = r.subscriber.Unsubscribe(sub.id); err != nil {
						return err
					}
				}
			}
		}

		return nil
	})
}

func NewStreamReceiver(subscriber Subscriber) *StreamReceiver {
	return &StreamReceiver{
		converter:  &FeedConverterImpl{},
		subscriber: subscriber,
	}
}
