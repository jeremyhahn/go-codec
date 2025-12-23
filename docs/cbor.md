# CBOR Codec

CBOR (Concise Binary Object Representation) is a compact binary data format.

## Import

```go
import "github.com/jeremyhahn/go-codec/pkg/cbor"
```

## Usage

```go
codec := cbor.New[MyStruct]()

data, err := codec.Marshal(myStruct)
err = codec.Unmarshal(data, &result)
```

## Struct Tags

Use the `cbor` tag to control field names:

```go
type Message struct {
    Version int    `cbor:"v"`
    Type    string `cbor:"t"`
    Payload []byte `cbor:"p"`
}
```

## Performance

| Operation | Time | Memory | Allocs |
|-----------|------|--------|--------|
| Marshal | 668 ns/op | 144 B/op | 2 |
| Unmarshal | 792 ns/op | 24 B/op | 2 |

## Notes

- Uses `github.com/fxamacker/cbor/v2`
- IETF standard (RFC 8949)
- Very compact encoding
- Good for IoT, embedded systems, and constrained environments
- Supports streaming and indefinite-length items
