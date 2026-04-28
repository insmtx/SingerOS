package events

import runtimeevents "github.com/insmtx/SingerOS/backend/runtime/events"

// EventSink receives run events emitted during execution.
type EventSink = runtimeevents.Sink

// Sink is kept as a short compatibility alias for EventSink.
//
// Deprecated: use EventSink.
type Sink = EventSink
