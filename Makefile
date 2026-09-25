# Makefile for go-codec
#
# This Makefile uses dynamic package discovery and templates for all targets.
# All packages in pkg/ are automatically included.

# ==============================================================================
# Codec Selection
# ==============================================================================
# There are no codec build tags. Each codec package (pkg/json, pkg/yaml, ...)
# always compiles its real implementation and registers itself in init() when
# a program imports it, so the import is the opt-in. Every target below builds
# and tests with no -tags.

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
# Usage: make test-json, make test-cbor, make test-yaml, etc.
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
# Usage: make integration-test-json, make integration-test-cbor, etc.
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
# Usage: make coverage-json, make coverage-cbor, etc.
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
# Usage: make bench-json, make bench-cbor, make bench-msgpack, etc.
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

# Security scanning — the full devops.md §8 suite.
# gosec, govulncheck, go mod verify, trivy fs, trivy image.
.PHONY: security
security: gosec vuln modverify trivy trivy-image
	@echo ""
	@echo "Security suite passed"

# gosec + govulncheck. Kept as its own target so `make security` composes it
# with the trivy scans below rather than duplicating the invocations.
#
# GOSEC_FLAGS stays identical to the gosec step in .github/workflows/ci.yml --
# devops.md §8 makes `make security` the contract CI runs, so the two drifting
# apart means a developer's "clean" run proves nothing. No -tags: every codec
# implementation compiles unconditionally, so the scanners see all of it.
# No -exclude rules: the only unsafe usage in the tree is in protoc-generated
# *.pb.go, which -exclude-generated already drops, so a blanket -exclude=G103
# would buy nothing and hide hand-written unsafe if any ever landed.
GOSEC_FLAGS := -exclude-generated -severity medium -confidence medium \
	-exclude-dir=testdata

# Each target resolves the binary once and runs it once. The previous
# `(which tool && tool ...) || $(GOPATH)/bin/tool ...` form treated a scanner
# that ran and *found something* as "not on PATH" and scanned the whole tree a
# second time, doubling the output of every real failure.
.PHONY: gosec vuln
gosec:
	@echo "Running security scan with gosec..."
	@BIN=$$(command -v gosec || echo $$(go env GOPATH)/bin/gosec); \
	test -x "$$BIN" || { echo "gosec not installed. Install with: go install github.com/securego/gosec/v2/cmd/gosec@latest"; exit 1; }; \
	"$$BIN" $(GOSEC_FLAGS) -fmt=text ./...

vuln:
	@echo ""
	@echo "Checking for known vulnerabilities..."
	@BIN=$$(command -v govulncheck || echo $$(go env GOPATH)/bin/govulncheck); \
	test -x "$$BIN" || { echo "govulncheck not installed. Install with: go install golang.org/x/vuln/cmd/govulncheck@latest"; exit 1; }; \
	"$$BIN" ./...

# Module integrity against go.sum.
.PHONY: modverify
modverify:
	@echo ""
	@echo "Verifying module integrity..."
	@go mod verify

# Filesystem CVE/secret scan, gated on CRITICAL/HIGH. examples/ is skipped:
# it is demo code that never ships to a consumer of the library.
.PHONY: trivy
trivy:
	@command -v trivy >/dev/null 2>&1 || { \
		echo "trivy not installed - skipping filesystem scan (see https://trivy.dev/latest/getting-started/installation/)"; \
		exit 0; \
	}; \
	echo "Running trivy filesystem scan..."; \
	trivy fs --scanners vuln,secret --severity CRITICAL,HIGH --exit-code 1 --skip-dirs examples .

# Image scan of the only container this repo builds: the local development /
# CI-tooling image (golang:<stable>-bookworm + protoc + the linters), which
# `make ci` builds, runs `make ci-local` inside, and discards. It is never
# pushed to a registry.
#
# Two scoping decisions, both deliberate:
#
#   --ignore-unfixed  The Debian package set is upstream's. A CVE with no
#                     released fix cannot be actioned by rebuilding and would
#                     pin this gate permanently red; a CVE that HAS a fix
#                     means the base is stale, which is a real finding.
#   --skip-dirs       go/bin, go/pkg/mod and usr/local/go hold the upstream
#                     release builds of golangci-lint/gosec/govulncheck and
#                     their module cache, baked in only so the container can
#                     run the checks. Their embedded stdlib version is a
#                     property of someone else's release build, not of
#                     go-codec, and gosec's own hardcoded_credentials.go rule
#                     fixture trips the secret scanner. Nothing under them
#                     reaches a consumer.
#
# Everything go-codec actually ships is gated unconditionally by `make trivy`
# and govulncheck.
.PHONY: trivy-image
trivy-image:
	@command -v docker >/dev/null 2>&1 || { echo "docker not installed - skipping image scan"; exit 0; }; \
	command -v trivy >/dev/null 2>&1 || { echo "trivy not installed - skipping image scan"; exit 0; }; \
	echo "Building and scanning container image..."; \
	docker build --pull -t go-codec:security . && \
	trivy image --scanners vuln,secret --severity CRITICAL,HIGH --ignore-unfixed --exit-code 1 \
		--skip-dirs go/bin --skip-dirs go/pkg/mod --skip-dirs usr/local/go go-codec:security

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

# CI target - runs all checks in Docker (mirrors GitHub Actions). The
# releasability gate runs first, on the host: it reads git's list of tracked
# files, and the image carries no .git (.dockerignore drops it).
.PHONY: ci
ci: releasability
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

# Public-releasability gate (go-codec#26): fails on a lab address, a fleet
# hostname, an owner path or an internal domain name in any tracked file. The
# GitHub tree is byte-identical to the Gitea tag (version-control.md §5a), so
# this is checked in the source, before tagging. Needs a git work tree.
.PHONY: releasability
releasability:
	@go run ./internal/cmd/releasability

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
	@echo "No codec build tags: a program links and registers exactly the"
	@echo "codec packages it imports (pkg/factory imports all seven)."
	@echo ""
	@echo "All targets support any package in pkg/ via templates."
	@echo ""
	@echo "Test targets:"
	@echo "  make test                    - Run all unit tests"
	@echo "  make test-<pkg>              - Run tests for package (e.g., test-json, test-cbor)"
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
	@echo "  make releasability           - Fail on lab addresses, fleet hostnames or owner paths in tracked files"
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

SUBMODULE_TAGS := 
# Builds never resolve through a workspace: modules come from Gitea tags
# (dependency-resolution.md §2). Immunizes against stray go.work files.
export GOWORK=off

# ---- dependency-resolution.md targets (2026-08-13 cutover) ----
# SUBMODULE_TAGS lists nested modules tagged in lockstep with the root
# (release-versioning.md P2). deps-latest runs before compiling in a
# consumer repo; release-alpha (below) cuts an explicitly named -alpha tag.
# deps-latest picks the highest tag INCLUDING -alpha prereleases — bare
# 'go get @latest' ignores prereleases and would downgrade to legacy v0.0.0.

.PHONY: deps-latest

# ---- release-alpha: handbook software/templates/make/release-alpha.mk v2.3.9,
# adopted verbatim (go-codec #10) with ONE deviation: the VERSION grammar guard
# also rejects a value containing a newline. grep is line-oriented, so v2.3.9
# accepts "vX.Y.Z-alpha" followed by a newline and anything else; the same
# guard in release-alpha-changelog would then commit a multiline heading and
# delete the fragments before the CHANGELOG guard refused. Report upstream.
# Usage: make release-alpha VERSION=vX.Y.Z-alpha
# SUBMODULE_TAGS is empty, so there are no lockstep tags. The one prerequisite
# is release-alpha-changelog, below.
# agent-git.sh is resolved from PATH (template v2.3.6: never bake a path in).
# It becomes an absolute path here because the recipe's `test -x` guard does
# not search PATH, and it is simply-expanded because the recipe reads it with
# $(value ...). Pure Make functions only, deliberately: under GNU Make 4.4 a
# parse-time $(shell ...) builds an exported environment BEFORE the template's
# `unexport` line is read, which expands command-line variables and re-opens
# the injection v2.3.7/v2.3.8 closed (measured: RELEASE_REMOTE, REL_REMOTE
# canaries fired). $(value PATH) keeps PATH itself unexpanded. Not found =>
# empty => the recipe fails closed before VERSION is written.
# Override: make release-alpha VERSION=vX.Y.Z-alpha AGENT_GIT=/path/agent-git.sh
ifeq ($(origin AGENT_GIT),undefined)
AGENT_GIT := $(firstword $(wildcard $(addsuffix /agent-git.sh,$(subst :, ,$(value PATH)))))
endif
REPO_NAME ?= go-codec
RELEASE_REMOTE ?= origin

# ---- release-alpha-changelog: the changelog assembly release-versioning.md §5
# requires as "a committed step performed before tagging". PRs add
# changelog.d/<issue>-<slug>.md; this folds [Unreleased] and every fragment into
# a "## [X.Y.Z-alpha] - <date>" section (internal/cmd/assemble-changelog, built
# to a temporary directory OUTSIDE the tree, because `go run` reports every
# non-zero exit as 1 and this recipe needs the tool's exit 3), deletes the
# fragments and commits both through agent-git.sh. The template recipe then
# re-runs its guards, finds the heading, commits VERSION, tags, and pushes HEAD —
# which carries this commit too — in one atomic push.
#
# It runs as a PREREQUISITE so the template body below stays verbatim (bar the
# newline guard noted above). It reads only the REL_* values release-alpha
# sanitises (target-specific variables are inherited by prerequisites; see the
# template's v2.3.3/v2.3.7/v2.3.8 notes), so
# run on its own it has no version and refuses. Its guards repeat the template's
# pre-write guards on purpose: this step commits, so everything that can refuse
# the release must refuse BEFORE it, or a refused release leaves a changelog
# commit on main. Nothing pending => no commit, and the template's CHANGELOG
# guard decides. Tests: internal/releasetest.
.PHONY: release-alpha-changelog
release-alpha: release-alpha-changelog
release-alpha-changelog:
	@test -n "$$REL_VERSION" || { echo 'ERROR: state the version explicitly: make release-alpha VERSION=vX.Y.Z-alpha (release-versioning.md decides patch/minor/major — the recipe refuses to guess)'; exit 1; }
	@test "$$(printf '%s' "$$REL_VERSION" | wc -l)" -eq 0 && printf '%s' "$$REL_VERSION" | grep -Eq '^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)-alpha$$' || { echo "ERROR: VERSION is not of the form vX.Y.Z-alpha (no leading zeroes): $$REL_VERSION"; exit 1; }
	@test "$$(git rev-parse --abbrev-ref HEAD)" = main || { echo "ERROR: releases are cut from main; you are on $$(git rev-parse --abbrev-ref HEAD). 'push $$REL_REMOTE HEAD' would publish this branch."; exit 1; }
	@test -n "$$REL_REPO" || { echo 'ERROR: REPO_NAME is empty — set it to the literal repo name in the Makefile'; exit 1; }
	@test -x "$$REL_AGENT_GIT" || { echo 'ERROR: AGENT_GIT does not point at an executable agent-git.sh — every git write in this recipe signs through it, and a bad path must fail before anything is committed'; exit 1; }
	@test -n "$$REL_REMOTE" || { echo 'ERROR: RELEASE_REMOTE is empty — set it to the Gitea remote'; exit 1; }
	@git remote get-url "$$REL_REMOTE" >/dev/null 2>&1 || { echo "ERROR: no git remote named $$REL_REMOTE. Set RELEASE_REMOTE to the Gitea remote"; exit 1; }
	@! git rev-parse -q --verify "refs/tags/$$REL_VERSION" >/dev/null || { echo "ERROR: tag $$REL_VERSION already exists locally"; exit 1; }
	@rc=0; git ls-remote --exit-code --tags "$$REL_REMOTE" "refs/tags/$$REL_VERSION" >/dev/null 2>&1 || rc=$$?; \
	case "$$rc" in \
	  0) echo "ERROR: tag $$REL_VERSION already exists on $$REL_REMOTE"; exit 1 ;; \
	  2) : ;; \
	  *) echo "ERROR: could not reach $$REL_REMOTE to check whether tag $$REL_VERSION exists. Refusing to continue: an unreachable remote is not an absent tag"; exit 1 ;; \
	esac
	@test -z "$$(git status --porcelain)" || { echo 'ERROR: working tree not clean'; exit 1; }
	@set -e; rc=0; tool=$$(mktemp -d); trap 'rm -rf "$$tool"' EXIT; \
	go build -o "$$tool/assemble-changelog" ./internal/cmd/assemble-changelog || { echo "ERROR: could not build the changelog assembler; nothing was committed"; exit 1; }; \
	"$$tool/assemble-changelog" CHANGELOG.md "$${REL_VERSION#v}" "$$(date -u +%Y-%m-%d)" changelog.d > CHANGELOG.md.assembled || rc=$$?; \
	case "$$rc" in \
	  0) : ;; \
	  3) rm -f CHANGELOG.md.assembled; echo "changelog: nothing pending to assemble for $$REL_VERSION"; exit 0 ;; \
	  *) rm -f CHANGELOG.md.assembled; echo "ERROR: changelog assembly for $$REL_VERSION refused (reason above); nothing was committed"; exit 1 ;; \
	esac; \
	mv CHANGELOG.md.assembled CHANGELOG.md; \
	for f in changelog.d/*.md; do \
	  if [ -e "$$f" ]; then "$$REL_AGENT_GIT" "$$REL_REPO" rm -q -- "$$f"; fi; \
	done; \
	"$$REL_AGENT_GIT" "$$REL_REPO" add CHANGELOG.md; \
	"$$REL_AGENT_GIT" "$$REL_REPO" commit -m "docs(changelog): assemble $$REL_VERSION"; \
	echo "changelog: assembled $$REL_VERSION and removed the consumed fragments"

# release-alpha-template: v2.3.9 (2026-09-22 — two recipe defects, go-quicbaas#967,
# found by codex adopting v2.3.8 verbatim in go-truststrap PR #72; regression
# tests in testdata/release-alpha-tests.sh run this recipe against a real remote):
#  (a) --no-follow-tags on the release push. With push.followTags=true (a common
#      GLOBAL setting) git also pushes every missing annotated tag reachable from
#      HEAD, whatever the refspecs name — and --atomic does not narrow that set.
#      "Pushes ONLY the named tag" was false on any host with that setting.
#  (b) The CHANGELOG guard counts a heading only OUTSIDE a fenced code block. A
#      fenced example quoting "## [X.Y.Z-alpha]" at column 1 satisfied the literal
#      match, so a release with no real section was cut and pushed. Fences follow
#      CommonMark: 0-3 spaces of indent, 3+ backticks or tildes; closed only by
#      the SAME character at least as long with nothing after it; an unclosed
#      fence runs to end of file. A naive toggle is not enough — a ``` line inside
#      a ```` block would "close" it and re-admit the quoted heading.
#  This header is on line 1 so the first line always names the CURRENT revision;
#  v2.3.6-v2.3.8 were stamped lower down and #967 read this file as v2.3.5.
# release-alpha-template: v2.3.5 (guidance only). Two release-WORKFLOW safety
# patterns from go-blockstore #8, for any repo whose pipeline re-validates VERSION
# or splits validate/publish across jobs:
#  (a) VERIFY VERSION EXACTLY, never normalize. `tr -d '[:space:]'` accepts
#      '0.3. 1-alpha', a split-across-lines file, or a missing trailing newline —
#      all publish a VERSION that does not match the tag. Compare BYTES against the
#      writer inverted: `printf '%s\n' "$VER" | cmp -s - VERSION` (writer and
#      checker one statement in two directions); on mismatch dump `od -c VERSION`.
#  (b) SPLIT validate/publish jobs share a TOCTOU: a tag is immutable only by
#      convention, and a force-move BETWEEN the jobs passes any self-consistency
#      check (tag and checkout agree — on the NEW commit). validate must EXPORT the
#      commit it approved and publish must REFUSE anything else (and refuse an
#      ABSENT value — an empty guard is not a passing guard); FAIL, do not fall back
#      to the validated commit — a mid-release tag move is an immutability breach a
#      human should see.
# release-alpha-template: v2.3.4 (guidance only — no recipe change). If your repo
# ALSO ships a release WORKFLOW that walks tags to compute a previous release for
# notes/compare links, the set the validator ACCEPTS must equal the set the
# sorter can ORDER. git's `--sort=v:refname` is NOT SemVer for prerelease
# identifiers (orders alpha2 before alpha10; SemVer says alpha10 first), and
# `versionsort.suffix=-` only places prereleases before their stable — it does
# NOT make identifier comparison SemVer-correct (measured go-blockstore #8;
# corrects that lane's own round-3 overstatement). Since release-versioning.md §1
# says the sanctioned TAG vocabulary is BARE (X.Y.Z, -alpha, -beta, -rc; numbered
# -alpha.N is never tagged), narrow BOTH the release-form gate AND the walk to it:
#   num='(0|[1-9][0-9]*)'; sanctioned="^v${num}\.${num}\.${num}(-(alpha|beta|rc))?$"
#   PREV="$(git -c versionsort.suffix=- tag --list 'v*' --sort=v:refname \
#           | grep -E "$sanctioned" | awk -v cur="$TAG" '$0==cur{exit}{print}' | tail -n1 || true)"
# Excluding an unorderable/off-vocabulary tag from the walk only ever WIDENS the
# notes range (cosmetic dupes), never truncates it — the safe failure direction.
# If a repo ever needs numbered prereleases, the gate widens and the walk needs a
# real SemVer §11 comparator FIRST.
# release-alpha-template: v2.3.8 (2026-09-01 — SECURITY: `override` + unexport on the
# internal REL_* names too. v2.3.7 sanitised the four public inputs and left their
# four internal aliases settable from the command line, where they both executed
# and outranked the sanitised value. Measured; see the block at the recipe.)
# v2.3.7 (2026-09-01 — SECURITY: $(value ...) on all four
# release inputs, not just VERSION, AND `unexport` of those four — $(value) alone
# does not close the injection, because a command-line variable is auto-exported
# into every recipe environment and building that environment expands it. Measured
# GNU Make 4.3; automatethethings#14. Do not re-measure this with `make -n` — under
# -n no recipe runs, no environment is built, and an injectable Makefile looks clean).
# v2.3.6 (2026-09-01 — no baked AGENT_GIT path: it published a developer path, a private repo name and its layout into every mirrored adopter; plus the pre-existing-VERSION shadowing check that made a guard unreachable. Measured go-luks2 #19/#20). v2.3.3 (SECURITY — Make-variable command-injection fix; see the v2.3.3 block at the recipe). v2.3.2 (--atomic on the release push — without it git permits PARTIAL success: an update-hook/branch-protection rejection of main can land the tag alone, unretractable; measured in go-blockstore #8 round 3, and note the harness subtlety: a pre-receive hook CANNOT exhibit this bug — it runs once per push — the correct model of branch protection is the per-ref update hook). v2.3.1 (adds the AGENT_GIT-executable guard from go-blockstore — a bad AGENT_GIT path must fail before VERSION is written, not at the first signed write). v2.3 — replaces the defective v1 recipe found in all 16
# product repos on 2026-08-26 (go-enclave #10). The v1 recipe: auto-incremented
# PATCH only (an API addition ships labeled as a bugfix), created LIGHTWEIGHT
# UNSIGNED tags as the OWNER's local identity (violating the per-agent-principal
# ratification and signed-tag discipline), pushed EVERY local tag, and `|| true`
# swallowed a failed VERSION commit so a tag could disagree with its own VERSION
# file. None of it failed loudly; it was wrong only when you looked.
#
# v2 contract:
#   - The version is EXPLICIT: `make release-alpha VERSION=v0.3.0-alpha`.
#     No auto-increment — the human/agent cutting a release decides
#     patch-vs-minor-vs-major per conventions/release-versioning.md and
#     states it. The recipe refuses to guess.
#   - All git write operations go through agent-git.sh (identity + SSH
#     signing as agent-<repo>); the tag is ANNOTATED and signed.
#   - Pushes ONLY the release commit and the named tag. Never --tags.
#   - Cut from `main` ONLY — a branch guard, since `push <remote> HEAD`
#     publishes whatever branch is checked out (go-enclave, 2026-08-26: v2
#     caught this residual before the 15-repo adoption).
#   - REPO_NAME is the LITERAL repo name (identity/signing), never derived
#     from the directory; a duplicate-tag check hits the REMOTE not just
#     local refs; the CHANGELOG guard anchors on a section heading, not a
#     substring (all three: reviewer-enclave-9, 2026-08-26, before the
#     v0.3.0-alpha cut).
#   - Fail-closed end to end: no `|| true`; any failing step aborts before
#     the tag exists.
#   - Error messages are SINGLE-quoted, deliberately. A double-quoted `echo`
#     is a live command-substitution context, and error strings are exactly
#     what people edit casually — a backtick or $(...) added there gets
#     EXECUTED at the moment a release is already failing. Never put backticks
#     or $(...) inside a recipe error message (reviewer-enclave-9, 2026-08-26:
#     the v2.2 branch-guard shipped with backticks and ran `push dev HEAD`).
#   - CHANGELOG discipline: refuses to cut if CHANGELOG.md lacks an entry
#     for $(VERSION) (repos without a CHANGELOG delete that guard when
#     adopting, consciously).
#   - Repos with a sdk/go module keep their existing lockstep checks
#     (check-tag-lockstep etc.) as prerequisites of this target.
#
# ------------------------------------------------------------------------------
# v2.3 (2026-08-27) — FOUR CORRECTIONS. v2.2 COULD NOT WORK AS WRITTEN in any
# fleet repo; each defect below was measured on go-blockstore #8, whose Makefile
# is the reference implementation, before the six queued adoptions (go-auth #22,
# go-frost #17, go-codec #10, go-libsignal #39, go-objstore #32, go-luks2 #19).
# None of the four is a preference:
#
#   1. VERSION is written UNPREFIXED — `$(VERSION:v%=%)`. v2.2 wrote the
#      v-prefixed tag string verbatim into the VERSION file, contradicting
#      conventions/release-versioning.md §2, whose table is normative: VERSION
#      is unprefixed, the TAG carries the `v`. Every repo in the portfolio
#      follows the table, and the v-prefixed form also breaks the standard
#      reconciliation `VERSION=$${GITHUB_REF#refs/tags/v}` that release.yml
#      validates against (§5).
#
#   2. The CHANGELOG guard matches the UNPREFIXED heading LITERALLY, via an
#      index-based awk predicate anchored at column 1. v2.2's
#      `grep -Eq '^#+ +\[?$(VERSION)\]?'` fails two ways at once. It demands the
#      `v` prefix, which no Keep-a-Changelog heading carries — so on its own it
#      REFUSES EVERY WELL-FORMED RELEASE. And it is a REGEX: the dots of a
#      version are wildcards and `#+ +\[?` admits heading shapes the release
#      workflow does not, so `## [0x3x2-alpha]` and `### 0.3.2-alpha` pass
#      (measured), the tag gets cut and pushed, and only then does the
#      workflow's literal lookup fail — on a tag that can no longer be
#      retracted (§3: published tags are immutable). A pre-flight check that is
#      not IDENTICAL to the thing it clears for is not a check.
#
#   3. The remote duplicate-tag check reads the EXIT CODE, not `!`.
#      `git ls-remote --exit-code` answers three different things and v2.2's
#      `! git ls-remote … || { error }` collapsed them to two: 0 = the tag
#      exists, 2 = it does not, anything else (128 on an unreachable,
#      unauthenticated or NONEXISTENT remote) = the question was never
#      answered. Under `!` that third case read as "tag absent" — the check
#      FAILED OPEN — and the recipe went on to commit VERSION and create the
#      tag before the push failed, which is the exact half-released state the
#      check exists to prevent. Codes verified against Gitea 2026-08-27, not
#      assumed.
#
#   4. RELEASE_REMOTE ?= origin replaces the hardcoded `dev`. NO FLEET REPO HAS
#      A `dev` REMOTE — go-blockstore, go-objstore, go-qrdb, go-quicbaas and
#      go-xkms all reach Gitea through `origin` (measured 2026-08-27), whatever
#      conventions/version-control.md calls the remote. Hardcoded, `push dev`
#      failed AFTER VERSION was already committed, and defect 3 made the
#      duplicate-tag check pass vacuously against a remote name that does not
#      resolve. It stays a VARIABLE, not a renamed constant, so a future
#      dev-remote rollout is one line per repo. Its existence is asserted
#      before it is trusted (also from go-blockstore #8): defect 3 already
#      blocks on rc 128, and this turns that into a diagnosis.
#
# SUPPLY CHAIN — for the release WORKFLOW that pairs with this target: any
# third-party action it uses MUST be pinned to a full commit SHA, never a
# mutable tag. A tag is a ref whoever controls the action can move, and a
# release job holds write-scoped tokens. All four existing fleet release.yml
# files use mutable tags today; tracked on go-objstore #37. When bumping a pin,
# read the ANNOTATED-TAG PEEL TRAP note in
# templates/github-actions/nightly.yml — `git ls-remote --tags --refs` returns
# the TAG OBJECT's sha for an annotated tag, which does not resolve as a pin
# and fails at checkout; take the `^{}` peeled line instead.
# ------------------------------------------------------------------------------
#
# AGENT_GIT resolves the signing wrapper; repos adopting this template set these
# near the top of their Makefile. REPO_NAME MUST be the literal repo name — NOT
# $(notdir $(CURDIR)), which derives the signing IDENTITY from the checkout
# directory, so a worktree/CI-workspace/clone named anything but the repo signs
# and attributes the release as the wrong agent (measured go-enclave 2026-08-26:
# a dir named `mk` → agent-mk):
#   AGENT_GIT ?= agent-git.sh      # RESOLVED FROM PATH — see v2.3.6 below
#   REPO_NAME ?= go-enclave        # the literal repo name; never $(notdir $(CURDIR))
#   RELEASE_REMOTE ?= origin       # the Gitea remote; see v2.3 correction 4
#
# v2.3.6 (2026-09-01, measured in go-luks2 #19/#20) — TWO defects every adopter
# inherits from earlier revisions. Both were found by adopting this template into
# a repo with a PUBLIC MIRROR, which is the condition that makes the first one
# matter and the second one visible.
#
#  (a) DO NOT BAKE AN ABSOLUTE PATH INTO AGENT_GIT. Earlier revisions suggested
#      a literal `/home/<user>/sources/automatethethings/software/scripts/...`
#      default. In a mirrored repo that single line publishes a developer's
#      username, the private company repo's name, and its checkout layout —
#      exactly the class of leak go-luks2 #20 exists to remove, shipped BY the
#      template into every adopter. Resolve from PATH instead. The v2.3.1
#      `test -x "$$REL_AGENT_GIT"` guard already makes an unresolvable value
#      fail closed BEFORE VERSION is written, so nothing is lost by dropping the
#      default: the failure mode is a clear error, not a silent wrong signer.
#
#  (b) CHECK FOR A PRE-EXISTING `VERSION` VARIABLE BEFORE ADOPTING. This recipe
#      takes the release version as `make release-alpha VERSION=v0.3.0-alpha`.
#      Many Makefiles ALREADY define `VERSION` for `-ldflags -X main.Version=`.
#      When they do, that definition shadows the argument: `make release-alpha`
#      with no argument inherits the build version, sails straight past the
#      "state the version explicitly" guard — the guard written to give the
#      clearest possible message becomes UNREACHABLE — and dies several steps
#      later on a confusing grammar error about a value the operator never
#      typed. Rename the build variable (`BUILD_VERSION`), not this one.
#      Measured in go-luks2, where `make release-alpha` did exactly this.
#
#      Grep before adopting:  grep -n '^VERSION' Makefile

# v2.3.3 — SECURITY: every settable Make variable crosses into the recipe ONCE,
# via a target-specific export, and the body references only shell parameters
# ($$REL_*). Before this, $(VERSION) was expanded into double-quoted recipe TEXT
# before the grammar guard ran, so `make release-alpha 'VERSION=v0.0.0-alpha`cmd`'`
# executed the backtick at the first guard line (measured go-blockstore #8 r4:
# canaries via VERSION backtick, VERSION embedded-quote, and RELEASE_REMOTE
# backtick — quoting is no boundary against a value that becomes recipe source).
# `$(value VERSION)` also stops the double-evaluation the export otherwise adds.
#
# v2.3.7 — the above was true but INCOMPLETE on two counts, both measured on
# GNU Make 4.3 (automatethethings#14; the issue's own analysis was wrong in the
# other direction and the correction is recorded there).
#
#  1. `$(value ...)` MUST be on all four inputs, not just VERSION. It was on
#     VERSION alone; RELEASE_REMOTE, REPO_NAME and AGENT_GIT were plain `$(...)`
#     and so were live injection points on a release path:
#       make release-alpha VERSION=v1.2.3-alpha 'RELEASE_REMOTE=$(shell id >&2)origin'
#     ran `id` before any guard.
#
#  2. `$(value ...)` ALONE DOES NOT CLOSE IT. A variable set on the make command
#     line is automatically exported into EVERY recipe's environment, and
#     BUILDING that environment expands it — so the payload fires even though
#     the assignment never expanded it, and even in a Makefile that names the
#     variable NOWHERE. `unexport` is the half that closes it:
#
#       makefile shape                          payload fires?
#       ----------------------------------      --------------
#       plain $(RELEASE_REMOTE)                  yes (twice)
#       $(value RELEASE_REMOTE)                  yes (once, via the env export)
#       unexport + $(value RELEASE_REMOTE)       NO
#       variable not mentioned in the makefile   yes (any recipe that runs)
#
#     Hence the `unexport` line below. It does not affect ordinary use: the
#     guards still receive the operator's literal text, and a value carrying a
#     $(shell ...) simply arrives as raw characters and fails the grammar,
#     remote-resolves or -x guard instead of executing.
#
# MEASURE IT WITH A REAL RECIPE, NOT `make -n`. Under `-n` no recipe runs, so no
# recipe environment is built and NOTHING fires — `-n` makes an injectable
# Makefile look clean. That is precisely the mistake that made this look
# unreproducible and nearly got this warning deleted.
#
# ADOPTER CHECK: `grep -P '(?<!\$)\$\((?!\()' ` over the recipe BODY must be empty
# — no $(...) carrying a settable value below the export lines — AND every
# settable input named on the `unexport` line.
.PHONY: release-alpha
# Command-line variables are auto-exported; the export expands them. See above.
# BOTH name families must be listed: the public inputs an operator sets, AND the
# internal REL_* names, which are just as settable from a command line.
unexport VERSION RELEASE_REMOTE REPO_NAME AGENT_GIT
unexport REL_VERSION REL_REMOTE REL_REPO REL_AGENT_GIT
# `override` is load-bearing, not decoration. A command-line variable outranks a
# target-specific assignment, so WITHOUT it an attacker simply uses the internal
# name and skips the sanitiser entirely:
#
#   make release-alpha VERSION=v1.2.3-alpha 'REL_REMOTE=$(shell id >&2)origin'
#
# Measured on GNU Make 4.3 against the v2.3.7 recipe: that fired the payload AND
# won the assignment (REL_REMOTE came out as the attacker's plain 'origin', the
# $(value ...) discarded). v2.3.7 closed the four public names and left their
# four internal aliases open — a fix that moves the hole rather than shutting
# it. With `override` the same command yields an EMPTY REL_REMOTE and no
# execution, so the existing empty-value guard rejects it.
release-alpha: override export REL_VERSION   := $(value VERSION)
release-alpha: override export REL_REMOTE    := $(value RELEASE_REMOTE)
release-alpha: override export REL_REPO      := $(value REPO_NAME)
release-alpha: override export REL_AGENT_GIT := $(value AGENT_GIT)
release-alpha:
	@test -n "$$REL_VERSION" || { echo 'ERROR: state the version explicitly: make release-alpha VERSION=vX.Y.Z-alpha (release-versioning.md decides patch/minor/major — the recipe refuses to guess)'; exit 1; }
	@test "$$(printf '%s' "$$REL_VERSION" | wc -l)" -eq 0 && printf '%s' "$$REL_VERSION" | grep -Eq '^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)-alpha$$' || { echo "ERROR: VERSION is not of the form vX.Y.Z-alpha (no leading zeroes): $$REL_VERSION"; exit 1; }
	@test "$$(git rev-parse --abbrev-ref HEAD)" = main || { echo "ERROR: releases are cut from main; you are on $$(git rev-parse --abbrev-ref HEAD). 'push $$REL_REMOTE HEAD' would publish this branch."; exit 1; }
	@test -n "$$REL_REPO" || { echo 'ERROR: REPO_NAME is empty — set it to the literal repo name in the Makefile, never $$(notdir $$(CURDIR))'; exit 1; }
	@test -x "$$REL_AGENT_GIT" || { echo 'ERROR: AGENT_GIT does not point at an executable agent-git.sh — every git write in this recipe signs through it, and a bad path must fail before VERSION is written'; exit 1; }
	@test -n "$$REL_REMOTE" || { echo 'ERROR: RELEASE_REMOTE is empty — set it to the Gitea remote (origin in every fleet repo as of 2026-08-27)'; exit 1; }
	@git remote get-url "$$REL_REMOTE" >/dev/null 2>&1 || { echo "ERROR: no git remote named $$REL_REMOTE. Set RELEASE_REMOTE to the Gitea remote; the duplicate-tag check below cannot answer against a remote that does not resolve"; exit 1; }
	@! git rev-parse -q --verify "refs/tags/$$REL_VERSION" >/dev/null || { echo "ERROR: tag $$REL_VERSION already exists locally"; exit 1; }
	@rc=0; git ls-remote --exit-code --tags "$$REL_REMOTE" "refs/tags/$$REL_VERSION" >/dev/null 2>&1 || rc=$$?; \
	case "$$rc" in \
	  0) echo "ERROR: tag $$REL_VERSION already exists on $$REL_REMOTE (a local check alone would miss this and fail the push AFTER committing VERSION)"; exit 1 ;; \
	  2) : ;; \
	  *) echo "ERROR: could not reach $$REL_REMOTE to check whether tag $$REL_VERSION exists. Refusing to continue: an unreachable remote is not an absent tag"; exit 1 ;; \
	esac
	@test -z "$$(git status --porcelain)" || { echo 'ERROR: working tree not clean'; exit 1; }
	@ver="## [$${REL_VERSION#v}]"; awk -v ver="$$ver" ' \
	  function fence(s,   i, n) { \
	    i = 1; while (i <= 3 && substr(s, i, 1) == " ") i++; \
	    fch = substr(s, i, 1); if (fch != "~" && fch != "\140") return 0; \
	    n = 0; while (substr(s, i + n, 1) == fch) n++; \
	    info = substr(s, i + n); return n >= 3 ? n : 0 } \
	  { n = fence($$0) } \
	  open { if (n >= open && fch == opench && info ~ /^[ \t]*$$/) open = 0; next } \
	  n && !(fch == "\140" && index(info, "\140")) { open = n; opench = fch; next } \
	  index($$0, ver) == 1 { f = 1 } \
	  END { exit f ? 0 : 1 }' CHANGELOG.md || { echo "ERROR: CHANGELOG.md has no \"$$ver\" section heading. Matched LITERALLY and anchored at the line start, exactly as .github/workflows/release.yml matches it — an unanchored prose mention does not count, neither does a near-miss heading, and neither does a heading quoted inside a fenced code block"; exit 1; }
	@printf '%s\n' "$${REL_VERSION#v}" > VERSION
	"$$REL_AGENT_GIT" "$$REL_REPO" add VERSION
	"$$REL_AGENT_GIT" "$$REL_REPO" commit -m "chore(release): $$REL_VERSION"
	"$$REL_AGENT_GIT" "$$REL_REPO" tag -a "$$REL_VERSION" -m "$$REL_VERSION"
	"$$REL_AGENT_GIT" "$$REL_REPO" push --atomic --no-follow-tags "$$REL_REMOTE" HEAD "refs/tags/$$REL_VERSION"
	@echo "Released $$REL_VERSION (signed, annotated, single tag pushed)"

deps-latest:
	@GOWORK=off sh -c ' \
	SELF=$$(go list -m); \
	for m in $$(go list -m -f "{{if not .Indirect}}{{.Path}}{{end}}" all 2>/dev/null | grep "^github.com/jeremyhahn/" | grep -vx "$$SELF"); do \
	  case "$$(go list -m -f "{{if .Replace}}{{.Replace.Path}}{{end}}" "$$m" 2>/dev/null)" in .*|/*) continue;; esac; \
	  V=$$(go list -m -versions "$$m" 2>/dev/null | tr " " "\n" | tail -1); \
	  if [ -n "$$V" ] && [ "$$V" != "$$m" ]; then go get "$$m@$$V" || exit 1; fi; \
	done; go mod tidy'
