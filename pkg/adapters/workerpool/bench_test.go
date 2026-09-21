package workerpool

import (
	"context"
	"testing"

	"github.com/armineyvazi/common.git/pkg/ports"
)

// BenchmarkWorkerPool_Throughput measures end-to-end job dispatch and result
// collection through a pool with 4 workers and a 10 000-slot queue.
func BenchmarkWorkerPool_Throughput(b *testing.B) {
	wp := New(&noopLogger{})
	results := make(chan ports.TaskResult, b.N+1)

	wp.RegisterTask("noop", func(req ports.JobRequest) ports.TaskResult {
		return ports.TaskResult{}
	}, 4, 10000)

	wp.Run()
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		wp.PushJob(ctx, "noop", results, nil)
	}
	for i := 0; i < b.N; i++ {
		<-results
	}
}

// BenchmarkWorkerPool_Dispatch_NoResult measures pure dispatch cost when
// the caller doesn't need a result (nil result channel).
func BenchmarkWorkerPool_Dispatch_NoResult(b *testing.B) {
	wp := New(&noopLogger{})
	wp.RegisterTask("noop", func(req ports.JobRequest) ports.TaskResult {
		return ports.TaskResult{}
	}, 8, 10000)
	wp.Run()

	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		wp.PushJob(ctx, "noop", nil, nil)
	}
}

// BenchmarkWorkerPool_Parallel stresses the pool from multiple goroutines.
func BenchmarkWorkerPool_Parallel(b *testing.B) {
	wp := New(&noopLogger{})
	results := make(chan ports.TaskResult, 100000)

	wp.RegisterTask("noop", func(req ports.JobRequest) ports.TaskResult {
		return ports.TaskResult{}
	}, 8, 100000)
	wp.Run()

	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			wp.PushJob(ctx, "noop", results, nil)
			<-results
		}
	})
}
