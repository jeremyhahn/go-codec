//go:build codec_avro

package avro

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

type TestStruct struct {
	Name  string `avro:"name"`
	Age   int    `avro:"age"`
	Email string `avro:"email"`
}

func TestNew(t *testing.T) {
	codec := New[TestStruct]()
	if codec == nil {
		t.Fatal("expected non-nil codec")
	}
	if codec.schema == nil {
		t.Fatal("expected non-nil schema")
	}
}

func TestNewWithSchema(t *testing.T) {
	schemaJSON := `{
		"type": "record",
		"name": "TestStruct",
		"fields": [
			{"name": "name", "type": "string"},
			{"name": "age", "type": "long"},
			{"name": "email", "type": "string"}
		]
	}`

	codec, err := NewWithSchema[TestStruct](schemaJSON)
	if err != nil {
		t.Fatalf("NewWithSchema failed: %v", err)
	}
	if codec == nil {
		t.Fatal("expected non-nil codec")
	}
}

func TestNewWithSchema_Invalid(t *testing.T) {
	_, err := NewWithSchema[TestStruct]("invalid json")
	if err == nil {
		t.Fatal("expected error for invalid schema, got nil")
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
}

func TestCodec_Unmarshal(t *testing.T) {
	codec := New[TestStruct]()
	data := TestStruct{
		Name:  "Jane Doe",
		Age:   25,
		Email: "jane@example.com",
	}

	encoded, err := codec.Marshal(data)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var result TestStruct
	err = codec.Unmarshal(encoded, &result)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if result.Name != data.Name {
		t.Errorf("expected name '%s', got '%s'", data.Name, result.Name)
	}
	if result.Age != data.Age {
		t.Errorf("expected age %d, got %d", data.Age, result.Age)
	}
	if result.Email != data.Email {
		t.Errorf("expected email '%s', got '%s'", data.Email, result.Email)
	}
}

func TestCodec_Unmarshal_Invalid(t *testing.T) {
	// Use a []byte codec since it validates length vs actual data
	codec := New[[]byte]()
	// Claims 8 bytes of data (0x10 = 16 in varint = 8 bytes) but provides none
	invalidData := []byte{0x10}

	var result []byte
	err := codec.Unmarshal(invalidData, &result)
	if err == nil {
		t.Fatal("expected error for invalid Avro data, got nil")
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
}

func TestCodec_Decode(t *testing.T) {
	codec := New[TestStruct]()
	data := TestStruct{
		Name:  "Alice Johnson",
		Age:   28,
		Email: "alice@example.com",
	}

	var buf bytes.Buffer
	err := codec.Encode(&buf, data)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	var result TestStruct
	err = codec.Decode(&buf, &result)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if result.Name != data.Name {
		t.Errorf("expected name '%s', got '%s'", data.Name, result.Name)
	}
	if result.Age != data.Age {
		t.Errorf("expected age %d, got %d", data.Age, result.Age)
	}
}

func TestCodec_Decode_Invalid(t *testing.T) {
	codec := New[TestStruct]()
	invalidData := strings.NewReader("\xff\xff\xff")

	var result TestStruct
	err := codec.Decode(invalidData, &result)
	if err == nil {
		t.Fatal("expected error for invalid Avro data, got nil")
	}
}

func TestCodec_Schema(t *testing.T) {
	codec := New[TestStruct]()
	schema := codec.Schema()
	if schema == nil {
		t.Fatal("expected non-nil schema")
	}
}

func TestCodec_SchemaJSON(t *testing.T) {
	codec := New[TestStruct]()
	schemaJSON := codec.SchemaJSON()
	if schemaJSON == "" {
		t.Error("expected non-empty schema JSON")
	}
	if !strings.Contains(schemaJSON, "name") {
		t.Error("expected schema JSON to contain 'name' field")
	}
}

func TestSchemaCache(t *testing.T) {
	// Clear cache first
	clearSchemaCache()

	// First creation generates schema
	codec1 := New[TestStruct]()
	schema1 := codec1.SchemaJSON()

	// Second creation should use cache
	codec2 := New[TestStruct]()
	schema2 := codec2.SchemaJSON()

	// Both should have equivalent schemas
	if schema1 != schema2 {
		t.Error("cached schema should match")
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

	t.Run("float64", func(t *testing.T) {
		codec := New[float64]()
		data := 3.14159

		marshaled, err := codec.Marshal(data)
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}

		var result float64
		err = codec.Unmarshal(marshaled, &result)
		if err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		if result != data {
			t.Errorf("expected %f, got %f", data, result)
		}
	})

	t.Run("bool", func(t *testing.T) {
		codec := New[bool]()
		data := true

		marshaled, err := codec.Marshal(data)
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}

		var result bool
		err = codec.Unmarshal(marshaled, &result)
		if err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		if result != data {
			t.Errorf("expected %v, got %v", data, result)
		}
	})

	t.Run("bytes", func(t *testing.T) {
		codec := New[[]byte]()
		data := []byte{0x01, 0x02, 0x03, 0x04}

		marshaled, err := codec.Marshal(data)
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}

		var result []byte
		err = codec.Unmarshal(marshaled, &result)
		if err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		if !bytes.Equal(result, data) {
			t.Errorf("expected %v, got %v", data, result)
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

	t.Run("nested_struct", func(t *testing.T) {
		type Inner struct {
			Value string `avro:"value"`
		}
		type Outer struct {
			Inner Inner `avro:"inner"`
		}

		codec := New[Outer]()
		data := Outer{Inner: Inner{Value: "nested"}}

		marshaled, err := codec.Marshal(data)
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}

		var result Outer
		err = codec.Unmarshal(marshaled, &result)
		if err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		if result.Inner.Value != data.Inner.Value {
			t.Errorf("expected %s, got %s", data.Inner.Value, result.Inner.Value)
		}
	})
}

func TestCodec_WithTimeField(t *testing.T) {
	type WithTime struct {
		Name      string    `avro:"name"`
		CreatedAt time.Time `avro:"created_at"`
	}

	codec := New[WithTime]()
	now := time.Now().Truncate(time.Microsecond) // Avro uses microsecond precision
	data := WithTime{
		Name:      "Test",
		CreatedAt: now,
	}

	marshaled, err := codec.Marshal(data)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var result WithTime
	err = codec.Unmarshal(marshaled, &result)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if result.Name != data.Name {
		t.Errorf("expected name '%s', got '%s'", data.Name, result.Name)
	}
}

func TestCodec_WithPointerField(t *testing.T) {
	type WithPointer struct {
		Name     string  `avro:"name"`
		Optional *string `avro:"optional"`
	}

	codec := New[WithPointer]()

	t.Run("nil_pointer", func(t *testing.T) {
		data := WithPointer{Name: "Test", Optional: nil}

		marshaled, err := codec.Marshal(data)
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}

		var result WithPointer
		err = codec.Unmarshal(marshaled, &result)
		if err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		if result.Name != data.Name {
			t.Errorf("expected name '%s', got '%s'", data.Name, result.Name)
		}
		if result.Optional != nil {
			t.Error("expected nil optional, got non-nil")
		}
	})

	t.Run("non_nil_pointer", func(t *testing.T) {
		value := "optional value"
		data := WithPointer{Name: "Test", Optional: &value}

		marshaled, err := codec.Marshal(data)
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}

		var result WithPointer
		err = codec.Unmarshal(marshaled, &result)
		if err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		if result.Name != data.Name {
			t.Errorf("expected name '%s', got '%s'", data.Name, result.Name)
		}
		if result.Optional == nil {
			t.Fatal("expected non-nil optional, got nil")
		}
		if *result.Optional != value {
			t.Errorf("expected optional '%s', got '%s'", value, *result.Optional)
		}
	})
}

func TestCodec_WithAvroTag(t *testing.T) {
	type WithAvroTag struct {
		Name  string `avro:"custom_name"`
		Value int    `avro:"custom_value"`
	}

	codec := New[WithAvroTag]()
	data := WithAvroTag{Name: "Test", Value: 42}

	marshaled, err := codec.Marshal(data)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var result WithAvroTag
	err = codec.Unmarshal(marshaled, &result)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if result.Name != data.Name {
		t.Errorf("expected name '%s', got '%s'", data.Name, result.Name)
	}
	if result.Value != data.Value {
		t.Errorf("expected value %d, got %d", data.Value, result.Value)
	}

	// Verify schema contains custom field names
	schemaJSON := codec.SchemaJSON()
	if !strings.Contains(schemaJSON, "custom_name") {
		t.Error("expected schema to contain 'custom_name'")
	}
	if !strings.Contains(schemaJSON, "custom_value") {
		t.Error("expected schema to contain 'custom_value'")
	}
}

func TestCodec_SkippedField(t *testing.T) {
	type TestSkipStruct struct {
		Name    string `avro:"name"`
		Ignored string `avro:"-"`
	}

	codec := New[TestSkipStruct]()
	schemaJSON := codec.SchemaJSON()

	// The schema should contain 'name' field
	if !strings.Contains(schemaJSON, `"name":"name"`) {
		t.Error("schema should contain 'name' field")
	}

	// The schema should not contain 'Ignored' as a field name
	// (checking for the field definition, not the struct name)
	if strings.Contains(schemaJSON, `"name":"Ignored"`) || strings.Contains(schemaJSON, `"name":"-"`) {
		t.Error("schema should not contain skipped field")
	}
}
