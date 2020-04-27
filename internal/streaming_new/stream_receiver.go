package streaming_new

import (
	"gitlab.flora.loc/mills/tondb/internal/ton"
	"gitlab.flora.loc/mills/tondb/internal/ton/view/feed"
)

type FeedConverter interface {
	ConvertBlock(*ton.Block) (*feed.BlockInFeed, error)
	ConvertTransactions([]*ton.Transaction) ([]*feed.TransactionInFeed, error)
	ConvertMessages([]*ton.TransactionMessage) ([]*feed.MessageInFeed, error)
}

type StreamReceiver struct {
	converter  FeedConverter
	subscriber Subscriber
	publisher  Publisher
}

func (r *StreamReceiver) HandleBlock(block *ton.Block) error {
	defer r.publisher.ClearCache()

	//blockJsonCache := JSON()
	//blockJsonCache := JSON()

	index := Index{ // TODO NewIndex()
		block:                     nil,
		transactions:              nil,
		messages:                  nil,
		transactionsByAddr:        nil,
		messagesByAddr:            nil,
		transactionsTotalNanogram: nil,
		messagesValueNanogram:     nil,
	}

	blockInFeed, _ := r.converter.ConvertBlock(block)
	transactionsInFeed, _ := r.converter.ConvertTransactions(block.Transactions)
	//transactionInFeed, _ := r.converter.ConvertMessages(block.) // TODO:

	index.IndexBlock(blockInFeed)
	// for transactionsInFeed
	//index.IndexTransaction(transactionInFeed)

	// index

	//r.subscriber.Fetch()

	for _, subscriptions := range r.subscriber.Subscriptions() {
		switch subscriptions.filter.FeedName {
		case FeedNameBlocks:
			block, _ := index.FetchBlocks(subscriptions.filter)

			// Async Send block to all clients
			for _, sub := range subscriptions.subs {
				err := r.publisher.PublishBlock(sub.client, block)
				if err != nil {
					r.subscriber.Unsubscribe(sub.id)
					close(sub.client.writeChan)
					// destroy sub.client
				}
			}

		case FeedNameTransactions:
			trxs, _ := index.FetchTransactions(subscriptions.filter)
			// Async Send trxs to all clients
			//for _, sub := range subscriptions.subs {
			//  sub.client.conn.Write(trxs)
			//}

			//subscriptions[0].client.conn.Write({sub_id:"", data: [feed_name, block:{}]})
			//subscriptions[0].client.conn.Write({sub_id:"", data: [feed_name, transactions:[]]})
			//

		case FeedNameMessages:
			msgs, _ := index.FetchMessage(subscriptions.filter)
			// Async Send msgs to all clients

		}
	}

}
