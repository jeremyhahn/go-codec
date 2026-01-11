//go:build !codec_avro

package avro

import (
	"io"

	codec "github.com/jeremyhahn/go-codec"
)

var errNotSupported = codec.ErrCodecNotSupported{CodecType: codec.Avro}

// Codec is a stub that returns errors when Avro codec is not compiled in.
type Codec[T any] struct{}

// New returns an Avro codec stub that will error on all operations.
func New[T any]() *Codec[T] {
	return &Codec[T]{}
}

// NewWithSchema returns an Avro codec stub that will error on all operations.
func NewWithSchema[T any](schemaJSON string) (*Codec[T], error) {
	return nil, errNotSupported
}

// Encode returns an error indicating Avro codec is not supported.
func (c *Codec[T]) Encode(w io.Writer, data T) error {
	return errNotSupported
}

// Decode returns an error indicating Avro codec is not supported.
func (c *Codec[T]) Decode(r io.Reader, data *T) error {
	return errNotSupported
}

// Marshal returns an error indicating Avro codec is not supported.
func (c *Codec[T]) Marshal(data T) ([]byte, error) {
	return nil, errNotSupported
}

// Unmarshal returns an error indicating Avro codec is not supported.
func (c *Codec[T]) Unmarshal(data []byte, v *T) error {
	return errNotSupported
}

// Schema returns nil when Avro codec is not supported.
func (c *Codec[T]) Schema() interface{} {
	return nil
}

// SchemaJSON returns an empty string when Avro codec is not supported.
func (c *Codec[T]) SchemaJSON() string {
	return ""
}
