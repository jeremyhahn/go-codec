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

// Codec implements the codec.Codec interface for CBOR serialization.
//
// A Codec from New uses fxamacker/cbor's default options. A Codec from
// NewWithOptions uses the decode and encode modes built from the caller's
// options.
type Codec[T any] struct {
	dm cbor.DecMode
	em cbor.EncMode
}

// New creates a new CBOR codec using fxamacker/cbor's default options.
//
// The defaults are lenient: duplicate map keys are accepted, and which value
// is retained depends on the destination type; indefinite-length items and
// tags are accepted. Use NewWithOptions to
// decode attacker-controlled input.
func New[T any]() *Codec[T] {
	return &Codec[T]{}
}

// NewWithOptions creates a CBOR codec whose decoding and encoding are governed
// by dec and enc. The modes are built once here; invalid options return an
// *OptionsError and a nil codec, so a misconfigured codec never reaches input.
//
// For attacker-reachable input, set at least DupMapKey to
// cbor.DupMapKeyEnforcedAPF, IndefLength to cbor.IndefLengthForbidden,
// UTF8 to cbor.UTF8RejectInvalid, MaxNestedLevels to a caller-chosen bound,
// and TagsMd to cbor.TagsForbidden where the container already identifies
// the content.
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
