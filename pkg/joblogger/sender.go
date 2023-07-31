package joblogger

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"neo/internal/client"
	neopb "neo/lib/genproto/neo"
)

const (
	sendInterval = time.Second
)

type Sender interface {
	Add(lines ...*neopb.LogLine)
}

func NewDummySender() *DummySender {
	return &DummySender{}
}

type DummySender struct{}

func (s *DummySender) Add(...*neopb.LogLine) {
}

func NewRemoteSender(client *client.Client) *RemoteSender {
	return &RemoteSender{
		client: client,
		queue:  make([]*neopb.LogLine, 0, 1000),
	}
}

type RemoteSender struct {
	client *client.Client
	queue  []*neopb.LogLine
	mu     sync.Mutex
}

func (s *RemoteSender) Add(lines ...*neopb.LogLine) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.queue = append(s.queue, lines...)
}

func (s *RemoteSender) Start(ctx context.Context) {
	ticker := time.NewTicker(sendInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := s.send(ctx); err != nil {
				logrus.Errorf("Error sending logs: %v", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (s *RemoteSender) send(ctx context.Context) error {
	s.mu.Lock()
	batch := make([]*neopb.LogLine, len(s.queue))
	copy(batch, s.queue)
	s.queue = s.queue[:0]
	s.mu.Unlock()

	if len(batch) == 0 {
		logrus.Debugf("Sending %d logs", len(batch))
	} else {
		logrus.Infof("Sending %d logs", len(batch))
	}
	if err := s.client.AddLogLines(ctx, batch...); err != nil {
		return fmt.Errorf("sending batch to server: %w", err)
	}
	return nil
}
