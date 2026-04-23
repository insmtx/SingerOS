package context

import (
	"context"
	"time"
)

type ExecutionContext struct {
	TaskID    string
	StartedAt time.Time
	Deadline  time.Time
	Metadata  map[string]string
}

func NewExecutionContext(taskID string, timeout time.Duration) *ExecutionContext {
	now := time.Now()
	return &ExecutionContext{
		TaskID:    taskID,
		StartedAt: now,
		Deadline:  now.Add(timeout),
		Metadata:  make(map[string]string),
	}
}

func (c *ExecutionContext) WithContext(ctx context.Context) (context.Context, context.CancelFunc) {
	if c.Deadline.IsZero() {
		return ctx, func() {}
	}
	return context.WithDeadline(ctx, c.Deadline)
}

func (c *ExecutionContext) Elapsed() time.Duration {
	return time.Since(c.StartedAt)
}

func (c *ExecutionContext) Remaining() time.Duration {
	return time.Until(c.Deadline)
}
