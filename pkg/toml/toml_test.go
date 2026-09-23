package toml

import (
	"bytes"
	"strings"
	"testing"
)

type TestStruct struct {
	Name  string `toml:"name"`
	Age   int    `toml:"age"`
	Email string `toml:"email"`
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
	if !strings.Contains(resultStr, `name = "John Doe"`) {
		t.Errorf("expected TOML to contain 'name = \"John Doe\"', got %s", resultStr)
	}
	if !strings.Contains(resultStr, "age = 30") {
		t.Errorf("expected TOML to contain 'age = 30', got %s", resultStr)
	}
	if !strings.Contains(resultStr, `email = "john@example.com"`) {
		t.Errorf("expected TOML to contain 'email = \"john@example.com\"', got %s", resultStr)
	}
}

func TestCodec_Unmarshal(t *testing.T) {
	codec := New[TestStruct]()
	tomlData := []byte(`name = "Jane Doe"
age = 25
email = "jane@example.com"`)

	var result TestStruct
	err := codec.Unmarshal(tomlData, &result)
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
	tomlData := []byte(`invalid toml = [[[`)

	var result TestStruct
	err := codec.Unmarshal(tomlData, &result)
	if err == nil {
		t.Fatal("expected error for invalid TOML, got nil")
	}
}

func TestCodec_Marshal_Invalid(t *testing.T) {
	type InvalidStruct struct {
		Ch chan int `toml:"ch"`
	}
	codec := New[InvalidStruct]()
	data := InvalidStruct{
		Ch: make(chan int),
	}

	_, err := codec.Marshal(data)
	if err == nil {
		t.Fatal("expected error for invalid TOML type, got nil")
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
	if !strings.Contains(resultStr, `name = "Bob Smith"`) {
		t.Errorf("expected TOML to contain 'name = \"Bob Smith\"', got %s", resultStr)
	}
	if !strings.Contains(resultStr, "age = 35") {
		t.Errorf("expected TOML to contain 'age = 35', got %s", resultStr)
	}
	if !strings.Contains(resultStr, `email = "bob@example.com"`) {
		t.Errorf("expected TOML to contain 'email = \"bob@example.com\"', got %s", resultStr)
	}
}

func TestCodec_Decode(t *testing.T) {
	codec := New[TestStruct]()
	tomlData := `name = "Alice Johnson"
age = 28
email = "alice@example.com"`
	reader := strings.NewReader(tomlData)

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
	tomlData := `invalid toml = [[[`
	reader := strings.NewReader(tomlData)

	var result TestStruct
	err := codec.Decode(reader, &result)
	if err == nil {
		t.Fatal("expected error for invalid TOML, got nil")
	}
}

func TestCodec_WithDifferentTypes(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		type StringWrapper struct {
			Value string `toml:"value"`
		}
		codec := New[StringWrapper]()
		data := StringWrapper{Value: "hello world"}

		marshaled, err := codec.Marshal(data)
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}

		var result StringWrapper
		err = codec.Unmarshal(marshaled, &result)
		if err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		if result.Value != data.Value {
			t.Errorf("expected %s, got %s", data.Value, result.Value)
		}
	})

	t.Run("int", func(t *testing.T) {
		type IntWrapper struct {
			Value int `toml:"value"`
		}
		codec := New[IntWrapper]()
		data := IntWrapper{Value: 42}

		marshaled, err := codec.Marshal(data)
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}

		var result IntWrapper
		err = codec.Unmarshal(marshaled, &result)
		if err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		if result.Value != data.Value {
			t.Errorf("expected %d, got %d", data.Value, result.Value)
		}
	})

	t.Run("slice", func(t *testing.T) {
		type SliceWrapper struct {
			Values []string `toml:"values"`
		}
		codec := New[SliceWrapper]()
		data := SliceWrapper{Values: []string{"one", "two", "three"}}

		marshaled, err := codec.Marshal(data)
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}

		var result SliceWrapper
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
		type MapWrapper struct {
			Values map[string]int `toml:"values"`
		}
		codec := New[MapWrapper]()
		data := MapWrapper{Values: map[string]int{"one": 1, "two": 2, "three": 3}}

		marshaled, err := codec.Marshal(data)
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}

		var result MapWrapper
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
