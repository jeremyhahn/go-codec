# Development and examples image on alpine.
FROM golang:1.27.2-alpine

# apk upgrade first so a pinned base picks up fixed packages.
RUN apk upgrade --no-cache && \
    apk add --no-cache protobuf protobuf-dev make curl git build-base

# Install Go protobuf plugin
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

# golangci-lint pinned at v2.12.2, installed with go install; the curl|sh installer fails under busybox sh.
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
