package streaming

import (
	"encoding/json"
	"fmt"
	"gitlab.flora.loc/mills/tondb/internal/ton/view/feed"
	"strconv"
)

type Publisher interface {
	PublishBlock(*Subscription, *feed.BlockInFeed) error
	PublishTransactions(*Subscription, []*feed.TransactionInFeed) error
	PublishMessages(*Subscription, []*feed.MessageInFeed) error
	ClearCache()
}

type JSONPublisher struct {
	// TODO: it's jsons cache per block, so there should be only one block's json cached. Do we really need map here?
	jsonCacheBlocks       map[string]JSON
	jsonCacheTransactions map[string]JSON
	jsonCacheMessages     map[string]JSON
}

func (p *JSONPublisher) PublishBlock(sub *Subscription, block *feed.BlockInFeed) error {
	key := fmt.Sprintf("%d,%d,%d", block.WorkchainId, block.Shard, block.SeqNo)

	if blockJson, ok := p.jsonCacheBlocks[key]; ok {
		return sub.client.WriteAsync(toWsFeed(sub.id, blockJson, "block"))
	}

	blockJson, err := json.Marshal(block)
	if err != nil {
		return err
	}

	p.jsonCacheBlocks[key] = blockJson

	return sub.client.WriteAsync(toWsFeed(sub.id, blockJson, "block"))
}

func (p *JSONPublisher) PublishTransactions(sub *Subscription, transactions []*feed.TransactionInFeed) error {
	trxJsons := make([]JSON, 0, len(transactions))

	for _, trx := range transactions {
		if trxJson, ok := p.jsonCacheTransactions[trx.TrxHash]; ok {
			trxJsons = append(trxJsons, trxJson)
		}

		trxJson, err := json.Marshal(trx)
		if err != nil {
			return err
		}

		p.jsonCacheTransactions[trx.TrxHash] = trxJson
	}

	mergedJsons := mergeJsons(trxJsons)

	return sub.client.WriteAsync(toWsFeed(sub.id, mergedJsons, "transactions"))
}

func (p *JSONPublisher) PublishMessages(sub *Subscription, messages []*feed.MessageInFeed) error {
	msgJsons := make([]JSON, 0, len(messages))
	key := ""

	for _, msg := range messages {
		key = msg.TrxHash + "," + strconv.FormatUint(msg.MessageLt, 64)

		if msgJson, ok := p.jsonCacheTransactions[key]; ok {
			msgJsons = append(msgJsons, msgJson)
		}

		msgJson, err := json.Marshal(msg)
		if err != nil {
			return err
		}

		p.jsonCacheTransactions[key] = msgJson
	}

	mergedJsons := mergeJsons(msgJsons)

	return sub.client.WriteAsync(toWsFeed(sub.id, mergedJsons, "messages"))
}

func (p *JSONPublisher) ClearCache() {
	p.jsonCacheBlocks = make(map[string]JSON)
	p.jsonCacheTransactions = make(map[string]JSON, 32)
	p.jsonCacheMessages = make(map[string]JSON, 64)
}

func toWsFeed(id SubscriptionID, json JSON, fieldName string) JSON {
	res := JSON(`{"subscription_id": ` + string(id) + `, "` + fieldName + `": `)
	res = append(res, json...)
	res = append(res, []byte("}")...)
	return res
}

func mergeJsons(jsons []JSON) JSON {
	res := make(JSON, 0, len(jsons)*len(jsons[0])+len(jsons)*16)
	comma := []byte(",")

	res = append(res, []byte("[")...)
	for i, v := range jsons {
		res = append(res, v...)
		if i < len(jsons)-1 {
			res = append(res, comma...)
		}
	}
	res = append(res, []byte("]")...)

	return res
}

func NewJSONPublisher() *JSONPublisher {
	return &JSONPublisher{
		jsonCacheBlocks:       make(map[string]JSON),
		jsonCacheTransactions: make(map[string]JSON, 32),
		jsonCacheMessages:     make(map[string]JSON, 64),
	}
}
