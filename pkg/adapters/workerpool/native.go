package workerpool

import (
	"context"

	"github.com/armineyvazi/common.git/pkg/ports"
)

type NativeWorkerpool struct {
	log   ports.Logger
	tasks map[string]*task
}

type task struct {
	log          ports.Logger
	name         string
	concurrency  int
	queueLength  int
	requestQueue chan ports.JobRequest
	handler      func(ports.JobRequest) ports.TaskResult
}

func New(log ports.Logger) ports.WorkerPool {
	return &NativeWorkerpool{
		log:   log,
		tasks: map[string]*task{},
	}
}

func (t *task) worker() {
	for job := range t.requestQueue {
		res := t.handler(job)
		if job.ResultChannel != nil {
			job.ResultChannel <- res
		}
	}

	panic("worker is dead :(")
}

func (nw *NativeWorkerpool) RegisterTask(taskName string, handler func(job ports.JobRequest) ports.TaskResult, concurrency int, queueLength int) {
	nw.tasks[taskName] = &task{
		log:          nw.log,
		name:         taskName,
		concurrency:  concurrency,
		queueLength:  queueLength,
		requestQueue: make(chan ports.JobRequest, 1000), // worker bug is here, for unbuffered channel, set 1000 for queue size for not blocked, it should be fixed.
		handler:      handler,
	}
	nw.log.Info("task %s registered", taskName)
}

func (nw *NativeWorkerpool) PushJob(ctx context.Context, taskName string, resultChannel chan ports.TaskResult, params interface{}) {
	nw.tasks[taskName].requestQueue <- ports.JobRequest{
		Ctx:           ctx,
		ResultChannel: resultChannel,
		Params:        params,
	}
}

func (nw *NativeWorkerpool) Run() {
	for _, task := range nw.tasks {
		for i := 0; i < task.concurrency; i++ {
			go task.worker()
		}
	}
}

func (nw *NativeWorkerpool) GetMetrics() map[string]any {
	return map[string]any{}
}
