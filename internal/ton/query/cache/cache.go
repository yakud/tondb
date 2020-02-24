package cache

import (
	"context"
	"time"
)

type QueryUpdater interface{
	UpdateQuery() error
}

type Cache interface {
	Get(key string) (interface{}, error)
	Set(key string, value interface{}) error
}

type Ticker interface {
	RunTicker(ctx context.Context, interval time.Duration)
	AddQuery(q QueryUpdater)
	SetQueries(queries []QueryUpdater)
	GetQueries() []QueryUpdater
}
