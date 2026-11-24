package codec

import (
	"fmt"
	"sort"
	"sync"
)

var (
	supportedCodecs = make(map[Type]bool)
	codecMu         sync.RWMutex
)

// RegisterCodec registers a codec as supported. Each codec package calls it
// from its init function, so a codec is registered exactly when the program
// imports its package (the database/sql driver pattern). No build tag is
// involved.
func RegisterCodec(t Type) {
	codecMu.Lock()
	defer codecMu.Unlock()
	supportedCodecs[t] = true
}

// SupportedCodecs returns a sorted list of all codecs registered in this
// program, that is, the codec packages it imports.
func SupportedCodecs() []Type {
	codecMu.RLock()
	defer codecMu.RUnlock()

	result := make([]Type, 0, len(supportedCodecs))
	for t := range supportedCodecs {
		result = append(result, t)
	}

	// Sort for consistent ordering
	sort.Slice(result, func(i, j int) bool {
		return string(result[i]) < string(result[j])
	})

	return result
}

// IsSupported returns true if the given codec type is registered, that is,
// its codec package is imported by this program.
func IsSupported(t Type) bool {
	codecMu.RLock()
	defer codecMu.RUnlock()
	return supportedCodecs[t]
}

// ErrCodecNotSupported is returned when attempting to use a codec that is not
// registered: its package is not imported, or the type is unknown.
type ErrCodecNotSupported struct {
	CodecType Type
}

func (e ErrCodecNotSupported) Error() string {
	return fmt.Sprintf("codec %q is not supported in this build; import its package (e.g., github.com/jeremyhahn/go-codec/pkg/%s) to register it", e.CodecType, e.CodecType)
}
