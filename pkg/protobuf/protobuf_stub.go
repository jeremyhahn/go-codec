//go:build !codec_protobuf

package protobuf

import (
	"io"

	codec "github.com/jeremyhahn/go-codec"
)

var errNotSupported = codec.ErrCodecNotSupported{CodecType: codec.ProtoBuf}

// ProtoMessage is a constraint for proto.Message compatibility.
// When protobuf is not compiled in, this is an empty interface.
type ProtoMessage interface{}

// Codec is a stub that returns errors when Protocol Buffers codec is not compiled in.
type Codec[T ProtoMessage] struct{}

// New returns a Protocol Buffers codec stub that will error on all operations.
func New[T ProtoMessage]() *Codec[T] {
	return &Codec[T]{}
}

// Encode returns an error indicating Protocol Buffers codec is not supported.
func (c *Codec[T]) Encode(w io.Writer, data T) error {
	return errNotSupported
}

// Decode returns an error indicating Protocol Buffers codec is not supported.
func (c *Codec[T]) Decode(r io.Reader, data *T) error {
	return errNotSupported
}

// Marshal returns an error indicating Protocol Buffers codec is not supported.
func (c *Codec[T]) Marshal(data T) ([]byte, error) {
	return nil, errNotSupported
}

// Unmarshal returns an error indicating Protocol Buffers codec is not supported.
func (c *Codec[T]) Unmarshal(data []byte, v *T) error {
	return errNotSupported
}
