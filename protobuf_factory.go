package codec

import (
	protobufcodec "github.com/jeremyhahn/go-codec/pkg/protobuf"
)

// NewProtoBuf creates a new Protocol Buffers codec
// Note: T must be a protobuf-generated type that implements proto.Message
func NewProtoBuf[T protobufcodec.ProtoMessage]() Codec[T] {
	return protobufcodec.New[T]()
}
