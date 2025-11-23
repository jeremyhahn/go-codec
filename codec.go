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

// OptimizedCodec extends Codec with zero-allocation methods
type OptimizedCodec[T any] interface {
	Codec[T]

	// MarshalTo marshals data into the provided buffer and returns the used portion.
	// If buf is nil or too small, a new buffer is allocated.
	// The returned slice may be a sub-slice of buf or a newly allocated buffer.
	MarshalTo(buf []byte, data T) ([]byte, error)

	// AppendMarshal appends the marshaled data to buf and returns the extended buffer.
	// This allows for efficient concatenation of multiple marshaled values.
	AppendMarshal(buf []byte, data T) ([]byte, error)

	// UnmarshalFrom unmarshals data using a pre-allocated buffer for temporary storage.
	// The scratch buffer can be reused across multiple calls to reduce allocations.
	UnmarshalFrom(data []byte, v *T, scratch []byte) error
}
