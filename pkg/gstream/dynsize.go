package gstream

import (
	"context"
	"fmt"
	"sync"
)

type SizerFunc[T any] func(T) int

type BatcherFunc[S, D any] func(t []S) (D, error)

// VTProtoSizer returns a sizer function for types that implement the SizeVT() method from vtproto.
// The size is overestimated by 25% to account for protobuf encoding overhead.
func VTProtoSizer[T interface{ SizeVT() int }]() SizerFunc[T] {
	return func(t T) int {
		return t.SizeVT() * 5 / 4
	}
}

func NewDynamicSizeCache[T, M any](s WStream[M], maxSize, maxCount int, sizer SizerFunc[T], bf BatcherFunc[T, *M]) *DynamicSizeCache[T, M] {
	return &DynamicSizeCache[T, M]{
		stream:   s,
		batcher:  bf,
		sizer:    sizer,
		maxSize:  maxSize,
		maxCount: maxCount,
	}
}

type DynamicSizeCache[T any, M any] struct {
	stream   WStream[M]
	batcher  BatcherFunc[T, *M]
	sizer    SizerFunc[T]
	maxSize  int
	maxCount int
	curSize  int
	queue    []T
	mu       sync.Mutex
}

func (d *DynamicSizeCache[T, M]) Queue(ts ...T) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	for _, t := range ts {
		d.curSize += d.sizer(t)
		d.queue = append(d.queue, t)
		if d.curSize >= d.maxSize || len(d.queue) >= d.maxCount {
			if err := d.flushUnlocked(); err != nil {
				return fmt.Errorf("flushing batch: %w", err)
			}
		}
	}
	return nil
}

func (d *DynamicSizeCache[T, M]) Flush() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.flushUnlocked()
}

func (d *DynamicSizeCache[T, M]) Context() context.Context {
	return d.stream.Context()
}

// flushUnlocked expects the lock to be held.
func (d *DynamicSizeCache[T, M]) flushUnlocked() error {
	if len(d.queue) == 0 {
		return nil
	}

	m, err := d.batcher(d.queue)
	if err != nil {
		return fmt.Errorf("error converting batch: %w", err)
	}
	if err := d.stream.Send(m); err != nil {
		return fmt.Errorf("sending batch to stream: %w", err)
	}
	d.curSize = 0
	d.queue = d.queue[:0]
	return nil
}
