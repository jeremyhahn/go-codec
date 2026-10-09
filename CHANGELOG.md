# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.5.0-alpha] - 2026-10-09

### Changed — untagged builds now work: the import is the opt-in (#11)

- **Codec build tags removed** (owner ruling 2026-09-22). Every codec package
  (`pkg/json`, `pkg/yaml`, `pkg/toml`, `pkg/msgpack`, `pkg/bson`, `pkg/cbor`,
  `pkg/protobuf`) now always compiles its real implementation and registers
  itself in `init()` when imported, like `database/sql` drivers. Previously a
  plain `go build` / `go test` with no `-tags codec_*` compiled every codec as
  an inert stub, so `SupportedCodecs()` was empty and every constructor
  returned `ErrCodecNotSupported`.
- Removed: the `*_stub.go` fallbacks, the `codec_json` … `codec_protobuf` and
  `codec_none` build constraints, and the Makefile `WITH_CODEC_*` variables.
  Go links only imported packages, so the tags never saved binary size:
  import just the codec packages you use, or `pkg/factory` for all seven.
- `ErrCodecNotSupported` now points at importing the codec package instead of
  a build tag. `factory.NewProtoBuf` always succeeds, because importing
  `pkg/factory` registers protobuf.
- Consumers that still pass `-tags codec_*` keep working; unknown tags are
  ignored, and the flags can be dropped.

### Changed — `make release-alpha` replaced with handbook template v2.3.9 (#10)

- The version is now explicit: `make release-alpha VERSION=vX.Y.Z-alpha`. The
  old recipe bumped PATCH only, made lightweight unsigned tags as the local git
  identity, pushed every local tag with `--tags` to a `dev` remote this repo
  does not have, and used `|| true` to hide a failed VERSION commit.
- The new recipe signs every git write through `agent-git.sh` (found on
  `PATH`, or pass `AGENT_GIT=`) and creates an annotated tag. It pushes only the
  release commit and that one tag to `origin`, atomically. It releases only
  from `main`, and it refuses to cut a version that ends up with no
  `## [X.Y.Z-alpha]` section in `CHANGELOG.md`.
- Releasing now also assembles the changelog, as `release-versioning.md` §5
  requires: before tagging, `make release-alpha` moves everything under
  `## [Unreleased]` and every `changelog.d/*.md` fragment into a dated
  `## [X.Y.Z-alpha]` section (`scripts/assemble-changelog.sh`). It then deletes
  the fragments and commits the result through `agent-git.sh`. It refuses when
  the version already has a section but entries are still pending.
- Regression tests for the recipe (`internal/releasetest`) run this repo's
  own Makefile against a throwaway remote.
- A `VERSION=` value containing a newline is refused before anything is
  committed. The template's line-oriented `grep` check accepted
  `vX.Y.Z-alpha` followed by a newline and more text.

### Added — `cbor.NewWithOptions` for decode strictness (#24)

- `cbor.NewWithOptions[T](dec cbor.DecOptions, enc cbor.EncOptions)` builds
  the fxamacker/cbor decode and encode modes once and returns a codec that
  uses them for `Marshal`, `Unmarshal`, `Encode` and `Decode`. Callers
  decoding attacker-controlled CBOR can now refuse duplicate map keys,
  indefinite-length items, tags and invalid UTF-8, and set their own nesting
  bound.
- Invalid options fail at construction with a typed `*cbor.OptionsError`
  (`errors.Is(err, cbor.ErrInvalidOptions)`), never per call.
- `cbor.New[T]()` keeps fxamacker's default options, so existing behavior
  is unchanged.

### Changed — docs index, status banners and doc corrections (#33)

- Added `docs/README.md`, an index of the seven per-format codec guides.
- Every `docs/*.md` now carries a REFERENCE status banner and a
  `Last verified:` stamp, checked against `pkg/`.
- `docs/protobuf.md`: the example now uses `protobuf.New[*User]()` and a
  non-nil destination message for `Unmarshal`. It documents that
  `factory.NewProtoBuf` returns `(codec.Codec[T], error)` with an always-nil
  error, and that importing `pkg/factory` links all seven codecs.
- `docs/msgpack.md`: removed the unmeasured "30-50% smaller" claim.
- Documentation only; no code or behavior change.

### Fixed — local CI gates can no longer pass vacuously (#36)

- `make lint` now also runs `actionlint` over `.github/workflows/` and fails
  when `actionlint` is not installed, printing the `go install` hint.
- `make trivy` and `make trivy-image` exit non-zero with the install URL when
  `trivy` is missing, instead of printing "skipping" and exiting 0.
  `make trivy-image` also fails when `docker` is missing.
- `.github/workflows/ci.yml`: quoted the shell expansions that actionlint's
  shellcheck pass flagged (`basename "$pkg"`, `"$GITHUB_OUTPUT"`) and read the
  coverage threshold from the `COVERAGE_THRESHOLD` shell variable instead of
  inlining `${{ env.COVERAGE_THRESHOLD }}` into the `-lt` test (SC2170).

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
