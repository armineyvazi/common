package workerpool_test

import (
	"context"
	"fmt"

	"github.com/armineyvazi/common.git/pkg/adapters/workerpool"
	"github.com/armineyvazi/common.git/pkg/ports"
)

// silentLogger satisfies ports.Logger without producing any output.
type silentLogger struct{}

func (l *silentLogger) Info(msg string, params ...any)  {}
func (l *silentLogger) Warn(msg string, params ...any)  {}
func (l *silentLogger) Error(msg string, params ...any) {}
func (l *silentLogger) Panic(msg string, params ...any) {}

// ExampleNew shows registering a task handler, starting the pool, and
// collecting results from a buffered channel.
func ExampleNew() {
	wp := workerpool.New(&silentLogger{})

	// Register a "double" task: receives an int, returns its double.
	wp.RegisterTask("double",
		func(req ports.JobRequest) ports.TaskResult {
			n, _ := req.Params.(int)
			return ports.TaskResult{Id: n * 2}
		},
		2,    // concurrency (worker goroutines)
		1000, // queue length
	)

	wp.Run()

	results := make(chan ports.TaskResult, 3)
	ctx := context.Background()

	wp.PushJob(ctx, "double", results, 5)
	wp.PushJob(ctx, "double", results, 10)
	wp.PushJob(ctx, "double", results, 15)

	sum := 0
	for i := 0; i < 3; i++ {
		r := <-results
		sum += r.Id
	}
	fmt.Println("sum:", sum)
	// Output:
	// sum: 60
}

// ExampleNew_multipleTaskTypes shows registering several task handlers and
// dispatching jobs to each independently.
func ExampleNew_multipleTaskTypes() {
	wp := workerpool.New(&silentLogger{})

	wp.RegisterTask("greet",
		func(req ports.JobRequest) ports.TaskResult {
			name, _ := req.Params.(string)
			return ports.TaskResult{Data: "hello " + name}
		},
		1, 100,
	)

	wp.RegisterTask("square",
		func(req ports.JobRequest) ports.TaskResult {
			n, _ := req.Params.(int)
			return ports.TaskResult{Id: n * n}
		},
		2, 100,
	)

	wp.Run()

	greetCh := make(chan ports.TaskResult, 1)
	squareCh := make(chan ports.TaskResult, 1)
	ctx := context.Background()

	wp.PushJob(ctx, "greet", greetCh, "world")
	wp.PushJob(ctx, "square", squareCh, 7)

	fmt.Println((<-greetCh).Data)
	fmt.Println((<-squareCh).Id)
	// Output:
	// hello world
	// 49
}

// ExampleNew_fireAndForget shows dispatching a job when no result is needed.
// Passing nil as the result channel avoids channel allocation and blocking.
func ExampleNew_fireAndForget() {
	wp := workerpool.New(&silentLogger{})

	processed := make(chan struct{}, 1)
	wp.RegisterTask("notify",
		func(req ports.JobRequest) ports.TaskResult {
			processed <- struct{}{}
			return ports.TaskResult{}
		},
		1, 10,
	)
	wp.Run()

	// Pass nil result channel — fire and forget.
	wp.PushJob(context.Background(), "notify", nil, nil)
	<-processed
	fmt.Println("notification sent")
	// Output:
	// notification sent
}
