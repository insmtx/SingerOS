package events

import runtimeevents "github.com/insmtx/SingerOS/backend/runtime/events"

// SinkFunc adapts a function into an EventSink.
type SinkFunc = runtimeevents.SinkFunc

// NewNoopSink returns a sink that drops all events.
var NewNoopSink = runtimeevents.NewNoopSink

// ChannelSink sends events to a channel and is useful for SSE/WebSocket adapters.
type ChannelSink = runtimeevents.ChannelSink
