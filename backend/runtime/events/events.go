// Package events defines the shared runtime event contract.
package events

import "time"

// EventType identifies one observable runtime event emitted during execution.
type EventType string

const (
	// EventStarted indicates that a runtime run has started.
	EventStarted EventType = "run.started"
	// EventCompleted indicates that a runtime run completed successfully.
	EventCompleted EventType = "run.completed"
	// EventFailed indicates that a runtime run failed.
	EventFailed EventType = "run.failed"
	// EventCancelled indicates that a runtime run was cancelled.
	EventCancelled EventType = "run.cancelled"

	// EventMessageDelta contains human-readable assistant output.
	EventMessageDelta EventType = "message.delta"
	// EventReasoningDelta contains reasoning output when available.
	EventReasoningDelta EventType = "reasoning.delta"
	// EventResult contains the final assistant result for a runtime run.
	EventResult EventType = "message.result"

	// EventToolCallStarted indicates that a tool call started.
	EventToolCallStarted EventType = "tool_call.started"
	// EventToolCallArguments contains streamed tool call arguments.
	EventToolCallArguments EventType = "tool_call.arguments"
	// EventToolCallOutput contains tool output.
	EventToolCallOutput EventType = "tool_call.output"
	// EventToolCallCompleted indicates that a tool call completed.
	EventToolCallCompleted EventType = "tool_call.completed"
	// EventToolCallFailed indicates that a tool call failed.
	EventToolCallFailed EventType = "tool_call.failed"

	// EventUsage contains token usage or provider usage metadata.
	EventUsage EventType = "run.usage"
)

// Event is the stable runtime event envelope for streaming execution.
type Event struct {
	ID        string    `json:"id,omitempty"`
	RunID     string    `json:"run_id,omitempty"`
	TraceID   string    `json:"trace_id,omitempty"`
	Seq       int64     `json:"seq,omitempty"`
	Type      EventType `json:"type"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	Content   string    `json:"content,omitempty"`
}

// UsagePayload describes model token usage when available.
type UsagePayload struct {
	InputTokens  int `json:"input_tokens,omitempty"`
	OutputTokens int `json:"output_tokens,omitempty"`
	TotalTokens  int `json:"total_tokens,omitempty"`
}
