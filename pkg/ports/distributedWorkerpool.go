package ports

import (
	"context"
	"time"

	"github.com/hibiken/asynq"
)

type Task = asynq.Task

type DistributedWorkerpoolWorker interface {
	RegisterTask(taskName string, handler func(ctx context.Context, t *Task) error)
	ActiveWebConsole(monitoringAddress string) error
	StartWorker() error
}

type DistributedWorkerpoolClient interface {
	PushJob(queueName, taskName string, params interface{}, maxRetry int, timeOut time.Duration) error
}
