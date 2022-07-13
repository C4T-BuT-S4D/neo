package queue

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/sirupsen/logrus"
)

// Compile-time type checks
var _ Queue = (*endlessQueue)(nil)
var _ Factory = NewEndlessQueue

type endlessQueue struct {
	out     chan *Output
	c       chan Task
	maxJobs int
	wg      sync.WaitGroup
	done    chan struct{}
}

func NewEndlessQueue(maxJobs int) Queue {
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

func (eq *endlessQueue) Add(task Task) error {
	select {
	case eq.c <- task:
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
		case task := <-eq.c:
			for {
				err := eq.runExploit(ctx, task)
				// eq.done is closed
				if errors.Is(err, context.Canceled) {
					return
				}
				// context expired
				if ctx.Err() != nil {
					return
				}
				if err != nil {
					task.logger.Errorf("Unexpected error returned from endless exploit: %v", err)
				} else {
					task.logger.Errorf("Endless exploit terminated unexpectedly")
				}
			}
		}
	}
}

func (eq *endlessQueue) runExploit(ctx context.Context, task Task) error {
	cmdCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	task.logger.Infof("Going to run endlessly: %s %s", task.executable, task.teamIP)
	cmd := task.Command(cmdCtx)

	r, w := io.Pipe()
	errC := make(chan error, 1)
	go func() {
		defer func(w *io.PipeWriter) {
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
			eq.out <- &Output{
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
	case <-eq.done:
		cancel()
		<-errC
		return context.Canceled
	case <-ctx.Done():
		<-errC
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("context done: %w", err)
		}
		return nil
	case err := <-errC:
		task.logger.Errorf("Endless sploit %v terminated: %v", task, err)
		if err != nil {
			return fmt.Errorf("unexpected error in endless exploit %v: %w", task, err)
		}
		return nil
	}
}
