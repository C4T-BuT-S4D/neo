package queue

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	jobBufferSize    = 1000
	resultBufferSize = 1000
)

// Compile-time type checks.
var (
	_ Queue = (*simpleQueue)(nil)
)

type simpleQueue struct {
	out     chan *Output
	c       chan *Job
	maxJobs int
	metrics *Metrics
	logger  *zap.Logger
}

func NewSimpleQueue(maxJobs int) Queue {
	id := uuid.NewString()[:8]

	return &simpleQueue{
		out:     make(chan *Output, resultBufferSize),
		c:       make(chan *Job, jobBufferSize),
		maxJobs: maxJobs,
		metrics: NewMetrics("neo", id, TypeSimple),
		logger:  zap.L().Named("simple_queue").With(zap.String("id", id)),
	}
}

func (q *simpleQueue) Type() Type {
	return TypeSimple
}

func (q *simpleQueue) Size() int {
	return len(q.c)
}

// Start is synchronous.
// Cancel the start's context to stop the queue.
func (q *simpleQueue) Start(ctx context.Context) {
	q.logger.Info("Starting", zap.Int("jobs", q.maxJobs))

	q.metrics.MaxJobs.Add(float64(q.maxJobs))
	defer q.metrics.MaxJobs.Sub(float64(q.maxJobs))

	wg := sync.WaitGroup{}
	wg.Add(q.maxJobs)
	for range q.maxJobs {
		go func() {
			defer wg.Done()
			q.worker(ctx)
		}()
	}
	wg.Wait()

	q.logger.Info("Stopped")
}

func (q *simpleQueue) Results() <-chan *Output {
	return q.out
}

func (q *simpleQueue) Add(task *Job) error {
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

			exploitLabels := task.Exploit.MetricLabels()

			var exitErr *exec.ExitError
			switch {
			case err == nil:
				task.logger.Infof("Successfully run")
				task.logger.Debugf("Output: %s", res)
				q.metrics.ExploitsFinished.With(exploitLabels).Inc()
			case errors.Is(err, context.Canceled):
				// Expected error on queue restart.
				q.metrics.ExploitsFinished.With(exploitLabels).Inc()
				return
			case errors.As(err, &exitErr):
				if exitErr.ExitCode() == -1 {
					task.logger.Warningf("Exploit exited with signal: %v", exitErr)
					task.logger.Debugf("Output: %s", res)
				} else {
					task.logger.Warningf("Unexpected error in exploit: %v", err)
					task.logger.Debugf("Output: %s", res)
				}
				q.metrics.ExploitsFailed.With(exploitLabels).Inc()

			default:
				task.logger.Errorf("Failed to run: %v", err)
				task.logger.Debugf("Output: %s", res)
				q.metrics.ExploitsFailed.With(exploitLabels).Inc()
			}
			q.out <- NewOutput(task, res)
		}
	}
}

func (q *simpleQueue) runExploit(ctx context.Context, task *Job) ([]byte, error) {
	cmdCtx, cancel := context.WithTimeout(ctx, task.timeout)
	defer cancel()

	cmd := task.Command(cmdCtx)
	task.logger.Infof("Going to run: %v", cmd)

	start := time.Now()
	exploitLabels := task.Exploit.MetricLabels()
	q.metrics.ExploitInstancesRunning.With(exploitLabels).Inc()
	defer func() {
		q.metrics.ExploitInstancesRunning.With(exploitLabels).Dec()
		q.metrics.ExploitRunTime.With(exploitLabels).Observe(time.Since(start).Seconds())
	}()

	// Will be terminated on context cancellation.
	out, err := cmd.CombinedOutput()
	if err != nil {
		err = fmt.Errorf("error running command: %w", err)
	}
	return out, err
}
