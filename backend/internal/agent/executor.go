package agent

import (
	"context"

	"github.com/insmtx/SingerOS/backend/internal/execution"
	"github.com/insmtx/SingerOS/backend/interaction"
	"github.com/ygpkg/yg-go/logs"
)

type AgentExecutor struct {
	runtime Runtime
}

func NewAgentExecutor(runtime Runtime) *AgentExecutor {
	return &AgentExecutor{runtime: runtime}
}

func (e *AgentExecutor) Execute(ctx context.Context, task *execution.Task) *execution.Result {
	if task == nil {
		return &execution.Result{
			Success: false,
			Error:   nil,
		}
	}

	logs.InfoContextf(ctx, "Executing agent task: %+v", task)

	eventRaw, ok := task.Payload["event"]
	if !ok {
		return &execution.Result{
			Success: false,
			Error:   nil,
		}
	}

	interactionEvent, ok := eventRaw.(*interaction.Event)
	if !ok {
		return &execution.Result{
			Success: false,
			Error:   nil,
		}
	}

	err := e.runtime.HandleEvent(ctx, interactionEvent)
	if err != nil {
		return &execution.Result{
			Success: false,
			Error:   err,
		}
	}

	return &execution.Result{
		Success: true,
		Output:  "Agent execution completed",
	}
}
