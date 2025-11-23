package pool

import (
	"bytes"
	"sync"
)

const (
	// Size64K maximum buffer size to pool (prevent memory bloat)
	Size64K = 65536
)

var (
	// bytesBufferPool is a pool for bytes.Buffer objects
	bytesBufferPool = &sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}
)

// GetBytesBuffer retrieves a bytes.Buffer from the pool
func GetBytesBuffer() *bytes.Buffer {
	buf := bytesBufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	return buf
}

// PutBytesBuffer returns a bytes.Buffer to the pool
func PutBytesBuffer(buf *bytes.Buffer) {
	if buf == nil {
		return
	}
	// Only pool buffers under a reasonable size to prevent memory bloat
	if buf.Cap() > Size64K {
		return
	}
	buf.Reset()
	bytesBufferPool.Put(buf)
}
