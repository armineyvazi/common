package optimus_test

import (
	"fmt"

	"github.com/armineyvazi/common.git/pkg/adapters/encoder/optimus"
)

// ExampleNewOptimusEncoder shows encoding and decoding a database row ID.
// This technique obfuscates sequential IDs in public URLs (e.g. /orders/482)
// so that record counts are not guessable.
func ExampleNewOptimusEncoder() {
	enc := optimus.NewOptimusEncoder(optimus.Config{
		// Use unique prime, mod-inverse, and random values per application.
		// The defaults are used here for demonstration only.
		Prime:      1580030173,
		ModInverse: 59260789,
		Random:     1163945558,
	})

	original := uint64(42)
	encoded := enc.Encode(original)
	decoded := enc.Decode(encoded)

	fmt.Println("original:", original)
	fmt.Println("encoded != original:", encoded != original)
	fmt.Println("decoded:", decoded)
	// Output:
	// original: 42
	// encoded != original: true
	// decoded: 42
}

// ExampleNewOptimusEncoder_defaults shows that zero Config values fall back
// to built-in defaults, which is convenient for local development.
func ExampleNewOptimusEncoder_defaults() {
	enc := optimus.NewOptimusEncoder(optimus.Config{})

	id := uint64(1)
	fmt.Println(enc.Decode(enc.Encode(id)) == id)
	// Output:
	// true
}
