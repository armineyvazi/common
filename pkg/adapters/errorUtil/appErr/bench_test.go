package appErr

import (
	"context"
	"errors"
	"testing"

	"github.com/armineyvazi/common.git/pkg/ports"
)

var sink ports.AppError // prevents dead-code elimination

func BenchmarkNewInternalErr(b *testing.B) {
	cause := errors.New("db timeout")
	b.ReportAllocs()
	for b.Loop() {
		sink = NewInternalErr(cause)
	}
}

func BenchmarkNewNotFoundErr(b *testing.B) {
	cause := errors.New("row not found")
	b.ReportAllocs()
	for b.Loop() {
		sink = NewNotFoundErr(cause)
	}
}

func BenchmarkAppErr_WithTrackId_FromContext(b *testing.B) {
	ctx := context.WithValue(context.Background(), ports.TraceID{}, "trace-bench")
	cause := errors.New("x")
	b.ReportAllocs()
	for b.Loop() {
		sink = NewInternalErr(cause).WithTrackId(ctx)
	}
}

func BenchmarkAppErr_WithTrackId_Generated(b *testing.B) {
	cause := errors.New("x")
	b.ReportAllocs()
	for b.Loop() {
		sink = NewInternalErr(cause).WithTrackId(context.Background())
	}
}

func BenchmarkAppErr_WithExtraInfo(b *testing.B) {
	cause := errors.New("x")
	extra := map[string]interface{}{"user_id": 42, "tenant": "acme"}
	b.ReportAllocs()
	for b.Loop() {
		sink = NewInternalErr(cause).WithExtraInfo(extra)
	}
}

func BenchmarkAppErr_WithErrorList(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		sink = New(nil).
			WithErrorList("email", "invalid format").
			WithErrorList("email", "too long").
			WithErrorList("name", "required")
	}
}

func BenchmarkAppErr_Error(b *testing.B) {
	e := NewInternalErr(errors.New("cause")).
		WithTrackId(context.WithValue(context.Background(), ports.TraceID{}, "t-bench"))
	b.ReportAllocs()
	for b.Loop() {
		_ = e.Error()
	}
}

func BenchmarkIsErrorType(b *testing.B) {
	e := NewNotFoundErr(errors.New("x"))
	b.ReportAllocs()
	for b.Loop() {
		_ = IsErrorType(e, ports.TypeNotFound)
	}
}

func BenchmarkHandleError_AppError(b *testing.B) {
	e := NewInternalErr(errors.New("db"))
	b.ReportAllocs()
	for b.Loop() {
		sink = HandleError(e)
	}
}

func BenchmarkHandleError_PlainError(b *testing.B) {
	e := errors.New("plain")
	b.ReportAllocs()
	for b.Loop() {
		sink = HandleError(e)
	}
}
