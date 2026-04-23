package events

import "context"

// SinkFunc adapts a function into an EventSink.
type SinkFunc func(ctx context.Context, event *RunEvent) error

// Emit calls f(ctx, event).
func (f SinkFunc) Emit(ctx context.Context, event *RunEvent) error {
	return f(ctx, event)
}

type noopSink struct{}

// NewNoopSink returns a sink that drops all events.
func NewNoopSink() EventSink {
	return noopSink{}
}

func (noopSink) Emit(context.Context, *RunEvent) error {
	return nil
}

// ChannelSink sends events to a channel and is useful for SSE/WebSocket adapters.
type ChannelSink struct {
	C chan<- *RunEvent
}

// Emit sends an event without blocking forever on disconnected consumers.
func (s ChannelSink) Emit(ctx context.Context, event *RunEvent) error {
	if s.C == nil {
		return nil
	}
	select {
	case s.C <- event:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}
