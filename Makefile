.PHONY: bench-ringbuffer bench-listbatcher

bench-ringbuffer:
	go test -bench=BenchmarkAddRemoveRingBuffer -run=^$ -benchmem -benchtime=10s ./ -trace=./trace.out
	go tool trace ./trace.out