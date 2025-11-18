# go-codec

A flexible, generic codec/serialization library for Go that provides a uniform API for encoding and decoding data across multiple formats.

## Features

- **Generic API**: Works with any Go type using generics
- **Multiple Formats**: Support for JSON, YAML, TOML, MessagePack, Protocol Buffers, and BSON
- **Factory Pattern**: Easy codec instantiation through a unified factory
- **Uniform Interface**: Consistent API across all codec implementations
- **100% Test Coverage**: All codecs are thoroughly tested
- **Simple to Use**: Intuitive encode/decode operations

## Supported Codecs

- **JSON** - JavaScript Object Notation
- **YAML** - YAML Ain't Markup Language
- **TOML** - Tom's Obvious, Minimal Language
- **MessagePack** - Efficient binary serialization format
- **Protocol Buffers** - Google's efficient binary protocol
- **BSON** - Binary JSON (MongoDB format)

## Installation

```bash
go get github.com/jeremyhahn/go-codec
```

## Usage

### Basic Example

```go
package main

import (
	"bytes"
	"fmt"
	"log"

	"github.com/jeremyhahn/go-codec"
)

type Person struct {
	Name  string `json:"name" yaml:"name" toml:"name" msgpack:"name"`
	Age   int    `json:"age" yaml:"age" toml:"age" msgpack:"age"`
	Email string `json:"email" yaml:"email" toml:"email" msgpack:"email"`
}

func main() {
	person := Person{
		Name:  "John Doe",
		Age:   30,
		Email: "john@example.com",
	}

	// Create a JSON codec
	jsonCodec, err := codec.New[Person](codec.JSON)
	if err != nil {
		log.Fatal(err)
	}

	// Marshal to JSON bytes
	jsonBytes, err := jsonCodec.Marshal(person)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("JSON: %s\n", string(jsonBytes))

	// Unmarshal from JSON bytes
	var decodedPerson Person
	err = jsonCodec.Unmarshal(jsonBytes, &decodedPerson)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Decoded: %+v\n", decodedPerson)
}
```

### Using Different Codecs

```go
// YAML codec
yamlCodec, _ := codec.New[Person](codec.YAML)
yamlBytes, _ := yamlCodec.Marshal(person)

// TOML codec
tomlCodec, _ := codec.New[Person](codec.TOML)
tomlBytes, _ := tomlCodec.Marshal(person)

// MessagePack codec
msgpackCodec, _ := codec.New[Person](codec.MsgPack)
msgpackBytes, _ := msgpackCodec.Marshal(person)

// BSON codec
bsonCodec, _ := codec.New[Person](codec.BSON)
bsonBytes, _ := bsonCodec.Marshal(person)
```

### Stream Encoding/Decoding

```go
var buf bytes.Buffer

// Encode to a writer
err = jsonCodec.Encode(&buf, person)

// Decode from a reader
var person2 Person
err = jsonCodec.Decode(&buf, &person2)
```

### Using Protocol Buffers

Protocol Buffers requires a different approach since types must be generated from `.proto` files:

1. **Create a .proto file** (person.proto):
```protobuf
syntax = "proto3";

package main;

message Person {
  string name = 1;
  int32 age = 2;
  string email = 3;
}
```

2. **Generate Go code**:
```bash
protoc --go_out=. --go_opt=paths=source_relative person.proto
```

3. **Use the codec**:
```go
// Note: Use NewProtoBuf instead of New for protobuf types
protobufCodec := codec.NewProtoBuf[*Person]()

person := &Person{
    Name:  "John Doe",
    Age:   30,
    Email: "john@example.com",
}

// Marshal and unmarshal
data, _ := protobufCodec.Marshal(person)
decoded := &Person{}
protobufCodec.Unmarshal(data, &decoded)
```

See `examples/protobuf/` for a complete working example.

## API Reference

### Codec Interface

```go
type Codec[T any] interface {
    // Encode serializes the given data to the writer
    Encode(w io.Writer, data T) error

    // Decode deserializes data from the reader into the provided type
    Decode(r io.Reader, data *T) error

    // Marshal serializes the given data to bytes
    Marshal(data T) ([]byte, error)

    // Unmarshal deserializes bytes into the provided type
    Unmarshal(data []byte, v *T) error
}
```

### Codec Types

```go
const (
    JSON     Type = "json"
    YAML     Type = "yaml"
    TOML     Type = "toml"
    MsgPack  Type = "msgpack"
    ProtoBuf Type = "protobuf"
    BSON     Type = "bson"
)
```

### Factory Functions

```go
// For JSON, YAML, TOML, and MessagePack
func New[T any](codecType Type) (Codec[T], error)

// For Protocol Buffers (requires proto.Message types)
func NewProtoBuf[T proto.Message]() Codec[T]
```

## Testing

Run all tests:

```bash
make test
```

Run tests for a specific codec:

```bash
make test-json
make test-yaml
make test-toml
make test-msgpack
make test-protobuf
make test-bson
```

Check test coverage:

```bash
make coverage
```

## Project Structure

```
go-codec/
├── codec.go              # Core interface definitions
├── factory.go            # Codec factory
├── protobuf_factory.go   # Protocol Buffers factory
├── pkg/
│   ├── json/            # JSON codec implementation
│   ├── yaml/            # YAML codec implementation
│   ├── toml/            # TOML codec implementation
│   ├── msgpack/         # MessagePack codec implementation
│   ├── protobuf/        # Protocol Buffers codec implementation
│   └── bson/            # BSON codec implementation
├── examples/
│   ├── main.go          # Basic example
│   └── protobuf/        # Protocol Buffers example
├── Makefile             # Build and test targets
└── README.md            # This file
```

## Dependencies

- [gopkg.in/yaml.v3](https://gopkg.in/yaml.v3) - YAML support
- [github.com/BurntSushi/toml](https://github.com/BurntSushi/toml) - TOML support
- [github.com/vmihailenco/msgpack/v5](https://github.com/vmihailenco/msgpack) - MessagePack support
- [google.golang.org/protobuf](https://google.golang.org/protobuf) - Protocol Buffers support
- [go.mongodb.org/mongo-driver/bson](https://go.mongodb.org/mongo-driver) - BSON support

## License

Apache 2.0
