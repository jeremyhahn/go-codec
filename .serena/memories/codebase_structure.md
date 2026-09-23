# Codebase Structure

```
go-codec/
├── codec.go              # Root codec type definitions
├── supported.go          # Supported codec registry (build-tag gated)
├── supported_test.go
├── pkg/
│   ├── json/             # JSON codec
│   ├── yaml/             # YAML codec
│   ├── toml/             # TOML codec
│   ├── msgpack/          # MessagePack codec
│   ├── bson/             # BSON codec
│   ├── cbor/             # CBOR codec
│   ├── protobuf/         # Protocol Buffers codec
│   └── factory/          # Runtime codec selection
├── examples/
│   ├── main.go           # Multi-codec example
│   └── protobuf/         # Protobuf-specific example
├── docs/
├── .github/              # CI workflows
├── Dockerfile
├── Makefile
└── VERSION
```

## Key Design Patterns
- Go generics for type-safe `Codec[T]` interface
- Factory pattern for runtime selection
- No build tags: each codec package registers itself in `init()` when imported
