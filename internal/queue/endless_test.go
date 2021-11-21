package queue

import (
	"context"
	"errors"
	"testing"
	"time"

	"neo/pkg/testutils"
)

func Test_endlessQueue_Add(t *testing.T) {
	makeQueue := func(s int) *endlessQueue {
		q := NewEndlessQueue(1).(*endlessQueue)
		q.c = make(chan Task, s)
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
		err := tc.q.Add(*tc.t)
		if !errors.Is(err, tc.wantErr) {
			t.Errorf("endlessQueue.Add(): got error = %v, want = %v", err, tc.wantErr)
			continue
		}
		if err != nil {
			continue
		}
		if tk := <-tc.q.c; tk.executable != tc.t.executable {
			t.Errorf("endlessQueue.Add(): got unexpected data = %v, want = %v", tk, tc.t)
		}
	}
}

func Test_endlessQueue_Start(t *testing.T) {
	q := NewEndlessQueue(10)
	task := Task{
		name:       "kek",
		executable: "echo",
		dir:        "",
		teamID:     "id",
		teamIP:     "ip",
		timeout:    time.Second * 2,
		logger:     testutils.DummyTaskLogger("echo", "ip"),
	}
	if err := q.Add(task); err != nil {
		t.Errorf("endlessQueue.Add(): got unexpected error = %v", err)
	}

	var out *Output
	ctx, cancel := context.WithCancel(context.Background())
	q.Start(ctx)
	defer q.Stop()
	out = <-q.Results()
	cancel()

	if out.Name != task.name {
		t.Errorf("endlessQueue.Start(): got unexpected result name: got = %v, want = %v", out.Name, task.name)
	}
	if out.Team != task.teamID {
		t.Errorf("endlessQueue.Start(): got unexpected result team: got = %v, want = %v", out.Team, task.teamID)
	}
	if string(out.Out) != task.teamIP {
		t.Errorf("endlessQueue.Start(): got unexpected result: got = %v, want = %v", out.Out, task.teamIP)
	}
}
