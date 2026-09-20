package workerpool

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/armineyvazi/common.git/pkg/ports"
)

type noopLogger struct{}

func (l *noopLogger) Info(msg string, params ...any)  {}
func (l *noopLogger) Warn(msg string, params ...any)  {}
func (l *noopLogger) Error(msg string, params ...any) {}
func (l *noopLogger) Panic(msg string, params ...any) {}

func TestRegisterAndRun(t *testing.T) {
	wp := New(&noopLogger{})
	results := make(chan ports.TaskResult, 10)

	wp.RegisterTask("double", func(req ports.JobRequest) ports.TaskResult {
		n, _ := req.Params.(int)
		return ports.TaskResult{Id: n * 2}
	}, 2, 10)

	wp.Run()

	for i := 1; i <= 5; i++ {
		wp.PushJob(context.Background(), "double", results, i)
	}

	seen := make(map[int]bool)
	timeout := time.After(2 * time.Second)
	for i := 0; i < 5; i++ {
		select {
		case r := <-results:
			seen[r.Id] = true
		case <-timeout:
			t.Fatal("timed out waiting for results")
		}
	}

	for _, want := range []int{2, 4, 6, 8, 10} {
		if !seen[want] {
			t.Errorf("missing result %d", want)
		}
	}
}

func TestConcurrentJobs(t *testing.T) {
	wp := New(&noopLogger{})
	results := make(chan ports.TaskResult, 100)

	var mu sync.Mutex
	var callCount int

	wp.RegisterTask("count", func(req ports.JobRequest) ports.TaskResult {
		mu.Lock()
		callCount++
		mu.Unlock()
		return ports.TaskResult{}
	}, 4, 200)

	wp.Run()

	const n = 50
	for i := 0; i < n; i++ {
		wp.PushJob(context.Background(), "count", results, nil)
	}

	timeout := time.After(3 * time.Second)
	for i := 0; i < n; i++ {
		select {
		case <-results:
		case <-timeout:
			t.Fatalf("timed out after %d results", i)
		}
	}

	mu.Lock()
	got := callCount
	mu.Unlock()
	if got != n {
		t.Errorf("handler called %d times, want %d", got, n)
	}
}

func TestGetMetrics(t *testing.T) {
	wp := New(&noopLogger{})
	m := wp.GetMetrics()
	if m == nil {
		t.Error("GetMetrics returned nil")
	}
}
