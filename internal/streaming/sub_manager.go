package streaming

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/panjf2000/ants/v2"
	"gitlab.flora.loc/mills/tondb/internal/ton"
	"sync"
	"time"
)

type SubManager struct {
	subs map[Filter][]*SubHandler
	rw   sync.RWMutex
	pool *ants.Pool
}

func (m *SubManager) HandleBlock(block *ton.Block) error {
	blockJson, err := json.Marshal(block)
	if err != nil {
		return err
	}
	transactionsJson, err := json.Marshal(block.Transactions)
	if err != nil {
		return err
	}

	messages := make([]*ton.TransactionMessage, 0, 64)
	for _, trx := range block.Transactions {
		if trx.InMsg != nil {
			messages = append(messages, trx.InMsg)
		}
		if trx.OutMsgs != nil {
			messages = append(messages, trx.OutMsgs...)
		}
	}

	messagesJson, err := json.Marshal(messages)
	if err != nil {
		return err
	}

	m.rw.RLock()
	for filter, subs := range m.subs {
		if filter.Match(block) {
			for _, sub := range subs {
				if sub != nil && !sub.Abandoned {
					switch sub.Sub.Filter.FeedName {
					case "blocks":
						return m.poolSubmit(sub, blockJson)
					case "transactions":
						return m.poolSubmit(sub, transactionsJson)
					case "messages":
						return m.poolSubmit(sub, messagesJson)
					}
				}
			}
		}
	}
	m.rw.RUnlock()

	return nil
}

func (m *SubManager) Add(handler *SubHandler) {
	m.rw.Lock()
	defer m.rw.Unlock()

	subs, ok := m.subs[handler.Sub.Filter]
	if !ok {
		subs = make([]*SubHandler, 0, 8)
	}

	subs = append(subs, handler)
	m.subs[handler.Sub.Filter] = subs
}

func (m *SubManager) Get(key Filter) (subs []*SubHandler, ok bool) {
	m.rw.RLock()
	defer m.rw.RUnlock()

	subs, ok = m.subs[key]
	return
}

func (m *SubManager) GarbageCollection(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			m.collectGarbage()
		case <-ctx.Done():
			return
		}
	}
}

func (m *SubManager) collectGarbage() {
	m.rw.Lock()
	defer m.rw.Unlock()

	for key, subs := range m.subs {
		newSubs := make([]*SubHandler, 0, 8)

		for _, v := range subs {
			if v != nil && !v.Abandoned {
				newSubs = append(newSubs, v)
			}
		}

		if len(newSubs) == 0 {
			delete(m.subs, key)
		} else {
			m.subs[key] = newSubs
		}
	}
}

func (m *SubManager) poolSubmit(sub *SubHandler, res []byte) error {
	if m.pool == nil {
		return fmt.Errorf("sub_manager goroutine pool not initialized")
	}

	var err error
	err = m.pool.Submit(func() {
		err = sub.Handle(res)
	})

	return err
}

func NewSubManager(pool *ants.Pool) *SubManager {
	return &SubManager{
		subs: make(map[Filter][]*SubHandler),
		rw:   sync.RWMutex{},
		pool: pool,
	}
}

