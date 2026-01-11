//go:build codec_bson

package bson

import (
	"io"

	codec "github.com/jeremyhahn/go-codec"
	"go.mongodb.org/mongo-driver/bson"
)

func init() {
	codec.RegisterCodec(codec.BSON)
}

// Codec implements the codec.Codec interface for BSON serialization
type Codec[T any] struct{}

// New creates a new BSON codec
func New[T any]() *Codec[T] {
	return &Codec[T]{}
}

// Encode serializes the given data to the writer using BSON
func (c *Codec[T]) Encode(w io.Writer, data T) error {
	bytes, err := bson.Marshal(data)
	if err == nil {
		_, err = w.Write(bytes)
	}
	return err
}

// Decode deserializes BSON data from the reader into the provided type
func (c *Codec[T]) Decode(r io.Reader, data *T) error {
	bytes, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	return bson.Unmarshal(bytes, data)
}

// Marshal serializes the given data to BSON bytes
func (c *Codec[T]) Marshal(data T) ([]byte, error) {
	return bson.Marshal(data)
}

// Unmarshal deserializes BSON bytes into the provided type
func (c *Codec[T]) Unmarshal(data []byte, v *T) error {
	return bson.Unmarshal(data, v)
}
