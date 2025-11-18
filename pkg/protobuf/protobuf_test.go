package protobuf

import (
	"bytes"
	"testing"

	"github.com/jeremyhahn/go-codec/pkg/protobuf/testdata"
)

func TestNew(t *testing.T) {
	codec := New[*testdata.TestMessage]()
	if codec == nil {
		t.Fatal("expected non-nil codec")
	}
}

func TestCodec_Marshal(t *testing.T) {
	codec := New[*testdata.TestMessage]()
	data := &testdata.TestMessage{
		Name:  "John Doe",
		Age:   30,
		Email: "john@example.com",
	}

	result, err := codec.Marshal(data)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	if len(result) == 0 {
		t.Error("expected non-empty result")
	}

	// Verify we can unmarshal it back
	unmarshaled := &testdata.TestMessage{}
	err = codec.Unmarshal(result, &unmarshaled)
	if err != nil {
		t.Fatalf("Unmarshal verification failed: %v", err)
	}

	if unmarshaled.Name != data.Name {
		t.Errorf("expected name '%s', got '%s'", data.Name, unmarshaled.Name)
	}
	if unmarshaled.Age != data.Age {
		t.Errorf("expected age %d, got %d", data.Age, unmarshaled.Age)
	}
	if unmarshaled.Email != data.Email {
		t.Errorf("expected email '%s', got '%s'", data.Email, unmarshaled.Email)
	}
}

func TestCodec_Unmarshal(t *testing.T) {
	codec := New[*testdata.TestMessage]()
	data := &testdata.TestMessage{
		Name:  "Jane Doe",
		Age:   25,
		Email: "jane@example.com",
	}

	// First marshal the data
	protobufData, err := codec.Marshal(data)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Now unmarshal it
	result := &testdata.TestMessage{}
	err = codec.Unmarshal(protobufData, &result)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if result.Name != "Jane Doe" {
		t.Errorf("expected name 'Jane Doe', got '%s'", result.Name)
	}
	if result.Age != 25 {
		t.Errorf("expected age 25, got %d", result.Age)
	}
	if result.Email != "jane@example.com" {
		t.Errorf("expected email 'jane@example.com', got '%s'", result.Email)
	}
}

func TestCodec_Unmarshal_Invalid(t *testing.T) {
	codec := New[*testdata.TestMessage]()
	protobufData := []byte{0xff, 0xff, 0xff} // Invalid protobuf data

	result := &testdata.TestMessage{}
	err := codec.Unmarshal(protobufData, &result)
	if err == nil {
		t.Fatal("expected error for invalid protobuf, got nil")
	}
}

func TestCodec_Encode(t *testing.T) {
	codec := New[*testdata.TestMessage]()
	data := &testdata.TestMessage{
		Name:  "Bob Smith",
		Age:   35,
		Email: "bob@example.com",
	}

	var buf bytes.Buffer
	err := codec.Encode(&buf, data)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	if buf.Len() == 0 {
		t.Error("expected non-empty buffer")
	}

	// Verify we can decode it back
	decoded := &testdata.TestMessage{}
	err = codec.Decode(&buf, &decoded)
	if err != nil {
		t.Fatalf("Decode verification failed: %v", err)
	}

	if decoded.Name != data.Name {
		t.Errorf("expected name '%s', got '%s'", data.Name, decoded.Name)
	}
	if decoded.Age != data.Age {
		t.Errorf("expected age %d, got %d", data.Age, decoded.Age)
	}
	if decoded.Email != data.Email {
		t.Errorf("expected email '%s', got '%s'", data.Email, decoded.Email)
	}
}

func TestCodec_Decode(t *testing.T) {
	codec := New[*testdata.TestMessage]()
	data := &testdata.TestMessage{
		Name:  "Alice Johnson",
		Age:   28,
		Email: "alice@example.com",
	}

	// First encode the data
	var buf bytes.Buffer
	err := codec.Encode(&buf, data)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	// Now decode it
	result := &testdata.TestMessage{}
	err = codec.Decode(&buf, &result)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if result.Name != "Alice Johnson" {
		t.Errorf("expected name 'Alice Johnson', got '%s'", result.Name)
	}
	if result.Age != 28 {
		t.Errorf("expected age 28, got %d", result.Age)
	}
	if result.Email != "alice@example.com" {
		t.Errorf("expected email 'alice@example.com', got '%s'", result.Email)
	}
}

func TestCodec_Decode_Invalid(t *testing.T) {
	codec := New[*testdata.TestMessage]()
	buf := bytes.NewBuffer([]byte{0xff, 0xff, 0xff}) // Invalid protobuf data

	result := &testdata.TestMessage{}
	err := codec.Decode(buf, &result)
	if err == nil {
		t.Fatal("expected error for invalid protobuf, got nil")
	}
}

func TestCodec_Decode_ReadError(t *testing.T) {
	codec := New[*testdata.TestMessage]()
	// Use an error reader
	errorReader := &errorReader{}

	result := &testdata.TestMessage{}
	err := codec.Decode(errorReader, &result)
	if err == nil {
		t.Fatal("expected error from reader, got nil")
	}
}

// errorReader is a reader that always returns an error
type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, bytes.ErrTooLarge
}

func TestCodec_Encode_WriteError(t *testing.T) {
	codec := New[*testdata.TestMessage]()
	data := &testdata.TestMessage{
		Name:  "Test",
		Age:   30,
		Email: "test@example.com",
	}

	// Use an error writer
	errorWriter := &errorWriter{}

	err := codec.Encode(errorWriter, data)
	if err == nil {
		t.Fatal("expected error from writer, got nil")
	}
}

// errorWriter is a writer that always returns an error
type errorWriter struct{}

func (e *errorWriter) Write(p []byte) (n int, err error) {
	return 0, bytes.ErrTooLarge
}

func TestCodec_WithDifferentTypes(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		codec := New[*testdata.StringValue]()
		data := &testdata.StringValue{Value: "hello world"}

		marshaled, err := codec.Marshal(data)
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}

		result := &testdata.StringValue{}
		err = codec.Unmarshal(marshaled, &result)
		if err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		if result.Value != data.Value {
			t.Errorf("expected %s, got %s", data.Value, result.Value)
		}
	})

	t.Run("int", func(t *testing.T) {
		codec := New[*testdata.IntValue]()
		data := &testdata.IntValue{Value: 42}

		marshaled, err := codec.Marshal(data)
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}

		result := &testdata.IntValue{}
		err = codec.Unmarshal(marshaled, &result)
		if err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		if result.Value != data.Value {
			t.Errorf("expected %d, got %d", data.Value, result.Value)
		}
	})

	t.Run("slice", func(t *testing.T) {
		codec := New[*testdata.StringList]()
		data := &testdata.StringList{Values: []string{"one", "two", "three"}}

		marshaled, err := codec.Marshal(data)
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}

		result := &testdata.StringList{}
		err = codec.Unmarshal(marshaled, &result)
		if err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		if len(result.Values) != len(data.Values) {
			t.Errorf("expected length %d, got %d", len(data.Values), len(result.Values))
		}
		for i, v := range data.Values {
			if result.Values[i] != v {
				t.Errorf("expected %s at index %d, got %s", v, i, result.Values[i])
			}
		}
	})

	t.Run("map", func(t *testing.T) {
		codec := New[*testdata.IntMap]()
		data := &testdata.IntMap{Values: map[string]int32{"one": 1, "two": 2, "three": 3}}

		marshaled, err := codec.Marshal(data)
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}

		result := &testdata.IntMap{}
		err = codec.Unmarshal(marshaled, &result)
		if err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		if len(result.Values) != len(data.Values) {
			t.Errorf("expected length %d, got %d", len(data.Values), len(result.Values))
		}
		for k, v := range data.Values {
			if result.Values[k] != v {
				t.Errorf("expected %d for key %s, got %d", v, k, result.Values[k])
			}
		}
	})
}
