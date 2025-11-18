package toml

import (
	"bytes"
	"io"

	"github.com/BurntSushi/toml"
)

// Codec implements the codec.Codec interface for TOML serialization
type Codec[T any] struct{}

// New creates a new TOML codec
func New[T any]() *Codec[T] {
	return &Codec[T]{}
}

// Encode serializes the given data to the writer using TOML
func (c *Codec[T]) Encode(w io.Writer, data T) error {
	encoder := toml.NewEncoder(w)
	return encoder.Encode(data)
}

// Decode deserializes TOML data from the reader into the provided type
func (c *Codec[T]) Decode(r io.Reader, data *T) error {
	decoder := toml.NewDecoder(r)
	_, err := decoder.Decode(data)
	return err
}

// Marshal serializes the given data to TOML bytes
func (c *Codec[T]) Marshal(data T) ([]byte, error) {
	var buf bytes.Buffer
	encoder := toml.NewEncoder(&buf)
	if err := encoder.Encode(data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Unmarshal deserializes TOML bytes into the provided type
func (c *Codec[T]) Unmarshal(data []byte, v *T) error {
	return toml.Unmarshal(data, v)
}
