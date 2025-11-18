.PHONY: test test-json test-yaml test-toml test-msgpack test-protobuf test-bson coverage coverage-json coverage-yaml coverage-toml coverage-msgpack coverage-protobuf coverage-bson clean

# Run all tests
test: test-json test-yaml test-toml test-msgpack test-protobuf test-bson

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

# Show coverage for all codecs
coverage: coverage-json coverage-yaml coverage-toml coverage-msgpack coverage-protobuf coverage-bson

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

# Clean coverage files
clean:
	@rm -f coverage-*.out
