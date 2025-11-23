.PHONY: test test-json test-yaml test-toml test-msgpack test-protobuf test-bson test-pool \
        coverage coverage-json coverage-yaml coverage-toml coverage-msgpack coverage-protobuf coverage-bson coverage-pool \
        bench bench-json bench-msgpack bench-pool bench-all bench-comparison \
        lint security fmt vet build ci clean

# Run all tests
test: test-json test-yaml test-toml test-msgpack test-protobuf test-bson test-pool

# Run tests for JSON codec
test-json:
	@echo "Running tests for JSON codec..."
	@go test -v -race -coverprofile=coverage-json.out ./pkg/json

# Run tests for YAML codec
test-yaml:
	@echo "Running tests for YAML codec..."
	@go test -v -race -coverprofile=coverage-yaml.out ./pkg/yaml

# Run tests for TOML codec
test-toml:
	@echo "Running tests for TOML codec..."
	@go test -v -race -coverprofile=coverage-toml.out ./pkg/toml

# Run tests for msgpack codec
test-msgpack:
	@echo "Running tests for msgpack codec..."
	@go test -v -race -coverprofile=coverage-msgpack.out ./pkg/msgpack

# Run tests for protobuf codec
test-protobuf:
	@echo "Running tests for protobuf codec..."
	@go test -v -race -coverprofile=coverage-protobuf.out ./pkg/protobuf

# Run tests for bson codec
test-bson:
	@echo "Running tests for BSON codec..."
	@go test -v -race -coverprofile=coverage-bson.out ./pkg/bson

# Run tests for buffer pool
test-pool:
	@echo "Running tests for buffer pool..."
	@go test -v -race -coverprofile=coverage-pool.out ./pkg/pool

# Show coverage for all codecs
coverage: coverage-json coverage-yaml coverage-toml coverage-msgpack coverage-protobuf coverage-bson coverage-pool

# Show coverage for JSON codec
coverage-json:
	@echo "Coverage for JSON codec:"
	@go test -coverprofile=coverage-json.out ./pkg/json > /dev/null 2>&1
	@go tool cover -func=coverage-json.out | grep total

# Show coverage for YAML codec
coverage-yaml:
	@echo "Coverage for YAML codec:"
	@go test -coverprofile=coverage-yaml.out ./pkg/yaml > /dev/null 2>&1
	@go tool cover -func=coverage-yaml.out | grep total

# Show coverage for TOML codec
coverage-toml:
	@echo "Coverage for TOML codec:"
	@go test -coverprofile=coverage-toml.out ./pkg/toml > /dev/null 2>&1
	@go tool cover -func=coverage-toml.out | grep total

# Show coverage for msgpack codec
coverage-msgpack:
	@echo "Coverage for msgpack codec:"
	@go test -coverprofile=coverage-msgpack.out ./pkg/msgpack > /dev/null 2>&1
	@go tool cover -func=coverage-msgpack.out | grep total

# Show coverage for protobuf codec
coverage-protobuf:
	@echo "Coverage for protobuf codec:"
	@go test -coverprofile=coverage-protobuf.out ./pkg/protobuf > /dev/null 2>&1
	@go tool cover -func=coverage-protobuf.out | grep total

# Show coverage for BSON codec
coverage-bson:
	@echo "Coverage for BSON codec:"
	@go test -coverprofile=coverage-bson.out ./pkg/bson > /dev/null 2>&1
	@go tool cover -func=coverage-bson.out | grep total

# Show coverage for buffer pool
coverage-pool:
	@echo "Coverage for buffer pool:"
	@go test -coverprofile=coverage-pool.out ./pkg/pool > /dev/null 2>&1
	@go tool cover -func=coverage-pool.out | grep total

# Benchmark targets

# Run all benchmarks
bench-all: bench-json bench-msgpack bench-pool

# Benchmark JSON codec (all benchmarks)
bench-json:
	@echo "Running JSON codec benchmarks..."
	@go test -bench=. -benchmem -run=^$$ ./pkg/json

# Benchmark MessagePack codec (all benchmarks)
bench-msgpack:
	@echo "Running MessagePack codec benchmarks..."
	@go test -bench=. -benchmem -run=^$$ ./pkg/msgpack

# Benchmark buffer pool
bench-pool:
	@echo "Running buffer pool benchmarks..."
	@go test -bench=. -benchmem -run=^$$ ./pkg/pool

# Run comparison benchmarks (Standard vs Optimized)
bench-comparison:
	@echo "Running comparison benchmarks..."
	@echo ""
	@echo "=== JSON Codec Comparison ==="
	@go test -bench=BenchmarkComparison -benchmem -run=^$$ ./pkg/json
	@echo ""
	@echo "=== MessagePack Codec Comparison ==="
	@go test -bench=BenchmarkComparison -benchmem -run=^$$ ./pkg/msgpack

# Run optimized API benchmarks only
bench-optimized:
	@echo "Running optimized API benchmarks..."
	@echo ""
	@echo "=== JSON Optimized ==="
	@go test -bench=BenchmarkOptimized -benchmem -run=^$$ ./pkg/json
	@echo ""
	@echo "=== MessagePack Optimized ==="
	@go test -bench=BenchmarkOptimized -benchmem -run=^$$ ./pkg/msgpack

# Run baseline (standard API) benchmarks only
bench-baseline:
	@echo "Running baseline API benchmarks..."
	@echo ""
	@echo "=== JSON Standard ==="
	@go test -bench='^BenchmarkCodec_' -benchmem -run=^$$ ./pkg/json
	@echo ""
	@echo "=== MessagePack Standard ==="
	@go test -bench='^BenchmarkCodec_' -benchmem -run=^$$ ./pkg/msgpack

# Run benchmarks with CPU profiling
bench-cpu:
	@echo "Running benchmarks with CPU profiling..."
	@go test -bench=. -benchmem -cpuprofile=cpu.prof -run=^$$ ./pkg/json
	@echo "CPU profile saved to cpu.prof"
	@echo "Analyze with: go tool pprof cpu.prof"

# Run benchmarks with memory profiling
bench-mem:
	@echo "Running benchmarks with memory profiling..."
	@go test -bench=. -benchmem -memprofile=mem.prof -run=^$$ ./pkg/json
	@echo "Memory profile saved to mem.prof"
	@echo "Analyze with: go tool pprof mem.prof"

# Quick benchmark comparison (small subset for fast feedback)
bench-quick:
	@echo "Quick benchmark comparison..."
	@go test -bench='Marshal$$' -benchmem -run=^$$ ./pkg/json
	@go test -bench='Marshal$$' -benchmem -run=^$$ ./pkg/msgpack

# Linting
lint:
	@echo "Running golangci-lint..."
	@which golangci-lint > /dev/null || test -f $$(go env GOPATH)/bin/golangci-lint || (echo "golangci-lint not installed. Install with: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin" && exit 1)
	@(which golangci-lint > /dev/null && golangci-lint run --timeout=5m ./...) || $$(go env GOPATH)/bin/golangci-lint run --timeout=5m ./...

# Security scanning
security:
	@echo "Running security scan with gosec..."
	@which gosec > /dev/null || test -f $$(go env GOPATH)/bin/gosec || (echo "gosec not installed. Install with: go install github.com/securego/gosec/v2/cmd/gosec@latest" && exit 1)
	@(which gosec > /dev/null && gosec -exclude=G103 -exclude-dir=testdata -tests=false -fmt=text ./...) || $$(go env GOPATH)/bin/gosec -exclude=G103 -exclude-dir=testdata -tests=false -fmt=text ./...
	@echo ""
	@echo "Checking for known vulnerabilities..."
	@which govulncheck > /dev/null || test -f $$(go env GOPATH)/bin/govulncheck || (echo "govulncheck not installed. Install with: go install golang.org/x/vuln/cmd/govulncheck@latest" && exit 1)
	@(which govulncheck > /dev/null && govulncheck ./...) || $$(go env GOPATH)/bin/govulncheck ./...

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Vet code
vet:
	@echo "Running go vet..."
	@go vet ./...

# Build all packages
build:
	@echo "Building all packages..."
	@go build ./...
	@echo "Building examples..."
	@go build -o /tmp/codec-example ./examples/main.go
	@go build -o /tmp/codec-protobuf-example ./examples/protobuf/...

# CI target - runs all checks locally
ci: fmt vet lint security test build
	@echo ""
	@echo "========================================="
	@echo "✅ All CI checks passed!"
	@echo "========================================="

# Clean coverage and profile files
clean:
	@rm -f coverage-*.out cpu.prof mem.prof
