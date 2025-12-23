package factory

import (
	"fmt"

	"github.com/jeremyhahn/go-codec"
	avrocodec "github.com/jeremyhahn/go-codec/pkg/avro"
	bsoncodec "github.com/jeremyhahn/go-codec/pkg/bson"
	cborcodec "github.com/jeremyhahn/go-codec/pkg/cbor"
	jsoncodec "github.com/jeremyhahn/go-codec/pkg/json"
	msgpackcodec "github.com/jeremyhahn/go-codec/pkg/msgpack"
	protobufcodec "github.com/jeremyhahn/go-codec/pkg/protobuf"
	tomlcodec "github.com/jeremyhahn/go-codec/pkg/toml"
	yamlcodec "github.com/jeremyhahn/go-codec/pkg/yaml"
)

// New creates a new codec of the specified type.
// Returns an error if the codec type is not supported.
//
// Note: For Protocol Buffers, use NewProtoBuf instead as it requires
// types that implement proto.Message.
func New[T any](codecType codec.Type) (codec.Codec[T], error) {
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
	case codec.Avro:
		return avrocodec.New[T](), nil
	case codec.ProtoBuf:
		return nil, fmt.Errorf("use NewProtoBuf for Protocol Buffers (requires proto.Message)")
	default:
		return nil, fmt.Errorf("unsupported codec type: %s", codecType)
	}
}

// NewProtoBuf creates a new Protocol Buffers codec.
// T must be a protobuf-generated type that implements proto.Message.
func NewProtoBuf[T protobufcodec.ProtoMessage]() codec.Codec[T] {
	return protobufcodec.New[T]()
}
