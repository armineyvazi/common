package appErr

import (
	"errors"
	"testing"

	"github.com/armineyvazi/common.git/pkg/ports"
)

func TestIsErrorType(t *testing.T) {
	appError := NewNotFoundErr(errors.New("x"))
	if !IsErrorType(appError, ports.TypeNotFound) {
		t.Error("expected TypeNotFound")
	}
	if IsErrorType(appError, ports.TypeInternal) {
		t.Error("should not match TypeInternal")
	}
	if IsErrorType(errors.New("plain"), ports.TypeNotFound) {
		t.Error("plain errors should not match")
	}
}

func TestIsNotFoundError(t *testing.T) {
	if !IsNotFoundError(NewNotFoundErr(nil)) {
		t.Error("expected true for not-found error")
	}
	if IsNotFoundError(NewInternalErr(nil)) {
		t.Error("expected false for internal error")
	}
	if IsNotFoundError(errors.New("generic")) {
		t.Error("expected false for generic error")
	}
}

func TestIsDuplicate(t *testing.T) {
	if !IsDuplicate(NewAlreadyExistsErr(nil)) {
		t.Error("expected true")
	}
	if IsDuplicate(NewNotFoundErr(nil)) {
		t.Error("expected false")
	}
}

func TestIsInternalError(t *testing.T) {
	if !IsInternalError(NewInternalErr(nil)) {
		t.Error("expected true")
	}
}

func TestIsUnauthorizedError(t *testing.T) {
	if !IsUnauthorizedError(NewUnauthenticatedErr(nil)) {
		t.Error("expected true")
	}
}

func TestIsForbiddenError(t *testing.T) {
	if !IsForbiddenError(NewPermissionDeniedErr(nil)) {
		t.Error("expected true")
	}
}

func TestIsValidationError(t *testing.T) {
	if !IsValidationError(NewInvalidArgumentErr(nil)) {
		t.Error("expected true")
	}
}

func TestHandleError_WrapsNonAppError(t *testing.T) {
	err := errors.New("some error")
	e := HandleError(err)
	if e == nil {
		t.Fatal("expected non-nil")
	}
	if e.GetType() != ports.TypeValidation {
		t.Errorf("got type %q, want TypeValidation", e.GetType())
	}
}

func TestHandleError_PassesThroughAppError(t *testing.T) {
	appError := NewNotFoundErr(nil)
	result := HandleError(appError)
	if result.GetType() != ports.TypeNotFound {
		t.Errorf("got type %q, want TypeNotFound", result.GetType())
	}
}

func TestHandleError_NilInput(t *testing.T) {
	if HandleError(nil) != nil {
		t.Error("expected nil for nil input")
	}
}

func TestHandleInternalError_WrapsNonAppError(t *testing.T) {
	err := errors.New("db error")
	e := HandleInternalError(err)
	if e.GetType() != ports.TypeInternal {
		t.Errorf("got type %q, want TypeInternal", e.GetType())
	}
}
