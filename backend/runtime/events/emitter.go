package events

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"
)

// Emitter fills common event metadata and forwards events to a Sink.
type Emitter struct {
	runID   string
	traceID string
	sink    Sink
	seq     atomic.Int64
}

// NewEmitter creates a run-scoped event emitter.
func NewEmitter(runID string, traceID string, sink Sink) *Emitter {
	if sink == nil {
		sink = NewNoopSink()
	}
	return &Emitter{
		runID:   runID,
		traceID: traceID,
		sink:    sink,
	}
}

// Emit forwards one event after filling stable metadata.
func (e *Emitter) Emit(ctx context.Context, event *Event) error {
	if e == nil || event == nil {
		return nil
	}
	event.Seq = e.seq.Add(1)
	if event.RunID == "" {
		event.RunID = e.runID
	}
	if event.TraceID == "" {
		event.TraceID = e.traceID
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now().UTC()
	}
	if event.ID == "" {
		event.ID = fmt.Sprintf("%s:%d", event.RunID, event.Seq)
	}
	return e.sink.Emit(ctx, event)
}
