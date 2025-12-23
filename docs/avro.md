# Avro Codec

Apache Avro is a compact binary format with schema support.

## Import

```go
import "github.com/jeremyhahn/go-codec/pkg/avro"
```

## Usage

```go
// Automatic schema inference
codec := avro.New[MyStruct]()

data, err := codec.Marshal(myStruct)
err = codec.Unmarshal(data, &result)
```

## Struct Tags

Use the `avro` tag to control field names:

```go
type Record struct {
    ID        int64   `avro:"id"`
    Name      string  `avro:"name"`
    Value     float64 `avro:"value"`
    Optional  *string `avro:"optional"` // nullable field
}
```

## Explicit Schema

For advanced use cases, provide an explicit schema:

```go
schema := `{
    "type": "record",
    "name": "User",
    "fields": [
        {"name": "id", "type": "long"},
        {"name": "name", "type": "string"}
    ]
}`

codec, err := avro.NewWithSchema[User](schema)
```

## Type Mapping

| Go Type | Avro Type |
|---------|-----------|
| `bool` | boolean |
| `int`, `int64` | long |
| `int32`, `int16`, `int8` | int |
| `float32` | float |
| `float64` | double |
| `string` | string |
| `[]byte` | bytes |
| `[]T` | array |
| `map[string]T` | map |
| `struct` | record |
| `*T` | union (null, T) |
| `time.Time` | long (timestamp-micros) |

## Performance

| Operation | Time | Memory | Allocs |
|-----------|------|--------|--------|
| Marshal | 420 ns/op | 112 B/op | 2 |
| Unmarshal | 221 ns/op | 24 B/op | 0 |

## Notes

- Uses `github.com/hamba/avro/v2`
- Schema is inferred automatically from Go types
- Schemas are cached for performance
- Very fast serialization/deserialization
- Ideal for data pipelines and event streaming
