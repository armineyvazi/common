package ports

import (
	"context"
)

type TaskResult struct {
	Id   int         `json:"id"`
	Key  string      `json:"key"`
	Data interface{} `json:"data"`
}

type GrouperWorkerPool interface {
	WorkerPool
	PushGroupJob(
		ctx context.Context,
		groupName string,
		taskName string,
		resultChannel chan TaskResult,
		params interface{})
	MakeGroup(groupName string)
	WaitGroupTasksCompletion(groupName string) error
}

type WorkerPool interface {
	RegisterTask(taskName string, handler func(JobRequest) TaskResult, concurrency int, queueLength int)
	PushJob(ctx context.Context, taskName string, resultChannel chan TaskResult, params interface{})
	Run()
	GetMetrics() map[string]any
}

type JobRequest struct {
	Ctx           context.Context
	ResultChannel chan TaskResult
	Params        interface{}
}
