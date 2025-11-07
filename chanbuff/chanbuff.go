package chanbuff

import (
	"context"
)

// ChanBuff represents a generic buffered channel wrapper for type T.
// It does exactly the same as a ring buffer at comparable efficiency.
// Adding the select statements pushed the ns/ops to much higher than the base ringbuffer implementation
type ChanBuff[T any] struct {
	buffer chan T
}

func NewChanBuff[T any](len int) *ChanBuff[T] {
	return &ChanBuff[T]{
		buffer: make(chan T, len),
	}
}

func (cb *ChanBuff[T]) Add(ctx context.Context, item T) error {
	select {
	case cb.buffer <- item:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (cb *ChanBuff[T]) Read(ctx context.Context) (T, error) {
	select {
	case i := <-cb.buffer:
		return i, nil
	case <-ctx.Done():
		var t T
		return t, ctx.Err()
	}
}
