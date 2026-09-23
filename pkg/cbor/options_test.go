package cbor

import (
	"bytes"
	"errors"
	"testing"

	fxcbor "github.com/fxamacker/cbor/v2"
)

func strictDecOptions() fxcbor.DecOptions {
	return fxcbor.DecOptions{
		DupMapKey:       fxcbor.DupMapKeyEnforcedAPF,
		IndefLength:     fxcbor.IndefLengthForbidden,
		TagsMd:          fxcbor.TagsForbidden,
		UTF8:            fxcbor.UTF8RejectInvalid,
		MaxNestedLevels: 4,
	}
}

// Hand-encoded CBOR inputs a strict verifier must refuse.
var (
	// {"a": 1, "a": 2}
	dupKeyMap = []byte{0xa2, 0x61, 0x61, 0x01, 0x61, 0x61, 0x02}
	// {_ "a": 1} (indefinite-length map)
	indefMap = []byte{0xbf, 0x61, 0x61, 0x01, 0xff}
	// 1(0) — tag 1 (epoch time) wrapping 0
	taggedInt = []byte{0xc1, 0x00}
	// [[[[[1]]]]] — five levels of nesting
	deepArray = []byte{0x81, 0x81, 0x81, 0x81, 0x81, 0x01}
	// text string with invalid UTF-8 byte 0xff
	badUTF8 = []byte{0x61, 0xff}
)

func TestNewWithOptions_ValidOptions(t *testing.T) {
	c, err := NewWithOptions[TestStruct](strictDecOptions(), fxcbor.EncOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil codec")
	}
}

func TestNewWithOptions_InvalidDecOptions(t *testing.T) {
	c, err := NewWithOptions[TestStruct](fxcbor.DecOptions{DupMapKey: fxcbor.DupMapKeyMode(99)}, fxcbor.EncOptions{})
	if err == nil {
		t.Fatal("expected error for invalid DecOptions")
	}
	if c != nil {
		t.Fatal("expected nil codec on error")
	}
	var oe *OptionsError
	if !errors.As(err, &oe) {
		t.Fatalf("expected *OptionsError, got %T", err)
	}
	if oe.Mode != "decode" {
		t.Errorf("expected mode decode, got %q", oe.Mode)
	}
	if !errors.Is(err, ErrInvalidOptions) {
		t.Error("expected errors.Is(err, ErrInvalidOptions)")
	}
	if oe.Unwrap() == nil {
		t.Error("expected wrapped cause")
	}
}

func TestNewWithOptions_InvalidEncOptions(t *testing.T) {
	_, err := NewWithOptions[TestStruct](fxcbor.DecOptions{}, fxcbor.EncOptions{Sort: fxcbor.SortMode(99)})
	var oe *OptionsError
	if !errors.As(err, &oe) {
		t.Fatalf("expected *OptionsError, got %T (%v)", err, err)
	}
	if oe.Mode != "encode" {
		t.Errorf("expected mode encode, got %q", oe.Mode)
	}
	if !errors.Is(err, ErrInvalidOptions) {
		t.Error("expected errors.Is(err, ErrInvalidOptions)")
	}
	if oe.Error() == "" {
		t.Error("expected non-empty message")
	}
}

func TestNewWithOptions_StrictRefusesAdversarialInput(t *testing.T) {
	cases := []struct {
		name  string
		input []byte
	}{
		{"duplicate map key", dupKeyMap},
		{"indefinite-length map", indefMap},
		{"cbor tag", taggedInt},
		{"nesting beyond bound", deepArray},
		{"invalid utf-8", badUTF8},
	}
	c, err := NewWithOptions[any](strictDecOptions(), fxcbor.EncOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var v any
			if err := c.Unmarshal(tc.input, &v); err == nil {
				t.Errorf("Unmarshal accepted %x, want refusal", tc.input)
			}
			var d any
			if err := c.Decode(bytes.NewReader(tc.input), &d); err == nil {
				t.Errorf("Decode accepted %x, want refusal", tc.input)
			}
		})
	}
}

// The default codec keeps fxamacker's default (lenient) behavior.
func TestNew_DefaultBehaviourUnchanged(t *testing.T) {
	c := New[any]()
	for _, in := range [][]byte{dupKeyMap, indefMap, taggedInt} {
		var v any
		if err := c.Unmarshal(in, &v); err != nil {
			t.Errorf("default Unmarshal refused %x: %v", in, err)
		}
	}
}

func TestNewWithOptions_RoundTrip(t *testing.T) {
	c, err := NewWithOptions[TestStruct](strictDecOptions(), fxcbor.CoreDetEncOptions())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	in := TestStruct{Name: "Ada", Age: 36, Email: "ada@example.com"}

	b, err := c.Marshal(in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out TestStruct
	if err := c.Unmarshal(b, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out != in {
		t.Errorf("round trip mismatch: got %+v want %+v", out, in)
	}

	var buf bytes.Buffer
	if err := c.Encode(&buf, in); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	if !bytes.Equal(buf.Bytes(), b) {
		t.Errorf("Encode and Marshal disagree: %x vs %x", buf.Bytes(), b)
	}
	var dec TestStruct
	if err := c.Decode(&buf, &dec); err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if dec != in {
		t.Errorf("decode mismatch: got %+v want %+v", dec, in)
	}
}

// Core deterministic encoding sorts map keys; the default encoder does not
// guarantee it. This proves EncOptions reach the encoder.
func TestNewWithOptions_EncOptionsApplied(t *testing.T) {
	c, err := NewWithOptions[map[string]int](fxcbor.DecOptions{}, fxcbor.CoreDetEncOptions())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := map[string]int{"bb": 1, "a": 2, "c": 3}
	want := []byte{0xa3, 0x61, 0x61, 0x02, 0x61, 0x63, 0x03, 0x62, 0x62, 0x62, 0x01}
	for i := 0; i < 10; i++ {
		got, err := c.Marshal(m)
		if err != nil {
			t.Fatalf("Marshal: %v", err)
		}
		if !bytes.Equal(got, want) {
			t.Fatalf("got %x want %x", got, want)
		}
	}
}

// Strict options do not enforce canonical input: a non-shortest encoding of
// integer 1 (0x18 0x01) is accepted. docs/cbor.md states this limitation;
// this test keeps the documentation honest.
func TestNewWithOptions_StrictAcceptsNonCanonicalInput(t *testing.T) {
	c, err := NewWithOptions[int](strictDecOptions(), fxcbor.CoreDetEncOptions())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var v int
	if err := c.Unmarshal([]byte{0x18, 0x01}, &v); err != nil {
		t.Fatalf("Unmarshal refused non-shortest encoding: %v", err)
	}
	if v != 1 {
		t.Fatalf("got %d want 1", v)
	}
}
