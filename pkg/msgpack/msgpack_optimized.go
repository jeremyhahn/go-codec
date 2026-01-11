//go:build codec_msgpack

package msgpack

import (
	"bytes"

	"github.com/jeremyhahn/go-codec/pkg/pool"
	"github.com/vmihailenco/msgpack/v5"
)

// OptimizedCodec implements zero-allocation MessagePack encoding/decoding
type OptimizedCodec[T any] struct {
	*Codec[T]
}

// NewPool creates a new optimized MessagePack codec with buffer pooling
func NewPool[T any]() *OptimizedCodec[T] {
	return &OptimizedCodec[T]{
		Codec: New[T](),
	}
}

// MarshalTo marshals data into the provided buffer
func (c *OptimizedCodec[T]) MarshalTo(buf []byte, data T) ([]byte, error) {
	// Use a bytes.Buffer from pool for marshaling
	byteBuf := pool.GetBytesBuffer()
	defer pool.PutBytesBuffer(byteBuf)

	encoder := msgpack.NewEncoder(byteBuf)
	if err := encoder.Encode(data); err != nil {
		return nil, err
	}

	encoded := byteBuf.Bytes()

	// If provided buffer is large enough, copy into it
	if cap(buf) >= len(encoded) {
		buf = buf[:len(encoded)]
		copy(buf, encoded)
		return buf, nil
	}

	// Otherwise allocate new buffer
	result := make([]byte, len(encoded))
	copy(result, encoded)
	return result, nil
}

// AppendMarshal appends marshaled data to the provided buffer
func (c *OptimizedCodec[T]) AppendMarshal(buf []byte, data T) ([]byte, error) {
	// Use a bytes.Buffer from pool for marshaling
	byteBuf := pool.GetBytesBuffer()
	defer pool.PutBytesBuffer(byteBuf)

	encoder := msgpack.NewEncoder(byteBuf)
	if err := encoder.Encode(data); err != nil {
		return buf, err
	}

	// Append to existing buffer
	return append(buf, byteBuf.Bytes()...), nil
}

// UnmarshalFrom unmarshals data with optional scratch buffer
func (c *OptimizedCodec[T]) UnmarshalFrom(data []byte, v *T, scratch []byte) error {
	// Use bytes.Reader from the data directly to avoid allocation
	reader := bytes.NewReader(data)
	decoder := msgpack.NewDecoder(reader)
	return decoder.Decode(v)
}
