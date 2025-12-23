# BSON Codec

BSON (Binary JSON) is the binary format used by MongoDB.

## Import

```go
import "github.com/jeremyhahn/go-codec/pkg/bson"
```

## Usage

```go
codec := bson.New[MyStruct]()

data, err := codec.Marshal(myStruct)
err = codec.Unmarshal(data, &result)
```

## Struct Tags

Use the `bson` tag to control field names:

```go
type Document struct {
    ID        string `bson:"_id"`
    Name      string `bson:"name"`
    CreatedAt int64  `bson:"created_at"`
    Tags      []string `bson:"tags,omitempty"`
}
```

## Performance

| Operation | Time | Memory | Allocs |
|-----------|------|--------|--------|
| Marshal | 1,191 ns/op | 162 B/op | 2 |
| Unmarshal | 1,785 ns/op | 289 B/op | 12 |

## Notes

- Uses `go.mongodb.org/mongo-driver/bson`
- Native MongoDB format
- Supports MongoDB-specific types (ObjectID, Timestamp, etc.)
- Larger than MessagePack for general data
