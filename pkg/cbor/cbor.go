package cbor

import (
	"errors"
	"io"

	"github.com/fxamacker/cbor/v2"
	codec "github.com/jeremyhahn/go-codec"
)

func init() {
	codec.RegisterCodec(codec.CBOR)
}

// ErrInvalidOptions is matched (via errors.Is) by every *OptionsError
// returned from NewWithOptions.
var ErrInvalidOptions = errors.New("cbor: invalid codec options")

// OptionsError reports that the DecOptions or EncOptions passed to
// NewWithOptions could not be built into a mode.
type OptionsError struct {
	// Mode is "decode" or "encode".
	Mode string
	// Err is the cause reported by fxamacker/cbor.
	Err error
}

// Error implements the error interface.
func (e *OptionsError) Error() string {
	return "cbor: invalid " + e.Mode + " options: " + e.Err.Error()
}

// Unwrap returns the underlying cause.
func (e *OptionsError) Unwrap() error { return e.Err }

// Is reports whether target is ErrInvalidOptions.
func (e *OptionsError) Is(target error) bool { return target == ErrInvalidOptions }

// Codec implements codec.Codec for CBOR with default or caller-supplied options.
type Codec[T any] struct {
	dm cbor.DecMode
	em cbor.EncMode
}

// New creates a CBOR codec with fxamacker/cbor's lenient defaults; use NewWithOptions for attacker-controlled input.
func New[T any]() *Codec[T] {
	return &Codec[T]{}
}

// NewWithOptions creates a CBOR codec from dec and enc, returning *OptionsError for invalid options; for untrusted input enforce DupMapKey, IndefLength, UTF8, MaxNestedLevels and TagsMd.
func NewWithOptions[T any](dec cbor.DecOptions, enc cbor.EncOptions) (*Codec[T], error) {
	dm, err := dec.DecMode()
	if err != nil {
		return nil, &OptionsError{Mode: "decode", Err: err}
	}
	em, err := enc.EncMode()
	if err != nil {
		return nil, &OptionsError{Mode: "encode", Err: err}
	}
	return &Codec[T]{dm: dm, em: em}, nil
}

// Encode serializes the given data to the writer using CBOR
func (c *Codec[T]) Encode(w io.Writer, data T) error {
	if c.em != nil {
		return c.em.NewEncoder(w).Encode(data)
	}
	return cbor.NewEncoder(w).Encode(data)
}

// Decode deserializes CBOR data from the reader into the provided type
func (c *Codec[T]) Decode(r io.Reader, data *T) error {
	if c.dm != nil {
		return c.dm.NewDecoder(r).Decode(data)
	}
	return cbor.NewDecoder(r).Decode(data)
}

// Marshal serializes the given data to CBOR bytes
func (c *Codec[T]) Marshal(data T) ([]byte, error) {
	if c.em != nil {
		return c.em.Marshal(data)
	}
	return cbor.Marshal(data)
}

// Unmarshal deserializes CBOR bytes into the provided type
func (c *Codec[T]) Unmarshal(data []byte, v *T) error {
	if c.dm != nil {
		return c.dm.Unmarshal(data, v)
	}
	return cbor.Unmarshal(data, v)
}
