package cache

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

type Background struct {
	values  *sync.Map
	queries []QueryUpdater
}

func NewBackground() *Background {
	return &Background{
		values:  &sync.Map{},
		queries: make([]QueryUpdater, 0, 10),
	}
}

func (v *Background) Set(key string, value interface{}) error {
	if value != nil {
		v.values.Store(key, value)
		return nil
	}

	return errors.New("value is empty")
}

func (v *Background) Get(key string) (interface{}, error) {
	if value, ok := v.values.Load(key); ok {
		return value, nil
	} else {
		return nil, fmt.Errorf("couldn't get value for key %s", key)
	}
}

func (v *Background) RunTicker(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			for _, updater := range v.queries {
				if err := updater.UpdateQuery(); err != nil {
					log.Println("Got an error while updating cache with query. err: ", err)
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

func (v *Background) AddQuery(updater QueryUpdater) {
	v.queries = append(v.queries, updater)
}

func (v *Background) SetQueries(queries []QueryUpdater) {
	v.queries = queries
}

func (v *Background) GetQueries() []QueryUpdater {
	return v.queries
}




