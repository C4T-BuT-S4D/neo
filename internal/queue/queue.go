package queue

import (
	"context"
	"errors"
	"fmt"
)

var (
	ErrQueueFull = errors.New("queue channel is full")
)

type Type string

func (t Type) String() string {
	return string(t)
}

const (
	TypeSimple  Type = "simple"
	TypeEndless Type = "endless"
)

type Factory interface {
	Create(maxJobs int) Queue
	Type() Type
}

type Queue interface {
	Start(context.Context)
	Add(*Job) error
	Results() <-chan *Output
	Type() Type

	fmt.Stringer
}
