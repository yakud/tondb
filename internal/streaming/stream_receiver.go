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
	var publisher Publisher

	index := NewIndex() // TODO: Maybe we need to pass built index when building StreamReceiver?

	blockInFeed, err := r.converter.ConvertBlock(block)
	if err != nil {
		return err
	}

	err = index.IndexBlock(blockInFeed)
	if err != nil {
		return err
	}

	transactionsInFeed := make([]*feed.TransactionInFeed, 0, len(block.Transactions))
	messagesInFeed := make([]*feed.MessageInFeed, 0, len(transactionsInFeed)*5)

	trxByTotalNanogram := map[uint64][]*feed.TransactionInFeed{}

	for _, trx := range block.Transactions {
		trxInFeed, err := r.converter.ConvertTransaction(trx)
		if err != nil {
			return err
		}
		transactionsInFeed = append(transactionsInFeed, trxInFeed)

		// Todo: проверки
		trxByTotalNanogram[trxInFeed.TotalNanograms] = append(trxByTotalNanogram[trxInFeed.TotalNanograms], trxInFeed)

		//err = index.IndexTransaction(trxInFeed)
		//if err != nil {
		//	return err
		//}

		inMsgInFeed, err := r.converter.ConvertMessage(block, trx.InMsg, "in", trx.Lt)
		if err != nil {
			return err
		}
		messagesInFeed = append(messagesInFeed, inMsgInFeed)
		err = index.IndexMessage(inMsgInFeed)
		if err != nil {
			return err
		}

		for _, msg := range trx.OutMsgs {
			msgInFeed, err := r.converter.ConvertMessage(block, msg, "out", trx.Lt)
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

	// todo: for messages same
	for totalNanogram, trxs := range trxByTotalNanogram {
		index.transactionsTotalNanogram.ReplaceOrInsert(NewUInt64TrxsIndex(totalNanogram, trxs))
	}

	return r.subscriber.IterateSubscriptions(func(subscriptions *Subscriptions) error {
		switch subscriptions.filter.FeedName {
		case FeedNameBlocks:
			block, err := index.FetchBlocks(subscriptions.filter)
			if err != nil {
				return err // TODO: handle error
			}

			// Async Send block to all clients
			for _, sub := range subscriptions.subs {
				if sub.abandoned { // todo: atomic
					continue
				}

				err := publisher.PublishBlock(sub, block)
				if err != nil { // TODO: check error if we need to destroy sub client (we need to destroy it if it is slow)
					err = r.subscriber.Unsubscribe(sub.id)
					if err != nil {
						return err // TODO: handle error
					}
				}
			}

		case FeedNameTransactions:
			trxs, err := index.FetchTransactions(subscriptions.filter)
			if err != nil {
				return err // TODO: handle error
			}

			// Async Send trxs to all clients
			for _, sub := range subscriptions.subs {
				if sub.abandoned {
					continue
				}

				err := publisher.PublishTransactions(sub, trxs)
				if err != nil { // TODO: check error if we need to destroy sub client (we need to destroy it if it is slow)
					err = r.subscriber.Unsubscribe(sub.id)
					if err != nil {
						return err // TODO: handle error
					}
				}
			}

		case FeedNameMessages:
			msgs, err := index.FetchMessage(subscriptions.filter)
			if err != nil {
				return err // TODO: handle error
			}

			// Async Send msgs to all clients
			for _, sub := range subscriptions.subs {
				if sub.abandoned {
					continue
				}

				err := publisher.PublishMessages(sub, msgs)
				if err != nil { // TODO: check error if we need to destroy sub client (we need to destroy it if it is slow)
					err = r.subscriber.Unsubscribe(sub.id)
					if err != nil {
						return err // TODO: handle error
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
