# go-codec

Generic codec library for Go with a unified API across multiple serialization formats.

## Purpose
Provides type-safe encoding/decoding via Go generics with a unified interface. Supports runtime codec selection via factory pattern and conditional compilation via build tags.

## Supported Formats
JSON, YAML, TOML, MessagePack, BSON, CBOR, Protocol Buffers

## Tech Stack
- Language: Go 1.25.5
- Module: `github.com/jeremyhahn/go-codec`
- Build: Makefile with build tags (`codec_json`, `codec_yaml`, etc.)
- CI: GitHub Actions + Docker
- Linting: golangci-lint
- Security: gosec, govulncheck
- Protobuf: protoc + protoc-gen-go
- Coverage threshold: 95%

## Architecture
- `codec.go` / `supported.go` - Root package with codec type definitions
- `pkg/{json,yaml,toml,msgpack,bson,cbor,protobuf}/` - Per-format implementations
- `pkg/factory/` - Runtime codec selection factory
- `examples/` - Usage examples including protobuf
- Each codec can be excluded via `WITH_CODEC_X=0` build flags
