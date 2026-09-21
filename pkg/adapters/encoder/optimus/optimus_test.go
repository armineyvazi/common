package optimus

import (
	"testing"
)

func TestEncodeDecodeRoundtrip(t *testing.T) {
	enc := NewOptimusEncoder(Config{})
	// The optimus library works within 32-bit integer space.
	inputs := []uint64{0, 1, 42, 1000, 1_000_000}
	for _, in := range inputs {
		encoded := enc.Encode(in)
		decoded := enc.Decode(encoded)
		if decoded != in {
			t.Errorf("roundtrip(%d): encoded=%d, decoded=%d", in, encoded, decoded)
		}
	}
}

func TestEncode_DifferentOutputs(t *testing.T) {
	enc := NewOptimusEncoder(Config{})
	a := enc.Encode(1)
	b := enc.Encode(2)
	if a == b {
		t.Error("distinct inputs should produce distinct outputs")
	}
	if a == 1 {
		t.Error("encoded value should differ from input")
	}
}

func TestZeroConfigFallback(t *testing.T) {
	enc := NewOptimusEncoder(Config{})
	for _, n := range []uint64{1, 100, 999_999} {
		if enc.Decode(enc.Encode(n)) != n {
			t.Errorf("roundtrip failed for %d with zero config", n)
		}
	}
}

func TestDefaultsApplied(t *testing.T) {
	// Zero config falls back to built-in defaults: encoded != original.
	enc := NewOptimusEncoder(Config{})
	if enc.Encode(12345) == 12345 {
		t.Error("encode with defaults should produce a different value")
	}
}
