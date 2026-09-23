package msgpack

import (
	"bytes"
	"testing"
)

type TestStruct struct {
	Name  string `msgpack:"name"`
	Age   int    `msgpack:"age"`
	Email string `msgpack:"email"`
}

func TestNew(t *testing.T) {
	codec := New[TestStruct]()
	if codec == nil {
		t.Fatal("expected non-nil codec")
	}
}

func TestCodec_Marshal(t *testing.T) {
	codec := New[TestStruct]()
	data := TestStruct{
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
	var unmarshaled TestStruct
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
	codec := New[TestStruct]()
	data := TestStruct{
		Name:  "Jane Doe",
		Age:   25,
		Email: "jane@example.com",
	}

	// First marshal the data
	msgpackData, err := codec.Marshal(data)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Now unmarshal it
	var result TestStruct
	err = codec.Unmarshal(msgpackData, &result)
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
	codec := New[TestStruct]()
	msgpackData := []byte{0xff, 0xff, 0xff} // Invalid msgpack data

	var result TestStruct
	err := codec.Unmarshal(msgpackData, &result)
	if err == nil {
		t.Fatal("expected error for invalid MessagePack, got nil")
	}
}

func TestCodec_Encode(t *testing.T) {
	codec := New[TestStruct]()
	data := TestStruct{
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
	var decoded TestStruct
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
	codec := New[TestStruct]()
	data := TestStruct{
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
	var result TestStruct
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
	codec := New[TestStruct]()
	buf := bytes.NewBuffer([]byte{0xff, 0xff, 0xff}) // Invalid msgpack data

	var result TestStruct
	err := codec.Decode(buf, &result)
	if err == nil {
		t.Fatal("expected error for invalid MessagePack, got nil")
	}
}

func TestCodec_WithDifferentTypes(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		codec := New[string]()
		data := "hello world"

		marshaled, err := codec.Marshal(data)
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}

		var result string
		err = codec.Unmarshal(marshaled, &result)
		if err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		if result != data {
			t.Errorf("expected %s, got %s", data, result)
		}
	})

	t.Run("int", func(t *testing.T) {
		codec := New[int]()
		data := 42

		marshaled, err := codec.Marshal(data)
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}

		var result int
		err = codec.Unmarshal(marshaled, &result)
		if err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		if result != data {
			t.Errorf("expected %d, got %d", data, result)
		}
	})

	t.Run("slice", func(t *testing.T) {
		codec := New[[]string]()
		data := []string{"one", "two", "three"}

		marshaled, err := codec.Marshal(data)
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}

		var result []string
		err = codec.Unmarshal(marshaled, &result)
		if err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		if len(result) != len(data) {
			t.Errorf("expected length %d, got %d", len(data), len(result))
		}
		for i, v := range data {
			if result[i] != v {
				t.Errorf("expected %s at index %d, got %s", v, i, result[i])
			}
		}
	})

	t.Run("map", func(t *testing.T) {
		codec := New[map[string]int]()
		data := map[string]int{"one": 1, "two": 2, "three": 3}

		marshaled, err := codec.Marshal(data)
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}

		var result map[string]int
		err = codec.Unmarshal(marshaled, &result)
		if err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		if len(result) != len(data) {
			t.Errorf("expected length %d, got %d", len(data), len(result))
		}
		for k, v := range data {
			if result[k] != v {
				t.Errorf("expected %d for key %s, got %d", v, k, result[k])
			}
		}
	})
}
