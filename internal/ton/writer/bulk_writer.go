package writer

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"github.com/yakud/ton-blocks-stream-receiver/internal/ton"
	"github.com/yakud/ton-blocks-stream-receiver/internal/ton/storage"
)

type BulkWriter struct {
	blocksStorage       *storage.Blocks
	transactionsStorage *storage.Transactions
	shardsDescrStorage  *storage.ShardsDescr

	in chan *Bulk

	isWorking int32
}

func (t *BulkWriter) IsWorking() bool {
	return atomic.LoadInt32(&t.isWorking) == 1
}

func (t *BulkWriter) Run(ctx context.Context, wg *sync.WaitGroup, onHandle func(b *Bulk, delay time.Duration)) error {
	defer wg.Done()

	for {
		select {
		case bulk, ok := <-t.in:
			if !ok {
				return nil
			}

			if atomic.LoadInt32(&t.isWorking) == 0 {
				atomic.StoreInt32(&t.isWorking, 1)
			}

			start := time.Now()

			blocksInsert := make([]*ton.Block, 0, bulk.Length())
			transactionsInsert := make([]*ton.Transaction, 0, bulk.LengthTransactions())
			shardsDescr := make([]*ton.ShardDescr, 0, bulk.LengthShardDescr())

			for _, block := range bulk.Blocks() {
				blocksInsert = append(blocksInsert, block)
				transactionsInsert = append(transactionsInsert, block.Transactions...)
				shardsDescr = append(shardsDescr, block.ShardDescr...)
			}

			if len(blocksInsert) > 0 {
				if err := t.blocksStorage.InsertMany(blocksInsert); err != nil {
					// @TODO: retry, no return
					return errors.WithStack(err)
				}
			}

			if len(transactionsInsert) > 0 {
				if err := t.transactionsStorage.InsertMany(transactionsInsert); err != nil {
					// @TODO: retry, no return
					return errors.WithStack(err)
				}
			}

			if len(shardsDescr) > 0 {
				if err := t.shardsDescrStorage.InsertMany(shardsDescr); err != nil {
					// @TODO: retry, no return
					return errors.WithStack(err)
				}
			}

			if onHandle != nil {
				onHandle(bulk, time.Now().Sub(start))
			}

		case <-ctx.Done():
			return nil

		default:
			<-time.After(time.Millisecond * 100)
			if atomic.LoadInt32(&t.isWorking) == 1 {
				atomic.StoreInt32(&t.isWorking, 0)
			}
		}
	}
}

func NewBulkWriter(
	blocksStorage *storage.Blocks,
	transactionsStorage *storage.Transactions,
	shardsDescrStorage *storage.ShardsDescr,
	in chan *Bulk,
) *BulkWriter {
	return &BulkWriter{
		blocksStorage:       blocksStorage,
		transactionsStorage: transactionsStorage,
		shardsDescrStorage:  shardsDescrStorage,
		in:                  in,
	}
}
