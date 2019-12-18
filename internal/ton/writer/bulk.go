package writer

import (
	"sync"

	"gitlab.flora.loc/mills/tondb/internal/ton"
)

// Bulk collect multiply blocks for bulked insert
type Bulk struct {
	closed        bool
	mutex         *sync.Mutex
	blocks        []*ton.Block
	accountStates []*ton.AccountState
}

func (t *Bulk) AddBlocks(blocks ...*ton.Block) {
	t.blocks = append(t.blocks, blocks...)
}

func (t *Bulk) AddAccountState(state ...*ton.AccountState) {
	t.accountStates = append(t.accountStates, state...)
}

func (t *Bulk) Length() int {
	return len(t.blocks)
}

func (t *Bulk) LengthAccountStates() int {
	return len(t.accountStates)
}

func (t *Bulk) LengthTransactions() int {
	totalLength := 0
	for _, b := range t.blocks {
		totalLength += len(b.Transactions)
	}

	return totalLength
}

func (t *Bulk) LengthShardDescr() int {
	totalLength := 0
	for _, b := range t.blocks {
		totalLength += len(b.ShardDescr)
	}

	return totalLength
}

func (t *Bulk) Blocks() []*ton.Block {
	return t.blocks
}

func (t *Bulk) AccountStates() []*ton.AccountState {
	return t.accountStates
}

func (t *Bulk) Close() {
	t.closed = true
}

func (t *Bulk) Open() {
	t.closed = false
}

func (t *Bulk) IsOpen() bool {
	return t.closed == false
}

func (t *Bulk) Mutex() *sync.Mutex {
	return t.mutex
}

func NewBulk(size int) *Bulk {
	return &Bulk{
		mutex:         &sync.Mutex{},
		blocks:        make([]*ton.Block, 0, size),
		accountStates: make([]*ton.AccountState, 0, size),
	}
}
