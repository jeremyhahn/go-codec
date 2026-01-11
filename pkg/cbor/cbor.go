//go:build codec_cbor

package cbor

import (
	"io"

	"github.com/fxamacker/cbor/v2"
	codec "github.com/jeremyhahn/go-codec"
)

func init() {
	codec.RegisterCodec(codec.CBOR)
}

// Codec implements the codec.Codec interface for CBOR serialization
type Codec[T any] struct{}

// New creates a new CBOR codec
func New[T any]() *Codec[T] {
	return &Codec[T]{}
}

// Encode serializes the given data to the writer using CBOR
func (c *Codec[T]) Encode(w io.Writer, data T) error {
	encoder := cbor.NewEncoder(w)
	return encoder.Encode(data)
}

// Decode deserializes CBOR data from the reader into the provided type
func (c *Codec[T]) Decode(r io.Reader, data *T) error {
	decoder := cbor.NewDecoder(r)
	return decoder.Decode(data)
}

// Marshal serializes the given data to CBOR bytes
func (c *Codec[T]) Marshal(data T) ([]byte, error) {
	return cbor.Marshal(data)
}

// Unmarshal deserializes CBOR bytes into the provided type
func (c *Codec[T]) Unmarshal(data []byte, v *T) error {
	return cbor.Unmarshal(data, v)
}
