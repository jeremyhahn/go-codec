package yaml

import (
	"bytes"
	"strings"
	"testing"
)

type TestStruct struct {
	Name  string `yaml:"name"`
	Age   int    `yaml:"age"`
	Email string `yaml:"email"`
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

	resultStr := string(result)
	if !strings.Contains(resultStr, "name: John Doe") {
		t.Errorf("expected YAML to contain 'name: John Doe', got %s", resultStr)
	}
	if !strings.Contains(resultStr, "age: 30") {
		t.Errorf("expected YAML to contain 'age: 30', got %s", resultStr)
	}
	if !strings.Contains(resultStr, "email: john@example.com") {
		t.Errorf("expected YAML to contain 'email: john@example.com', got %s", resultStr)
	}
}

func TestCodec_Unmarshal(t *testing.T) {
	codec := New[TestStruct]()
	yamlData := []byte(`name: Jane Doe
age: 25
email: jane@example.com`)

	var result TestStruct
	err := codec.Unmarshal(yamlData, &result)
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
	yamlData := []byte(`invalid: yaml: data:
  - broken
    - structure`)

	var result TestStruct
	err := codec.Unmarshal(yamlData, &result)
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
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

	resultStr := buf.String()
	if !strings.Contains(resultStr, "name: Bob Smith") {
		t.Errorf("expected YAML to contain 'name: Bob Smith', got %s", resultStr)
	}
	if !strings.Contains(resultStr, "age: 35") {
		t.Errorf("expected YAML to contain 'age: 35', got %s", resultStr)
	}
	if !strings.Contains(resultStr, "email: bob@example.com") {
		t.Errorf("expected YAML to contain 'email: bob@example.com', got %s", resultStr)
	}
}

func TestCodec_Decode(t *testing.T) {
	codec := New[TestStruct]()
	yamlData := `name: Alice Johnson
age: 28
email: alice@example.com`
	reader := strings.NewReader(yamlData)

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
	yamlData := `invalid: yaml: data:
  - broken
    - structure`
	reader := strings.NewReader(yamlData)

	var result TestStruct
	err := codec.Decode(reader, &result)
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
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
