package zap_test

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/armineyvazi/common/pkg/adapters/logger/zap"
	"github.com/armineyvazi/common/pkg/ports"
)

// captureStdout runs fn with os.Stdout redirected and returns each JSON line.
// The logger binds os.Stdout at construction, so fn must construct it.
func captureStdout(t *testing.T, fn func()) []map[string]any {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	orig := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = orig }()

	fn()
	_ = w.Close()

	var lines []map[string]any
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		var m map[string]any
		if err := json.Unmarshal(sc.Bytes(), &m); err != nil {
			t.Fatalf("log line %q is not JSON: %v", sc.Text(), err)
		}
		lines = append(lines, m)
	}
	return lines
}

func TestStructuredMethodsEmitFields(t *testing.T) {
	tests := []struct {
		name  string
		level string
		log   func(l ports.LoggerWithTraceID, ctx context.Context)
	}{
		{name: "Debugw", level: "debug", log: func(l ports.LoggerWithTraceID, ctx context.Context) {
			l.Debugw(ctx, "cache miss", "key", "user:42", "attempt", 2)
		}},
		{name: "Infow", level: "info", log: func(l ports.LoggerWithTraceID, ctx context.Context) {
			l.Infow(ctx, "cache miss", "key", "user:42", "attempt", 2)
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), ports.TraceID{}, "trace-1")
			lines := captureStdout(t, func() {
				l := zap.New(ports.Debug, nil)
				tt.log(l, ctx)
				_ = l.Flush()
			})
			if len(lines) != 1 {
				t.Fatalf("got %d lines, want 1", len(lines))
			}
			got := lines[0]
			if got["msg"] != "cache miss" {
				t.Errorf("msg = %v, want %q (fields must not be folded into msg)", got["msg"], "cache miss")
			}
			if got["level"] != tt.level || got["key"] != "user:42" || got["attempt"] != float64(2) || got[ports.TraceIDKey] != "trace-1" {
				t.Errorf("fields = %v", got)
			}
		})
	}
}
