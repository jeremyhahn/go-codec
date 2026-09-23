# go-codec

A generic codec library for Go with a unified API across multiple serialization formats.

## Features

- **Unified API** - Same interface for all formats
- **Go Generics** - Type-safe encoding/decoding
- **7 Formats** - JSON, YAML, TOML, MessagePack, BSON, CBOR, Protocol Buffers
- **Modular Imports** - Import a codec package to register it; no build tags
- **100% Test Coverage**

## Installation

```bash
go get github.com/jeremyhahn/go-codec
```

## Quick Start

### Direct Import (Recommended)

Import only the codec you need:

```go
import "github.com/jeremyhahn/go-codec/pkg/json"

type User struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}

func main() {
    codec := json.New[User]()

    // Marshal
    data, _ := codec.Marshal(User{Name: "Alice", Email: "alice@example.com"})

    // Unmarshal
    var user User
    codec.Unmarshal(data, &user)
}
```

### Factory (All Codecs)

Use the factory when you need runtime codec selection. Importing
`pkg/factory` links and registers all seven codecs:

```go
import (
    "github.com/jeremyhahn/go-codec"
    "github.com/jeremyhahn/go-codec/pkg/factory"
)

c, err := factory.New[User](codec.JSON)
if err != nil {
    log.Fatal(err) // unknown codec type
}
data, _ := c.Marshal(user)
```

## Available Codecs

| Format | Import | Struct Tag |
|--------|--------|------------|
| JSON | `pkg/json` | `json:"name"` |
| YAML | `pkg/yaml` | `yaml:"name"` |
| TOML | `pkg/toml` | `toml:"name"` |
| MessagePack | `pkg/msgpack` | `msgpack:"name"` |
| BSON | `pkg/bson` | `bson:"name"` |
| CBOR | `pkg/cbor` | `cbor:"name"` |
| Protocol Buffers | `pkg/protobuf` | (generated) |

## Choosing Codecs: the Import Is the Opt-In

There are no codec build tags. Each codec lives in its own package and
registers itself in `init()` when imported, the same pattern as
`database/sql` drivers and `image/png`. A plain `go build` / `go test ./...`
works; no `-tags` are needed.

Go links only the packages a program imports, so to keep a binary small,
import just the codec packages you use (for example only `pkg/json`) rather
than `pkg/factory`, which imports all seven.

Builds that still pass the old `-tags codec_json,...` flags keep working:
the tags no longer mean anything and are ignored.

Query available codecs at runtime:

```go
import "github.com/jeremyhahn/go-codec"

for _, c := range codec.SupportedCodecs() {
    fmt.Println(c) // prints only the codecs whose packages are imported
}

if codec.IsSupported(codec.CBOR) {
    // use CBOR
}
```

## API

All codecs implement the same interface:

```go
type Codec[T any] interface {
    Marshal(data T) ([]byte, error)
    Unmarshal(data []byte, v *T) error
    Encode(w io.Writer, data T) error
    Decode(r io.Reader, data *T) error
}
```

## Examples

### Stream Encoding

```go
codec := json.New[User]()

// Write to any io.Writer
var buf bytes.Buffer
codec.Encode(&buf, user)

// Read from any io.Reader
var result User
codec.Decode(&buf, &result)
```

### Protocol Buffers

```go
import "github.com/jeremyhahn/go-codec/pkg/factory"

// Requires generated protobuf code
codec, err := factory.NewProtoBuf[*pb.User]()
if err != nil {
    log.Fatal(err)
}
data, _ := codec.Marshal(user)
```

### High-Performance (Buffer Reuse)

JSON and MessagePack support buffer reuse for high-throughput scenarios:

```go
codec := json.NewPool[User]()
buf := make([]byte, 0, 1024)

for _, user := range users {
    buf = buf[:0]
    result, _ := codec.MarshalTo(buf, user)
    process(result)
}
```

## Performance

Benchmark results (ns/op, lower is better):

| Format | Marshal | Unmarshal |
|--------|---------|-----------|
| Protocol Buffers | 219 | 245 |
| CBOR | 668 | 792 |
| JSON | 874 | 2,036 |
| MessagePack | 1,062 | 803 |
| BSON | 1,191 | 1,785 |
| TOML | 16,721 | 19,253 |
| YAML | 20,104 | 31,589 |

Run benchmarks: `make bench`

## Documentation

See the [docs/](./docs/) directory for detailed documentation on each codec:

- [JSON](./docs/json.md) - With buffer reuse methods
- [YAML](./docs/yaml.md)
- [TOML](./docs/toml.md)
- [MessagePack](./docs/msgpack.md) - With buffer reuse methods
- [BSON](./docs/bson.md)
- [CBOR](./docs/cbor.md)
- [Protocol Buffers](./docs/protobuf.md)

## Development

```bash
make test      # Run all tests
make bench     # Run benchmarks
make coverage  # Check coverage
make lint      # Run linter
make ci        # Run all CI checks
```

## License

Apache 2.0
