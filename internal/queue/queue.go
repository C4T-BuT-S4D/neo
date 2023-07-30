package queue

import (
	"context"
	"errors"
	"fmt"
)

var (
	ErrQueueFull = errors.New("queue channel is full")
)

type Factory func(maxJobs int) Queue

type Queue interface {
	Start(context.Context)
	Add(*Task) error
	Results() <-chan *Output

	fmt.Stringer
}
