package events

import runtimeevents "github.com/insmtx/SingerOS/backend/runtime/events"

type RunEventType = runtimeevents.EventType
type RunEvent = runtimeevents.Event
type UsagePayload = runtimeevents.UsagePayload

const (
	RunEventStarted   = runtimeevents.EventStarted
	RunEventCompleted = runtimeevents.EventCompleted
	RunEventFailed    = runtimeevents.EventFailed
	RunEventCancelled = runtimeevents.EventCancelled

	RunEventMessageDelta   = runtimeevents.EventMessageDelta
	RunEventReasoningDelta = runtimeevents.EventReasoningDelta
	RunEventResult         = runtimeevents.EventResult

	RunEventToolCallStarted   = runtimeevents.EventToolCallStarted
	RunEventToolCallArguments = runtimeevents.EventToolCallArguments
	RunEventToolCallOutput    = runtimeevents.EventToolCallOutput
	RunEventToolCallCompleted = runtimeevents.EventToolCallCompleted
	RunEventToolCallFailed    = runtimeevents.EventToolCallFailed

	RunEventUsage = runtimeevents.EventUsage
)
