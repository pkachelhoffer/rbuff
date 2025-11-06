package rbuff

import (
	"sync"
	"time"
)

const defaultSleep = time.Millisecond

type RingBuffer[T any] struct {
	buffer []T

	// reader is the index of the buffer that's been read previously.
	reader     int
	readerNext int
	readerOdd  bool
	readerInit bool

	// writer is the index of the buffer that's been written to previously.
	writer     int
	writerNext int
	writerOdd  bool
	writerInit bool

	muWriter sync.RWMutex
	muReader sync.RWMutex

	mu sync.Mutex

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

func (rb *RingBuffer[T]) Add(item T) {
	for {
		ok := rb.checkAddNext(item)
		if !ok {
			time.Sleep(rb.sleep)
			continue
		}

		return
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

func (rb *RingBuffer[T]) Read() T {
	for {
		val, ok := rb.checkReadNext()
		if !ok {
			time.Sleep(rb.sleep)
			continue
		}

		return val
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
