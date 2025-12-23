# JSON Codec

JSON (JavaScript Object Notation) is a lightweight text-based data format.

## Import

```go
import "github.com/jeremyhahn/go-codec/pkg/json"
```

## Usage

```go
// Standard codec
codec := json.New[MyStruct]()

// With buffer reuse (high-performance)
codec := json.NewPool[MyStruct]()
```

## Struct Tags

Use the `json` tag to control field names and behavior:

```go
type User struct {
    ID        int    `json:"id"`
    Name      string `json:"name"`
    Email     string `json:"email,omitempty"` // omit if empty
    Internal  string `json:"-"`               // skip field
}
```

## Buffer Reuse Methods

For high-throughput scenarios, use `NewPool()` to access buffer reuse methods:

```go
codec := json.NewPool[MyStruct]()
buf := make([]byte, 0, 1024)

// Reuse buffer across operations
for _, item := range items {
    buf = buf[:0]
    result, err := codec.MarshalTo(buf, item)
    process(result)
}
```

Available methods:
- `MarshalTo(buf, data)` - Marshal into provided buffer
- `AppendMarshal(buf, data)` - Append marshaled data to buffer
- `UnmarshalFrom(data, v, scratch)` - Unmarshal with scratch buffer

## Performance

| Operation | Time | Memory | Allocs |
|-----------|------|--------|--------|
| Marshal | 874 ns/op | 160 B/op | 2 |
| Unmarshal | 2,036 ns/op | 240 B/op | 6 |
| MarshalTo (reuse) | 655 ns/op | 64 B/op | 1 |

## Notes

- Uses Go's standard `encoding/json` package
- Supports all standard JSON types
- UTF-8 encoded output
