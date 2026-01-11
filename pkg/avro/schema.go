//go:build codec_avro

package avro

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hamba/avro/v2"
)

var (
	// schemaCache stores generated schemas by type to avoid regeneration
	schemaCache sync.Map // map[reflect.Type]avro.Schema

	// schemaNameIndex is an atomic counter for generating unique anonymous record names
	schemaNameIndex int64

	// schemaMutex protects schema generation to prevent duplicate work
	schemaMutex sync.Mutex

	// timeType is cached for comparison
	timeType = reflect.TypeOf(time.Time{})
)

// getOrCreateSchema returns a cached schema or creates a new one
func getOrCreateSchema(t reflect.Type) avro.Schema {
	// Handle nil type
	if t == nil {
		return avro.NewPrimitiveSchema(avro.Null, nil)
	}

	// Handle pointer types - get underlying type
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Check cache first (fast path)
	if cached, ok := schemaCache.Load(t); ok {
		return cached.(avro.Schema)
	}

	// Generate schema with mutex protection
	schemaMutex.Lock()
	defer schemaMutex.Unlock()

	// Double-check after acquiring lock
	if cached, ok := schemaCache.Load(t); ok {
		return cached.(avro.Schema)
	}

	schema := generateSchema(t)
	schemaCache.Store(t, schema)
	return schema
}

// generateSchema creates an Avro schema from a Go type
func generateSchema(t reflect.Type) avro.Schema {
	// Handle pointer types
	if t.Kind() == reflect.Ptr {
		// Nullable type: union of null and the element type
		elemSchema := generateSchema(t.Elem())
		union, _ := avro.NewUnionSchema([]avro.Schema{
			&avro.NullSchema{},
			elemSchema,
		})
		return union
	}

	switch t.Kind() {
	case reflect.Bool:
		return avro.NewPrimitiveSchema(avro.Boolean, nil)

	case reflect.Int, reflect.Int64, reflect.Uint, reflect.Uint64:
		return avro.NewPrimitiveSchema(avro.Long, nil)

	case reflect.Int32, reflect.Int16, reflect.Int8, reflect.Uint32, reflect.Uint16, reflect.Uint8:
		return avro.NewPrimitiveSchema(avro.Int, nil)

	case reflect.Float32:
		return avro.NewPrimitiveSchema(avro.Float, nil)

	case reflect.Float64:
		return avro.NewPrimitiveSchema(avro.Double, nil)

	case reflect.String:
		return avro.NewPrimitiveSchema(avro.String, nil)

	case reflect.Slice:
		// []byte is treated as Avro bytes
		if t.Elem().Kind() == reflect.Uint8 {
			return avro.NewPrimitiveSchema(avro.Bytes, nil)
		}
		// Other slices become arrays
		elemSchema := generateSchema(t.Elem())
		return avro.NewArraySchema(elemSchema)

	case reflect.Array:
		// Fixed-size arrays treated as bytes if byte array, otherwise as array
		if t.Elem().Kind() == reflect.Uint8 {
			return avro.NewPrimitiveSchema(avro.Bytes, nil)
		}
		elemSchema := generateSchema(t.Elem())
		return avro.NewArraySchema(elemSchema)

	case reflect.Map:
		// Avro maps require string keys
		if t.Key().Kind() != reflect.String {
			// Fall back to string representation for non-string keys
			return avro.NewPrimitiveSchema(avro.String, nil)
		}
		valueSchema := generateSchema(t.Elem())
		return avro.NewMapSchema(valueSchema)

	case reflect.Struct:
		return generateRecordSchema(t)

	case reflect.Interface:
		// Interface types default to string
		return avro.NewPrimitiveSchema(avro.String, nil)

	default:
		// Fallback to string for unsupported types
		return avro.NewPrimitiveSchema(avro.String, nil)
	}
}

// generateRecordSchema creates a record schema from a struct type
func generateRecordSchema(t reflect.Type) avro.Schema {
	// Handle time.Time specially
	if t == timeType {
		return avro.NewPrimitiveSchema(avro.Long, avro.NewPrimitiveLogicalSchema(avro.TimestampMicros))
	}

	fields := make([]*avro.Field, 0, t.NumField())

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Get field name from tags
		fieldName := getFieldName(field)
		if fieldName == "-" {
			continue // Skip fields marked with "-"
		}

		fieldSchema := generateSchema(field.Type)

		// Create field with default handling for pointers (nullable)
		var avroField *avro.Field
		var err error

		if field.Type.Kind() == reflect.Ptr {
			// Pointer fields have null as default
			avroField, err = avro.NewField(fieldName, fieldSchema, avro.WithDefault(nil))
		} else {
			avroField, err = avro.NewField(fieldName, fieldSchema)
		}

		if err != nil {
			// If field creation fails, skip this field
			continue
		}

		fields = append(fields, avroField)
	}

	// Generate record name
	recordName := t.Name()
	if recordName == "" {
		recordName = fmt.Sprintf("AnonymousRecord%d", atomic.AddInt64(&schemaNameIndex, 1))
	}

	// Generate namespace from package path
	namespace := t.PkgPath()
	if namespace == "" {
		namespace = "go.codec.generated"
	}
	// Convert Go package path to valid Avro namespace
	namespace = strings.ReplaceAll(namespace, "/", ".")
	namespace = strings.ReplaceAll(namespace, "-", "_")

	// Record name is always valid (Go struct name or "AnonymousRecord#")
	// Namespace is sanitized (/ → ., - → _)
	// So NewRecordSchema will not error with our inputs
	schema, _ := avro.NewRecordSchema(recordName, namespace, fields)
	return schema
}

// getFieldName extracts the Avro field name from struct tags
// Priority: avro tag > field name (matches hamba/avro library behavior)
// Note: json tags are NOT used because hamba/avro only recognizes avro tags
func getFieldName(field reflect.StructField) string {
	// Check avro tag
	if tag := field.Tag.Get("avro"); tag != "" {
		parts := strings.Split(tag, ",")
		if parts[0] != "" {
			return parts[0]
		}
	}

	// Default to field name (matches hamba/avro behavior)
	return field.Name
}

// clearSchemaCache clears the schema cache (useful for testing)
func clearSchemaCache() {
	schemaCache = sync.Map{}
}
