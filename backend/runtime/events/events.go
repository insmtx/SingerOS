package events

import (
	"time"
)

// RunEventType identifies one observable event emitted during an agent run.
type RunEventType string

const (
	RunEventStarted   RunEventType = "run.started"
	RunEventCompleted RunEventType = "run.completed"
	RunEventFailed    RunEventType = "run.failed"
	RunEventCancelled RunEventType = "run.cancelled"

	RunEventMessageDelta   RunEventType = "message.delta"
	RunEventReasoningDelta RunEventType = "reasoning.delta"

	RunEventToolCallStarted   RunEventType = "tool_call.started"
	RunEventToolCallArguments RunEventType = "tool_call.arguments"
	RunEventToolCallOutput    RunEventType = "tool_call.output"
	RunEventToolCallCompleted RunEventType = "tool_call.completed"
	RunEventToolCallFailed    RunEventType = "tool_call.failed"

	RunEventUsage RunEventType = "run.usage"
)

// RunEvent is the stable runtime event envelope for streaming agent execution.
type RunEvent struct {
	ID        string       `json:"id"`
	RunID     string       `json:"run_id"`
	TraceID   string       `json:"trace_id,omitempty"`
	Seq       int64        `json:"seq"`
	Type      RunEventType `json:"type"`
	CreatedAt time.Time    `json:"created_at"`
	Content   string       `json:"content,omitempty"`
}

// UsagePayload describes model token usage when available.
type UsagePayload struct {
	InputTokens  int `json:"input_tokens,omitempty"`
	OutputTokens int `json:"output_tokens,omitempty"`
	TotalTokens  int `json:"total_tokens,omitempty"`
}
