//go:build codec_yaml

package yaml

import (
	"io"

	codec "github.com/jeremyhahn/go-codec"
	"gopkg.in/yaml.v3"
)

func init() {
	codec.RegisterCodec(codec.YAML)
}

// Codec implements the codec.Codec interface for YAML serialization
type Codec[T any] struct{}

// New creates a new YAML codec
func New[T any]() *Codec[T] {
	return &Codec[T]{}
}

// Encode serializes the given data to the writer using YAML
func (c *Codec[T]) Encode(w io.Writer, data T) error {
	encoder := yaml.NewEncoder(w)
	if err := encoder.Encode(data); err != nil {
		_ = encoder.Close()
		return err
	}
	return encoder.Close()
}

// Decode deserializes YAML data from the reader into the provided type
func (c *Codec[T]) Decode(r io.Reader, data *T) error {
	decoder := yaml.NewDecoder(r)
	return decoder.Decode(data)
}

// Marshal serializes the given data to YAML bytes
func (c *Codec[T]) Marshal(data T) ([]byte, error) {
	return yaml.Marshal(data)
}

// Unmarshal deserializes YAML bytes into the provided type
func (c *Codec[T]) Unmarshal(data []byte, v *T) error {
	return yaml.Unmarshal(data, v)
}
