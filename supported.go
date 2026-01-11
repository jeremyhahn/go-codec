//go:build !codec_none

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

// RegisterCodec registers a codec as supported. This is called by codec packages
// during initialization to indicate they are compiled in.
func RegisterCodec(t Type) {
	codecMu.Lock()
	defer codecMu.Unlock()
	supportedCodecs[t] = true
}

// SupportedCodecs returns a sorted list of all codecs that are compiled into the build.
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

// IsSupported returns true if the given codec type is compiled into the build.
func IsSupported(t Type) bool {
	codecMu.RLock()
	defer codecMu.RUnlock()
	return supportedCodecs[t]
}

// ErrCodecNotSupported is returned when attempting to use a codec that
// was not compiled into the build.
type ErrCodecNotSupported struct {
	CodecType Type
}

func (e ErrCodecNotSupported) Error() string {
	return fmt.Sprintf("codec %q is not supported in this build; rebuild with the appropriate build tag (e.g., -tags codec_%s)", e.CodecType, e.CodecType)
}
