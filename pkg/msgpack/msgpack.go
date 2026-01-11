//go:build codec_msgpack

package msgpack

import (
	"io"

	codec "github.com/jeremyhahn/go-codec"
	"github.com/vmihailenco/msgpack/v5"
)

func init() {
	codec.RegisterCodec(codec.MsgPack)
}

// Codec implements the codec.Codec interface for MessagePack serialization
type Codec[T any] struct{}

// New creates a new MessagePack codec
func New[T any]() *Codec[T] {
	return &Codec[T]{}
}

// Encode serializes the given data to the writer using MessagePack
func (c *Codec[T]) Encode(w io.Writer, data T) error {
	encoder := msgpack.NewEncoder(w)
	return encoder.Encode(data)
}

// Decode deserializes MessagePack data from the reader into the provided type
func (c *Codec[T]) Decode(r io.Reader, data *T) error {
	decoder := msgpack.NewDecoder(r)
	return decoder.Decode(data)
}

// Marshal serializes the given data to MessagePack bytes
func (c *Codec[T]) Marshal(data T) ([]byte, error) {
	return msgpack.Marshal(data)
}

// Unmarshal deserializes MessagePack bytes into the provided type
func (c *Codec[T]) Unmarshal(data []byte, v *T) error {
	return msgpack.Unmarshal(data, v)
}
