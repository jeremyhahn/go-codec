// Package factory is the "every codec" convenience import: it constructs any
// codec by its codec.Type. Importing it links and registers all seven codec
// packages (json, yaml, toml, msgpack, bson, cbor and protobuf), together with
// their third-party dependencies. A program that needs only some formats
// should import those codec packages directly (for example
// github.com/jeremyhahn/go-codec/pkg/json) instead, and links only those.
package factory

import (
	"fmt"

	"github.com/jeremyhahn/go-codec"
	bsoncodec "github.com/jeremyhahn/go-codec/pkg/bson"
	cborcodec "github.com/jeremyhahn/go-codec/pkg/cbor"
	jsoncodec "github.com/jeremyhahn/go-codec/pkg/json"
	msgpackcodec "github.com/jeremyhahn/go-codec/pkg/msgpack"
	protobufcodec "github.com/jeremyhahn/go-codec/pkg/protobuf"
	tomlcodec "github.com/jeremyhahn/go-codec/pkg/toml"
	yamlcodec "github.com/jeremyhahn/go-codec/pkg/yaml"
)

// New creates a new codec of the specified type.
// Returns an error if the codec type is not registered.
//
// Note: For Protocol Buffers, use NewProtoBuf instead as it requires
// types that implement proto.Message.
//
// Use codec.IsSupported() to check if a codec is available before calling this.
// Use codec.SupportedCodecs() to get a list of all available codecs.
func New[T any](codecType codec.Type) (codec.Codec[T], error) {
	// Reject types no codec package registered (unknown types).
	if !codec.IsSupported(codecType) {
		return nil, codec.ErrCodecNotSupported{CodecType: codecType}
	}

	switch codecType {
	case codec.JSON:
		return jsoncodec.New[T](), nil
	case codec.YAML:
		return yamlcodec.New[T](), nil
	case codec.TOML:
		return tomlcodec.New[T](), nil
	case codec.MsgPack:
		return msgpackcodec.New[T](), nil
	case codec.BSON:
		return bsoncodec.New[T](), nil
	case codec.CBOR:
		return cborcodec.New[T](), nil
	case codec.ProtoBuf:
		return nil, fmt.Errorf("use NewProtoBuf for Protocol Buffers (requires proto.Message)")
	default:
		return nil, fmt.Errorf("unsupported codec type: %s", codecType)
	}
}

// NewProtoBuf creates a new Protocol Buffers codec.
// T must be a protobuf-generated type that implements proto.Message.
// Importing this package registers the protobuf codec, so the error is always
// nil; the error return is kept for API compatibility.
func NewProtoBuf[T protobufcodec.ProtoMessage]() (codec.Codec[T], error) {
	return protobufcodec.New[T](), nil
}
