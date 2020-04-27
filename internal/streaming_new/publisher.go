package streaming_new

import (
	"encoding/json"

	"gitlab.flora.loc/mills/tondb/internal/ton/view/feed"
)

type Publisher interface {
	PublishBlock(*Client, *feed.BlockInFeed) error
	PublishTransactions(*Client, []*feed.TransactionInFeed) error
	PublishMessages(*Client, []*feed.MessageInFeed) error
	ClearCache()
}

type JSONPublisher struct {
	jsonCacheBlocks map[string]JSON // TODO:
}

func (p *JSONPublisher) PublishBlock(client *Client, block *feed.BlockInFeed) error {
	blockJson, _ := json.Marshal(block) // TODO: lazy
	return client.WriteAsync(blockJson)
}

func (p *JSONPublisher) PublishTransactions(*Client, []*feed.TransactionInFeed) error {
	panic("implement me")
}

func (p *JSONPublisher) PublishMessages(*Client, []*feed.MessageInFeed) error {
	panic("implement me")
}
