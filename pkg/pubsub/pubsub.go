package pubsub

import "sync"

type PubSub[T any] struct {
	subs map[string]*Subscription[T]
	mu   sync.RWMutex
}

func NewPubSub[T any]() *PubSub[T] {
	return &PubSub[T]{
		subs: make(map[string]*Subscription[T]),
	}
}

func (ps *PubSub[T]) Publish(msg T) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	for _, sub := range ps.subs {
		sub.Push(msg)
	}
}

func (ps *PubSub[T]) Subscribe(onMsg MessageHandler[T]) *Subscription[T] {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	sub := NewSubscription(onMsg)
	ps.subs[sub.GetID()] = sub
	return sub
}

func (ps *PubSub[T]) Unsubscribe(sub *Subscription[T]) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	delete(ps.subs, sub.GetID())
}
