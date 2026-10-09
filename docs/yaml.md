# YAML Codec

> **Status: REFERENCE** · Last verified: 2026-10-09
> As-built usage of the YAML codec in `pkg/yaml/`. The code is authoritative.

YAML (YAML Ain't Markup Language) is a human-readable data serialization format.

## Import

```go
import "github.com/jeremyhahn/go-codec/pkg/yaml"
```

Importing the package registers the codec in `init()`; no build tags are
needed. Import only the codec packages you use: Go links just those.

## Usage

```go
codec := yaml.New[MyStruct]()

data, err := codec.Marshal(myStruct)
err = codec.Unmarshal(data, &result)
```

## Struct Tags

Use the `yaml` tag to control field names and behavior:

```go
type Config struct {
    Host     string `yaml:"host"`
    Port     int    `yaml:"port"`
    Debug    bool   `yaml:"debug,omitempty"`
    Internal string `yaml:"-"`
}
```

## Performance

| Operation | Time | Memory | Allocs |
|-----------|------|--------|--------|
| Marshal | 20,104 ns/op | 7,056 B/op | 37 |
| Unmarshal | 31,589 ns/op | 9,264 B/op | 94 |

## Notes

- Uses `gopkg.in/yaml.v3`
- Best for configuration files and human-editable data
- Supports comments in source (not preserved on round-trip)
- Slower than binary formats - use JSON/MsgPack for high-throughput
