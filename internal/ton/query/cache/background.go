package cache

import (
	"context"
	"database/sql"
	"sync"
	"time"
)

type Updater func(conn *sql.DB) interface{}

type Background struct {
	value   interface{}
	m       *sync.RWMutex
	updater Updater
}

func NewBackground(updater Updater, conn *sql.DB, dur time.Duration, ctx context.Context) *Background {
	bg := &Background{
		value:   nil,
		m:       &sync.RWMutex{},
		updater: updater,
	}

	bg.Set(bg.updater(conn))
	go bg.update(conn, time.NewTicker(dur), ctx)

	return bg
}

func (v *Background) update(conn *sql.DB, ticker *time.Ticker, ctx context.Context)  {
	for {
		select {
		case <-ticker.C:
			v.Set(v.updater(conn))
		case <-ctx.Done():
			return
		}
	}

}

func (v *Background) Set(value interface{}) {
	if value != nil {
		v.m.Lock()
		v.value = value
		v.m.Unlock()
	}
}

func (v *Background) Get() (interface{}, bool) {
	v.m.RLock()
	if v.value != nil {
		v.m.RUnlock()
		return v.value, true
	}

	v.m.RUnlock()
	return nil, false
}




