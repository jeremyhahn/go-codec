//go:build !codec_json

package json

import (
	"io"

	codec "github.com/jeremyhahn/go-codec"
)

var errNotSupported = codec.ErrCodecNotSupported{CodecType: codec.JSON}

// Codec is a stub that returns errors when JSON codec is not compiled in.
type Codec[T any] struct{}

// New returns a JSON codec stub that will error on all operations.
func New[T any]() *Codec[T] {
	return &Codec[T]{}
}

// Encode returns an error indicating JSON codec is not supported.
func (c *Codec[T]) Encode(w io.Writer, data T) error {
	return errNotSupported
}

// Decode returns an error indicating JSON codec is not supported.
func (c *Codec[T]) Decode(r io.Reader, data *T) error {
	return errNotSupported
}

// Marshal returns an error indicating JSON codec is not supported.
func (c *Codec[T]) Marshal(data T) ([]byte, error) {
	return nil, errNotSupported
}

// Unmarshal returns an error indicating JSON codec is not supported.
func (c *Codec[T]) Unmarshal(data []byte, v *T) error {
	return errNotSupported
}

// OptimizedCodec is a stub for the optimized JSON codec.
type OptimizedCodec[T any] struct {
	*Codec[T]
}

// NewPool returns an optimized JSON codec stub that will error on all operations.
func NewPool[T any]() *OptimizedCodec[T] {
	return &OptimizedCodec[T]{
		Codec: New[T](),
	}
}

// MarshalTo returns an error indicating JSON codec is not supported.
func (c *OptimizedCodec[T]) MarshalTo(buf []byte, data T) ([]byte, error) {
	return nil, errNotSupported
}

// AppendMarshal returns an error indicating JSON codec is not supported.
func (c *OptimizedCodec[T]) AppendMarshal(buf []byte, data T) ([]byte, error) {
	return nil, errNotSupported
}

// UnmarshalFrom returns an error indicating JSON codec is not supported.
func (c *OptimizedCodec[T]) UnmarshalFrom(data []byte, v *T, scratch []byte) error {
	return errNotSupported
}
