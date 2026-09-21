package appErr

import (
	"context"
	"errors"
	"testing"

	"github.com/armineyvazi/common.git/pkg/ports"
)

func TestNewAppError(t *testing.T) {
	inner := errors.New("inner")
	e := NewAppError(ports.InternalErrorCode, ports.InternalErrorMessage, inner)
	if e.GetCode() != ports.InternalErrorCode {
		t.Errorf("code: got %d, want %d", e.GetCode(), ports.InternalErrorCode)
	}
	if e.GetNestedError() != inner {
		t.Error("nested error not set")
	}
}

func TestWithType_SetsDefaultMessage(t *testing.T) {
	e := New(nil).WithType(ports.TypeNotFound)
	if e.GetMessage() != ports.NotFoundMessage {
		t.Errorf("message: got %q, want %q", e.GetMessage(), ports.NotFoundMessage)
	}
	if e.GetType() != ports.TypeNotFound {
		t.Errorf("type: got %q, want %q", e.GetType(), ports.TypeNotFound)
	}
}

func TestWithMessage_OverridesDefault(t *testing.T) {
	e := New(nil).WithType(ports.TypeInternal).WithMessage("custom msg")
	if e.GetMessage() != "custom msg" {
		t.Errorf("got %q, want %q", e.GetMessage(), "custom msg")
	}
}

func TestWithTrackId_FromContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), ports.TraceID{}, "trace-abc")
	e := New(nil).WithTrackId(ctx)
	if e.GetTrackId() != "trace-abc" {
		t.Errorf("track id: got %q, want %q", e.GetTrackId(), "trace-abc")
	}
}

func TestWithTrackId_GeneratedWhenMissing(t *testing.T) {
	e := New(nil).WithTrackId(context.Background())
	if e.GetTrackId() == "" {
		t.Error("expected generated track id, got empty string")
	}
}

func TestWithTrackId_NotOverwritten(t *testing.T) {
	ctx := context.WithValue(context.Background(), ports.TraceID{}, "first-id")
	e := New(nil).WithTrackId(ctx)
	_ = e.WithTrackId(context.WithValue(context.Background(), ports.TraceID{}, "second-id"))
	if e.GetTrackId() != "first-id" {
		t.Errorf("track id should not be overwritten: got %q", e.GetTrackId())
	}
}

func TestWithErrorList(t *testing.T) {
	e := New(nil).
		WithErrorList("email", "invalid format").
		WithErrorList("email", "too long").
		WithErrorList("name", "required")

	lists := e.GetErrorLists()
	if len(lists["email"]) != 2 {
		t.Errorf("email errors: got %d, want 2", len(lists["email"]))
	}
	if len(lists["name"]) != 1 {
		t.Errorf("name errors: got %d, want 1", len(lists["name"]))
	}
}

func TestWithExtraInfo_MergesKeys(t *testing.T) {
	e := New(nil).
		WithExtraInfo(map[string]interface{}{"a": 1}).
		WithExtraInfo(map[string]interface{}{"b": 2, "a": 99})

	info := e.GetExtraInfo()
	if info["a"] != 99 {
		t.Errorf("expected a=99, got %v", info["a"])
	}
	if info["b"] != 2 {
		t.Errorf("expected b=2, got %v", info["b"])
	}
}

func TestIsDebugDisabled(t *testing.T) {
	e := New(nil)
	if e.IsDebugDisabled() {
		t.Error("debug should be enabled by default")
	}
	_ = e.WithDisableDebug()
	if !e.IsDebugDisabled() {
		t.Error("debug should be disabled after WithDisableDebug")
	}
}

func TestError_String(t *testing.T) {
	e := &appErr{
		TrackId: "tid",
		Message: "oops",
		Detail:  "detail here",
	}
	s := e.Error()
	for _, substr := range []string{"tid", "oops", "detail here"} {
		if !contains(s, substr) {
			t.Errorf("Error() = %q, missing %q", s, substr)
		}
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStr(s, sub))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func TestConstructors(t *testing.T) {
	inner := errors.New("cause")
	tests := []struct {
		name     string
		err      ports.AppError
		wantType ports.ErrorType
		wantCode int
	}{
		{"NewInternalErr", NewInternalErr(inner), ports.TypeInternal, ports.InternalErrorCode},
		{"NewNotFoundErr", NewNotFoundErr(inner), ports.TypeNotFound, ports.NotFoundErrorCode},
		{"NewInvalidArgumentErr", NewInvalidArgumentErr(inner), ports.TypeValidation, ports.InvalidArgumentErrorCode},
		{"NewPermissionDeniedErr", NewPermissionDeniedErr(inner), ports.TypeForbidden, ports.PermissionDeniedErrorCode},
		{"NewAlreadyExistsErr", NewAlreadyExistsErr(inner), ports.TypeDuplicate, ports.AlreadyExistsErrorCode},
		{"NewUnauthenticatedErr", NewUnauthenticatedErr(inner), ports.TypeUnAuthorized, ports.UnauthenticatedErrorCode},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.GetType() != tt.wantType {
				t.Errorf("type: got %q, want %q", tt.err.GetType(), tt.wantType)
			}
			if tt.err.GetCode() != tt.wantCode {
				t.Errorf("code: got %d, want %d", tt.err.GetCode(), tt.wantCode)
			}
		})
	}
}
