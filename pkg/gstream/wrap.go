package gstream

import (
	"context"
	"fmt"
)

type wrappedWStream[In, Out any] struct {
	original WStream[Out]
	mapper   func(*In) *Out
}

func Wrap[In, Out any](original WStream[Out], mapper func(*In) *Out) WStream[In] {
	return &wrappedWStream[In, Out]{
		original: original,
		mapper:   mapper,
	}
}

func (w *wrappedWStream[In, Out]) Send(t *In) error {
	if err := w.original.Send(w.mapper(t)); err != nil {
		return fmt.Errorf("sending to original stream: %w", err)
	}
	return nil
}

func (w *wrappedWStream[In, Out]) Context() context.Context {
	return w.original.Context()
}
