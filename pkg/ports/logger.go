package ports

import "context"

type Logger interface {
	Info(msg string, params ...any)
	Warn(msg string, params ...any)
	Error(msg string, params ...any)
	Panic(msg string, params ...any)
}

type TraceID struct{}
type SourceName struct{}

const TraceIDKey = "trace_id"
const SourceNameKey = "source"

type LoggerWithTraceID interface {
	Debug(ctx context.Context, msg string, params ...any)
	Info(ctx context.Context, msg string, params ...any)
	Warn(ctx context.Context, msg string, params ...any)
	Error(ctx context.Context, msg string, params ...any)
	Panic(ctx context.Context, msg string, params ...any)
	Errorw(ctx context.Context, msg string, params ...any)
	Warnw(ctx context.Context, msg string, params ...any)
	Flush() error
}

type LogLevel string

const (
	Debug LogLevel = "debug"
	Info  LogLevel = "info"
	Warn  LogLevel = "warn"
	Error LogLevel = "error"
	Panic LogLevel = "panic"
)
