package writer

import (
	"sync"

	"gitlab.flora.loc/mills/tondb/internal/ton"
)

// Bulk collect multiply blocks for bulked insert
type Bulk struct {
	closed bool
	mutex  *sync.Mutex
	blocks []*ton.Block
}

func (t *Bulk) Add(blocks ...*ton.Block) {
	t.blocks = append(t.blocks, blocks...)
}

func (t *Bulk) Length() int {
	return len(t.blocks)
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
		mutex:  &sync.Mutex{},
		blocks: make([]*ton.Block, 0, size),
	}
}
