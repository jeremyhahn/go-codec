# go-codec API Reference

## Quick Reference

### Standard Codec API

All codecs implement the standard `Codec[T]` interface:

```go
import "github.com/jeremyhahn/go-codec/pkg/json"

// Create codec
codec := json.New[MyStruct]()

// Marshal to bytes
data, err := codec.Marshal(myStruct)

// Unmarshal from bytes
var result MyStruct
err = codec.Unmarshal(data, &result)

// Encode to writer
err = codec.Encode(writer, myStruct)

// Decode from reader
err = codec.Decode(reader, &result)
```

**Performance:** 635 ns/op, 160 B/op, 2 allocs/op (marshal)

### Buffer Reuse Methods (JSON & MessagePack)

For high-throughput scenarios, JSON and MessagePack codecs provide additional methods:

```go
import "github.com/jeremyhahn/go-codec/pkg/json"

// Use NewPool to get the extended type
codec := json.NewPool[MyStruct]()

// Reuse buffer (best performance)
buf := make([]byte, 0, 1024)
result, err := codec.MarshalTo(buf, myStruct)

// Append to buffer
buf, err = codec.AppendMarshal(buf, myStruct)

// Unmarshal (scratch buffer not used by JSON)
var result MyStruct
scratch := make([]byte, 1024)
err = codec.UnmarshalFrom(data, &result, scratch)
```

**Performance:** 596 ns/op, 64 B/op, 1 allocs/op (buffer reuse)
**Improvement:** 6% faster, 60% less memory

## Core Interfaces

### Codec[T any]

Standard codec interface implemented by all codecs.

```go
type Codec[T any] interface {
    // Encode serializes data to a writer
    Encode(w io.Writer, data T) error

    // Decode deserializes data from a reader
    Decode(r io.Reader, data *T) error

    // Marshal serializes data to bytes
    Marshal(data T) ([]byte, error)

    // Unmarshal deserializes bytes into data
    Unmarshal(data []byte, v *T) error
}
```

### OptimizedCodec[T any] (JSON & MessagePack only)

The `OptimizedCodec` type extends the standard codec with buffer reuse methods:

```go
// Available in pkg/json and pkg/msgpack
type OptimizedCodec[T any] struct {
    *Codec[T]  // Embeds standard codec
}

// Create with NewPool
codec := json.NewPool[T]()

// Additional methods:
// MarshalTo marshals into provided buffer
// Returns view into buf or new allocation if buf is too small
func (c *OptimizedCodec[T]) MarshalTo(buf []byte, data T) ([]byte, error)

// AppendMarshal appends marshaled data to buf
func (c *OptimizedCodec[T]) AppendMarshal(buf []byte, data T) ([]byte, error)

// UnmarshalFrom unmarshals data (scratch buffer unused)
func (c *OptimizedCodec[T]) UnmarshalFrom(data []byte, v *T, scratch []byte) error
```

## Available Codecs

### JSON

```go
import "github.com/jeremyhahn/go-codec/pkg/json"

// Standard codec
codec := json.New[T]()

// With buffer reuse methods
codec := json.NewPool[T]()
```

### MessagePack

```go
import "github.com/jeremyhahn/go-codec/pkg/msgpack"

// Standard codec
codec := msgpack.New[T]()

// With buffer reuse methods
codec := msgpack.NewPool[T]()
```

### YAML

```go
import "github.com/jeremyhahn/go-codec/pkg/yaml"

codec := yaml.New[T]()
```

### TOML

```go
import "github.com/jeremyhahn/go-codec/pkg/toml"

codec := toml.New[T]()
```

### BSON

```go
import "github.com/jeremyhahn/go-codec/pkg/bson"

codec := bson.New[T]()
```

### Protocol Buffers

```go
import "github.com/jeremyhahn/go-codec"

codec := codec.NewProtoBuf[*MyProtoMessage]()
```

**Note:** Requires generated protobuf code

## Internal Buffer Pool

The buffer reuse methods (`MarshalTo`, `AppendMarshal`) use an internal `bytes.Buffer` pool to reduce allocations. This is handled automatically - you just provide your own byte slice for the output.

**You don't need to interact with the pool directly** - just use `MarshalTo()` and `AppendMarshal()`.

If you need a `bytes.Buffer` for other purposes:

```go
import "github.com/jeremyhahn/go-codec/pkg/pool"

buf := pool.GetBytesBuffer()
defer pool.PutBytesBuffer(buf)

buf.WriteString("data")
```

## Common Patterns

### Pattern 1: Single Item Marshal

```go
codec := json.NewPool[Item]()
buf := make([]byte, 0, 256)

result, err := codec.MarshalTo(buf, item)
```

### Pattern 2: Batch Processing

```go
codec := json.NewPool[Item]()
buf := make([]byte, 0, 1024)

for _, item := range items {
    buf = buf[:0] // Reset
    result, err := codec.MarshalTo(buf, item)
    if err != nil {
        return err
    }
    process(result)
}
```

### Pattern 3: Concatenate Multiple Items

```go
codec := json.NewPool[Item]()
buf := make([]byte, 0, 4096)

for _, item := range items {
    buf, err = codec.AppendMarshal(buf, item)
    if err != nil {
        return err
    }
    buf = append(buf, '\n')
}
```

### Pattern 4: HTTP Handler

```go
type Handler struct {
    codec *json.OptimizedCodec[Response]
    buf   []byte
}

func NewHandler() *Handler {
    return &Handler{
        codec: json.NewPool[Response](),
        buf:   make([]byte, 0, 4096),
    }
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    response := processRequest(r)

    h.buf = h.buf[:0]
    data, err := h.codec.MarshalTo(h.buf, response)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.Write(data)
}
```

## Error Handling

All codec methods return standard Go errors:

```go
data, err := codec.Marshal(item)
if err != nil {
    // Handle error
    return fmt.Errorf("failed to marshal: %w", err)
}
```

Common errors:
- Invalid JSON/YAML/TOML syntax
- Type mismatches
- I/O errors (for Encode/Decode)

## Type Requirements

### Standard Codecs (JSON, YAML, TOML, MessagePack, BSON)

Any Go type with proper struct tags:

```go
type Person struct {
    Name  string `json:"name" yaml:"name" msgpack:"name"`
    Age   int    `json:"age" yaml:"age" msgpack:"age"`
    Email string `json:"email" yaml:"email" msgpack:"email"`
}
```

### Protocol Buffers

Must implement `proto.Message`:

```go
// Generated from .proto file
type Person struct {
    Name  string
    Age   int32
    Email string
}

codec := codec.NewProtoBuf[*Person]()
```

## Performance Comparison

Based on actual benchmarks:

| Metric | Standard (Marshal) | Buffer Reuse (MarshalTo) | Improvement |
|--------|-------------------|-------------------------|-------------|
| Time | 635 ns/op | 596 ns/op | 6% faster |
| Memory | 160 B/op | 64 B/op | 60% less |
| Allocations | 2 allocs/op | 1 allocs/op | 50% fewer |

**Large batch (100 items):** Memory improvement scales to 99%+ reduction.

## When to Use Buffer Reuse

### Use Standard Codec When:
- Code simplicity is the priority
- One-time or infrequent operations
- Configuration file loading
- Performance is not critical

### Use Buffer Reuse (NewPool) When:
- Processing >1000 requests/sec
- Hot request paths (HTTP handlers)
- Batch processing loops
- Memory allocations show up in profiling
- Low-latency requirements

## Testing

### Running Tests

```bash
# All tests
make test

# Specific codec
make test-json
make test-msgpack
make test-pool
```

### Running Benchmarks

```bash
# All benchmarks
make bench-all

# Comparison benchmarks
make bench-comparison

# Quick benchmarks
make bench-quick
```

### Example Test

```go
func TestMarshalUnmarshal(t *testing.T) {
    codec := json.New[Person]()

    person := Person{
        Name:  "John Doe",
        Age:   30,
        Email: "john@example.com",
    }

    data, err := codec.Marshal(person)
    if err != nil {
        t.Fatalf("Marshal failed: %v", err)
    }

    var result Person
    err = codec.Unmarshal(data, &result)
    if err != nil {
        t.Fatalf("Unmarshal failed: %v", err)
    }

    if result.Name != person.Name {
        t.Errorf("expected %s, got %s", person.Name, result.Name)
    }
}
```

### Example Benchmark

```go
func BenchmarkMarshal(b *testing.B) {
    codec := json.NewPool[Person]()
    person := Person{Name: "John", Age: 30}
    buf := make([]byte, 0, 256)

    b.ResetTimer()
    b.ReportAllocs()

    for i := 0; i < b.N; i++ {
        buf = buf[:0]
        _, err := codec.MarshalTo(buf, person)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

## Additional Resources

- [Performance Guide](./PERFORMANCE.md) - Detailed performance information
- [Optimization Report](./OPTIMIZATION_REPORT.md) - Technical deep dive
- [README](./README.md) - Getting started guide
- [Examples](./examples/) - Code examples

## Quick Tips

1. **Pre-allocate buffers** based on expected data size
2. **Reuse buffers** across iterations with `buf = buf[:0]`
3. **Benchmark your code** to validate improvements
4. **Start simple** - only optimize hot paths
5. **Measure first** - profile before optimizing
