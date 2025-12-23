package json

import (
	"testing"
)

func TestOptimizedCodec_MarshalTo(t *testing.T) {
	codec := NewPool[TestStruct]()
	data := TestStruct{
		Name:  "John Doe",
		Age:   30,
		Email: "john@example.com",
	}

	buf := make([]byte, 0, 256)
	result, err := codec.MarshalTo(buf, data)
	if err != nil {
		t.Fatalf("MarshalTo failed: %v", err)
	}

	// Verify result is using the provided buffer
	if cap(result) != cap(buf) {
		t.Errorf("expected result to use provided buffer, got different capacity")
	}

	// Verify we can unmarshal it back
	var unmarshaled TestStruct
	err = codec.Unmarshal(result, &unmarshaled)
	if err != nil {
		t.Fatalf("Unmarshal verification failed: %v", err)
	}

	if unmarshaled.Name != data.Name {
		t.Errorf("expected name '%s', got '%s'", data.Name, unmarshaled.Name)
	}
}

func TestOptimizedCodec_MarshalTo_NilBuffer(t *testing.T) {
	codec := NewPool[TestStruct]()
	data := TestStruct{
		Name:  "John Doe",
		Age:   30,
		Email: "john@example.com",
	}

	result, err := codec.MarshalTo(nil, data)
	if err != nil {
		t.Fatalf("MarshalTo with nil buffer failed: %v", err)
	}

	if len(result) == 0 {
		t.Error("expected non-empty result")
	}
}

func TestOptimizedCodec_AppendMarshal(t *testing.T) {
	codec := NewPool[TestStruct]()
	data := TestStruct{
		Name:  "John Doe",
		Age:   30,
		Email: "john@example.com",
	}

	buf := []byte("prefix:")
	result, err := codec.AppendMarshal(buf, data)
	if err != nil {
		t.Fatalf("AppendMarshal failed: %v", err)
	}

	// Verify prefix is still there
	if string(result[:7]) != "prefix:" {
		t.Errorf("expected prefix to be preserved, got %s", string(result[:7]))
	}
}

func TestOptimizedCodec_UnmarshalFrom(t *testing.T) {
	codec := NewPool[TestStruct]()
	data := TestStruct{
		Name:  "Jane Doe",
		Age:   25,
		Email: "jane@example.com",
	}

	marshaled, err := codec.Marshal(data)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var result TestStruct
	scratch := make([]byte, 256)
	err = codec.UnmarshalFrom(marshaled, &result, scratch)
	if err != nil {
		t.Fatalf("UnmarshalFrom failed: %v", err)
	}

	if result.Name != data.Name {
		t.Errorf("expected name '%s', got '%s'", data.Name, result.Name)
	}
	if result.Age != data.Age {
		t.Errorf("expected age %d, got %d", data.Age, result.Age)
	}
}

func TestOptimizedCodec_InvalidData(t *testing.T) {
	codec := NewPool[TestStruct]()
	invalidData := []byte(`{invalid json}`)

	var result TestStruct
	scratch := make([]byte, 256)
	err := codec.UnmarshalFrom(invalidData, &result, scratch)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestOptimizedCodec_BufferReuse(t *testing.T) {
	codec := NewPool[TestStruct]()
	data := TestStruct{
		Name:  "Test User",
		Age:   30,
		Email: "test@example.com",
	}

	// Use the same buffer multiple times
	buf := make([]byte, 0, 256)
	for i := 0; i < 10; i++ {
		buf = buf[:0] // Reset length
		result, err := codec.MarshalTo(buf, data)
		if err != nil {
			t.Fatalf("MarshalTo iteration %d failed: %v", i, err)
		}

		// Verify buffer is being reused
		if cap(result) != cap(buf) {
			t.Errorf("iteration %d: expected buffer to be reused", i)
		}
	}
}

// unmarshalableType is a type that cannot be marshaled to JSON
type unmarshalableType struct {
	Ch chan int `json:"ch"`
}

func TestOptimizedCodec_MarshalTo_Error(t *testing.T) {
	codec := NewPool[unmarshalableType]()
	data := unmarshalableType{Ch: make(chan int)}

	buf := make([]byte, 0, 256)
	_, err := codec.MarshalTo(buf, data)
	if err == nil {
		t.Fatal("expected error for unmarshalable type, got nil")
	}
}

func TestOptimizedCodec_MarshalTo_SmallBuffer(t *testing.T) {
	codec := NewPool[TestStruct]()
	data := TestStruct{
		Name:  "John Doe",
		Age:   30,
		Email: "john@example.com",
	}

	// Use a buffer that is too small to hold the result
	buf := make([]byte, 0, 1)
	result, err := codec.MarshalTo(buf, data)
	if err != nil {
		t.Fatalf("MarshalTo failed: %v", err)
	}

	// Result should be a new buffer, not the provided one
	if cap(result) == cap(buf) {
		t.Error("expected result to use new buffer when provided buffer is too small")
	}

	// Verify we can unmarshal it back
	var unmarshaled TestStruct
	err = codec.Unmarshal(result, &unmarshaled)
	if err != nil {
		t.Fatalf("Unmarshal verification failed: %v", err)
	}

	if unmarshaled.Name != data.Name {
		t.Errorf("expected name '%s', got '%s'", data.Name, unmarshaled.Name)
	}
}

func TestOptimizedCodec_AppendMarshal_Error(t *testing.T) {
	codec := NewPool[unmarshalableType]()
	data := unmarshalableType{Ch: make(chan int)}

	buf := []byte("prefix:")
	result, err := codec.AppendMarshal(buf, data)
	if err == nil {
		t.Fatal("expected error for unmarshalable type, got nil")
	}

	// Verify original buffer is returned on error
	if string(result) != "prefix:" {
		t.Errorf("expected original buffer on error, got %s", string(result))
	}
}
