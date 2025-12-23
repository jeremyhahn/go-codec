# Protocol Buffers Codec

Protocol Buffers (protobuf) is Google's binary serialization format.

## Import

```go
import "github.com/jeremyhahn/go-codec/pkg/factory"
```

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
codec := factory.NewProtoBuf[*User]()

user := &User{Id: 1, Name: "John", Email: "john@example.com"}

data, err := codec.Marshal(user)
err = codec.Unmarshal(data, &result)
```

## Direct Package Import

```go
import "github.com/jeremyhahn/go-codec/pkg/protobuf"

codec := protobuf.New[*User]()
```

## Performance

| Operation | Time | Memory | Allocs |
|-----------|------|--------|--------|
| Marshal | 219 ns/op | 32 B/op | 1 |
| Unmarshal | 245 ns/op | 24 B/op | 2 |

## Notes

- Uses `google.golang.org/protobuf`
- Requires code generation from `.proto` files
- Type must implement `proto.Message`
- Fastest serialization format
- Ideal for RPC and microservices
- Schema evolution with backward compatibility
