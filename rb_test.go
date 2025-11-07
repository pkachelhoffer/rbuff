package rbuff

import (
	"sync"
	"testing"
	"time"

	"golang.org/x/sync/errgroup"
)

type iType string

const (
	iTypeRead  iType = "read"
	iTypeWrite iType = "write"
)

type instruction struct {
	iType      iType
	count      int
	sleep      time.Duration
	sleepStart time.Duration
}

func TestScenarios(t *testing.T) {
	tcs := []struct {
		name         string
		bufferLength int
		instructions []instruction
		expResults   []int
	}{
		{
			name:         "simple in bounds",
			bufferLength: 100,
			instructions: []instruction{
				{
					iType: iTypeWrite,
					count: 10,
				},
				{
					iType: iTypeRead,
					count: 10,
				},
			},
			expResults: sliceCount(10),
		},
		{
			name:         "on bounds",
			bufferLength: 10,
			instructions: []instruction{
				{
					iType: iTypeWrite,
					count: 10,
				},
				{
					iType: iTypeRead,
					count: 10,
				},
			},
			expResults: sliceCount(10),
		},
		{
			name:         "over bounds writer",
			bufferLength: 10,
			instructions: []instruction{
				{
					iType: iTypeWrite,
					count: 14,
				},
				{
					iType:      iTypeRead,
					count:      14,
					sleepStart: time.Millisecond * 5,
				},
			},
			expResults: sliceCount(14),
		},
		{
			name:         "slow writer",
			bufferLength: 10,
			instructions: []instruction{
				{
					iType: iTypeWrite,
					count: 10,
					sleep: time.Millisecond,
				},
				{
					iType: iTypeRead,
					count: 10,
				},
			},
			expResults: sliceCount(10),
		},
		{
			name:         "slow writer wrap",
			bufferLength: 10,
			instructions: []instruction{
				{
					iType: iTypeWrite,
					count: 10,
					sleep: time.Millisecond,
				},
				{
					iType: iTypeRead,
					count: 10,
				},
				{
					iType: iTypeWrite,
					count: 10,
					sleep: time.Millisecond,
				},
				{
					iType: iTypeRead,
					count: 10,
				},
			},
			expResults: append(sliceCount(10), sliceCount(10)...),
		},
		{
			name:         "slow reader wrap",
			bufferLength: 10,
			instructions: []instruction{
				{
					iType: iTypeWrite,
					count: 10,
				},
				{
					iType: iTypeRead,
					count: 10,
					sleep: time.Millisecond,
				},
				{
					iType: iTypeWrite,
					count: 10,
				},
				{
					iType: iTypeRead,
					count: 10,
					sleep: time.Millisecond,
				},
			},
			expResults: append(sliceCount(10), sliceCount(10)...),
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			rb := NewRB[int](tc.bufferLength)

			var (
				results []int
				mu      sync.Mutex
			)

			var eg errgroup.Group

			// writers
			eg.Go(func() error {
				for _, in := range tc.instructions {
					if in.iType != iTypeWrite {
						continue
					}

					time.Sleep(in.sleepStart)

					for i := 0; i < in.count; i++ {
						time.Sleep(in.sleep)
						err := rb.Add(t.Context(), i)
						if err != nil {
							return err
						}
					}
				}

				return nil
			})

			// readers
			eg.Go(func() error {
				for _, in := range tc.instructions {
					if in.iType != iTypeRead {
						continue
					}

					time.Sleep(in.sleepStart)

					for i := 0; i < in.count; i++ {
						time.Sleep(in.sleep)
						mu.Lock()
						res, err := rb.Read(t.Context())
						if err != nil {
							return err
						}
						results = append(results, res)
						mu.Unlock()
					}
				}

				return nil
			})

			err := eg.Wait()
			if err != nil {
				t.Errorf("error running loops: %s", err.Error())
				t.Fail()
			}

			if len(tc.expResults) != len(results) {
				t.Errorf("expected result length differs")
				t.Fail()
			}

			for i := 0; i < len(tc.expResults); i++ {
				if tc.expResults[i] != results[i] {
					t.Errorf("expected %d (%d), got %d", tc.expResults[i], i, results[i])
				}
			}
		})
	}
}

func sliceCount(n int) []int {
	var s []int
	for i := 0; i < n; i++ {
		s = append(s, i)
	}
	return s
}

func BenchmarkAddRemoveRingBuffer(b *testing.B) {
	rb := NewRB[int](30)
	for i := 0; i < b.N; i++ {
		runAddRemove(b, rb, 30)
	}
}

func runAddRemove(b *testing.B, rb *RingBuffer[int], cnt int) {
	var err error
	for i := 0; i < cnt; i++ {
		err = rb.Add(b.Context(), i)
		if err != nil {
			b.Errorf("err adding item")
			b.Fail()
		}
	}

	for i := 0; i < cnt; i++ {
		_, err = rb.Read(b.Context())
		if err != nil {
			b.Errorf("err adding item")
			b.Fail()
		}
	}
}
