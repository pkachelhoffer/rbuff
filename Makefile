.PHONY: bench-ringbuffer bench-channelbuffer

bench-ringbuffer:
	go test -bench=BenchmarkAddRemoveRingBuffer -run=^$ -benchmem -benchtime=10s ./ -trace=./trace.out
	go tool trace ./trace.out

bench-channelbuffer:
	go test -bench=BenchmarkAddRemoveChannelBuffer -run=^$ -benchmem -benchtime=10s ./chanbuff -trace=./chanbuff/trace.out
	go tool trace ./chanbuff/trace.out
