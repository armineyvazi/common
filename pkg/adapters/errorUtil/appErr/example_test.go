package appErr_test

import (
	"context"
	"errors"
	"fmt"

	"github.com/armineyvazi/common.git/pkg/adapters/errorUtil/appErr"
	"github.com/armineyvazi/common.git/pkg/ports"
)

// ExampleNewNotFoundErr shows how to create a typed not-found error and
// check its type later in the call stack.
func ExampleNewNotFoundErr() {
	// In a repository or service layer:
	err := appErr.NewNotFoundErr(errors.New("user with id=42 not found"))

	// In a handler or middleware:
	if appErr.IsErrorType(err, ports.TypeNotFound) {
		fmt.Println("resource not found")
	}
	// Output:
	// resource not found
}

// ExampleNewInternalErr shows wrapping an unexpected database error.
func ExampleNewInternalErr() {
	dbErr := errors.New("connection timeout")
	err := appErr.NewInternalErr(dbErr)

	fmt.Println(err.GetCode() > 0)
	// Output:
	// true
}

// ExampleNew_withChaining shows the builder API for attaching context to an error.
func ExampleNew_withChaining() {
	ctx := context.Background()

	err := appErr.New(errors.New("duplicate email")).
		WithType(ports.TypeDuplicate).
		WithCode(ports.AlreadyExistsErrorCode).
		WithTrackId(ctx).
		WithExtraInfo(map[string]interface{}{
			"field": "email",
			"value": "user@example.com",
		})

	if appErr.IsErrorType(err, ports.TypeDuplicate) {
		fmt.Println("duplicate detected")
	}
	// Output:
	// duplicate detected
}

// ExampleIsErrorType shows the canonical way to check an error's domain type
// across multiple call-stack layers without importing concrete error types.
func ExampleIsErrorType() {
	makeUserErr := func(id int) error {
		if id == 0 {
			return appErr.NewBadRequestErr(errors.New("id must be positive"))
		}
		return appErr.NewNotFoundErr(fmt.Errorf("user %d not found", id))
	}

	for _, id := range []int{0, 99} {
		err := makeUserErr(id)
		switch {
		case appErr.IsErrorType(err, ports.TypeValidation):
			fmt.Printf("id=%d: bad request\n", id)
		case appErr.IsErrorType(err, ports.TypeNotFound):
			fmt.Printf("id=%d: not found\n", id)
		}
	}
	// Output:
	// id=0: bad request
	// id=99: not found
}

// ExampleNewAlreadyExistsErr shows handling a database unique-constraint violation.
func ExampleNewAlreadyExistsErr() {
	err := appErr.NewAlreadyExistsErr(errors.New("pq: duplicate key value violates unique constraint"))

	var appError ports.AppError
	if errors.As(err, &appError) {
		fmt.Println("code:", appError.GetCode() > 0)
		fmt.Println("type duplicate:", appErr.IsErrorType(err, ports.TypeDuplicate))
	}
	// Output:
	// code: true
	// type duplicate: true
}
