package queue

import (
	"context"
	"testing"
	"time"

	"neo/pkg/testutils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_endlessQueue_Add(t *testing.T) {
	makeQueue := func(s int) *endlessQueue {
		q := NewEndlessQueue(1).(*endlessQueue)
		q.c = make(chan *Task, s)
		return q
	}
	for _, tc := range []struct {
		q        *endlessQueue
		t        *Task
		wantTask *Task
		wantErr  error
	}{
		{q: makeQueue(100), t: &Task{executable: "1"}, wantErr: nil},
		{q: makeQueue(1), t: &Task{executable: "1"}, wantErr: nil},
		{q: makeQueue(0), t: &Task{executable: "1"}, wantErr: ErrQueueFull},
	} {
		tc.t.logger = testutils.DummyTaskLogger("1", "127.0.0.1")
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

	var out *Output
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	q.Start(ctx)
	out = <-q.Results()
	cancel()

	assert.Equal(t, task.name, out.Name)
	assert.Equal(t, task.teamID, out.Team)
	assert.Equal(t, task.teamIP, string(out.Out))
}
