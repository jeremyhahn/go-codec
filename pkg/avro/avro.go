//go:build codec_avro

package avro

import (
	"io"
	"reflect"

	"github.com/hamba/avro/v2"
	codec "github.com/jeremyhahn/go-codec"
)

func init() {
	codec.RegisterCodec(codec.Avro)
}

// Codec implements the codec.Codec interface for Avro serialization
type Codec[T any] struct {
	schema avro.Schema
}

// New creates a new Avro codec with automatic schema inference from the type parameter
func New[T any]() *Codec[T] {
	var zero T
	schema := getOrCreateSchema(reflect.TypeOf(zero))
	return &Codec[T]{
		schema: schema,
	}
}

// NewWithSchema creates a new Avro codec with an explicit schema
func NewWithSchema[T any](schemaJSON string) (*Codec[T], error) {
	schema, err := avro.Parse(schemaJSON)
	if err != nil {
		return nil, err
	}
	return &Codec[T]{schema: schema}, nil
}

// Encode serializes the given data to the writer using Avro
func (c *Codec[T]) Encode(w io.Writer, data T) error {
	encoder := avro.NewEncoderForSchema(c.schema, w)
	return encoder.Encode(data)
}

// Decode deserializes Avro data from the reader into the provided type
func (c *Codec[T]) Decode(r io.Reader, data *T) error {
	decoder := avro.NewDecoderForSchema(c.schema, r)
	return decoder.Decode(data)
}

// Marshal serializes the given data to Avro bytes
func (c *Codec[T]) Marshal(data T) ([]byte, error) {
	return avro.Marshal(c.schema, data)
}

// Unmarshal deserializes Avro bytes into the provided type
func (c *Codec[T]) Unmarshal(data []byte, v *T) error {
	return avro.Unmarshal(c.schema, data, v)
}

// Schema returns the Avro schema used by this codec
func (c *Codec[T]) Schema() avro.Schema {
	return c.schema
}

// SchemaJSON returns the Avro schema as a JSON string
func (c *Codec[T]) SchemaJSON() string {
	return c.schema.String()
}
