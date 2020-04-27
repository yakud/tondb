package streaming_new

import (
	"github.com/google/btree"
	"gitlab.flora.loc/mills/tondb/internal/ton/view/feed"
)

/*
                  |
client -> conn -> | SocketServer -> conn
                  |

client -> sub(filter) | subs []{Client, Filter}

client -> sub({blocks, workchain = -1}) | subs []{Client, {blocks, workchain = -1}}
client -> sub({trx, workchain = -1}) | subs []{Client, {blocks, workchain = -1}}, {Client, {trx, workchain = -1}}
client -> sub({blocks, workchain = 0}) | subs []{Client, {blocks, workchain = -1}}, {Client, {trx, workchain = -1}, {Client, {blocks, workchain = 0}}

*/

const (
	FeedNameBlocks       FeedName = "blocks"
	FeedNameTransactions FeedName = "transactions"
	FeedNameMessages     FeedName = "messages"
)

type (
	FeedName   string
	Addr       string
	FilterHash string

	Filter struct {
		FeedName      FeedName       `json:"feed_name"`
		WorkchainId   *int32         `json:"workchain_id,omitempty"`
		Shard         *uint64        `json:"shard,omitempty"`
		AccountAddr   *string        `json:"account_addr,omitempty"`
		CustomFilters []CustomFilter `json:"custom_filters,omitempty"`
	}

	CustomFilter struct {
		Field       string `json:"field"`     // enum
		Operation   string `json:"operation"` // enum eq, lt, gt, range
		ValueString string `json:"value_string"`
	}
)

func (f *Filter) Hash() FilterHash {
	// TODO:
	// fmt.Sprintf("......", f.WorkchainId, f.Shard, f.AccountAddr)
	panic("implement me")
}

type Indexer interface {
	IndexBlock(*feed.BlockInFeed) error
	IndexTransaction(*feed.TransactionInFeed) error
	IndexMessage(*feed.MessageInFeed) error
}

type Fetcher interface {
	FetchBlocks(Filter) (*feed.BlockInFeed, error)
	FetchTransactions(Filter) ([]*feed.TransactionInFeed, error)
	FetchMessage(Filter) ([]*feed.MessageInFeed, error)
}

type Index struct {
	block        *feed.BlockInFeed
	transactions []*feed.TransactionInFeed
	messages     []*feed.MessageInFeed

	transactionsByAddr map[Addr][]*feed.TransactionInFeed
	messagesByAddr     map[Addr][]*feed.MessageInFeed

	transactionsTotalNanogram *btree.BTree
	messagesValueNanogram     *btree.BTree
}

func (i *Index) FetchBlocks(Filter) (*feed.BlockInFeed, error) {
	panic("implement me")
}

func (i *Index) FetchTransactions(Filter) ([]*feed.TransactionInFeed, error) {
	panic("implement me")
}

func (i *Index) FetchMessage(Filter) ([]*feed.MessageInFeed, error) {
	panic("implement me")
}

func (i *Index) IndexBlock(block *feed.BlockInFeed) error {
	// fill block
	//block.Shard

	panic("implement me")
}

func (i *Index) IndexTransaction(*feed.TransactionInFeed) error {
	// fill transactions, transactionsByAddr, transactionsTotalNanogram
	panic("implement me")
}

func (i *Index) IndexMessage(*feed.MessageInFeed) error {
	panic("implement me")
}
