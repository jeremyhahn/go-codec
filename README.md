# go-codec

A generic codec library for Go with a unified API across multiple serialization formats.

## Features

- **Unified API** - Same interface for all formats
- **Go Generics** - Type-safe encoding/decoding
- **8 Formats** - JSON, YAML, TOML, MessagePack, BSON, CBOR, Avro, Protocol Buffers
- **Modular Imports** - Only import the codecs you need
- **Conditional Compilation** - Exclude codecs via build tags
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

Use the factory when you need runtime codec selection:

```go
import (
    "github.com/jeremyhahn/go-codec"
    "github.com/jeremyhahn/go-codec/pkg/factory"
)

c, err := factory.New[User](codec.JSON)
if err != nil {
    log.Fatal(err) // codec not compiled in
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
| Avro | `pkg/avro` | `avro:"name"` |
| Protocol Buffers | `pkg/protobuf` | (generated) |

## Conditional Compilation

Exclude codecs at build time to reduce binary size:

```bash
# Build without Avro and Protobuf
make build WITH_CODEC_AVRO=0 WITH_CODEC_PROTOBUF=0

# Or with go build
go build -tags "codec_json,codec_yaml"
```

Query available codecs at runtime:

```go
import "github.com/jeremyhahn/go-codec"

for _, c := range codec.SupportedCodecs() {
    fmt.Println(c) // prints only compiled-in codecs
}

if codec.IsSupported(codec.Avro) {
    // use Avro
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
| Avro | 420 | 221 |
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
- [Avro](./docs/avro.md) - With automatic schema inference
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
