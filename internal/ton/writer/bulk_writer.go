package writer

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"gitlab.flora.loc/mills/tondb/internal/ton"
	"gitlab.flora.loc/mills/tondb/internal/ton/storage"
)

type BulkWriter struct {
	blocksStorage       *storage.Blocks
	transactionsStorage *storage.Transactions
	shardsDescrStorage  *storage.ShardsDescr
	accountStateStorage *storage.AccountState

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

			if err := t.handleBulk(bulk, onHandle); err != nil {
				return err
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

func (t *BulkWriter) handleBulk(bulk *Bulk, onHandle func(b *Bulk, delay time.Duration)) error {
	if atomic.LoadInt32(&t.isWorking) == 0 {
		atomic.StoreInt32(&t.isWorking, 1)
	}

	start := time.Now()

	blocksInsert := make([]*ton.Block, 0, bulk.Length())
	transactionsInsert := make([]*ton.Transaction, 0, bulk.LengthTransactions())
	shardsDescr := make([]*ton.ShardDescr, 0, bulk.LengthShardDescr())
	accountState := make([]*ton.AccountState, 0, bulk.LengthAccountStates())

	for _, block := range bulk.Blocks() {
		blocksInsert = append(blocksInsert, block)
		transactionsInsert = append(transactionsInsert, block.Transactions...)
		shardsDescr = append(shardsDescr, block.ShardDescr...)
	}
	accountState = append(accountState, bulk.AccountStates()...)

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

	if len(accountState) > 0 {
		if err := t.accountStateStorage.InsertMany(accountState); err != nil {
			// @TODO: retry, no return
			return errors.WithStack(err)
		}
	}

	if onHandle != nil {
		onHandle(bulk, time.Now().Sub(start))
	}

	return nil
}

func NewBulkWriter(
	blocksStorage *storage.Blocks,
	transactionsStorage *storage.Transactions,
	shardsDescrStorage *storage.ShardsDescr,
	accountStateStorage *storage.AccountState,
	in chan *Bulk,
) *BulkWriter {
	return &BulkWriter{
		blocksStorage:       blocksStorage,
		transactionsStorage: transactionsStorage,
		shardsDescrStorage:  shardsDescrStorage,
		accountStateStorage: accountStateStorage,
		in:                  in,
	}
}
