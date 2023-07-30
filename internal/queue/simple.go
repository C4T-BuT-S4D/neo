package queue

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"sync"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

const (
	jobBufferSize    = 1000
	resultBufferSize = 1000
)

// Compile-time type checks
var (
	_ Queue   = (*simpleQueue)(nil)
	_ Factory = NewSimpleQueue
)

type simpleQueue struct {
	out     chan *Output
	c       chan *Task
	maxJobs int
	logger  *logrus.Entry
}

func NewSimpleQueue(maxJobs int) Queue {
	return &simpleQueue{
		out:     make(chan *Output, resultBufferSize),
		c:       make(chan *Task, jobBufferSize),
		maxJobs: maxJobs,
		logger: logrus.WithFields(logrus.Fields{
			"component": "simple_queue",
			"id":        uuid.NewString()[:4],
		}),
	}
}

// Start is synchronous.
// Cancel the start's context to stop the queue.
func (q *simpleQueue) Start(ctx context.Context) {
	q.logger.WithField("jobs", q.maxJobs).Info("Starting")
	defer q.logger.Info("Stopped")

	wg := sync.WaitGroup{}
	wg.Add(q.maxJobs)
	for i := 0; i < q.maxJobs; i++ {
		go func() {
			defer wg.Done()
			q.worker(ctx)
		}()
	}
	wg.Wait()
}

func (q *simpleQueue) Results() <-chan *Output {
	return q.out
}

func (q *simpleQueue) Add(task *Task) error {
	select {
	case q.c <- task:
		return nil
	default:
		return ErrQueueFull
	}
}

func (q *simpleQueue) String() string {
	return "SimpleQueue"
}

func (q *simpleQueue) worker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case task := <-q.c:
			res, err := q.runExploit(ctx, task)

			var exitErr *exec.ExitError
			if err == nil {
				task.logger.Infof("Successfully run")
				task.logger.Debugf("Output: %s", res)
			} else if errors.Is(err, context.Canceled) || errors.As(err, &exitErr) {
				task.logger.Warningf("Task finished unsuccessfully: %v. Output: %s", err, res)
			} else {
				task.logger.Errorf("Failed to run: %v. Output: %s", err, res)
			}
			q.out <- &Output{
				Name: task.name,
				Out:  res,
				Team: task.teamID,
			}
		}
	}
}

func (q *simpleQueue) runExploit(ctx context.Context, task *Task) ([]byte, error) {
	cmdCtx, cancel := context.WithTimeout(ctx, task.timeout)
	defer cancel()

	task.logger.Infof("Going to run: %s %s", task.executable, task.teamIP)
	cmd := task.Command(cmdCtx)

	// Will be terminated on context cancellation.
	out, err := cmd.CombinedOutput()
	if err != nil {
		err = fmt.Errorf("error running command: %w", err)
	}
	return out, err
}
