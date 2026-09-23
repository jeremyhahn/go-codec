# CBOR Codec

CBOR (Concise Binary Object Representation) is a compact binary data format.

## Import

```go
import "github.com/jeremyhahn/go-codec/pkg/cbor"
```

Importing the package registers the codec in `init()`; no build tags are
needed. Import only the codec packages you use: Go links just those.

## Usage

```go
codec := cbor.New[MyStruct]()

data, err := codec.Marshal(myStruct)
err = codec.Unmarshal(data, &result)
```

## Decode strictness

`New` uses fxamacker/cbor's default options, which are lenient: a duplicate
map key is accepted without error, and which value is retained depends on the
destination type; indefinite-length items and tags are accepted. For attacker-controlled input, such as evidence inside a
peer-supplied certificate, build the codec from explicit options:

```go
import fxcbor "github.com/fxamacker/cbor/v2"

c, err := cbor.NewWithOptions[Claims](fxcbor.DecOptions{
    DupMapKey:       fxcbor.DupMapKeyEnforcedAPF, // refuse duplicate keys
    IndefLength:     fxcbor.IndefLengthForbidden,  // reject indefinite-length items
    TagsMd:          fxcbor.TagsForbidden,         // where the container identifies the content
    UTF8:            fxcbor.UTF8RejectInvalid,
    MaxNestedLevels: 16,                           // caller-set bound (4..65535)
}, fxcbor.CoreDetEncOptions())
if err != nil {
    // *cbor.OptionsError; errors.Is(err, cbor.ErrInvalidOptions) is true
}
```

These options do not enforce canonical (deterministic) input: a value with a
non-shortest encoding, such as integer 1 as `18 01` instead of `01`, is still
accepted. `CoreDetEncOptions` governs only what this codec encodes, not what it
accepts. A signature must be verified over the received bytes, not over a
re-encoding of the decoded value. With `DupMapKeyEnforcedAPF`, a
`*fxcbor.DupMapKeyError` may leave the destination partially filled; discard
it on error.

The modes are built once, at construction. Invalid options return an error
and a nil codec, so a misconfigured codec never reaches input.

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
