# go-codec specification

> **Status: DRAFT** · Scope: go-codec · Last verified: 2026-11-12

## Purpose

go-codec is a Layer 0 Go library: multi-format serialization behind one common generic `Codec[T]` interface.
It supports JSON, YAML, TOML, MessagePack, BSON, CBOR and Protocol Buffers.
Every consuming project in the portfolio uses it for consistent encoding, and it pins no in-house modules.

## Boundaries

- Owns: the `Codec[T]` and `OptimizedCodec[T]` interfaces, the codec type constants and runtime registry, one package per format, and the factory.
- Owns: buffer reuse helpers for high-throughput encoding (`pkg/pool`, `NewPool` in `pkg/json` and `pkg/msgpack`).
- Never: depends on another in-house module.
- Never: does transport, storage, schema management or protobuf code generation; protobuf types come generated from the consumer.
- Never: uses build tags to choose codecs; the import is the opt-in.

## Features

| Feature | Status | Requirements | Detail |
|---------|--------|--------------|--------|
| codec-interface | shipped | Every codec implements `Codec[T]` in [codec.go](../codec.go). It marshals to and unmarshals from bytes. It encodes to an `io.Writer` and decodes from an `io.Reader`. | none |
| codec-registry | shipped | Importing a codec package registers it in `init()` ([supported.go](../supported.go)). `SupportedCodecs` and `IsSupported` report the registered codecs. An unregistered codec returns `ErrCodecNotSupported`. | none |
| codec-factory | shipped | `factory.New` selects a codec by `codec.Type` at runtime. Importing `pkg/factory` registers all seven codecs. `factory.NewProtoBuf` builds a protobuf codec for generated messages. | none |
| json-codec | shipped | `pkg/json` encodes and decodes JSON with `json` struct tags. | [docs/json.md](json.md) |
| yaml-codec | shipped | `pkg/yaml` encodes and decodes YAML with `yaml` struct tags. | [docs/yaml.md](yaml.md) |
| toml-codec | shipped | `pkg/toml` encodes and decodes TOML with `toml` struct tags. | [docs/toml.md](toml.md) |
| msgpack-codec | shipped | `pkg/msgpack` encodes and decodes MessagePack with `msgpack` struct tags. | [docs/msgpack.md](msgpack.md) |
| bson-codec | shipped | `pkg/bson` encodes and decodes BSON with `bson` struct tags. | [docs/bson.md](bson.md) |
| cbor-codec | shipped | `pkg/cbor` encodes and decodes CBOR with `cbor` struct tags. `NewWithOptions` accepts strict decode and encode options. | [docs/cbor.md](cbor.md) |
| protobuf-codec | shipped | `pkg/protobuf` encodes and decodes generated Protocol Buffers messages. The type parameter must be a proto message. | [docs/protobuf.md](protobuf.md) |
| buffer-reuse | shipped | JSON and MessagePack offer `NewPool`, an `OptimizedCodec[T]` that marshals into a caller buffer. `pkg/pool` pools byte buffers. | [docs/json.md](json.md) |
| release-readiness | in progress | The repository meets public release readiness ([#26](../../../issues/26)). The readiness PR #29 merges. | none |
| docs-index-and-banners | in progress | Every doc carries the index, a status banner and a verified stamp ([#33](../../../issues/33)). | [docs/README.md](README.md) |

## Interfaces

- SDK: Go module `github.com/jeremyhahn/go-codec`; root package `codec` plus `pkg/json`, `pkg/yaml`, `pkg/toml`, `pkg/msgpack`, `pkg/bson`, `pkg/cbor`, `pkg/protobuf`, `pkg/factory` and `pkg/pool`.
- CLI: none.
- API: none; in-process library only.

## Release criteria

Milestone: Release candidate (M2).

- Public release readiness ([#26](../../../issues/26)) is merged.
- Docs carry the index, status banners and stamps ([#33](../../../issues/33)).
- CI is green on main.
