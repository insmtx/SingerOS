package events

import "context"

// EventSink receives run events emitted during execution.
type EventSink interface {
	Emit(ctx context.Context, event *RunEvent) error
}

// Sink is kept as a short compatibility alias for EventSink.
//
// Deprecated: use EventSink.
type Sink = EventSink
