package goccy_test

import (
	"bytes"
	"testing"

	"github.com/armineyvazi/common.git/pkg/adapters/json/goccy"
)

func TestCodec_MarshalUnmarshal(t *testing.T) {
	c := goccy.New()

	type payload struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	original := payload{Name: "armin", Age: 30}
	b, err := c.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var got payload
	if err := c.Unmarshal(b, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got != original {
		t.Fatalf("roundtrip mismatch: got %+v, want %+v", got, original)
	}
}

func TestCodec_Decoder(t *testing.T) {
	c := goccy.New()
	input := `{"name":"armin","age":30}`
	dec := c.NewDecoder(bytes.NewBufferString(input))

	var v map[string]any
	if err := dec.Decode(&v); err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if v["name"] != "armin" {
		t.Fatalf("expected name=armin, got %v", v["name"])
	}
}

func TestCodec_Encoder(t *testing.T) {
	c := goccy.New()
	var buf bytes.Buffer
	enc := c.NewEncoder(&buf)

	if err := enc.Encode(map[string]any{"key": "val"}); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatal("expected non-empty output from Encode")
	}
}

func TestCodec_InvalidUnmarshal(t *testing.T) {
	c := goccy.New()
	var v any
	if err := c.Unmarshal([]byte(`not-json`), &v); err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}
