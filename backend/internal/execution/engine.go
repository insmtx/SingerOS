package execution

import (
	"context"
	"time"
)

type TaskType string

const (
	TaskTypeAgent    TaskType = "agent"
	TaskTypeSkill    TaskType = "skill"
	TaskTypeWorkflow TaskType = "workflow"
)

type Task struct {
	Type       TaskType
	Payload    map[string]interface{}
	Timeout    time.Duration
	MaxRetries int
	Metadata   map[string]string
}

type Result struct {
	Output   string
	Metadata map[string]interface{}
	Success  bool
	Error    error
}

type Engine interface {
	Execute(ctx context.Context, task *Task) *Result
	RegisterExecutor(taskType TaskType, executor Executor)
}

type Executor interface {
	Execute(ctx context.Context, task *Task) *Result
}

type ExecutionEngine struct {
	executors map[TaskType]Executor
}

func NewExecutionEngine() *ExecutionEngine {
	return &ExecutionEngine{
		executors: make(map[TaskType]Executor),
	}
}
