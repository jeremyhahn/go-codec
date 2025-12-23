# Makefile for go-codec
#
# This Makefile uses dynamic package discovery and templates for all targets.
# All packages in pkg/ are automatically included.

# ==============================================================================
# Package Discovery
# ==============================================================================

# Discover all packages in pkg/ directory
PACKAGES := $(notdir $(wildcard pkg/*))

# Generate target names from discovered packages
TEST_TARGETS := $(addprefix test-,$(PACKAGES))
INTEGRATION_TEST_TARGETS := $(addprefix integration-test-,$(PACKAGES))
COVERAGE_TARGETS := $(addprefix coverage-,$(PACKAGES))
BENCH_TARGETS := $(addprefix bench-,$(PACKAGES))

# Proto files that need to be generated
PROTO_FILES := $(wildcard pkg/protobuf/testdata/*.proto)
PROTO_GEN_FILES := $(PROTO_FILES:.proto=.pb.go)

# Example proto files
EXAMPLE_PROTO_FILES := $(wildcard examples/protobuf/*.proto)
EXAMPLE_PROTO_GEN_FILES := $(EXAMPLE_PROTO_FILES:.proto=.pb.go)

# ==============================================================================
# Default Target
# ==============================================================================

.PHONY: all
all: build

# ==============================================================================
# Test Targets (Template-based)
# ==============================================================================

# Run all unit tests
.PHONY: test
test: $(TEST_TARGETS)

# Generic test target for any package in pkg/
# Usage: make test-json, make test-avro, make test-cbor, etc.
.PHONY: $(TEST_TARGETS)
$(TEST_TARGETS): test-%:
	@echo "Running tests for $*..."
	@go test -v -race -coverprofile=coverage-$*.out ./pkg/$*

# ==============================================================================
# Integration Test Targets (Template-based)
# ==============================================================================

# Run all integration tests
.PHONY: integration-test
integration-test: $(INTEGRATION_TEST_TARGETS)

# Generic integration test target for any package in pkg/
# Usage: make integration-test-json, make integration-test-avro, etc.
# Runs tests with integration build tag
.PHONY: $(INTEGRATION_TEST_TARGETS)
$(INTEGRATION_TEST_TARGETS): integration-test-%:
	@echo "Running integration tests for $*..."
	@go test -v -race -tags=integration -coverprofile=coverage-integration-$*.out ./pkg/$* 2>/dev/null || echo "No integration tests for $*"

# ==============================================================================
# Coverage Targets (Template-based)
# ==============================================================================

# Show coverage for all packages
.PHONY: coverage
coverage: $(COVERAGE_TARGETS)

# Generic coverage target for any package in pkg/
# Usage: make coverage-json, make coverage-avro, etc.
.PHONY: $(COVERAGE_TARGETS)
$(COVERAGE_TARGETS): coverage-%:
	@echo "Coverage for $*:"
	@go test -coverprofile=coverage-$*.out ./pkg/$* > /dev/null 2>&1
	@go tool cover -func=coverage-$*.out | grep total

# Combined coverage report
.PHONY: coverage-report
coverage-report: $(TEST_TARGETS)
	@echo "Generating combined coverage report..."
	@echo "mode: set" > coverage-all.out
	@for pkg in $(PACKAGES); do \
		tail -n +2 coverage-$$pkg.out >> coverage-all.out 2>/dev/null || true; \
	done
	@go tool cover -func=coverage-all.out | grep total
	@echo "HTML report: go tool cover -html=coverage-all.out"

# ==============================================================================
# Benchmark Targets (Template-based)
# ==============================================================================

# Run all benchmarks
.PHONY: bench
bench: $(BENCH_TARGETS)

# Alias for backwards compatibility
.PHONY: bench-all
bench-all: bench

# Generic benchmark target for any package in pkg/
# Usage: make bench-json, make bench-avro, make bench-msgpack, etc.
.PHONY: $(BENCH_TARGETS)
$(BENCH_TARGETS): bench-%:
	@echo "Running benchmarks for $*..."
	@go test -bench=. -benchmem -run=^$$ ./pkg/$* 2>/dev/null || echo "No benchmarks for $*"

# ==============================================================================
# Specialized Benchmark Targets
# ==============================================================================

# Run comparison benchmarks (Standard vs Optimized)
.PHONY: bench-comparison
bench-comparison:
	@echo "Running comparison benchmarks..."
	@for pkg in $(PACKAGES); do \
		if go test -list 'BenchmarkComparison' ./pkg/$$pkg 2>/dev/null | grep -q Benchmark; then \
			echo ""; \
			echo "=== $$pkg Comparison ==="; \
			go test -bench=BenchmarkComparison -benchmem -run=^$$ ./pkg/$$pkg; \
		fi \
	done

# Run optimized API benchmarks
.PHONY: bench-optimized
bench-optimized:
	@echo "Running optimized API benchmarks..."
	@for pkg in $(PACKAGES); do \
		if go test -list 'BenchmarkOptimized' ./pkg/$$pkg 2>/dev/null | grep -q Benchmark; then \
			echo ""; \
			echo "=== $$pkg Optimized ==="; \
			go test -bench=BenchmarkOptimized -benchmem -run=^$$ ./pkg/$$pkg; \
		fi \
	done

# Run baseline (standard API) benchmarks
.PHONY: bench-baseline
bench-baseline:
	@echo "Running baseline API benchmarks..."
	@for pkg in $(PACKAGES); do \
		if go test -list 'BenchmarkCodec_' ./pkg/$$pkg 2>/dev/null | grep -q Benchmark; then \
			echo ""; \
			echo "=== $$pkg Baseline ==="; \
			go test -bench='^BenchmarkCodec_' -benchmem -run=^$$ ./pkg/$$pkg; \
		fi \
	done

# Run benchmarks with CPU profiling (specify PKG=json)
.PHONY: bench-cpu
bench-cpu:
	@if [ -z "$(PKG)" ]; then \
		echo "Usage: make bench-cpu PKG=json"; \
		exit 1; \
	fi
	@echo "Running benchmarks with CPU profiling for $(PKG)..."
	@go test -bench=. -benchmem -cpuprofile=cpu-$(PKG).prof -run=^$$ ./pkg/$(PKG)
	@echo "CPU profile saved to cpu-$(PKG).prof"
	@echo "Analyze with: go tool pprof cpu-$(PKG).prof"

# Run benchmarks with memory profiling (specify PKG=json)
.PHONY: bench-mem
bench-mem:
	@if [ -z "$(PKG)" ]; then \
		echo "Usage: make bench-mem PKG=json"; \
		exit 1; \
	fi
	@echo "Running benchmarks with memory profiling for $(PKG)..."
	@go test -bench=. -benchmem -memprofile=mem-$(PKG).prof -run=^$$ ./pkg/$(PKG)
	@echo "Memory profile saved to mem-$(PKG).prof"
	@echo "Analyze with: go tool pprof mem-$(PKG).prof"

# Quick benchmark (Marshal only)
.PHONY: bench-quick
bench-quick:
	@echo "Quick benchmark comparison (Marshal only)..."
	@for pkg in $(PACKAGES); do \
		if go test -list 'Marshal' ./pkg/$$pkg 2>/dev/null | grep -q Benchmark; then \
			echo "=== $$pkg ==="; \
			go test -bench='Marshal$$' -benchmem -run=^$$ ./pkg/$$pkg 2>/dev/null | grep -E "(Benchmark|ns/op)" || true; \
		fi \
	done

# ==============================================================================
# Code Generation Targets
# ==============================================================================

# Generate all code (protobuf, etc.)
.PHONY: generate
generate: generate-proto

# Generate protobuf files
.PHONY: generate-proto
generate-proto: $(PROTO_GEN_FILES)

# Pattern rule for generating .pb.go from .proto files
pkg/protobuf/testdata/%.pb.go: pkg/protobuf/testdata/%.proto
	@echo "Generating protobuf code for $<..."
	@protoc --go_out=. --go_opt=paths=source_relative $<

# ==============================================================================
# Code Quality Targets
# ==============================================================================

# Linting
.PHONY: lint
lint:
	@echo "Running golangci-lint..."
	@which golangci-lint > /dev/null || test -f $$(go env GOPATH)/bin/golangci-lint || \
		(echo "golangci-lint not installed. Install with: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin" && exit 1)
	@(which golangci-lint > /dev/null && golangci-lint run --timeout=5m ./...) || \
		$$(go env GOPATH)/bin/golangci-lint run --timeout=5m ./...

# Security scanning
.PHONY: security
security:
	@echo "Running security scan with gosec..."
	@which gosec > /dev/null || test -f $$(go env GOPATH)/bin/gosec || \
		(echo "gosec not installed. Install with: go install github.com/securego/gosec/v2/cmd/gosec@latest" && exit 1)
	@(which gosec > /dev/null && gosec -exclude=G103 -exclude-dir=testdata -tests=false -fmt=text ./...) || \
		$$(go env GOPATH)/bin/gosec -exclude=G103 -exclude-dir=testdata -tests=false -fmt=text ./...
	@echo ""
	@echo "Checking for known vulnerabilities..."
	@which govulncheck > /dev/null || test -f $$(go env GOPATH)/bin/govulncheck || \
		(echo "govulncheck not installed. Install with: go install golang.org/x/vuln/cmd/govulncheck@latest" && exit 1)
	@(which govulncheck > /dev/null && govulncheck ./...) || \
		$$(go env GOPATH)/bin/govulncheck ./...

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Vet code
.PHONY: vet
vet:
	@echo "Running go vet..."
	@go vet ./...

# ==============================================================================
# Build Targets
# ==============================================================================

# Build all packages
.PHONY: build
build:
	@echo "Building all packages..."
	@go build ./...
	@echo "Building examples..."
	@go build -o /tmp/codec-example ./examples/main.go
	@go build -o /tmp/codec-protobuf-example ./examples/protobuf/...

# ==============================================================================
# Examples Targets
# ==============================================================================

# Generate example protobuf files
examples/protobuf/%.pb.go: examples/protobuf/%.proto
	@echo "Generating protobuf code for $<..."
	@protoc --go_out=. --go_opt=paths=source_relative $<

# Build and run all examples
.PHONY: examples
examples: $(EXAMPLE_PROTO_GEN_FILES)
	@echo "=== Building and Running Examples ==="
	@echo ""
	@echo "--- Main Example (all codecs) ---"
	@go run ./examples/main.go
	@echo ""
	@echo "--- Protocol Buffers Example ---"
	@go run ./examples/protobuf/...

# Run examples in Docker (includes protobuf tooling)
.PHONY: docker-examples
docker-examples:
	@echo "Building Docker image..."
	@docker build -t go-codec .
	@echo ""
	@docker run --rm go-codec make examples

# Alias for backwards compatibility
.PHONY: docker-ci
docker-ci: ci

# ==============================================================================
# CI/CD Targets
# ==============================================================================

# Coverage threshold (must match .github/workflows/ci.yml)
COVERAGE_THRESHOLD := 95

# Check for required CI tools
.PHONY: check-tools
check-tools:
	@echo "Checking required tools..."
	@command -v protoc >/dev/null 2>&1 || { echo "ERROR: protoc not found. Install with: sudo apt-get install protobuf-compiler"; exit 1; }
	@command -v protoc-gen-go >/dev/null 2>&1 || { echo "ERROR: protoc-gen-go not found. Install with: go install google.golang.org/protobuf/cmd/protoc-gen-go@latest"; exit 1; }
	@command -v golangci-lint >/dev/null 2>&1 || { echo "ERROR: golangci-lint not found. Install with: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$$(go env GOPATH)/bin"; exit 1; }
	@command -v gosec >/dev/null 2>&1 || { echo "ERROR: gosec not found. Install with: go install github.com/securego/gosec/v2/cmd/gosec@latest"; exit 1; }
	@command -v govulncheck >/dev/null 2>&1 || { echo "ERROR: govulncheck not found. Install with: go install golang.org/x/vuln/cmd/govulncheck@latest"; exit 1; }
	@echo "All required tools are installed"

# Install all required CI tools
.PHONY: install-tools
install-tools:
	@echo "Installing CI tools..."
	@echo "Installing protoc-gen-go..."
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@echo "Installing golangci-lint..."
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v2.7.2
	@echo "Installing gosec..."
	@go install github.com/securego/gosec/v2/cmd/gosec@latest
	@echo "Installing govulncheck..."
	@go install golang.org/x/vuln/cmd/govulncheck@latest
	@echo ""
	@echo "All tools installed. Note: protoc must be installed separately:"
	@echo "  Ubuntu/Debian: sudo apt-get install protobuf-compiler"
	@echo "  macOS: brew install protobuf"

# CI target - runs all checks in Docker (mirrors GitHub Actions)
.PHONY: ci
ci:
	@echo "Building Docker image..."
	@docker build -t go-codec .
	@echo ""
	@docker run --rm go-codec make ci-local

# CI-local target - runs all checks locally (used inside Docker or with local tools)
.PHONY: ci-local
ci-local: check-tools generate generate-examples fmt-check vet lint security test-coverage build
	@echo ""
	@echo "========================================="
	@echo "All CI checks passed!"
	@echo "========================================="

# Format check (fails if code is not formatted)
.PHONY: fmt-check
fmt-check:
	@echo "Checking code formatting..."
	@if [ -n "$$(gofmt -l .)" ]; then \
		echo "Code is not formatted:"; \
		gofmt -l .; \
		exit 1; \
	fi
	@echo "Code formatting OK"

# Test with coverage and threshold check
.PHONY: test-coverage
test-coverage: $(TEST_TARGETS)
	@echo ""
	@echo "Checking coverage threshold..."
	@echo "mode: set" > coverage-all.out
	@for pkg in $(PACKAGES); do \
		tail -n +2 coverage-$$pkg.out >> coverage-all.out 2>/dev/null || true; \
	done
	@COVERAGE=$$(go tool cover -func=coverage-all.out | grep total | awk '{print $$3}' | sed 's/%//'); \
	COVERAGE_INT=$$(echo "$$COVERAGE" | awk '{printf "%d", $$1}'); \
	echo "Total coverage: $${COVERAGE}%"; \
	if [ "$${COVERAGE_INT}" -lt "$(COVERAGE_THRESHOLD)" ]; then \
		echo "Coverage $${COVERAGE}% is below threshold $(COVERAGE_THRESHOLD)%"; \
		exit 1; \
	else \
		echo "Coverage $${COVERAGE}% meets threshold $(COVERAGE_THRESHOLD)%"; \
	fi

# Generate example protobuf files (for CI)
.PHONY: generate-examples
generate-examples: $(EXAMPLE_PROTO_GEN_FILES)

# ==============================================================================
# Clean Targets
# ==============================================================================

# Clean all generated and temporary files
.PHONY: clean
clean:
	@echo "Cleaning generated files..."
	@rm -f coverage-*.out cpu-*.prof mem-*.prof cpu.prof mem.prof *.log *.out
	@rm -f pkg/protobuf/testdata/*.pb.go
	@rm -f examples/protobuf/*.pb.go
	@rm -f /tmp/codec-example /tmp/codec-protobuf-example
	@echo "Clean complete."

# Clean only coverage files
.PHONY: clean-coverage
clean-coverage:
	@rm -f coverage-*.out

# Clean only profiling files
.PHONY: clean-prof
clean-prof:
	@rm -f cpu-*.prof mem-*.prof cpu.prof mem.prof

# ==============================================================================
# Help Target
# ==============================================================================

.PHONY: help
help:
	@echo "go-codec Makefile"
	@echo ""
	@echo "Discovered packages: $(PACKAGES)"
	@echo ""
	@echo "All targets support any package in pkg/ via templates."
	@echo ""
	@echo "Test targets:"
	@echo "  make test                    - Run all unit tests"
	@echo "  make test-<pkg>              - Run tests for package (e.g., test-json, test-avro)"
	@echo "  make integration-test        - Run all integration tests"
	@echo "  make integration-test-<pkg>  - Run integration tests for package"
	@echo ""
	@echo "Coverage targets:"
	@echo "  make coverage                - Show coverage for all packages"
	@echo "  make coverage-<pkg>          - Show coverage for package"
	@echo "  make coverage-report         - Generate combined coverage report"
	@echo ""
	@echo "Benchmark targets:"
	@echo "  make bench                   - Run all benchmarks"
	@echo "  make bench-<pkg>             - Run benchmarks for package (e.g., bench-json)"
	@echo "  make bench-comparison        - Run comparison benchmarks (all packages)"
	@echo "  make bench-optimized         - Run optimized API benchmarks"
	@echo "  make bench-baseline          - Run baseline API benchmarks"
	@echo "  make bench-cpu PKG=<pkg>     - Run benchmarks with CPU profiling"
	@echo "  make bench-mem PKG=<pkg>     - Run benchmarks with memory profiling"
	@echo "  make bench-quick             - Quick Marshal benchmark comparison"
	@echo ""
	@echo "Code generation:"
	@echo "  make generate                - Generate all code (protobuf, etc.)"
	@echo "  make generate-proto          - Generate protobuf files"
	@echo ""
	@echo "Code quality:"
	@echo "  make fmt                     - Format code"
	@echo "  make vet                     - Run go vet"
	@echo "  make lint                    - Run golangci-lint"
	@echo "  make security                - Run security scans"
	@echo ""
	@echo "Build:"
	@echo "  make build                   - Build all packages"
	@echo "  make examples                - Build and run all examples"
	@echo "  make docker-examples         - Run examples in Docker (includes protobuf)"
	@echo "  make ci                      - Run all CI checks in Docker (mirrors GitHub Actions)"
	@echo "  make ci-local                - Run all CI checks locally (requires tools)"
	@echo ""
	@echo "Clean:"
	@echo "  make clean                   - Clean all generated files"
	@echo "  make clean-coverage          - Clean only coverage files"
	@echo "  make clean-prof              - Clean only profiling files"
