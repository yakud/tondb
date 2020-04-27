package streaming_new

type Subscriber interface {
	Subscribe(*Client, Filter) (*Subscription, error)
	Unsubscribe(SubscriptionID) error
	Subscriptions() map[FilterHash]*Subscriptions
}

type (
	SubscriptionID string

	Subscription struct {
		id     SubscriptionID
		client *Client
		filter Filter
	}

	Subscriptions struct {
		filter Filter
		subs   []*Subscription
	}
)

type SubscriberImpl struct {
	clients map[FilterHash][]*Subscription
}

func (s *SubscriberImpl) Subscribe(client *Client, filter Filter) (*Subscription, error) {
	sub := &Subscription{
		id:     "", // todo generate
		client: client,
		filter: filter,
	}
	s.clients[filter.Hash()] = sub // TODO: append...
	panic("implement me")
}

func (s *SubscriberImpl) Unsubscribe(SubscriptionID) error {
	panic("implement me")
}

func (s *SubscriberImpl) Subscriptions() map[FilterHash]*Subscriptions {
	panic("implement me")
}
