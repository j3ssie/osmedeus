package utils

import (
	"testing"
)

func TestGetPutLargeBuffer(t *testing.T) {
	buf := GetLargeBuffer()
	if buf == nil {
		t.Fatal("expected non-nil buffer")
	}
	if len(*buf) != LargeBufferSize {
		t.Errorf("expected buffer size %d, got %d", LargeBufferSize, len(*buf))
	}

	// Write some data
	(*buf)[0] = 'x'

	// Return to pool
	PutLargeBuffer(buf)

	// Get another - might be same buffer
	buf2 := GetLargeBuffer()
	if len(*buf2) != LargeBufferSize {
		t.Errorf("expected buffer size %d after reuse, got %d", LargeBufferSize, len(*buf2))
	}
	PutLargeBuffer(buf2)
}

func TestPutNilBuffer(t *testing.T) {
	// Should not panic
	PutLargeBuffer(nil)
}

func TestGetScannerBuffer(t *testing.T) {
	initial, max := GetScannerBuffer()
	if len(initial) != 64*1024 {
		t.Errorf("expected initial size 64KB, got %d", len(initial))
	}
	if max != LargeBufferSize {
		t.Errorf("expected max size %d, got %d", LargeBufferSize, max)
	}
}

func BenchmarkBufferPool(b *testing.B) {
	b.Run("pool", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			buf := GetLargeBuffer()
			PutLargeBuffer(buf)
		}
	})

	b.Run("alloc", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = make([]byte, LargeBufferSize)
		}
	})
}
