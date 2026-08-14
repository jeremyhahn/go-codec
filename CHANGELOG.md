# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.4.0] - 2026-08-14

### Removed — BREAKING: the Avro codec is gone

- **Avro codec removed entirely** (2026-08-14, owner decision): the `pkg/avro`
  package, the `codec.Avro` type constant, the `codec_avro` build tag and
  `WITH_CODEC_AVRO` Makefile variable, `docs/avro.md`, and the
  `github.com/hamba/avro/v2` dependency.

  The upstream library, `github.com/hamba/avro/v2`, is **archived** — v2.31.0 is
  its final release — and carries three unfixed advisories that are reachable
  from this package's decode path: GO-2026-5046 (CPU exhaustion), GO-2026-5047
  (integer overflow) and GO-2026-5048 (unbounded map allocations), all reported
  by the Go vulnerability database as `Fixed in: N/A`. No fix will ever ship.

  Every alternative was evaluated and none is production grade: the only patched
  code is a single-maintainer fork of the archived project, and `linkedin/goavro`
  is in maintenance mode (LinkedIn itself migrated to hamba) and lacks struct
  marshaling. Rather than take a supply-chain dependency on an unmaintained or
  forked parser for untrusted input, Avro support is withdrawn.

  Callers needing Avro should depend on an Avro library directly, so the
  security trade-off is theirs to make explicitly.

#### Versioning note — a deliberate deviation from SemVer 2.0.0

Removing exported API (`codec.Avro` and the whole `pkg/avro` package) from a
`1.x` module is a MAJOR change under SemVer 2.0.0, which this project otherwise
follows. Under Go's semantic import versioning a major bump would also require
the module path to become `github.com/jeremyhahn/go-codec/v2`, changing every
import line in every consumer.

This release is **1.4.0 instead of 2.0.0**, by owner decision on 2026-08-14, on
two grounds: no consumer used the removed API — every dependent pinned
`go-codec v1.3.1` and none imported `pkg/avro` or referenced `codec.Avro`, so
real-world breakage is zero — and the module is pre-publication, with no
external users whose builds could break.

Recorded here explicitly so it does not become precedent. Once this module is
published, exported-API removals take a major bump and the `/v2` path with it.

## [1.3.0] - 2025-01-10

### Added
- **Conditional compilation** via build tags (`codec_json`, `codec_yaml`, etc.)
- **Makefile variables** (`WITH_CODEC_X=0/1`) to include/exclude codecs at build time
- **Runtime codec discovery** with `SupportedCodecs()` and `IsSupported()` functions
- **ErrCodecNotSupported** error type for disabled codecs

### Changed
- Factory functions now return errors when codec is unavailable
- All codec packages use build tags for conditional inclusion

## [1.2.0] - 2024-12-22

### Added
- **Avro codec** with automatic schema inference from Go types
- **CBOR codec** for compact binary serialization
- **Factory package** (`pkg/factory`) for modular imports
- **Docker support** with `make docker-examples` and `make docker-ci` targets
- **Documentation** in `docs/` for each codec format
- **Benchmarks** for all codec packages

### Changed
- Moved factory to `pkg/factory` for selective imports
- Updated examples to demonstrate all 8 codec formats
- Improved Makefile with template-based targets
- Simplified YAML encoder (removed unreachable error handling)

### Removed
- Root-level `factory.go` and `protobuf_factory.go` (moved to `pkg/factory`)
- `API_REFERENCE.md` (consolidated into `docs/`)

## [1.1.0] - 2024-11-23

### Added
- Buffer reuse optimization (`MarshalTo`, `AppendMarshal`, `UnmarshalFrom`)
- CI/CD pipeline with GitHub Actions
- Security scanning with gosec and govulncheck

## [1.0.0] - 2024-11-22

### Added
- Initial release
- JSON, YAML, TOML, MessagePack, Protocol Buffers, BSON codecs
- Generic `Codec[T]` interface
- Stream encoding/decoding with `io.Reader`/`io.Writer`
- 100% test coverage
