package pubsub

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewSubscription(t *testing.T) {
	sub := NewSubscription[int](nil)
	newSub := NewSubscription[int](nil)
	require.NotEqual(t, sub.GetID(), newSub.GetID(), "id is not random")
}

func Test_subscription_Push(t *testing.T) {
	sub := &Subscription[string]{
		id:     "test",
		queue:  nil,
		notify: make(chan struct{}),
		h:      nil,
	}
	sub.Push("test")
	require.Equal(t, []string{"test"}, sub.queue)
}

func Test_subscription_Run(t *testing.T) {
	var received string
	signal := make(chan struct{})
	handler := func(msg string) error {
		received = msg
		close(signal)
		return nil
	}
	sub := &Subscription[string]{
		id:     "test",
		queue:  nil,
		notify: make(chan struct{}, 1),
		h:      handler,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go sub.Run(ctx)

	const msg = "message"
	sub.Push(msg)

	select {
	case <-signal:
		break
	case <-time.After(time.Millisecond * 100):
		t.Errorf("Handler was not called in time")
	}

	require.Equal(t, msg, received)
}
