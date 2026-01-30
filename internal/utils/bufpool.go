// Package utils provides shared utility functions
package utils

import "sync"

// LargeBufferSize is the default size for large buffers (10MB)
const LargeBufferSize = 10 * 1024 * 1024

// largeBufferPool provides reusable 10MB buffers to reduce allocation overhead
var largeBufferPool = sync.Pool{
	New: func() interface{} {
		buf := make([]byte, LargeBufferSize)
		return &buf
	},
}

// GetLargeBuffer retrieves a 10MB buffer from the pool
func GetLargeBuffer() *[]byte {
	return largeBufferPool.Get().(*[]byte)
}

// PutLargeBuffer returns a buffer to the pool
// The buffer slice will be reset to full capacity
func PutLargeBuffer(buf *[]byte) {
	if buf == nil {
		return
	}
	// Reset to full capacity
	*buf = (*buf)[:cap(*buf)]
	largeBufferPool.Put(buf)
}

// GetScannerBuffer retrieves a buffer suitable for bufio.Scanner
// Returns both initial buffer and max size for Scanner.Buffer()
func GetScannerBuffer() ([]byte, int) {
	buf := GetLargeBuffer()
	return (*buf)[:64*1024], LargeBufferSize // initial 64KB, max 10MB
}
