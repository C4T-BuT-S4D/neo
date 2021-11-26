package queue

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"sync"

	"github.com/sirupsen/logrus"
)

const maxBufferSize = 10000

// Compile-time type checks
var _ Queue = (*simpleQueue)(nil)
var _ Factory = NewSimpleQueue

type simpleQueue struct {
	out     chan *Output
	c       chan Task
	maxJobs int
	wg      sync.WaitGroup
	done    chan struct{}
}

func NewSimpleQueue(maxJobs int) Queue {
	return &simpleQueue{
		out:     make(chan *Output, maxBufferSize),
		c:       make(chan Task, maxBufferSize),
		done:    make(chan struct{}),
		maxJobs: maxJobs,
	}
}

func (eq *simpleQueue) Start(ctx context.Context) {
	logrus.Infof("Starting a queue with %d jobs", eq.maxJobs)
	eq.wg.Add(eq.maxJobs)
	for i := 0; i < eq.maxJobs; i++ {
		go eq.worker(ctx)
	}
}

func (eq *simpleQueue) Results() <-chan *Output {
	return eq.out
}

func (eq *simpleQueue) Add(et Task) error {
	select {
	case eq.c <- et:
		return nil
	default:
		return ErrQueueFull
	}
}

func (eq *simpleQueue) Stop() {
	close(eq.done)
	eq.wg.Wait()
	close(eq.out)
}

func (eq *simpleQueue) worker(ctx context.Context) {
	defer eq.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case <-eq.done:
			return
		case job := <-eq.c:
			res, err := eq.runExploit(ctx, job)

			var exitErr *exec.ExitError
			if err == nil {
				job.logger.Infof("Successfully run")
				job.logger.Debugf("Output: %s", res)
			} else if errors.Is(err, context.Canceled) || errors.As(err, &exitErr) {
				job.logger.Warningf("Task finished unsuccessfully: %v. Output: %s", err, res)
			} else {
				job.logger.Errorf("Failed to run: %v. Output: %s", err, res)
			}
			eq.out <- &Output{
				Name: job.name,
				Out:  res,
				Team: job.teamID,
			}
		}
	}
}

func (eq *simpleQueue) runExploit(ctx context.Context, et Task) ([]byte, error) {
	cmdCtx, cancel := context.WithTimeout(ctx, et.timeout)
	defer cancel()

	et.logger.Infof("Going to run: %s %s", et.executable, et.teamIP)
	cmd := et.Command(cmdCtx)

	var out []byte
	errC := make(chan error, 1)
	go func() {
		var err error
		out, err = cmd.CombinedOutput()
		errC <- err
	}()
	select {
	case <-eq.done:
		cancel()
		<-errC
		return out, context.Canceled
	case <-ctx.Done():
		<-errC
		return out, fmt.Errorf("context terminated: %w", ctx.Err())
	case err := <-errC:
		return out, err
	}
}
