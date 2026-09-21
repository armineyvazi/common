// Package goccy provides a ports.JSONCodec backed by github.com/goccy/go-json,
// a high-performance, pure-Go drop-in replacement for encoding/json.
// It is API-compatible with encoding/json; swap it in by constructing this
// adapter and passing it wherever a ports.JSONCodec is required.
package goccy

import (
	"io"

	"github.com/goccy/go-json"

	"github.com/armineyvazi/common.git/pkg/ports"
)

// Codec implements ports.JSONCodec using goccy/go-json.
type Codec struct{}

// New returns a ports.JSONCodec backed by goccy/go-json.
func New() ports.JSONCodec {
	return &Codec{}
}

func (c *Codec) Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

func (c *Codec) Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

func (c *Codec) NewDecoder(r io.Reader) ports.JSONDecoder {
	return json.NewDecoder(r)
}

func (c *Codec) NewEncoder(w io.Writer) ports.JSONEncoder {
	return json.NewEncoder(w)
}
