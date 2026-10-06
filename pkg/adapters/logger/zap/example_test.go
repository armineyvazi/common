package zap_test

import (
	"context"

	"github.com/armineyvazi/common.git/pkg/adapters/logger/zap"
	"github.com/armineyvazi/common.git/pkg/ports"
)

// ExampleNew shows constructing a production-grade structured logger and
// attaching a trace ID to a request context before logging.
func ExampleNew() {
	log := zap.New(ports.Info, nil) // nil errHandler disables Sentry integration

	// Attach a trace ID to the context. Every log line emitted with this
	// context will include the trace_id field automatically.
	ctx := context.WithValue(context.Background(), ports.TraceID{}, "req-abc-123")

	log.Info(ctx, "server started", "addr", ":8080")
	log.Debug(ctx, "debug message") // omitted at Info level
}

// ExampleNew_withSentry shows wiring a Sentry-backed error handler so that
// Error-level log lines are forwarded to Sentry in addition to stdout.
func ExampleNew_withSentry() {
	// errHandler implements ports.ErrorHandler (io.Writer).
	// In production this would be a Sentry core or similar.
	// Here we pass nil to keep the example self-contained.
	log := zap.New(ports.Error, nil)

	ctx := context.Background()
	log.Error(ctx, "payment gateway unreachable", "attempt", 3)
}

// ExampleNew_levels shows that only messages at or above the configured level
// are emitted. This example is not executed (no // Output: comment).
func ExampleNew_levels() {
	log := zap.New(ports.Warn, nil)
	ctx := context.Background()

	log.Debug(ctx, "this is suppressed") // below Warn threshold
	log.Info(ctx, "this is also suppressed")
	log.Warn(ctx, "this is emitted")
	log.Error(ctx, "this is also emitted")
}
