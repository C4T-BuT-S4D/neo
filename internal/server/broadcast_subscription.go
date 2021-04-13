package server

import (
	"context"
	"sync"

	"github.com/sirupsen/logrus"

	neopb "neo/lib/genproto/neo"
)

type cmdHandler func(command *neopb.Command) error

type broadcastSubscription struct {
	id     string
	queue  []*neopb.Command
	mu     sync.Mutex
	notify chan struct{}
	h      cmdHandler
}

func newBroadcastSubscription(id string, onMsg cmdHandler) *broadcastSubscription {
	return &broadcastSubscription{
		id:     id,
		queue:  nil,
		notify: make(chan struct{}, 1),
		h:      onMsg,
	}
}

func (s *broadcastSubscription) Run(ctx context.Context) {
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

				logrus.Debugf("Handling command %v in subscription %s", cmd, s.id)
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

func (s *broadcastSubscription) Push(cmd *neopb.Command) {
	s.mu.Lock()
	s.queue = append(s.queue, cmd)
	s.mu.Unlock()

	select {
	case s.notify <- struct{}{}:
	default:
	}
}
