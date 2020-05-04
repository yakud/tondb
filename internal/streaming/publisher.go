package streaming

import (
	"encoding/json"
	"strconv"

	"gitlab.flora.loc/mills/tondb/internal/ton/view/feed"
)

type Publisher interface {
	PublishBlock(*Subscription, *feed.BlockInFeed) error
	PublishTransactions(*Subscription, []*feed.TransactionInFeed) error
	PublishMessages(*Subscription, []*feed.MessageInFeed) error
	ClearCache()
}

type JSONPublisher struct {
	jsonCacheBlocks       JSON
	jsonCacheTransactions map[string]JSON
	jsonCacheMessages     map[string]JSON
}

func (p *JSONPublisher) PublishBlock(sub *Subscription, block *feed.BlockInFeed) (err error) {
	if len(p.jsonCacheBlocks) == 0 {
		p.jsonCacheBlocks, err = json.Marshal(block)
		if err != nil {
			return err
		}
	}

	// todo: metric += len blocks

	return sub.client.WriteAsync(toWsFeed(sub.id, p.jsonCacheBlocks, "block"))
}

func (p *JSONPublisher) PublishTransactions(sub *Subscription, transactions []*feed.TransactionInFeed) error {
	trxJsons := make([]JSON, 0, len(transactions))

	for _, trx := range transactions {
		if trxJson, ok := p.jsonCacheTransactions[trx.TrxHash]; ok {
			trxJsons = append(trxJsons, trxJson)
			continue
		}

		trxJson, err := json.Marshal(trx)
		if err != nil {
			return err
		}
		trxJsons = append(trxJsons, trxJson)

		p.jsonCacheTransactions[trx.TrxHash] = trxJson
	}
	// todo: metric += len trxJsons

	mergedJsons := mergeJsons(trxJsons)

	return sub.client.WriteAsync(toWsFeed(sub.id, mergedJsons, "transactions"))
}

func (p *JSONPublisher) PublishMessages(sub *Subscription, messages []*feed.MessageInFeed) error {
	msgJsons := make([]JSON, 0, len(messages))

	for _, msg := range messages {
		key := msg.TrxHash + "," + strconv.FormatUint(msg.MessageLt, 10)

		if msgJson, ok := p.jsonCacheMessages[key]; ok {
			msgJsons = append(msgJsons, msgJson)
			continue
		}

		msgJson, err := json.Marshal(msg)
		if err != nil {
			return err
		}
		msgJsons = append(msgJsons, msgJson)

		p.jsonCacheMessages[key] = msgJson
	}

	// todo: metric += len messages

	mergedJsons := mergeJsons(msgJsons)

	return sub.client.WriteAsync(toWsFeed(sub.id, mergedJsons, "messages"))
}

func (p *JSONPublisher) ClearCache() {
	p.jsonCacheBlocks = make(JSON, 0, 128)
	p.jsonCacheTransactions = make(map[string]JSON, 32)
	p.jsonCacheMessages = make(map[string]JSON, 64)
}

func toWsFeed(id SubscriptionID, json JSON, fieldName string) JSON {
	res := JSON(`{"subscription_id":"` + string(id) + `","` + fieldName + `":`)
	res = append(res, json...)
	res = append(res, "}"...)
	return res
}

func mergeJsons(jsons []JSON) JSON {
	if len(jsons) == 0 {
		return JSON{}
	}

	res := make(JSON, 0, len(jsons)*len(jsons[0])+len(jsons)*16)

	res = append(res, "["...)
	for i, v := range jsons {
		res = append(res, v...)
		if i < len(jsons)-1 {
			res = append(res, ","...)
		}
	}
	res = append(res, "]"...)

	return res
}

func NewJSONPublisher() *JSONPublisher {
	return &JSONPublisher{
		jsonCacheBlocks:       make(JSON, 0, 128),
		jsonCacheTransactions: make(map[string]JSON, 32),
		jsonCacheMessages:     make(map[string]JSON, 64),
	}
}
