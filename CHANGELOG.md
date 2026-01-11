# Changelog

All notable changes to this project will be documented in this file.

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
