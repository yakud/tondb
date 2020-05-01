package streaming

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

type Subscriber interface {
	Subscribe(*Client, Filter) (*Subscription, error)
	Unsubscribe(SubscriptionID) error
	IterateSubscriptions(iterator func(*Subscriptions) error) error
}

type (
	SubscriptionID string

	Subscription struct {
		id        SubscriptionID
		client    *Client
		filter    Filter
		abandoned int32
	}

	Subscriptions struct {
		filter Filter
		subs   []*Subscription
	}
)

type SubscriberImpl struct {
	subsByFilterHashRWMutex sync.RWMutex
	subBySubIdMutex         sync.Mutex

	subsByFilterHash map[FilterHash]*Subscriptions
	subBySubId       map[SubscriptionID]*Subscription
}

func (s *SubscriberImpl) Subscribe(client *Client, filter Filter) (*Subscription, error) {
	sub := &Subscription{
		id:     SubscriptionID(uuid.New().String()),
		client: client,
		filter: filter,
	}

	client.AddSubscription(sub)

	s.subBySubIdMutex.Lock()
	s.subBySubId[sub.id] = sub
	s.subBySubIdMutex.Unlock()

	s.subsByFilterHashRWMutex.Lock()
	if cli, ok := s.subsByFilterHash[filter.Hash()]; ok {
		cli.subs = append(cli.subs, sub)
		s.subsByFilterHash[filter.Hash()] = cli
	} else {
		s.subsByFilterHash[filter.Hash()] = &Subscriptions{filter: filter, subs: []*Subscription{sub}}
	}
	s.subsByFilterHashRWMutex.Unlock()

	return sub, nil
}

func (s *SubscriberImpl) Unsubscribe(id SubscriptionID) error {
	s.subBySubIdMutex.Lock()
	if sub, ok := s.subBySubId[id]; ok {
		sub.Abandon()
		delete(s.subBySubId, id)
	}
	s.subBySubIdMutex.Unlock()
	return nil
}

func (s *SubscriberImpl) IterateSubscriptions(iterator func(subscriptions *Subscriptions) error) error {
	s.subsByFilterHashRWMutex.RLock()
	for _, subscriptions := range s.subsByFilterHash {
		if err := iterator(subscriptions); err != nil {
			return err
		}
	}
	s.subsByFilterHashRWMutex.RUnlock()

	return nil
}

func (s *SubscriberImpl) GarbageCollection(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			s.collectGarbage()
		case <-ctx.Done():
			return
		}
	}
}

func (s *SubscriberImpl) collectGarbage() {
	s.subsByFilterHashRWMutex.Lock()
	defer s.subsByFilterHashRWMutex.Unlock()

	for filterHash, subs := range s.subsByFilterHash {
		newSubs := make([]*Subscription, 0, 8)

		for _, v := range subs.subs {
			if v == nil {
				continue
			}

			if !v.GetAbandoned() {
				newSubs = append(newSubs, v)
			} else {
				s.subBySubIdMutex.Lock()
				delete(s.subBySubId, v.id)
				s.subBySubIdMutex.Unlock()
			}
		}

		subs.subs = newSubs

		if len(newSubs) == 0 {
			delete(s.subsByFilterHash, filterHash)
		}
	}
}

func NewSubscriber() *SubscriberImpl {
	return &SubscriberImpl{
		subsByFilterHashRWMutex: sync.RWMutex{},
		subBySubIdMutex:         sync.Mutex{},

		subsByFilterHash: make(map[FilterHash]*Subscriptions, 8192),
		subBySubId:       make(map[SubscriptionID]*Subscription, 8192),
	}
}

func (s *Subscription) GetAbandoned() bool {
	if atomic.LoadInt32(&(s.abandoned)) != 0 {
		return true
	}
	return false
}

func (s *Subscription) Abandon() {
	atomic.StoreInt32(&(s.abandoned), 1)
}
