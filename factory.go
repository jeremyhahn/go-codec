package codec

import (
	"fmt"

	bsoncodec "github.com/jeremyhahn/go-codec/pkg/bson"
	jsoncodec "github.com/jeremyhahn/go-codec/pkg/json"
	msgpackcodec "github.com/jeremyhahn/go-codec/pkg/msgpack"
	tomlcodec "github.com/jeremyhahn/go-codec/pkg/toml"
	yamlcodec "github.com/jeremyhahn/go-codec/pkg/yaml"
)

// New creates a new codec of the specified type
func New[T any](codecType Type) (Codec[T], error) {
	switch codecType {
	case JSON:
		return jsoncodec.New[T](), nil
	case YAML:
		return yamlcodec.New[T](), nil
	case TOML:
		return tomlcodec.New[T](), nil
	case MsgPack:
		return msgpackcodec.New[T](), nil
	case BSON:
		return bsoncodec.New[T](), nil
	default:
		return nil, fmt.Errorf("unsupported codec type: %s", codecType)
	}
}
