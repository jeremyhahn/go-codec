//go:build codec_json && !codec_protobuf

package factory

import (
	"bytes"
	"testing"

	"github.com/jeremyhahn/go-codec"
)

// DummyMessage is a type that satisfies protobufcodec.ProtoMessage (empty interface when stub)
type DummyMessage struct {
	Name string
}

// TestNewProtoBuf_NotCompiled tests the error path when protobuf codec is not compiled in.
// This test only runs when codec_protobuf build tag is NOT set.
func TestNewProtoBuf_NotCompiled(t *testing.T) {
	// Verify that protobuf is not supported in this build
	if codec.IsSupported(codec.ProtoBuf) {
		t.Skip("This test only runs when protobuf codec is NOT compiled in")
	}

	// When protobuf is not compiled in, NewProtoBuf should return ErrCodecNotSupported
	_, err := NewProtoBuf[*DummyMessage]()
	if err == nil {
		t.Fatal("Expected error when protobuf codec is not compiled in")
	}

	expectedMsg := "codec \"protobuf\" is not supported"
	if !bytes.Contains([]byte(err.Error()), []byte(expectedMsg)) {
		t.Errorf("Expected error containing %q, got %q", expectedMsg, err.Error())
	}
}
