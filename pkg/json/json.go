//go:build codec_json

package json

import (
	"encoding/json"
	"io"

	codec "github.com/jeremyhahn/go-codec"
)

func init() {
	codec.RegisterCodec(codec.JSON)
}

// Codec implements the codec.Codec interface for JSON serialization
type Codec[T any] struct{}

// New creates a new JSON codec
func New[T any]() *Codec[T] {
	return &Codec[T]{}
}

// Encode serializes the given data to the writer using JSON
func (c *Codec[T]) Encode(w io.Writer, data T) error {
	encoder := json.NewEncoder(w)
	return encoder.Encode(data)
}

// Decode deserializes JSON data from the reader into the provided type
func (c *Codec[T]) Decode(r io.Reader, data *T) error {
	decoder := json.NewDecoder(r)
	return decoder.Decode(data)
}

// Marshal serializes the given data to JSON bytes
func (c *Codec[T]) Marshal(data T) ([]byte, error) {
	return json.Marshal(data)
}

// Unmarshal deserializes JSON bytes into the provided type
func (c *Codec[T]) Unmarshal(data []byte, v *T) error {
	return json.Unmarshal(data, v)
}
