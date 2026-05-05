package workerpool

import (
	"context"
	"fmt"
	"github.com/alitto/pond"
	"github.com/go-errors/errors"

	"github.com/armineyvazi/common.git/pkg/ports"
)

const PondWorkerPoolName = "pond"
const PondStrategyBalanced = "balanced"
const PondStrategyLazy = "lazy"

type pondWP struct {
	log            ports.LoggerWithTraceID
	pond           *pond.WorkerPool
	taskHandlerMap map[string]func(request ports.JobRequest) ports.TaskResult
	groups         map[string]*pond.TaskGroup
}

func NewPond(
	log ports.LoggerWithTraceID,
	strategy string,
	maxWorker,
	maxCapacity int,
	taskHandler map[string]func(request ports.JobRequest,
	) ports.TaskResult) ports.WorkerPool {

	pondStrategy := pond.Strategy(pond.Eager())
	switch strategy {
	case PondStrategyBalanced:
		pondStrategy = pond.Strategy(pond.Balanced())
	case PondStrategyLazy:
		pondStrategy = pond.Strategy(pond.Lazy())
	}

	pond := pond.New(maxWorker, maxCapacity, pondStrategy)

	return &pondWP{
		log:            log,
		pond:           pond,
		taskHandlerMap: taskHandler,
	}
}

func NewGrouperWorkerPool(
	log ports.LoggerWithTraceID,
	strategy string,
	maxWorker,
	maxCapacity int,
	taskHandler map[string]func(request ports.JobRequest,
	) ports.TaskResult) ports.GrouperWorkerPool {

	pondStrategy := pond.Strategy(pond.Eager())
	switch strategy {
	case PondStrategyBalanced:
		pondStrategy = pond.Strategy(pond.Balanced())
	case PondStrategyLazy:
		pondStrategy = pond.Strategy(pond.Lazy())
	}
	pondWorker := pond.New(maxWorker, maxCapacity, pondStrategy)
	p := &pondWP{
		log:            log,
		pond:           pondWorker,
		taskHandlerMap: taskHandler,
		groups:         map[string]*pond.TaskGroup{},
	}
	return p
}

func (wp *pondWP) RegisterTask(taskName string, handler func(job ports.JobRequest) ports.TaskResult, concurrency int, queueLength int) {

}

func (wp *pondWP) Run() {

}

func (wp *pondWP) GetMetrics() map[string]any {
	return map[string]any{
		"workers_running":        wp.pond.RunningWorkers(),
		"workers_idle":           wp.pond.IdleWorkers(),
		"tasks_submitted_total":  wp.pond.SubmittedTasks(),
		"tasks_waiting_total":    wp.pond.WaitingTasks(),
		"tasks_successful_total": wp.pond.SuccessfulTasks(),
		"tasks_failed_total":     wp.pond.FailedTasks(),
		"tasks_completed_total":  wp.pond.CompletedTasks(),
	}
}

func (wp *pondWP) PushJob(ctx context.Context, taskName string, resultChannel chan ports.TaskResult, params interface{}) {
	wp.pond.Submit(func() {

		handler, ok := wp.taskHandlerMap[taskName]
		if !ok {
			wp.log.Error(ctx, fmt.Sprintf("%s has no handler", taskName))
			resultChannel <- ports.TaskResult{
				Key: "error",
			}
			return
		}

		result := handler(ports.JobRequest{
			Ctx:           ctx,
			ResultChannel: resultChannel,
			Params:        params,
		})

		resultChannel <- result
	})
}

func (wp *pondWP) MakeGroup(groupName string) {
	if _, ok := wp.groups[groupName]; !ok {
		wp.groups[groupName] = wp.pond.Group()
	}
}

func (wp *pondWP) PushGroupJob(
	ctx context.Context,
	groupName string,
	taskName string,
	resultChannel chan ports.TaskResult,
	params interface{}) {

	if groupName == "" {
		wp.log.Error(ctx, "groupName is not defined for PushGroupJob")
		resultChannel <- ports.TaskResult{
			Key: "error",
		}
		return
	}

	group, ok := wp.groups[groupName]
	if !ok {
		wp.log.Error(ctx, "invalid groupName")
		resultChannel <- ports.TaskResult{
			Key: "error",
		}
		return
	}

	group.Submit(func() {
		handler, ok := wp.taskHandlerMap[taskName]
		if !ok {
			wp.log.Error(ctx, fmt.Sprintf("%s has no handler", taskName))
			resultChannel <- ports.TaskResult{
				Key: "error",
			}
			return
		}

		result := handler(ports.JobRequest{
			Ctx:           ctx,
			ResultChannel: resultChannel,
			Params:        params,
		})
		resultChannel <- result
	})
}

func (wp *pondWP) WaitGroupTasksCompletion(
	groupName string) error {
	group, ok := wp.groups[groupName]
	if !ok {
		return errors.New("invalid groupName to Tasks Completion")
	}
	group.Wait()
	return nil
}
