//go:build codec_json

package json

import (
	"bytes"
	"strings"
	"testing"
)

type TestStruct struct {
	Name  string `json:"name"`
	Age   int    `json:"age"`
	Email string `json:"email"`
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

	expected := `{"name":"John Doe","age":30,"email":"john@example.com"}`
	actual := strings.TrimSpace(string(result))
	if actual != expected {
		t.Errorf("expected %s, got %s", expected, actual)
	}
}

func TestCodec_Unmarshal(t *testing.T) {
	codec := New[TestStruct]()
	jsonData := []byte(`{"name":"Jane Doe","age":25,"email":"jane@example.com"}`)

	var result TestStruct
	err := codec.Unmarshal(jsonData, &result)
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
	jsonData := []byte(`{invalid json}`)

	var result TestStruct
	err := codec.Unmarshal(jsonData, &result)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
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

	expected := `{"name":"Bob Smith","age":35,"email":"bob@example.com"}`
	actual := strings.TrimSpace(buf.String())
	if actual != expected {
		t.Errorf("expected %s, got %s", expected, actual)
	}
}

func TestCodec_Decode(t *testing.T) {
	codec := New[TestStruct]()
	jsonData := `{"name":"Alice Johnson","age":28,"email":"alice@example.com"}`
	reader := strings.NewReader(jsonData)

	var result TestStruct
	err := codec.Decode(reader, &result)
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
	jsonData := `{invalid json}`
	reader := strings.NewReader(jsonData)

	var result TestStruct
	err := codec.Decode(reader, &result)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
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
