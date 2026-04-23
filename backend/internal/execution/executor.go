package execution

import (
	"context"
	"fmt"
	"time"
)

func (e *ExecutionEngine) Execute(ctx context.Context, task *Task) *Result {
	if task == nil {
		return &Result{
			Success: false,
			Error:   fmt.Errorf("task is nil"),
		}
	}

	executor, exists := e.executors[task.Type]
	if !exists {
		return &Result{
			Success: false,
			Error:   fmt.Errorf("no executor registered for task type: %s", task.Type),
		}
	}

	timeout := task.Timeout
	if timeout == 0 {
		timeout = 5 * time.Minute
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	maxRetries := task.MaxRetries
	if maxRetries == 0 {
		maxRetries = 1
	}

	result := executor.Execute(ctx, task)
	return result
}

func (e *ExecutionEngine) RegisterExecutor(taskType TaskType, executor Executor) {
	e.executors[taskType] = executor
}
