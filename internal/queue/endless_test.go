package queue

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

	"github.com/c4t-but-s4d/neo/internal/logger"
	"github.com/c4t-but-s4d/neo/internal/models"
	"github.com/c4t-but-s4d/neo/pkg/testutils"
)

func TestMain(m *testing.M) {
	logger.Init()
	logrus.SetLevel(logrus.DebugLevel)
	goleak.VerifyTestMain(m)
}

func Test_endlessQueue_Add(t *testing.T) {
	makeQueue := func(s int) *endlessQueue {
		q := NewEndlessQueue(1).(*endlessQueue)
		q.c = make(chan *Job, s)
		return q
	}
	for _, tc := range []struct {
		q        *endlessQueue
		t        *Job
		wantTask *Job
		wantErr  error
	}{
		{q: makeQueue(100), t: &Job{executable: "1"}, wantErr: nil},
		{q: makeQueue(1), t: &Job{executable: "1"}, wantErr: nil},
		{q: makeQueue(0), t: &Job{executable: "1"}, wantErr: ErrQueueFull},
	} {
		tc.t.logger = testutils.DummyJobLogger("1", "127.0.0.1")
		err := tc.q.Add(tc.t)
		require.ErrorIs(t, err, tc.wantErr)
		if err != nil {
			continue
		}
		require.Equal(t, tc.t.executable, (<-tc.q.c).executable)
	}
}

func Test_endlessQueue_Start(t *testing.T) {
	q := NewEndlessQueue(10)
	task := &Job{
		Exploit: &models.Exploit{ID: "kek"},
		Target: &models.Target{
			ID: "id",
			IP: "ip",
		},
		executable: "echo",
		dir:        "",
		timeout:    time.Second * 2,
		logger:     testutils.DummyJobLogger("echo", "ip"),
	}
	require.NoError(t, q.Add(task))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		q.Start(ctx)
	}()

	select {
	case <-time.After(time.Second * 3):
		t.Fatal("timeout")
	case out := <-q.Results():
		assert.Equal(t, task.Exploit, out.Exploit)
		assert.Equal(t, task.Target, out.Target)
		assert.Equal(t, task.Target.IP, string(out.Out))
		break
	}

	cancel()
	wg.Wait()
}
