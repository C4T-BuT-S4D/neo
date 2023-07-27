package queue

import (
	"context"
	"os/exec"
	"testing"
	"time"

	"neo/pkg/testutils"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSimpleQueue_Add(t *testing.T) {
	makeQueue := func(s int) *simpleQueue {
		q := NewSimpleQueue(1).(*simpleQueue)
		q.c = make(chan *Task, s)
		return q
	}
	for _, tc := range []struct {
		q        *simpleQueue
		t        *Task
		wantTask *Task
		wantErr  error
	}{
		{q: makeQueue(100), t: &Task{executable: "1"}, wantErr: nil},
		{q: makeQueue(1), t: &Task{executable: "1"}, wantErr: nil},
		{q: makeQueue(0), t: &Task{executable: "1"}, wantErr: ErrQueueFull},
	} {
		tc.t.logger = testutils.DummyTaskLogger(tc.t.name, tc.t.teamIP)
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
		t        *Task
		ctx      context.Context
		wantErr  error
		wantData []byte
	}{
		{
			q:        NewSimpleQueue(1).(*simpleQueue),
			t:        &Task{name: "echo", executable: "echo", teamID: "id", teamIP: "ip", timeout: time.Second * 5},
			wantErr:  nil,
			wantData: []byte("ip\n"),
			ctx:      context.Background(),
		},
		{
			q:        NewSimpleQueue(1).(*simpleQueue),
			t:        &Task{name: "bad executable", executable: "notfoundcli", teamID: "id", teamIP: "ip", timeout: time.Second * 5},
			wantErr:  exec.ErrNotFound,
			wantData: nil,
			ctx:      context.Background(),
		},
		{
			q:        NewSimpleQueue(1).(*simpleQueue),
			t:        &Task{name: "cancelled ctx", executable: "echo", teamID: "id", teamIP: "ip", timeout: time.Second * 5},
			wantErr:  context.Canceled,
			wantData: nil,
			ctx:      testutils.CanceledContext(),
		},
		{
			q:        NewSimpleQueue(1).(*simpleQueue),
			t:        &Task{name: "timed out ctx", executable: "echo", teamID: "id", teamIP: "ip", timeout: time.Second * 5},
			wantErr:  context.DeadlineExceeded,
			wantData: nil,
			ctx:      testutils.TimedOutContext(),
		},
		{
			q:        NewSimpleQueue(1).(*simpleQueue),
			t:        &Task{name: "zero timeout", executable: "echo", teamID: "id", teamIP: "ip", timeout: time.Second * 0},
			wantErr:  context.DeadlineExceeded,
			wantData: nil,
			ctx:      context.Background(),
		},
	} {
		tc.t.logger = testutils.DummyTaskLogger(tc.t.name, tc.t.teamIP)
		data, err := tc.q.runExploit(tc.ctx, tc.t)
		require.ErrorIs(t, err, tc.wantErr)

		if diff := cmp.Diff(data, tc.wantData, cmpopts.EquateEmpty()); diff != "" {
			t.Errorf("simpleQueue.runExploit() returned data mismatch (-want +got):\n%s", diff)
		}
	}
}

func TestSimpleQueue_Start(t *testing.T) {
	q := NewSimpleQueue(10)
	task := &Task{
		name:       "kek",
		executable: "echo",
		dir:        "",
		teamID:     "id",
		teamIP:     "ip",
		timeout:    time.Second * 2,
		logger:     testutils.DummyTaskLogger("echo", "ip"),
	}
	require.NoError(t, q.Add(task))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go q.Start(ctx)
	select {
	case <-time.After(time.Second * 3):
		t.Fatal("timeout")
	case out := <-q.Results():
		assert.Equal(t, task.name, out.Name)
		assert.Equal(t, task.teamID, out.Team)
		assert.Equal(t, task.teamIP+"\n", string(out.Out))
	}
}
