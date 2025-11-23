package pool

import (
	"bytes"
	"testing"
)

func TestGetBytesBuffer(t *testing.T) {
	buf := GetBytesBuffer()
	if buf == nil {
		t.Fatal("GetBytesBuffer() returned nil")
	}
	if buf.Len() != 0 {
		t.Errorf("Buffer length = %d, want 0", buf.Len())
	}

	buf.WriteString("test")
	PutBytesBuffer(buf)

	// Get another buffer
	buf2 := GetBytesBuffer()
	if buf2.Len() != 0 {
		t.Errorf("Buffer not reset, length = %d, want 0", buf2.Len())
	}
	PutBytesBuffer(buf2)
}

func TestPutBytesBuffer_LargeBuffer(t *testing.T) {
	buf := bytes.NewBuffer(make([]byte, 0, Size64K*2))
	PutBytesBuffer(buf)
	// Should not panic, but won't actually pool the buffer
}

func BenchmarkBytesBufferPool(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		buf := GetBytesBuffer()
		buf.WriteString("some data to test")
		PutBytesBuffer(buf)
	}
}

func BenchmarkBytesBufferDirect(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		buf := new(bytes.Buffer)
		buf.WriteString("some data to test")
	}
}
