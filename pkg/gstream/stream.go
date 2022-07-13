package gstream

import (
	"context"
)

type RStream[T any] interface {
	Recv() (*T, error)
	Context() context.Context
}

type WStream[T any] interface {
	Send(*T) error
	Context() context.Context
}
