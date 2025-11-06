# rbuff — A tiny generic ring buffer for Go

rbuff is a minimal, thread-safe, generic ring buffer implemented in Go. It provides a simple API to add items and read them in FIFO order, using an internal mutex and backoff sleep when reads or writes must wait.

The package is intended as a small building block for producer/consumer style pipelines or anywhere you need an in-memory circular buffer with blocking read/write semantics.

## Features
- Generic type parameter `RingBuffer[T]`
- Blocking `Add` and `Read` with configurable sleep interval
- No allocations on steady state operations
- Simple, dependency-free

## Installation
```bash
go get github.com/pkachelhoffer/rbuff@latest
```

## Quick Start
```go
package main

import (
    "fmt"
    "time"

    "github.com/pkachelhoffer/rbuff"
)

func main() {
    // Create a ring buffer of capacity 10, with a custom sleep/backoff of 100µs
    rb := rbuff.NewRB[int](10, rbuff.WithSleep(100*time.Microsecond))

    // Producer
    go func() {
        for i := 0; i < 5; i++ {
            rb.Add(i)
        }
    }()

    // Consumer
    for i := 0; i < 5; i++ {
        v := rb.Read()
        fmt.Println(v)
    }
}
```

## API
- `func NewRB[T any](len int, opts ...FnOptions) *RingBuffer[T]`
  - Create a new ring buffer with the given length/capacity.
- `func WithSleep(d time.Duration) FnOptions`
  - Option to configure the sleep/backoff used while waiting to read/write.
- `(rb *RingBuffer[T]) Add(item T)`
  - Add an item, blocking until space is available.
- `(rb *RingBuffer[T]) Read() T`
  - Read the next item, blocking until an item becomes available.

## Scripts & Tooling
- Makefile targets:
  - `bench-ringbuffer`: run the ring buffer benchmark and open the Go trace UI.
    ```bash
    make bench-ringbuffer
    # Equivalent:
    # go test -bench=BenchmarkAddRemoveRingBuffer -run=^$ -benchmem -benchtime=10s ./ -trace=./trace.out
    # go tool trace ./trace.out
    ```

## Running Tests
- Unit tests:
  ```bash
  go test ./...
  ```
- Benchmarks (standalone):
  ```bash
  go test -bench=BenchmarkAddRemoveRingBuffer -run=^$ -benchmem -benchtime=10s ./
  ```

## Entry Points
There is no command-line application. Import it as:

```go
import "github.com/pkachelhoffer/rbuff"
```

## Performance Notes
- The buffer uses a `sync.Mutex` for internal synchronization and busy-sleeps for backoff when reads/writes must wait. Tune the backoff using `WithSleep` to balance latency and CPU usage for your workload.

## License
This project is licensed under the GNU General Public License v3.0. See the `LICENSE` file for details.
