package zap

import (
	"context"
	"fmt"
	"os"

	"github.com/armineyvazi/common.git/pkg/ports"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type logger struct {
	core       *zap.SugaredLogger
	errHandler ports.ErrorHandler
}

func New(level ports.LogLevel, errHandler ports.ErrorHandler) ports.LoggerWithTraceID {
	var logLevel zapcore.Level
	switch level {
	case ports.Debug:
		logLevel = zap.DebugLevel
	case ports.Info:
		logLevel = zap.InfoLevel
	case ports.Warn:
		logLevel = zap.WarnLevel
	case ports.Error:
		logLevel = zap.ErrorLevel
	case ports.Panic:
		logLevel = zap.PanicLevel
	}
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	stdoutEncoder := zapcore.NewJSONEncoder(encoderConfig)
	sentryEncoder := zapcore.EncoderConfig{
		MessageKey: "msg",
	}

	var encoder []zapcore.Core

	stdoutCore := zapcore.NewCore(stdoutEncoder, os.Stdout, logLevel)

	encoder = append(encoder, stdoutCore)
	var sentryCore zapcore.Core
	if errHandler != nil {
		sentryCore = zapcore.NewCore(zapcore.NewJSONEncoder(sentryEncoder), errHandler, zap.ErrorLevel)
		encoder = append(encoder, sentryCore)
	}

	core := zapcore.NewTee(encoder...)

	z := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1), zap.AddStacktrace(zap.ErrorLevel))
	sugaredLogger := z.Sugar()
	return &logger{
		core:       sugaredLogger,
		errHandler: errHandler,
	}
}

func (l *logger) Debug(ctx context.Context, msg string, params ...any) {
	traceID := ctx.Value(ports.TraceID{})
	source := ctx.Value(ports.SourceName{})
	l.core.Debugw(fmt.Sprintf(msg, params...),
		ports.SourceNameKey, source,
		ports.TraceIDKey, traceID)
}

func (l *logger) Info(ctx context.Context, msg string, params ...any) {
	traceID := ctx.Value(ports.TraceID{})
	source := ctx.Value(ports.SourceName{})
	l.core.Infow(fmt.Sprintf(msg, params...),
		ports.SourceNameKey, source,
		ports.TraceIDKey, traceID)
}

func (l *logger) Warn(ctx context.Context, msg string, params ...any) {
	traceID := ctx.Value(ports.TraceID{})
	source := ctx.Value(ports.SourceName{})
	l.core.Warnw(fmt.Sprintf(msg, params...),
		ports.SourceNameKey, source,
		ports.TraceIDKey, traceID)
}

func (l *logger) Error(ctx context.Context, msg string, params ...any) {
	traceID := ctx.Value(ports.TraceID{})
	source := ctx.Value(ports.SourceName{})
	l.core.Errorw(fmt.Sprintf(msg, params...),
		ports.SourceNameKey, source,
		ports.TraceIDKey, traceID)
}

func (l *logger) Panic(ctx context.Context, msg string, params ...any) {
	traceID := ctx.Value(ports.TraceID{})
	source := ctx.Value(ports.SourceName{})
	l.core.Panicw(fmt.Sprintf(msg, params...),
		ports.SourceNameKey, source,
		ports.TraceIDKey, traceID)
}

func (l *logger) Errorw(ctx context.Context, msg string, params ...any) {
	traceID := ctx.Value(ports.TraceID{})
	source := ctx.Value(ports.SourceName{})
	l.core.Errorw(msg, append(
		params, ports.SourceNameKey, source,
		ports.TraceIDKey, traceID)...)
}

func (l *logger) Warnw(ctx context.Context, msg string, params ...any) {
	traceID := ctx.Value(ports.TraceID{})
	source := ctx.Value(ports.SourceName{})
	l.core.Warnw(msg, append(
		params, ports.SourceNameKey, source,
		ports.TraceIDKey, traceID)...)
}
func (l *logger) Flush() error {
	return l.core.Sync()
}
