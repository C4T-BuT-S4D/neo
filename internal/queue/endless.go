package queue

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Compile-time type checks
var (
	_ Queue   = (*endlessQueue)(nil)
	_ Factory = NewEndlessQueue
)

type endlessQueue struct {
	out     chan *Output
	c       chan *Task
	maxJobs int
	logger  *logrus.Entry
}

func NewEndlessQueue(maxJobs int) Queue {
	return &endlessQueue{
		out:     make(chan *Output, resultBufferSize),
		c:       make(chan *Task, jobBufferSize),
		maxJobs: maxJobs,
		logger: logrus.WithFields(logrus.Fields{
			"component": "endless_queue",
			"id":        uuid.NewString()[:4],
		}),
	}
}

// Start is synchronous.
// Cancel the start's context to stop the queue.
func (q *endlessQueue) Start(ctx context.Context) {
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

func (q *endlessQueue) Results() <-chan *Output {
	return q.out
}

func (q *endlessQueue) Add(task *Task) error {
	select {
	case q.c <- task:
		return nil
	default:
		return ErrQueueFull
	}
}

func (q *endlessQueue) String() string {
	return "EndlessQueue"
}

func (q *endlessQueue) worker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case task := <-q.c:
			for {
				err := q.runExploit(ctx, task)
				switch {
				case errors.Is(err, context.Canceled):
					return
				case err != nil:
					task.logger.Errorf("Unexpected error returned from endless exploit: %v", err)
				default:
					task.logger.Errorf("Endless exploit terminated unexpectedly")
				}
			}
		}
	}
}

func (q *endlessQueue) runExploit(ctx context.Context, task *Task) error {
	cmdCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	task.logger.Infof("Going to run endlessly: %s %s", task.executable, task.teamIP)
	cmd := task.Command(cmdCtx)

	// os.Pipe performs better than io.Pipe.
	r, w, err := os.Pipe()
	if err != nil {
		return fmt.Errorf("creating pipe: %w", err)
	}

	errC := make(chan error, 1)
	go func() {
		defer func(w *os.File) {
			if err := w.Close(); err != nil {
				task.logger.Errorf("Error closing pipe: %v", err)
			}
		}(w)

		cmd.Stdout = w
		cmd.Stderr = w
		errC <- cmd.Run()
	}()

	readDone := make(chan struct{})
	defer func() {
		task.logger.Infof("Waiting for endless read to finish")
		<-readDone
	}()

	go func() {
		defer close(readDone)
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			q.out <- &Output{
				Name: task.name,
				Out:  scanner.Bytes(),
				Team: task.teamID,
			}
		}
		if err := scanner.Err(); err != nil && ctx.Err() == nil {
			task.logger.Errorf("Unexpected error reading endless script output: %v", err)
		}
	}()

	select {
	case <-ctx.Done():
		<-errC
		return nil
	case err := <-errC:
		// Context terminated, expected error.
		if ctx.Err() != nil {
			return nil
		}

		task.logger.Errorf("Endless sploit %v terminated: %v", task, err)
		if err != nil {
			return fmt.Errorf("unexpected error in endless exploit %v: %w", task, err)
		}
		return nil
	}
}
