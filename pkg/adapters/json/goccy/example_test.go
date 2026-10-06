package goccy_test

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/armineyvazi/common.git/pkg/adapters/json/goccy"
)

type event struct {
	ID     int    `json:"id"`
	Action string `json:"action"`
}

// ExampleNew_marshal shows marshalling a struct to JSON bytes.
func ExampleNew_marshal() {
	codec := goccy.New()

	data, err := codec.Marshal(event{ID: 1, Action: "user.created"})
	if err != nil {
		fmt.Println("marshal:", err)
		return
	}
	fmt.Println(string(data))
	// Output:
	// {"id":1,"action":"user.created"}
}

// ExampleNew_unmarshal shows deserialising JSON bytes back to a struct.
func ExampleNew_unmarshal() {
	codec := goccy.New()

	var e event
	if err := codec.Unmarshal([]byte(`{"id":7,"action":"order.placed"}`), &e); err != nil {
		fmt.Println("unmarshal:", err)
		return
	}
	fmt.Printf("id=%d action=%s\n", e.ID, e.Action)
	// Output:
	// id=7 action=order.placed
}

// ExampleNew_decoder shows streaming JSON decoding from a reader, which
// avoids holding the full payload in memory for large inputs.
func ExampleNew_decoder() {
	codec := goccy.New()

	src := strings.NewReader(`{"id":3,"action":"payment.confirmed"}`)
	dec := codec.NewDecoder(src)

	var e event
	if err := dec.Decode(&e); err != nil {
		fmt.Println("decode:", err)
		return
	}
	fmt.Printf("id=%d action=%s\n", e.ID, e.Action)
	// Output:
	// id=3 action=payment.confirmed
}

// ExampleNew_encoder shows streaming JSON encoding to a writer.
func ExampleNew_encoder() {
	codec := goccy.New()

	var buf bytes.Buffer
	enc := codec.NewEncoder(&buf)

	if err := enc.Encode(event{ID: 5, Action: "cart.cleared"}); err != nil {
		fmt.Println("encode:", err)
		return
	}
	fmt.Print(buf.String())
	// Output:
	// {"id":5,"action":"cart.cleared"}
}
