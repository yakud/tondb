package writer

import (
	"context"
	"log"
	"sync"
	"time"

	"gitlab.flora.loc/mills/tondb/internal/ton"
)

const DefaultBulkSize = 5000
const DefaultBulkTimeout = time.Second * 2

// IS NOT SAFE FOR CONCURRENCY!!!
type BulkBuffer struct {
	bulkSize    int
	bulkTimeout time.Duration

	bulk *Bulk
	out  chan *Bulk

	lastOp  time.Time
	firstOp time.Time
}

func (t *BulkBuffer) SetBulkSize(bulkSize int) {
	t.bulkSize = bulkSize
}

func (t *BulkBuffer) SetBulkTimeout(bulkTimeout time.Duration) {
	t.bulkTimeout = bulkTimeout
}

func (t *BulkBuffer) Out() chan *Bulk {
	return t.out
}

func (t *BulkBuffer) AddBlock(block *ton.Block) error {
	bulk := t.bulk

	bulk.Mutex().Lock()
	if !bulk.IsOpen() {
		bulk.Mutex().Unlock()
		return t.AddBlock(block)
	}

	bulk.AddBlocks(block)

	t.lastOp = time.Now()
	if t.firstOp.IsZero() {
		t.firstOp = t.lastOp
	}

	if bulk.Length() >= t.bulkSize {
		t.Reduce(bulk)
	}
	bulk.Mutex().Unlock()

	return nil
}
func (t *BulkBuffer) AddState(state *ton.AccountState) error {
	bulk := t.bulk

	bulk.Mutex().Lock()
	if !bulk.IsOpen() {
		bulk.Mutex().Unlock()
		return t.AddState(state)
	}

	bulk.AddAccountState(state)

	t.lastOp = time.Now()
	if t.firstOp.IsZero() {
		t.firstOp = t.lastOp
	}

	if bulk.Length() >= t.bulkSize {
		t.Reduce(bulk)
	}
	bulk.Mutex().Unlock()

	return nil
}

func (t *BulkBuffer) Reduce(b *Bulk) error {
	log.Printf("Reduce bulk: blocks: %d; transactions: %d; states: %d", b.Length(), b.LengthTransactions(), b.LengthAccountStates())

	b.Close()
	t.out <- b
	t.bulk = NewBulk(t.bulkSize)
	t.firstOp = time.Time{}

	return nil
}

func (t *BulkBuffer) Timeout(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	var lastBulkLenght = 0

	percentFromMaxSize := 0.05
	ticker := time.NewTicker(t.bulkTimeout)
	for {
		select {
		case <-ticker.C:
			// We can reduce bulk not often then bulkTimeout
			if t.firstOp.IsZero() || time.Now().Sub(t.firstOp) < t.bulkTimeout {
				continue
			}

			currentBulkLength := t.bulk.Length()

			// if from last check bulk length grow more N% then waiting
			if float64(currentBulkLength-lastBulkLenght) >= float64(t.bulkSize)*percentFromMaxSize {
				lastBulkLenght = currentBulkLength
				continue
			}

			lastBulkLenght = 0
			t.CheckReduce()

		case <-ctx.Done():
			t.CheckReduce()
			return
		}
	}
}

func (t *BulkBuffer) CheckReduce() {
	bulk := t.bulk
	bulk.Mutex().Lock()
	if bulk.IsOpen() && bulk.Length() > 0 {
		t.Reduce(bulk)
	}
	bulk.Mutex().Unlock()
}

func NewBulkBuffer(out chan *Bulk) *BulkBuffer {
	return &BulkBuffer{
		out:         out,
		bulk:        NewBulk(DefaultBulkSize),
		bulkSize:    DefaultBulkSize,
		bulkTimeout: DefaultBulkTimeout,
	}
}
