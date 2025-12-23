# Dockerfile for go-codec development and examples
FROM golang:1.25.5-bookworm

# Install protobuf compiler and dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    protobuf-compiler \
    make \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Install Go protobuf plugin
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

# Install golangci-lint
RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b /go/bin v2.7.2

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
