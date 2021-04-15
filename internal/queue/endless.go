package queue

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
)

const (
	truncateOut = 4096
)

type endlessQueue struct {
	out     chan *Output
	c       chan Task
	maxJobs int
	wg      sync.WaitGroup
	done    chan struct{}
}

func NewEndlessQueue(maxJobs int) Queue {
	maxJobs *= maxJobsMultiplier

	return &endlessQueue{
		out:     make(chan *Output, maxBufferSize),
		c:       make(chan Task, maxBufferSize),
		done:    make(chan struct{}),
		maxJobs: maxJobs,
	}
}

func (eq *endlessQueue) Start(ctx context.Context) {
	logrus.Infof("Starting a queue with %d jobs", eq.maxJobs)
	eq.wg.Add(eq.maxJobs)
	for i := 0; i < eq.maxJobs; i++ {
		go eq.worker(ctx)
	}
}

func (eq *endlessQueue) Results() <-chan *Output {
	return eq.out
}

func (eq *endlessQueue) Add(et Task) error {
	select {
	case eq.c <- et:
		return nil
	default:
		return ErrQueueFull
	}
}

func (eq *endlessQueue) Stop() {
	close(eq.done)
	eq.wg.Wait()
	close(eq.out)
}

func (eq *endlessQueue) worker(ctx context.Context) {
	defer eq.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case <-eq.done:
			return
		case job := <-eq.c:
			for {
				err := eq.runExploit(ctx, job)
				// eq.done is closed
				if errors.Is(err, context.Canceled) {
					return
				}
				// context expired
				if ctx.Err() != nil {
					return
				}
				// any other error -> retry
			}
		}
	}
}

func (eq *endlessQueue) runExploit(ctx context.Context, et Task) error {
	cmdCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	logrus.Infof("Going to run endlessly: %s %s", et.executable, et.teamIP)
	cmd := et.Command(cmdCtx)

	res := new(bytes.Buffer)
	errC := make(chan error, 1)
	go func() {
		cmd.Stdout = res
		cmd.Stderr = res
		errC <- cmd.Run()
	}()
	select {
	case <-eq.done:
		cancel()
		<-errC
		return context.Canceled
	case <-ctx.Done():
		<-errC
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("contextdone: %w", err)
		}
		return nil
	case err := <-errC:
		out := res.Bytes()
		if len(out) > truncateOut {
			out = out[len(out)-truncateOut:]
		}
		logrus.Errorf("Endless sploit %v terminated: %v. Out: %s", et, err, out)
		return fmt.Errorf("unexpected error in endless exploit %v: %w", et, err)
	}
}
