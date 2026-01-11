//go:build codec_json && codec_yaml && codec_toml && codec_msgpack && codec_bson && codec_cbor && codec_avro && codec_protobuf

package factory

import (
	"bytes"
	"testing"

	"github.com/jeremyhahn/go-codec"
	"github.com/jeremyhahn/go-codec/pkg/protobuf/testdata"
)

type TestData struct {
	Name  string `json:"name" yaml:"name" toml:"name" msgpack:"name" avro:"name"`
	Value int    `json:"value" yaml:"value" toml:"value" msgpack:"value" avro:"value"`
}

func TestNew_JSON(t *testing.T) {
	c, err := New[TestData](codec.JSON)
	if err != nil {
		t.Fatalf("Failed to create JSON codec: %v", err)
	}

	data := TestData{Name: "test", Value: 42}
	encoded, err := c.Marshal(data)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded TestData
	if err := c.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if decoded.Name != data.Name || decoded.Value != data.Value {
		t.Errorf("Data mismatch: got %+v, want %+v", decoded, data)
	}
}

func TestNew_YAML(t *testing.T) {
	c, err := New[TestData](codec.YAML)
	if err != nil {
		t.Fatalf("Failed to create YAML codec: %v", err)
	}

	data := TestData{Name: "test", Value: 42}
	encoded, err := c.Marshal(data)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded TestData
	if err := c.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if decoded.Name != data.Name || decoded.Value != data.Value {
		t.Errorf("Data mismatch: got %+v, want %+v", decoded, data)
	}
}

func TestNew_TOML(t *testing.T) {
	c, err := New[TestData](codec.TOML)
	if err != nil {
		t.Fatalf("Failed to create TOML codec: %v", err)
	}

	data := TestData{Name: "test", Value: 42}
	encoded, err := c.Marshal(data)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded TestData
	if err := c.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if decoded.Name != data.Name || decoded.Value != data.Value {
		t.Errorf("Data mismatch: got %+v, want %+v", decoded, data)
	}
}

func TestNew_MsgPack(t *testing.T) {
	c, err := New[TestData](codec.MsgPack)
	if err != nil {
		t.Fatalf("Failed to create MsgPack codec: %v", err)
	}

	data := TestData{Name: "test", Value: 42}
	encoded, err := c.Marshal(data)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded TestData
	if err := c.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if decoded.Name != data.Name || decoded.Value != data.Value {
		t.Errorf("Data mismatch: got %+v, want %+v", decoded, data)
	}
}

func TestNew_BSON(t *testing.T) {
	c, err := New[TestData](codec.BSON)
	if err != nil {
		t.Fatalf("Failed to create BSON codec: %v", err)
	}

	data := TestData{Name: "test", Value: 42}
	encoded, err := c.Marshal(data)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded TestData
	if err := c.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if decoded.Name != data.Name || decoded.Value != data.Value {
		t.Errorf("Data mismatch: got %+v, want %+v", decoded, data)
	}
}

func TestNew_CBOR(t *testing.T) {
	c, err := New[TestData](codec.CBOR)
	if err != nil {
		t.Fatalf("Failed to create CBOR codec: %v", err)
	}

	data := TestData{Name: "test", Value: 42}
	encoded, err := c.Marshal(data)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded TestData
	if err := c.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if decoded.Name != data.Name || decoded.Value != data.Value {
		t.Errorf("Data mismatch: got %+v, want %+v", decoded, data)
	}
}

func TestNew_Avro(t *testing.T) {
	c, err := New[TestData](codec.Avro)
	if err != nil {
		t.Fatalf("Failed to create Avro codec: %v", err)
	}

	data := TestData{Name: "test", Value: 42}
	encoded, err := c.Marshal(data)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded TestData
	if err := c.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if decoded.Name != data.Name || decoded.Value != data.Value {
		t.Errorf("Data mismatch: got %+v, want %+v", decoded, data)
	}
}

func TestNew_ProtoBuf_Error(t *testing.T) {
	_, err := New[TestData](codec.ProtoBuf)
	if err == nil {
		t.Fatal("Expected error for ProtoBuf type")
	}

	expectedMsg := "use NewProtoBuf for Protocol Buffers"
	if !bytes.Contains([]byte(err.Error()), []byte(expectedMsg)) {
		t.Errorf("Expected error containing %q, got %q", expectedMsg, err.Error())
	}
}

func TestNew_Unsupported(t *testing.T) {
	_, err := New[TestData]("unsupported")
	if err == nil {
		t.Fatal("Expected error for unsupported codec type")
	}

	// The error should indicate the codec is not supported
	expectedMsg := "not supported"
	if !bytes.Contains([]byte(err.Error()), []byte(expectedMsg)) {
		t.Errorf("Expected error containing %q, got %q", expectedMsg, err.Error())
	}
}

func TestNewProtoBuf(t *testing.T) {
	c, err := NewProtoBuf[*testdata.TestMessage]()
	if err != nil {
		t.Fatalf("Failed to create ProtoBuf codec: %v", err)
	}

	data := &testdata.TestMessage{
		Name:  "John Doe",
		Age:   30,
		Email: "john@example.com",
	}

	encoded, err := c.Marshal(data)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	decoded := &testdata.TestMessage{}
	if err := c.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if decoded.Name != data.Name || decoded.Age != data.Age || decoded.Email != data.Email {
		t.Errorf("Data mismatch: got %+v, want %+v", decoded, data)
	}
}

func TestNew_EncodeDecodeStream(t *testing.T) {
	c, err := New[TestData](codec.JSON)
	if err != nil {
		t.Fatalf("Failed to create JSON codec: %v", err)
	}

	data := TestData{Name: "stream-test", Value: 123}

	var buf bytes.Buffer
	if err := c.Encode(&buf, data); err != nil {
		t.Fatalf("Failed to encode: %v", err)
	}

	var decoded TestData
	if err := c.Decode(&buf, &decoded); err != nil {
		t.Fatalf("Failed to decode: %v", err)
	}

	if decoded.Name != data.Name || decoded.Value != data.Value {
		t.Errorf("Data mismatch: got %+v, want %+v", decoded, data)
	}
}

func TestSupportedCodecs(t *testing.T) {
	// Get supported codecs - since we're building with all codec tags,
	// all 8 codecs should be supported
	supported := codec.SupportedCodecs()

	// Verify we have at least one codec
	if len(supported) == 0 {
		t.Error("Expected at least one supported codec")
	}

	// Log supported codecs for debugging
	t.Logf("Supported codecs: %v", supported)

	// Verify IsSupported works for each
	for _, c := range supported {
		if !codec.IsSupported(c) {
			t.Errorf("IsSupported(%q) returned false but codec is in SupportedCodecs()", c)
		}
	}
}

func TestNew_RegisteredButUnhandledCodec(t *testing.T) {
	// This test covers the default case in the switch statement.
	// We register a custom codec type that isn't handled in the factory switch.
	customCodec := codec.Type("custom_test_codec")

	// Register the custom codec (simulating a new codec being added)
	codec.RegisterCodec(customCodec)

	// Now IsSupported returns true, but there's no case for it in the switch
	_, err := New[TestData](customCodec)
	if err == nil {
		t.Fatal("Expected error for unhandled codec type")
	}

	expectedMsg := "unsupported codec type"
	if !bytes.Contains([]byte(err.Error()), []byte(expectedMsg)) {
		t.Errorf("Expected error containing %q, got %q", expectedMsg, err.Error())
	}
}
