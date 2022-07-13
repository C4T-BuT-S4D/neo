package stream

import (
	"context"
	"fmt"
	"sync"
)

type Sizable interface {
	EstimateSize() int
}

type BatcherFunc[S, D any] func(t []S) D

func NewDynamicSizeCache[T Sizable, M any](s WStream[M], maxSize int, bf BatcherFunc[T, *M]) *DynamicSizeCache[T, M] {
	return &DynamicSizeCache[T, M]{
		stream:  s,
		batcher: bf,
		maxSize: maxSize,
	}
}

type DynamicSizeCache[T Sizable, M any] struct {
	stream  WStream[M]
	batcher BatcherFunc[T, *M]
	maxSize int
	curSize int
	queue   []T
	mu      sync.Mutex
}

func (d *DynamicSizeCache[T, M]) Queue(ts ...T) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	for _, t := range ts {
		d.curSize += t.EstimateSize()
		d.queue = append(d.queue, t)
		if d.curSize >= d.maxSize {
			if err := d.flushUnsafe(); err != nil {
				return fmt.Errorf("flushing batch: %w", err)
			}
		}
	}
	return nil
}

func (d *DynamicSizeCache[T, M]) Flush() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.flushUnsafe()
}

func (d *DynamicSizeCache[T, M]) Context() context.Context {
	return d.stream.Context()
}

// flushUnsafe expects the lock to be held.
func (d *DynamicSizeCache[T, M]) flushUnsafe() error {
	m := d.batcher(d.queue)
	if err := d.stream.Send(m); err != nil {
		return fmt.Errorf("sending batch to stream: %w", err)
	}
	d.curSize = 0
	d.queue = nil
	return nil
}
