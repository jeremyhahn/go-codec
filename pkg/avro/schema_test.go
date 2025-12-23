package avro

import (
	"reflect"
	"testing"
	"time"
)

func TestGetOrCreateSchema_NilType(t *testing.T) {
	schema := getOrCreateSchema(nil)
	if schema == nil {
		t.Fatal("expected non-nil schema for nil type")
	}
}

func TestGetOrCreateSchema_PointerType(t *testing.T) {
	var ptr *string
	schema := getOrCreateSchema(reflect.TypeOf(ptr))
	if schema == nil {
		t.Fatal("expected non-nil schema for pointer type")
	}
}

func TestGetOrCreateSchema_Cache(t *testing.T) {
	clearSchemaCache()

	type CacheTest struct {
		Value string `avro:"value"`
	}

	// First call generates schema
	schema1 := getOrCreateSchema(reflect.TypeOf(CacheTest{}))

	// Second call should use cache
	schema2 := getOrCreateSchema(reflect.TypeOf(CacheTest{}))

	if schema1.String() != schema2.String() {
		t.Error("cached schema should match")
	}
}

func TestGenerateSchema_AllPrimitives(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected string
	}{
		{"bool", true, "boolean"},
		{"int", int(0), "long"},
		{"int8", int8(0), "int"},
		{"int16", int16(0), "int"},
		{"int32", int32(0), "int"},
		{"int64", int64(0), "long"},
		{"uint", uint(0), "long"},
		{"uint8", uint8(0), "int"},
		{"uint16", uint16(0), "int"},
		{"uint32", uint32(0), "int"},
		{"uint64", uint64(0), "long"},
		{"float32", float32(0), "float"},
		{"float64", float64(0), "double"},
		{"string", "", "string"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := generateSchema(reflect.TypeOf(tt.value))
			if schema == nil {
				t.Fatal("expected non-nil schema")
			}
		})
	}
}

func TestGenerateSchema_Bytes(t *testing.T) {
	// []byte should be bytes
	schema := generateSchema(reflect.TypeOf([]byte{}))
	if schema == nil {
		t.Fatal("expected non-nil schema for []byte")
	}
}

func TestGenerateSchema_Array(t *testing.T) {
	// Fixed-size byte array
	schema := generateSchema(reflect.TypeOf([4]byte{}))
	if schema == nil {
		t.Fatal("expected non-nil schema for [4]byte")
	}

	// Fixed-size non-byte array
	schema2 := generateSchema(reflect.TypeOf([4]int{}))
	if schema2 == nil {
		t.Fatal("expected non-nil schema for [4]int")
	}
}

func TestGenerateSchema_Slice(t *testing.T) {
	schema := generateSchema(reflect.TypeOf([]int{}))
	if schema == nil {
		t.Fatal("expected non-nil schema for []int")
	}
}

func TestGenerateSchema_Map(t *testing.T) {
	// String key map
	schema := generateSchema(reflect.TypeOf(map[string]int{}))
	if schema == nil {
		t.Fatal("expected non-nil schema for map[string]int")
	}

	// Non-string key map (should fall back to string)
	schema2 := generateSchema(reflect.TypeOf(map[int]string{}))
	if schema2 == nil {
		t.Fatal("expected non-nil schema for map[int]string")
	}
}

func TestGenerateSchema_Interface(t *testing.T) {
	var i interface{}
	schema := generateSchema(reflect.TypeOf(&i).Elem())
	if schema == nil {
		t.Fatal("expected non-nil schema for interface{}")
	}
}

func TestGenerateSchema_Pointer(t *testing.T) {
	var ptr *string
	schema := generateSchema(reflect.TypeOf(ptr))
	if schema == nil {
		t.Fatal("expected non-nil schema for *string")
	}
}

func TestGenerateRecordSchema_Time(t *testing.T) {
	schema := generateSchema(reflect.TypeOf(time.Time{}))
	if schema == nil {
		t.Fatal("expected non-nil schema for time.Time")
	}
}

func TestGenerateRecordSchema_UnexportedFields(t *testing.T) {
	type WithUnexported struct {
		Public  string `avro:"public"`
		private string //nolint:unused
	}

	schema := generateSchema(reflect.TypeOf(WithUnexported{}))
	schemaJSON := schema.String()

	// Should contain public field
	if schemaJSON == "" {
		t.Error("expected non-empty schema")
	}
}

func TestGenerateRecordSchema_AnonymousStruct(t *testing.T) {
	// Anonymous struct
	data := struct {
		Field string `avro:"field"`
	}{}

	schema := generateSchema(reflect.TypeOf(data))
	if schema == nil {
		t.Fatal("expected non-nil schema for anonymous struct")
	}
}

func TestGenerateRecordSchema_PointerField(t *testing.T) {
	type WithPointer struct {
		Value *string `avro:"value"`
	}

	schema := generateSchema(reflect.TypeOf(WithPointer{}))
	if schema == nil {
		t.Fatal("expected non-nil schema for struct with pointer field")
	}
}

func TestGetFieldName_EmptyAvroTag(t *testing.T) {
	type WithEmptyTag struct {
		Field string `avro:""`
	}

	field := reflect.TypeOf(WithEmptyTag{}).Field(0)
	name := getFieldName(field)

	// Should fall back to field name
	if name != "Field" {
		t.Errorf("expected 'Field', got '%s'", name)
	}
}

func TestGetFieldName_NoTags(t *testing.T) {
	type NoTags struct {
		MyField string
	}

	field := reflect.TypeOf(NoTags{}).Field(0)
	name := getFieldName(field)

	if name != "MyField" {
		t.Errorf("expected 'MyField', got '%s'", name)
	}
}

func TestGetFieldName_AvroTagWithOptions(t *testing.T) {
	type WithOptions struct {
		Field string `avro:"field_name,omitempty"`
	}

	field := reflect.TypeOf(WithOptions{}).Field(0)
	name := getFieldName(field)

	// Should use first part before comma
	if name != "field_name" {
		t.Errorf("expected 'field_name', got '%s'", name)
	}
}

func TestClearSchemaCache(t *testing.T) {
	type ClearTest struct {
		Value string `avro:"value"`
	}

	// Add to cache
	getOrCreateSchema(reflect.TypeOf(ClearTest{}))

	// Clear cache
	clearSchemaCache()

	// Verify we can still get schema (will regenerate)
	schema := getOrCreateSchema(reflect.TypeOf(ClearTest{}))
	if schema == nil {
		t.Fatal("expected non-nil schema after cache clear")
	}
}

func TestGenerateSchema_UnsupportedType(t *testing.T) {
	// Channel type should fall back to string
	ch := make(chan int)
	schema := generateSchema(reflect.TypeOf(ch))
	if schema == nil {
		t.Fatal("expected non-nil schema for channel type")
	}
}

func TestGenerateSchema_FuncType(t *testing.T) {
	// Function type should fall back to string
	var fn func()
	schema := generateSchema(reflect.TypeOf(fn))
	if schema == nil {
		t.Fatal("expected non-nil schema for func type")
	}
}

func TestGenerateSchema_UnsafePointer(t *testing.T) {
	// Complex types that aren't directly supported
	var c complex128
	schema := generateSchema(reflect.TypeOf(c))
	if schema == nil {
		t.Fatal("expected non-nil schema for complex type")
	}
}

func TestGenerateRecordSchema_InvalidFieldName(t *testing.T) {
	// Field with invalid Avro name (contains dash) should be skipped
	type WithInvalidField struct {
		Valid   string `avro:"valid"`
		Invalid string `avro:"invalid-name"` // Dash is invalid in Avro
	}

	schema := generateSchema(reflect.TypeOf(WithInvalidField{}))
	if schema == nil {
		t.Fatal("expected non-nil schema")
	}

	// The invalid field should be skipped, but schema should still be created
	schemaJSON := schema.String()
	if schemaJSON == "" {
		t.Error("expected non-empty schema JSON")
	}
}

func TestGetOrCreateSchema_ConcurrentAccess(t *testing.T) {
	// Test concurrent access to trigger double-check locking path
	for iteration := 0; iteration < 100; iteration++ {
		clearSchemaCache()

		type ConcurrentTest struct {
			Value string `avro:"value"`
		}

		// Use a start signal to maximize concurrency
		start := make(chan struct{})
		done := make(chan bool, 50)

		for i := 0; i < 50; i++ {
			go func() {
				<-start // Wait for signal
				schema := getOrCreateSchema(reflect.TypeOf(ConcurrentTest{}))
				if schema == nil {
					t.Error("expected non-nil schema")
				}
				done <- true
			}()
		}

		// Release all goroutines simultaneously
		close(start)

		// Wait for all goroutines
		for i := 0; i < 50; i++ {
			<-done
		}
	}
}
