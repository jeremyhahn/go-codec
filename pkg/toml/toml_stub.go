//go:build !codec_toml

package toml

import (
	"io"

	codec "github.com/jeremyhahn/go-codec"
)

var errNotSupported = codec.ErrCodecNotSupported{CodecType: codec.TOML}

// Codec is a stub that returns errors when TOML codec is not compiled in.
type Codec[T any] struct{}

// New returns a TOML codec stub that will error on all operations.
func New[T any]() *Codec[T] {
	return &Codec[T]{}
}

// Encode returns an error indicating TOML codec is not supported.
func (c *Codec[T]) Encode(w io.Writer, data T) error {
	return errNotSupported
}

// Decode returns an error indicating TOML codec is not supported.
func (c *Codec[T]) Decode(r io.Reader, data *T) error {
	return errNotSupported
}

// Marshal returns an error indicating TOML codec is not supported.
func (c *Codec[T]) Marshal(data T) ([]byte, error) {
	return nil, errNotSupported
}

// Unmarshal returns an error indicating TOML codec is not supported.
func (c *Codec[T]) Unmarshal(data []byte, v *T) error {
	return errNotSupported
}
