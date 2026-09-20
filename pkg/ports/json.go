package ports

import "io"

// JSONCodec encodes and decodes JSON values.
// Implementations may use encoding/json or a faster drop-in replacement.
type JSONCodec interface {
	Marshal(v any) ([]byte, error)
	Unmarshal(data []byte, v any) error
	NewDecoder(r io.Reader) JSONDecoder
	NewEncoder(w io.Writer) JSONEncoder
}

// JSONDecoder reads and decodes JSON values from an input stream.
type JSONDecoder interface {
	Decode(v any) error
}

// JSONEncoder writes JSON values to an output stream.
type JSONEncoder interface {
	Encode(v any) error
}
