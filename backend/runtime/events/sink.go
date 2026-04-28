package events

import "context"

// Sink receives runtime events emitted during execution.
type Sink interface {
	Emit(ctx context.Context, event *Event) error
}

// SinkFunc adapts a function into a Sink.
type SinkFunc func(ctx context.Context, event *Event) error

// Emit calls f(ctx, event).
func (f SinkFunc) Emit(ctx context.Context, event *Event) error {
	return f(ctx, event)
}

type noopSink struct{}

// NewNoopSink returns a sink that drops all events.
func NewNoopSink() Sink {
	return noopSink{}
}

func (noopSink) Emit(context.Context, *Event) error {
	return nil
}

// ChannelSink sends events to a channel and is useful for SSE/WebSocket adapters.
type ChannelSink struct {
	C chan<- *Event
}

// Emit sends an event without blocking forever on disconnected consumers.
func (s ChannelSink) Emit(ctx context.Context, event *Event) error {
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
