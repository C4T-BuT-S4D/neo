package queue

import (
	"context"
	"os/exec"
	"testing"
	"time"

	"github.com/c4t-but-s4d/neo/internal/models"
	"github.com/c4t-but-s4d/neo/pkg/testutils"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSimpleQueue_Add(t *testing.T) {
	makeQueue := func(s int) *simpleQueue {
		q := NewSimpleQueue(1).(*simpleQueue)
		q.c = make(chan *Job, s)
		return q
	}
	for _, tc := range []struct {
		q        *simpleQueue
		t        *Job
		wantTask *Job
		wantErr  error
	}{
		{
			q: makeQueue(100),
			t: &Job{
				Exploit:    &models.Exploit{ID: "kek"},
				Target:     &models.Target{ID: "t", IP: "i"},
				executable: "1",
			},
			wantErr: nil,
		},
		{
			q: makeQueue(1),
			t: &Job{
				Exploit:    &models.Exploit{ID: "kek"},
				Target:     &models.Target{ID: "t", IP: "i"},
				executable: "1",
			},
			wantErr: nil,
		},
		{
			q: makeQueue(0),
			t: &Job{
				Exploit:    &models.Exploit{ID: "kek"},
				Target:     &models.Target{ID: "t", IP: "i"},
				executable: "1",
			},
			wantErr: ErrQueueFull,
		},
	} {
		tc.t.logger = testutils.DummyJobLogger(tc.t.Exploit.ID, tc.t.Target.IP)
		err := tc.q.Add(tc.t)
		require.ErrorIs(t, err, tc.wantErr)
		if err != nil {
			continue
		}
		require.Equal(t, tc.t.executable, (<-tc.q.c).executable)
	}
}

func TestSimpleQueue_runExploit(t *testing.T) {
	for _, tc := range []struct {
		q        *simpleQueue
		t        *Job
		ctx      context.Context
		wantErr  error
		wantData []byte
	}{
		{
			q: NewSimpleQueue(1).(*simpleQueue),
			t: &Job{
				Exploit:    &models.Exploit{ID: "echo"},
				executable: "echo",
				Target:     &models.Target{ID: "id", IP: "ip"},
				timeout:    time.Second * 5,
			},
			wantErr:  nil,
			wantData: []byte("ip\n"),
			ctx:      context.Background(),
		},
		{
			q: NewSimpleQueue(1).(*simpleQueue),
			t: &Job{
				Exploit:    &models.Exploit{ID: "bad executable"},
				executable: "notfoundcli",
				Target:     &models.Target{ID: "id", IP: "ip"},
				timeout:    time.Second * 5,
			},
			wantErr:  exec.ErrNotFound,
			wantData: nil,
			ctx:      context.Background(),
		},
		{
			q: NewSimpleQueue(1).(*simpleQueue),
			t: &Job{
				Exploit:    &models.Exploit{ID: "cancelled ctx"},
				executable: "echo",
				Target:     &models.Target{ID: "id", IP: "ip"},
				timeout:    time.Second * 5,
			},
			wantErr:  context.Canceled,
			wantData: nil,
			ctx:      testutils.CanceledContext(),
		},
		{
			q: NewSimpleQueue(1).(*simpleQueue),
			t: &Job{
				Exploit:    &models.Exploit{ID: "timed out ctx"},
				executable: "echo",
				Target:     &models.Target{ID: "id", IP: "ip"},
				timeout:    time.Second * 5,
			},
			wantErr:  context.DeadlineExceeded,
			wantData: nil,
			ctx:      testutils.TimedOutContext(),
		},
		{
			q: NewSimpleQueue(1).(*simpleQueue),
			t: &Job{
				Exploit:    &models.Exploit{ID: "zero timeout"},
				executable: "echo",
				Target:     &models.Target{ID: "id", IP: "ip"},
				timeout:    time.Second * 0,
			},
			wantErr:  context.DeadlineExceeded,
			wantData: nil,
			ctx:      context.Background(),
		},
	} {
		tc.t.logger = testutils.DummyJobLogger(tc.t.Exploit.ID, tc.t.Target.IP)
		data, err := tc.q.runExploit(tc.ctx, tc.t)
		require.ErrorIs(t, err, tc.wantErr)

		if diff := cmp.Diff(data, tc.wantData, cmpopts.EquateEmpty()); diff != "" {
			t.Errorf("simpleQueue.runExploit() returned data mismatch (-want +got):\n%s", diff)
		}
	}
}

func TestSimpleQueue_Start(t *testing.T) {
	q := NewSimpleQueue(10)
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

	go q.Start(ctx)
	select {
	case <-time.After(time.Second * 3):
		t.Fatal("timeout")
	case out := <-q.Results():
		assert.Equal(t, task.Exploit, out.Exploit)
		assert.Equal(t, task.Target, out.Target)
		assert.Equal(t, task.Target.IP+"\n", string(out.Out))
	}
}
