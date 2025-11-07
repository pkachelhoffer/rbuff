package rbuff

import (
	"context"
	"sync"
	"time"
)

const defaultSleep = time.Millisecond

type RingBuffer[T any] struct {
	buffer []T

	// reader is the index of the buffer that's been read previously.
	reader int

	// readerOdd tracks when the reader wraps around to the start of the buffer.
	readerOdd bool

	// readerInit indicates whether the reader has been initialized to start reading from the buffer.
	readerInit bool

	// writer is the index of the buffer that's been written to previously.
	writer int

	// writerNext is the next index of the buffer to be written to, computed during the write operation. Declared outside
	// the scope of the function to avoid allocations.
	writerNext int

	// writerOdd tracks when the writer wraps around to the start of the buffer. Used to check if the writer is ahead or
	// behind the reader.
	writerOdd bool

	// writerInit indicates whether the writer has been initialized to start writing to the buffer.
	writerInit bool

	mu sync.Mutex

	// sleep defines the duration for which the buffer sleeps while waiting for read or write operations to be possible.
	sleep time.Duration
}

func NewRB[T any](len int, opts ...FnOptions) *RingBuffer[T] {
	o := options{
		sleep: defaultSleep,
	}

	for _, opt := range opts {
		opt(&o)
	}

	rb := &RingBuffer[T]{
		buffer: make([]T, len),
		sleep:  o.sleep,
		reader: 0,
		writer: 0,
	}

	return rb
}

type options struct {
	sleep time.Duration
}

type FnOptions func(opts *options)

func WithSleep(sleep time.Duration) FnOptions {
	return func(opts *options) {
		opts.sleep = sleep
	}
}

// Add inserts an item into the ring buffer, waiting if necessary until space is available to write.
func (rb *RingBuffer[T]) Add(ctx context.Context, item T) error {
	for {
		ok := rb.checkAddNext(item)
		if !ok {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(rb.sleep):
				continue
			}
		}

		return nil
	}
}

func (rb *RingBuffer[T]) checkAddNext(item T) bool {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	if rb.writer == rb.reader {
		if rb.writerOdd != rb.readerOdd {
			return false
		}
	}

	rb.writerNext = rb.writer + 1

	// If the reader hasn't started yet, don't overwrite index 0
	if rb.writerNext == len(rb.buffer) && !rb.readerInit {
		return false
	}

	if rb.writerInit {
		rb.writer++
		if rb.writer == len(rb.buffer) {
			rb.writer = 0
			rb.writerOdd = !rb.writerOdd
		}
	} else {
		rb.writerInit = true
	}

	rb.buffer[rb.writer] = item

	return true
}

// Read retrieves the next item from the ring buffer, blocking and retrying if the buffer is empty or if the writer has
// added no new entries
func (rb *RingBuffer[T]) Read(ctx context.Context) (T, error) {
	for {
		val, ok := rb.checkReadNext()
		if !ok {
			select {
			case <-ctx.Done():
				var t T
				return t, ctx.Err()
			case <-time.After(rb.sleep):
				continue
			}
		}

		return val, nil
	}
}

func (rb *RingBuffer[T]) checkReadNext() (T, bool) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	if rb.reader == rb.writer {
		if rb.readerOdd == rb.writerOdd {
			var t T
			return t, false
		}
	}

	if rb.readerInit {
		rb.reader++
		if rb.reader == len(rb.buffer) {
			rb.reader = 0
			rb.readerOdd = !rb.readerOdd
		}
	} else {
		rb.readerInit = true
	}

	return rb.buffer[rb.reader], true
}
