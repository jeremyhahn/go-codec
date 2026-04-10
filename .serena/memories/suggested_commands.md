# Suggested Commands

## Testing
- `make test` - Run all unit tests (with race detection and coverage)
- `make test-<pkg>` - Test specific codec (e.g., `make test-json`, `make test-avro`)
- `make integration-test` - Run all integration tests
- `make integration-test-<pkg>` - Integration test specific codec

## Code Quality
- `make fmt` - Format code with gofmt
- `make vet` - Run go vet
- `make lint` - Run golangci-lint
- `make security` - Run gosec + govulncheck
- `make fmt-check` - Check formatting (CI-style, fails on unformatted)

## Coverage
- `make coverage` - Show coverage for all packages
- `make coverage-<pkg>` - Coverage for specific package
- `make coverage-report` - Combined HTML coverage report
- `make test-coverage` - Tests + threshold check (95%)

## Benchmarks
- `make bench` - All benchmarks
- `make bench-<pkg>` - Benchmark specific codec
- `make bench-comparison` - Standard vs Optimized comparison
- `make bench-cpu PKG=json` - CPU profiling
- `make bench-mem PKG=json` - Memory profiling

## Build
- `make build` - Build all packages
- `make generate` - Generate protobuf code
- `make examples` - Build and run examples
- `make ci-local` - Full CI pipeline locally
- `make ci` - CI in Docker

## Codec Selection
- `make build WITH_CODEC_AVRO=0 WITH_CODEC_PROTOBUF=0` - Exclude specific codecs

## Clean
- `make clean` - Clean all generated/temp files
