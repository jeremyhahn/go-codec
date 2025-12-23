# MessagePack Codec

MessagePack is an efficient binary serialization format.

## Import

```go
import "github.com/jeremyhahn/go-codec/pkg/msgpack"
```

## Usage

```go
// Standard codec
codec := msgpack.New[MyStruct]()

// With buffer reuse (high-performance)
codec := msgpack.NewPool[MyStruct]()
```

## Struct Tags

Use the `msgpack` tag to control field names:

```go
type Event struct {
    ID        int64  `msgpack:"id"`
    Type      string `msgpack:"type"`
    Timestamp int64  `msgpack:"ts"`
    Data      []byte `msgpack:"data,omitempty"`
}
```

## Buffer Reuse Methods

For high-throughput scenarios, use `NewPool()`:

```go
codec := msgpack.NewPool[MyStruct]()
buf := make([]byte, 0, 1024)

for _, item := range items {
    buf = buf[:0]
    result, err := codec.MarshalTo(buf, item)
    send(result)
}
```

## Performance

| Operation | Time | Memory | Allocs |
|-----------|------|--------|--------|
| Marshal | 1,062 ns/op | 305 B/op | 4 |
| Unmarshal | 803 ns/op | 72 B/op | 3 |
| MarshalTo (reuse) | 838 ns/op | 176 B/op | 3 |

## Notes

- Uses `github.com/vmihailenco/msgpack/v5`
- More compact than JSON (~30-50% smaller)
- Faster unmarshaling than JSON
- Good for network protocols and caching
