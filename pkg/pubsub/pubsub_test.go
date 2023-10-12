package pubsub

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

	"github.com/c4t-but-s4d/neo/v2/pkg/neosync"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestPubSub_single(t *testing.T) {
	p := NewPubSub[string]()

	wg := sync.WaitGroup{}
	wg.Add(1)

	sub := p.Subscribe(func(msg string) error {
		require.Equal(t, "blah-blah", msg)
		wg.Done()
		return nil
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go sub.Run(ctx)

	p.Publish("blah-blah")
	wg.Wait()
}

func TestPubSub_nonBlockPublish(t *testing.T) {
	p := NewPubSub[string]()

	wg := sync.WaitGroup{}
	wg.Add(11)

	sub := p.Subscribe(func(msg string) error {
		require.Equal(t, "pew-pew", msg)
		wg.Done()
		return nil
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go sub.Run(ctx)

	done := make(chan struct{})
	go func() {
		for i := 0; i < 11; i++ {
			p.Publish("pew-pew")
		}
		close(done)
	}()

	select {
	case <-time.After(1 * time.Second):
		t.Fatal("publish method must not be blocked")
	case <-done:
	}

	wg.Wait()
}

func TestPubSub_multipleSubscribers(t *testing.T) {
	p := NewPubSub[string]()

	wgFirst := sync.WaitGroup{}
	wgFirst.Add(1)
	wgSecond := sync.WaitGroup{}
	wgSecond.Add(1)

	sub1 := p.Subscribe(func(msg string) error {
		require.Equal(t, "blah", msg)
		wgFirst.Done()
		return nil
	})
	ctx1, cancel1 := context.WithCancel(context.Background())
	defer cancel1()
	go sub1.Run(ctx1)

	sub2 := p.Subscribe(func(msg string) error {
		require.Equal(t, "blah", msg)
		wgSecond.Done()
		return nil
	})
	ctx2, cancel2 := context.WithCancel(context.Background())
	defer cancel2()
	go sub2.Run(ctx2)

	p.Publish("blah")

	wgFirst.Wait()
	wgSecond.Wait()
}

func TestPubSub_slowpoke(t *testing.T) {
	p := NewPubSub[string]()

	const samples = 100

	wgSlow := sync.WaitGroup{}
	wgSlow.Add(samples)
	slowCtx, slowCancel := context.WithCancel(context.Background())
	defer func() {
		slowCancel()
		wgSlow.Wait()
	}()

	slowSub := p.Subscribe(func(msg string) error {
		defer wgSlow.Done()

		select {
		case <-slowCtx.Done():
			return nil
		default:
			time.Sleep(1 * time.Second)
		}
		return nil
	})
	go slowSub.Run(slowCtx)

	fastWg := sync.WaitGroup{}
	fastWg.Add(samples)

	fastSub := p.Subscribe(func(msg string) error {
		require.Equal(t, "pew-pew", msg)
		fastWg.Done()
		return nil
	})
	fastCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go fastSub.Run(fastCtx)

	for i := 0; i < samples; i++ {
		p.Publish("pew-pew")
	}

	select {
	case <-time.After(1 * time.Second):
		t.Fatal("publish blocks on slowpoke?")
	case <-neosync.AwaitWG(&fastWg):
		// ok
	}
}

func TestPubSub_unsubscribe(t *testing.T) {
	p := NewPubSub[string]()

	sub1 := p.Subscribe(func(msg string) error {
		t.Error("first subscriber must not be called")
		return nil
	})
	ctx1, cancel1 := context.WithCancel(context.Background())
	defer cancel1()
	go sub1.Run(ctx1)

	p.Unsubscribe(sub1)

	wg := sync.WaitGroup{}
	wg.Add(1)

	sub2 := p.Subscribe(func(msg string) error {
		require.Equal(t, "pew-pew", msg)
		wg.Done()
		return nil
	})
	ctx2, cancel2 := context.WithCancel(context.Background())
	defer cancel2()
	go sub2.Run(ctx2)

	p.Publish("pew-pew")

	wg.Wait()
}

func TestPubSub_sequencePublishers(t *testing.T) {
	p := NewPubSub[string]()

	wg := sync.WaitGroup{}
	wg.Add(10)

	sub := p.Subscribe(func(msg string) error {
		require.Equal(t, "pew-pew", msg)
		wg.Done()
		return nil
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go sub.Run(ctx)

	for i := 0; i < 10; i++ {
		p.Publish("pew-pew")
	}

	wg.Wait()
}

func TestPubSub_concurrentPublishers(t *testing.T) {
	p := NewPubSub[string]()

	wg := sync.WaitGroup{}
	wg.Add(10)

	sub := p.Subscribe(func(msg string) error {
		require.Equal(t, "pew-pew", msg)
		wg.Done()
		return nil
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go sub.Run(ctx)

	for i := 0; i < 10; i++ {
		go p.Publish("pew-pew")
	}

	wg.Wait()
}

func TestPubSub_msgOrder(t *testing.T) {
	p := NewPubSub[uint64]()

	wg := sync.WaitGroup{}
	wg.Add(15)

	c := uint64(0)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sub := p.Subscribe(func(msg uint64) error {
		expected := atomic.AddUint64(&c, 1)
		require.Equal(t, expected, msg)
		wg.Done()
		return nil
	})
	go sub.Run(ctx)

	for i := uint64(1); i < 11; i++ {
		if i == 6 {
			c := uint64(5)
			sub := p.Subscribe(func(msg uint64) error {
				expected := atomic.AddUint64(&c, 1)
				require.Equal(t, expected, msg)
				wg.Done()
				return nil
			})
			go sub.Run(ctx)
		}

		p.Publish(i)
	}

	wg.Wait()
}
