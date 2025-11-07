package chanbuff

import (
	"testing"
)

func BenchmarkAddRemoveChannelBuffer(b *testing.B) {
	rb := NewChanBuff[int](30)
	for i := 0; i < b.N; i++ {
		runAddRemove(b, rb, 30)
	}
}

func runAddRemove(b *testing.B, rb *ChanBuff[int], cnt int) {
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
