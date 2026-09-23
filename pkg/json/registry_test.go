package json_test

import (
	"reflect"
	"testing"

	codec "github.com/jeremyhahn/go-codec"
	jsoncodec "github.com/jeremyhahn/go-codec/pkg/json"
)

// TestImportRegistersOnlyJSON proves the import is the opt-in: this test
// binary links pkg/json and no other codec package, so the registry must hold
// JSON and nothing else — with no build tags.
func TestImportRegistersOnlyJSON(t *testing.T) {
	want := []codec.Type{codec.JSON}
	if got := codec.SupportedCodecs(); !reflect.DeepEqual(got, want) {
		t.Fatalf("SupportedCodecs() = %v, want %v", got, want)
	}
	for _, other := range []codec.Type{codec.YAML, codec.TOML, codec.MsgPack, codec.BSON, codec.CBOR, codec.ProtoBuf} {
		if codec.IsSupported(other) {
			t.Errorf("IsSupported(%q) = true, want false: pkg/json must not register other codecs", other)
		}
	}

	b, err := jsoncodec.New[map[string]int]().Marshal(map[string]int{"a": 1})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if string(b) != `{"a":1}` {
		t.Errorf("Marshal = %s, want {\"a\":1}", b)
	}
}
