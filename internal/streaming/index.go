package streaming

import (
	"errors"
	"github.com/google/btree"
	"gitlab.flora.loc/mills/tondb/internal/ton/view/feed"
	"strconv"
)

type (
	Indexer interface {
		IndexBlock(*feed.BlockInFeed) error
		IndexTransaction(*feed.TransactionInFeed) error
		IndexMessage(*feed.MessageInFeed) error
	}

	Fetcher interface {
		FetchBlocks(Filter) (*feed.BlockInFeed, error)
		FetchTransactions(Filter) ([]*feed.TransactionInFeed, error)
		FetchMessage(Filter) ([]*feed.MessageInFeed, error)
	}

	IndexIterator interface {
		Iterate(item btree.Item) bool
	}

	Index struct {
		block        *feed.BlockInFeed
		transactions []*feed.TransactionInFeed
		messages     []*feed.MessageInFeed

		transactionsByAddr map[Addr][]*feed.TransactionInFeed
		messagesByAddr     map[Addr][]*feed.MessageInFeed

		transactionsTotalNanogram *btree.BTree
		messagesValueNanogram     *btree.BTree
	}

	// TODO: Do we need some kind of interfaces for these two structs
	UInt64TrxIndex struct {
		key    uint64
		trxs   []*feed.TransactionInFeed
	}

	UInt64MsgIndex struct {
		key    uint64
		msgs   []*feed.MessageInFeed
	}

	TrxIndexIterator struct {
		trxs []*feed.TransactionInFeed
	}

	MsgIndexIterator struct {
		msgs []*feed.MessageInFeed
	}
)

func (i *Index) FetchBlocks(f Filter) (*feed.BlockInFeed, error) {
	if i.block == nil {
		return nil, errors.New("no block in index")
	}

	if f.MatchWorkchainAndShard(i.block) {
		return i.block, nil
	}

	return nil, nil
}

func (i *Index) FetchTransactions(f Filter) ([]*feed.TransactionInFeed, error) {
	if i.block == nil {
		return nil, errors.New("no block in index")
	}

	if !f.MatchWorkchainAndShard(i.block) {
		return []*feed.TransactionInFeed{}, nil
	}

	if f.AccountAddr == nil || len(f.CustomFilters) == 0 {
		return i.transactions, nil
	}

	trxsToIntersect := make([][]*feed.TransactionInFeed, 0, 8)
	if f.AccountAddr != nil {
		if trxs, ok := i.transactionsByAddr[Addr(*f.AccountAddr)]; ok {
			trxsToIntersect = append(trxsToIntersect, trxs)
		} else {
			return []*feed.TransactionInFeed{}, nil
		}
	}

	if len(f.CustomFilters) == 0 {
		return i.intersectTransactions(trxsToIntersect), nil
	}

	for _, cf := range f.CustomFilters {
		iter := NewTrxIndexIterator()
		if err := i.fetch(cf, iter, ConstructUInt64TrxIndex); err != nil {
		    return nil, err
		}
		trxsToIntersect = append(trxsToIntersect, iter.GetTransactions())
	}

	return i.intersectTransactions(trxsToIntersect), nil
}

func (i *Index) FetchMessage(f Filter) ([]*feed.MessageInFeed, error) {
	if i.block == nil {
		return nil, errors.New("no block in index")
	}

	if !f.MatchWorkchainAndShard(i.block) {
		return []*feed.MessageInFeed{}, nil
	}

	if f.AccountAddr == nil || len(f.CustomFilters) == 0 {
		return i.messages, nil
	}

	msgsToIntersect := make([][]*feed.MessageInFeed, 0, 8)
	if f.AccountAddr != nil {
		if msgs, ok := i.messagesByAddr[Addr(*f.AccountAddr)]; ok {
			msgsToIntersect = append(msgsToIntersect, msgs)
		} else {
			return []*feed.MessageInFeed{}, nil
		}
	}

	if len(f.CustomFilters) == 0 {
		return i.intersectMessages(msgsToIntersect), nil
	}

	for _, cf := range f.CustomFilters {
		iter := NewMsgIndexIterator()
		if err := i.fetch(cf, iter, ConstructUInt64MsgIndex); err != nil {
			return nil, err
		}
		msgsToIntersect = append(msgsToIntersect, iter.GetMessages())
	}

	return i.intersectMessages(msgsToIntersect), nil
}

func (i *Index) fetch(cf CustomFilter, iter IndexIterator, itemConstructor func(uint64) btree.Item) error {
	switch cf.Operation {
	case OpEq:
		v, err := strconv.ParseUint(cf.ValueString, 10, 64)
		if err != nil {
			return err
		}

		item := i.transactionsTotalNanogram.Get(itemConstructor(v))
		if item == nil {
			return nil
		}

		iter.Iterate(item)
	case OpLt:
		v, err := strconv.ParseUint(cf.ValueString, 10, 64)
		if err != nil {
			return err
		}

		i.transactionsTotalNanogram.AscendLessThan(itemConstructor(v), iter.Iterate)
	case OpGt:
		v, err := strconv.ParseUint(cf.ValueString, 10, 64)
		if err != nil {
			return err
		}

		i.transactionsTotalNanogram.DescendGreaterThan(itemConstructor(v), iter.Iterate)
	case OpRange:
		first, second, err := cf.ParseRange()
		if err == nil {
			return err
		}

		i.transactionsTotalNanogram.DescendRange(itemConstructor(first), itemConstructor(second), iter.Iterate)
	default:
		return errors.New("wrong operation code")
	}

	return nil
}

func (i *Index) intersectTransactions(trxsToIntersect [][]*feed.TransactionInFeed) []*feed.TransactionInFeed {
	smallestTrxs := trxsToIntersect[0]
	trxSets := make([]map[string]struct{}, 0, len(trxsToIntersect))
	for i, trxs := range trxsToIntersect {
		if len(trxs) < len(smallestTrxs) {
			smallestTrxs = trxs
		}

		trxSets[i] = make(map[string]struct{}, len(trxs))
		for _, trx := range trxs {
			trxSets[i][trx.TrxHash] = struct{}{}
		}
	}

	result := make([]*feed.TransactionInFeed, 0, len(smallestTrxs))
	for _, trx := range smallestTrxs {
		for i, set := range trxSets {
			if _, ok := set[trx.TrxHash]; ok {
				if i == len(trxSets) - 1 {
					result = append(result, trx)
				}
			} else {
				break
			}
		}
	}

	return result
}

func (i *Index) intersectMessages(msgsToIntersect [][]*feed.MessageInFeed) []*feed.MessageInFeed {
	smallestMsgs := msgsToIntersect[0]
	msgSets := make([]map[string]struct{}, 0, len(msgsToIntersect))
	for i, msgs := range msgsToIntersect {
		if len(msgs) < len(smallestMsgs) {
			smallestMsgs = msgs
		}

		msgSets[i] = make(map[string]struct{}, len(msgs))
		for _, msg := range msgs {
			msgSets[i][msg.TrxHash] = struct{}{}
		}
	}

	result := make([]*feed.MessageInFeed, 0, len(smallestMsgs))
	for _, msg := range smallestMsgs {
		for i, set := range msgSets {
			if _, ok := set[msg.TrxHash]; ok {
				if i == len(msgSets) - 1 {
					result = append(result, msg)
				}
			} else {
				break
			}
		}
	}

	return result
}

func (i *Index) IndexBlock(block *feed.BlockInFeed) error {
	i.block = block
	return nil
}

func (i *Index) IndexTransaction(trx *feed.TransactionInFeed) error {
	i.transactions = append(i.transactions, trx)

	if v, ok := i.transactionsByAddr[Addr(trx.AccountAddr)]; ok {
		v = append(v, trx)
	} else {
		i.transactionsByAddr[Addr(trx.AccountAddr)] = []*feed.TransactionInFeed{trx}
	}

	if rawIndex := i.transactionsTotalNanogram.ReplaceOrInsert(NewUInt64TrxIndex(trx.TotalNanograms, trx)); rawIndex != nil {
		index := rawIndex.(UInt64TrxIndex)
		index.trxs = append(index.trxs, trx)
		i.transactionsTotalNanogram.ReplaceOrInsert(index)
	}

	return nil
}

func (i *Index) IndexMessage(msg *feed.MessageInFeed) error {
	i.messages = append(i.messages, msg)

	if v, ok := i.messagesByAddr[Addr(msg.Src)]; ok {
		v = append(v, msg)
	} else {
		i.messagesByAddr[Addr(msg.Src)] = []*feed.MessageInFeed{msg}
	}

	if v, ok := i.messagesByAddr[Addr(msg.Dest)]; ok {
		v = append(v, msg)
	} else {
		i.messagesByAddr[Addr(msg.Dest)] = []*feed.MessageInFeed{msg}
	}

	if rawIndex := i.messagesValueNanogram.ReplaceOrInsert(NewUInt64MsgIndex(msg.ValueNanogram, msg)); rawIndex != nil {
		index := rawIndex.(UInt64MsgIndex)
		index.msgs = append(index.msgs, msg)
		i.messagesValueNanogram.ReplaceOrInsert(index)
	}

	return nil
}

func NewIndex() *Index {
	return &Index{
		transactions: make([]*feed.TransactionInFeed, 0, 16),
		messages:     make([]*feed.MessageInFeed, 0, 32),

		transactionsByAddr:        make(map[Addr][]*feed.TransactionInFeed),
		messagesByAddr:            make(map[Addr][]*feed.MessageInFeed),
		transactionsTotalNanogram: btree.New(2),
		messagesValueNanogram:     btree.New(2),
	}
}

func (i UInt64TrxIndex) Less(item btree.Item) bool {
	return i.key < item.(UInt64TrxIndex).key
}

func (i UInt64TrxIndex) GetTransactions() []*feed.TransactionInFeed {
	return i.trxs
}

func NewUInt64TrxIndex(key uint64, trx *feed.TransactionInFeed) UInt64TrxIndex {
	return UInt64TrxIndex{
		key:  key,
		trxs: []*feed.TransactionInFeed{trx},
	}
}

func ConstructUInt64TrxIndex(key uint64) btree.Item {
	return UInt64TrxIndex{key: key}
}

func (i UInt64MsgIndex) Less(item btree.Item) bool {
	return i.key < item.(UInt64TrxIndex).key
}

func (i UInt64MsgIndex) GetMessages() []*feed.MessageInFeed {
	return i.msgs
}

func NewUInt64MsgIndex(key uint64, msg *feed.MessageInFeed) UInt64MsgIndex {
	return UInt64MsgIndex{
		key:  key,
		msgs: []*feed.MessageInFeed{msg},
	}
}

func ConstructUInt64MsgIndex(key uint64) btree.Item {
	return UInt64MsgIndex{key: key}
}

func (it *TrxIndexIterator) Iterate(itemRaw btree.Item) bool {
	item := itemRaw.(UInt64TrxIndex)
	it.trxs = append(it.trxs, item.GetTransactions()...)
	return true
}

func (it *TrxIndexIterator) GetTransactions() []*feed.TransactionInFeed {
	return it.trxs
}

func NewTrxIndexIterator() *TrxIndexIterator {
	return &TrxIndexIterator{
		trxs: []*feed.TransactionInFeed{},
	}
}

func (it *MsgIndexIterator) Iterate(itemRaw btree.Item) bool {
	item := itemRaw.(UInt64MsgIndex)
	it.msgs = append(it.msgs, item.GetMessages()...)
	return true
}

func (it *MsgIndexIterator) GetMessages() []*feed.MessageInFeed {
	return it.msgs
}

func NewMsgIndexIterator() *MsgIndexIterator {
	return &MsgIndexIterator{
		msgs: []*feed.MessageInFeed{},
	}
}
