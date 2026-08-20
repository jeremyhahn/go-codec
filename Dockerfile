# Dockerfile for go-codec development and examples.
#
# alpine-first (owner ruling 2026-08-20, company task #68): alpine is the
# portfolio's default base — it also removes the Debian OS package set that
# kept tripping the Trivy Image gate (CVE-2026-34191, libaprutil1) for
# packages this library never uses. Debian bases require a documented reason
# (FIPS/OpenSSL builds, glibc-only deps); this image has neither.
FROM golang:1.26.7-alpine

# Build and codegen dependencies (apk, no recommends concept, tiny footprint)
RUN apk add --no-cache protobuf protobuf-dev make curl git build-base

# Install Go protobuf plugin
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

# Install golangci-lint — PINNED v2.12.2 (devops.md §9, company task #122:
# v2.13.0 ships a staticcheck rc whose nilness pass panics). Installed via
# `go install` (same as CI) — the curl|sh installer fails under busybox sh.
RUN go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2

# Install security tools
RUN go install github.com/securego/gosec/v2/cmd/gosec@latest && \
    go install golang.org/x/vuln/cmd/govulncheck@latest

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Default command
CMD ["make", "examples"]
