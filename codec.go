package codec

import "io"

// Type represents the supported codec types
type Type string

const (
	JSON     Type = "json"
	YAML     Type = "yaml"
	TOML     Type = "toml"
	MsgPack  Type = "msgpack"
	ProtoBuf Type = "protobuf"
	BSON     Type = "bson"
)

// Codec defines the interface for encoding and decoding operations
type Codec[T any] interface {
	// Encode serializes the given data to the writer
	Encode(w io.Writer, data T) error

	// Decode deserializes data from the reader into the provided type
	Decode(r io.Reader, data *T) error

	// Marshal serializes the given data to bytes
	Marshal(data T) ([]byte, error)

	// Unmarshal deserializes bytes into the provided type
	Unmarshal(data []byte, v *T) error
}
