# Protocol Buffers Codec

> **Status: REFERENCE** · Last verified: 2026-10-09
> As-built usage of the Protocol Buffers codec in `pkg/protobuf/`. The code is authoritative.

Protocol Buffers (protobuf) is Google's binary serialization format.

## Import

```go
import "github.com/jeremyhahn/go-codec/pkg/protobuf"
```

Importing the package registers the codec in `init()`; no build tags are
needed. Import only the codec packages you use: Go links just those.
`pkg/factory` links and registers all seven codecs; use it only when you
need every format.

## Usage

Protocol Buffers requires generated code from `.proto` files:

**1. Create a .proto file:**

```protobuf
syntax = "proto3";
package myapp;
option go_package = "./";

message User {
  int64 id = 1;
  string name = 2;
  string email = 3;
}
```

**2. Generate Go code:**

```bash
protoc --go_out=. --go_opt=paths=source_relative user.proto
```

**3. Use the codec:**

```go
codec := protobuf.New[*User]()

user := &User{Id: 1, Name: "John", Email: "john@example.com"}

data, err := codec.Marshal(user)

result := &User{} // must be non-nil: Unmarshal fills the message it points to
err = codec.Unmarshal(data, &result)
```

## Factory

`factory.NewProtoBuf[*User]()` (`pkg/factory/factory.go:60`) returns the same
codec as `(codec.Codec[*User], error)`. The error is always nil.
`factory.New` rejects `codec.ProtoBuf` with an error; use `NewProtoBuf`.

## Performance

| Operation | Time | Memory | Allocs |
|-----------|------|--------|--------|
| Marshal | 219 ns/op | 32 B/op | 1 |
| Unmarshal | 245 ns/op | 24 B/op | 2 |

## Notes

- Uses `google.golang.org/protobuf`
- Requires code generation from `.proto` files
- Type must satisfy `protobuf.ProtoMessage` (`proto.Message`)
- Fastest of the formats benchmarked in this directory
- Ideal for RPC and microservices
- Schema evolution with backward compatibility
