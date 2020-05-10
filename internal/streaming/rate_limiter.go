package streaming

import (
	"errors"
	"sync"
)

type RateLimiter struct {
	sync.RWMutex
	counters map[string]uint32
	limit    uint32
}

func (l *RateLimiter) CheckLimit(key string) error {
	l.RLock()
	defer l.RUnlock()

	if clientsCount, ok := l.counters[key]; ok {
		if clientsCount >= l.limit {
			return errors.New("rate limit exceeded")
		}
	}

	return nil
}

func (l *RateLimiter) CheckLimitAndInc(key string) error {
	l.Lock()
	defer l.Unlock()

	if clientsCount, ok := l.counters[key]; ok {
		if clientsCount >= l.limit {
			return errors.New("rate limit exceeded")
		}

		l.counters[key] = clientsCount + 1
	} else {
		l.counters[key] = 1
	}

	return nil
}

func (l *RateLimiter) Dec(key string) {
	l.Lock()
	if cnt, ok := l.counters[key]; ok {
		l.counters[key] = cnt - 1
	}
	l.Unlock()
}

func NewRateLimiter(limit uint32) *RateLimiter {
	return &RateLimiter{
		RWMutex:  sync.RWMutex{},
		counters: make(map[string]uint32, 32768),
		limit:    limit,
	}
}

