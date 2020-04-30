package streaming

import (
	"context"
	"sync"
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
		abandoned bool
	}

	Subscriptions struct {
		filter Filter
		subs   []*Subscription
	}
)

type SubscriberImpl struct {
	sync.RWMutex
	subsByFilterHash map[FilterHash]*Subscriptions
	subBySubId       map[SubscriptionID]*Subscription
}

func (s *SubscriberImpl) Subscribe(client *Client, filter Filter) (*Subscription, error) {
	sub := &Subscription{
		id:     SubscriptionID(uuid.New().String()),
		client: client,
		filter: filter,
	}

	s.Lock()
	if cli, ok := s.subsByFilterHash[filter.Hash()]; ok {
		cli.subs = append(cli.subs, sub)
	}
	s.Unlock()

	return sub, nil
}

func (s *SubscriberImpl) Unsubscribe(id SubscriptionID) error {
	// todo: mutex
	if sub, ok := s.subBySubId[id]; ok {
		sub.abandoned = true
		delete(s.subBySubId, id)
	}
	return nil
}

func (s *SubscriberImpl) IterateSubscriptions(iterator func(subscriptions *Subscriptions) error) error {
	s.RLock()
	for _, subscriptions := range s.subsByFilterHash {
		if err := iterator(subscriptions); err != nil {
			return err
		}
	}
	s.RUnlock()

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
	s.Lock()
	defer s.Unlock()

	for filterHash, subs := range s.subsByFilterHash {
		newSubs := make([]*Subscription, 0, 8)

		for _, v := range subs.subs {
			if v != nil {
				if v.abandoned {
					// TODO: How and when do we need to close client connection and channel?
					// as long as we one client can have many subs we need to check if it is last his sub and if it is,
					// we need to close conn (If we need to close conn at all)
					// if we need to close a con it would be helpful to use map[ClientID]int that contains noumber of subs for each client
					//if err := v.client.conn.Close(); err != nil {
					//	// TODO: handle error
					//}
				} else {
					newSubs = append(newSubs, v)
				}
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
		RWMutex:          sync.RWMutex{},
		subsByFilterHash: make(map[FilterHash]*Subscriptions, 8192),
		subBySubId:       make(map[SubscriptionID]*Subscription, 8192),
	}
}
