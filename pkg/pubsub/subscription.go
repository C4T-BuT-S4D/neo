package pubsub

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type MessageHandler[T any] func(T) error

type Subscription[T any] struct {
	id     string
	queue  []T
	mu     sync.Mutex
	notify chan struct{}
	h      MessageHandler[T]
}

func NewSubscription[T any](onMsg MessageHandler[T]) *Subscription[T] {
	id := uuid.NewString()
	return &Subscription[T]{
		id:     id,
		queue:  nil,
		notify: make(chan struct{}, 1),
		h:      onMsg,
	}
}

func (s *Subscription[T]) Run(ctx context.Context) {
	for {
		select {
		case _, ok := <-s.notify:
			if !ok {
				return
			}
			for {
				s.mu.Lock()
				if len(s.queue) == 0 {
					s.mu.Unlock()
					break
				}
				cmd := s.queue[0]
				s.queue = s.queue[1:]
				s.mu.Unlock()

				logrus.Debugf("Handling message %v in subscription %s", cmd, s.id)
				if err := s.h(cmd); err != nil {
					logrus.Errorf("Error in subscription message handler: %v", err)
					continue
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

func (s *Subscription[T]) Push(msg T) {
	s.mu.Lock()
	s.queue = append(s.queue, msg)
	s.mu.Unlock()

	select {
	case s.notify <- struct{}{}:
	default:
	}
}

func (s *Subscription[T]) GetID() string {
	return s.id
}
