package protobuf

import (
	"io"

	"google.golang.org/protobuf/proto"
)

// ProtoMessage is a constraint that requires types to implement proto.Message
type ProtoMessage interface {
	proto.Message
}

// Codec implements the codec.Codec interface for Protocol Buffers serialization
type Codec[T ProtoMessage] struct{}

// New creates a new Protocol Buffers codec
func New[T ProtoMessage]() *Codec[T] {
	return &Codec[T]{}
}

// Encode serializes the given data to the writer using Protocol Buffers
func (c *Codec[T]) Encode(w io.Writer, data T) error {
	bytes, err := proto.Marshal(data)
	if err == nil {
		_, err = w.Write(bytes)
	}
	return err
}

// Decode deserializes Protocol Buffers data from the reader into the provided type
func (c *Codec[T]) Decode(r io.Reader, data *T) error {
	bytes, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	return proto.Unmarshal(bytes, *data)
}

// Marshal serializes the given data to Protocol Buffers bytes
func (c *Codec[T]) Marshal(data T) ([]byte, error) {
	return proto.Marshal(data)
}

// Unmarshal deserializes Protocol Buffers bytes into the provided type
func (c *Codec[T]) Unmarshal(data []byte, v *T) error {
	return proto.Unmarshal(data, *v)
}
