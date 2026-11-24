// Package factory constructs any codec by codec.Type and registers all seven codecs; import single codec packages to link only those.
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

// New creates a codec of type t, or an error if it is not registered; use NewProtoBuf for Protocol Buffers.
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

// NewProtoBuf creates a Protocol Buffers codec for a proto.Message type; the error is always nil.
func NewProtoBuf[T protobufcodec.ProtoMessage]() (codec.Codec[T], error) {
	return protobufcodec.New[T](), nil
}
