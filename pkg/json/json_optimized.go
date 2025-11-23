package json

import (
	"encoding/json"

	"github.com/jeremyhahn/go-codec/pkg/pool"
)

// OptimizedCodec implements zero-allocation JSON encoding/decoding
type OptimizedCodec[T any] struct {
	*Codec[T]
}

// NewPool creates a new optimized JSON codec with buffer pooling
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

	encoder := json.NewEncoder(byteBuf)
	if err := encoder.Encode(data); err != nil {
		return nil, err
	}

	// Remove the trailing newline that json.Encoder adds
	encoded := byteBuf.Bytes()
	if len(encoded) > 0 && encoded[len(encoded)-1] == '\n' {
		encoded = encoded[:len(encoded)-1]
	}

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

	encoder := json.NewEncoder(byteBuf)
	if err := encoder.Encode(data); err != nil {
		return buf, err
	}

	// Remove the trailing newline that json.Encoder adds
	encoded := byteBuf.Bytes()
	if len(encoded) > 0 && encoded[len(encoded)-1] == '\n' {
		encoded = encoded[:len(encoded)-1]
	}

	// Append to existing buffer
	return append(buf, encoded...), nil
}

// UnmarshalFrom unmarshals data with optional scratch buffer
func (c *OptimizedCodec[T]) UnmarshalFrom(data []byte, v *T, scratch []byte) error {
	// For JSON, we don't need scratch buffer as json.Unmarshal doesn't benefit from it
	// But we keep the signature for interface compatibility
	return json.Unmarshal(data, v)
}
