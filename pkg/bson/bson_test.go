package bson

import (
	"bytes"
	"testing"
)

type TestStruct struct {
	Name  string `bson:"name"`
	Age   int    `bson:"age"`
	Email string `bson:"email"`
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
	bsonData, err := codec.Marshal(data)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Now unmarshal it
	var result TestStruct
	err = codec.Unmarshal(bsonData, &result)
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
	bsonData := []byte{0xff, 0xff, 0xff} // Invalid BSON data

	var result TestStruct
	err := codec.Unmarshal(bsonData, &result)
	if err == nil {
		t.Fatal("expected error for invalid BSON, got nil")
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
	buf := bytes.NewBuffer([]byte{0xff, 0xff, 0xff}) // Invalid BSON data

	var result TestStruct
	err := codec.Decode(buf, &result)
	if err == nil {
		t.Fatal("expected error for invalid BSON, got nil")
	}
}

func TestCodec_Decode_ReadError(t *testing.T) {
	codec := New[TestStruct]()
	// Use an error reader
	errorReader := &errorReader{}

	var result TestStruct
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
	codec := New[TestStruct]()
	data := TestStruct{
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
	t.Run("struct with embedded fields", func(t *testing.T) {
		type Address struct {
			Street string `bson:"street"`
			City   string `bson:"city"`
		}
		type Person struct {
			Name    string  `bson:"name"`
			Age     int     `bson:"age"`
			Address Address `bson:"address"`
		}

		codec := New[Person]()
		data := Person{
			Name: "John Doe",
			Age:  30,
			Address: Address{
				Street: "123 Main St",
				City:   "Springfield",
			},
		}

		marshaled, err := codec.Marshal(data)
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}

		var result Person
		err = codec.Unmarshal(marshaled, &result)
		if err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		if result.Name != data.Name {
			t.Errorf("expected name %s, got %s", data.Name, result.Name)
		}
		if result.Age != data.Age {
			t.Errorf("expected age %d, got %d", data.Age, result.Age)
		}
		if result.Address.Street != data.Address.Street {
			t.Errorf("expected street %s, got %s", data.Address.Street, result.Address.Street)
		}
	})

	t.Run("map", func(t *testing.T) {
		codec := New[map[string]interface{}]()
		data := map[string]interface{}{
			"name":  "Jane Doe",
			"age":   25,
			"email": "jane@example.com",
		}

		marshaled, err := codec.Marshal(data)
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}

		var result map[string]interface{}
		err = codec.Unmarshal(marshaled, &result)
		if err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		if result["name"] != data["name"] {
			t.Errorf("expected name %v, got %v", data["name"], result["name"])
		}
		// BSON stores numbers as int32/int64, so we need to compare differently
		if result["age"].(int32) != int32(data["age"].(int)) {
			t.Errorf("expected age %v, got %v", data["age"], result["age"])
		}
	})

	t.Run("struct with slice field", func(t *testing.T) {
		type ItemList struct {
			Items []string `bson:"items"`
			Count int      `bson:"count"`
		}

		codec := New[ItemList]()
		data := ItemList{
			Items: []string{"one", "two", "three"},
			Count: 3,
		}

		marshaled, err := codec.Marshal(data)
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}

		var result ItemList
		err = codec.Unmarshal(marshaled, &result)
		if err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		if len(result.Items) != len(data.Items) {
			t.Errorf("expected length %d, got %d", len(data.Items), len(result.Items))
		}
		for i, v := range data.Items {
			if result.Items[i] != v {
				t.Errorf("expected %s at index %d, got %s", v, i, result.Items[i])
			}
		}
		if result.Count != data.Count {
			t.Errorf("expected count %d, got %d", data.Count, result.Count)
		}
	})
}
