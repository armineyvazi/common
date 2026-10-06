package uuid_test

import (
	"fmt"
	"regexp"

	"github.com/armineyvazi/common.git/pkg/adapters/uuid"
)

var uuidRegex = regexp.MustCompile(
	`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`,
)

// ExampleNew_generateV4 shows generating a UUID v4 and validating its format.
func ExampleNew_generateV4() {
	gen := uuid.New()

	id, err := gen.GenV4()
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	// UUID v4 has the form xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx
	fmt.Println(uuidRegex.MatchString(id))
	// Output:
	// true
}

// ExampleNew_uniqueness shows that consecutive calls produce distinct values.
func ExampleNew_uniqueness() {
	gen := uuid.New()

	a, _ := gen.GenV4()
	b, _ := gen.GenV4()

	fmt.Println(a != b)
	// Output:
	// true
}

// ExampleNew_useAsRequestID shows a common production pattern: attaching a
// generated UUID to every incoming request as a correlation / trace ID.
func ExampleNew_useAsRequestID() {
	gen := uuid.New()

	handleRequest := func() {
		traceID, err := gen.GenV4()
		if err != nil {
			fmt.Println("generate trace id:", err)
			return
		}
		// Attach traceID to context, log it with every message, return it
		// in response headers (e.g. X-Request-ID: <traceID>).
		fmt.Println(uuidRegex.MatchString(traceID))
	}

	handleRequest()
	// Output:
	// true
}
