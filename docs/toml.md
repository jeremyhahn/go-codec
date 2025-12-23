# TOML Codec

TOML (Tom's Obvious, Minimal Language) is a configuration file format.

## Import

```go
import "github.com/jeremyhahn/go-codec/pkg/toml"
```

## Usage

```go
codec := toml.New[MyStruct]()

data, err := codec.Marshal(myStruct)
err = codec.Unmarshal(data, &result)
```

## Struct Tags

Use the `toml` tag to control field names:

```go
type Config struct {
    Database DatabaseConfig `toml:"database"`
    Server   ServerConfig   `toml:"server"`
}

type DatabaseConfig struct {
    Host string `toml:"host"`
    Port int    `toml:"port"`
}
```

## Performance

| Operation | Time | Memory | Allocs |
|-----------|------|--------|--------|
| Marshal | 16,721 ns/op | 5,387 B/op | 51 |
| Unmarshal | 19,253 ns/op | 4,343 B/op | 59 |

## Notes

- Uses `github.com/BurntSushi/toml`
- Ideal for configuration files
- Supports nested tables and arrays
- Not suitable for high-throughput serialization
