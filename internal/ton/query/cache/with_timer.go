package cache

import (
	"sync"
	"time"
)

type WithTimer struct {
	value interface{}
	m     *sync.RWMutex

	dur        time.Duration
	lastUpdate time.Time
}

func NewWithTimer(dur time.Duration) *WithTimer {
	return &WithTimer{
		value: nil,
		dur:   dur,
		m:     &sync.RWMutex{},
	}
}

func (v *WithTimer) Set(value interface{}) {
	v.m.Lock()
	v.value = value
	v.lastUpdate = time.Now()
	v.m.Unlock()
}

func (v *WithTimer) Get() (interface{}, bool) {
	v.m.RLock()
	if v.value != nil && time.Now().Sub(v.lastUpdate) <= v.dur {
		v.m.RUnlock()
		return v.value, true
	}

	v.m.RUnlock()
	return nil, false
}
