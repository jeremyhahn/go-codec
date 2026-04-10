//go:build !codec_none

package codec

import (
	"testing"
)

func TestRegisterCodec(t *testing.T) {
	// Register a test codec
	testCodec := Type("test_codec")
	RegisterCodec(testCodec)

	// Should be supported
	if !IsSupported(testCodec) {
		t.Errorf("Expected test codec to be supported after registration")
	}

	// Should appear in SupportedCodecs
	supported := SupportedCodecs()
	found := false
	for _, c := range supported {
		if c == testCodec {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected test codec to appear in SupportedCodecs()")
	}
}

func TestIsSupported_Unsupported(t *testing.T) {
	// An unknown codec type should not be supported
	if IsSupported("unknown") {
		t.Error("Expected 'unknown' codec to not be supported")
	}
}

func TestCodecsRegisteredWithBuildTags(t *testing.T) {
	// This test verifies the registration mechanism works.
	// When compiled with build tags (e.g., -tags codec_cbor,codec_json,codec_msgpack),
	// those codecs will be registered via their init() functions.
	// Without build tags, no codecs are registered by default.

	// Get currently registered codecs
	supported := SupportedCodecs()

	// Just verify the mechanism works - list what's registered
	t.Logf("Currently registered codecs: %v", supported)

	// If we registered a test codec above, it should still be there
	if !IsSupported("test_codec") {
		t.Error("Expected test_codec to still be registered from TestRegisterCodec")
	}
}

func TestErrCodecNotSupported(t *testing.T) {
	err := ErrCodecNotSupported{CodecType: JSON}
	msg := err.Error()

	if msg == "" {
		t.Error("Expected non-empty error message")
	}

	// Should contain the codec type
	if !contains(msg, string(JSON)) {
		t.Errorf("Error message should contain codec type %q, got: %s", JSON, msg)
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
