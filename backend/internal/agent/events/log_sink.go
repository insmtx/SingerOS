package events

import (
	"context"
	"encoding/json"

	"github.com/ygpkg/yg-go/logs"
)

type logSink struct{}

// NewLogSink returns a sink that writes run events to debug logs.
func NewLogSink() EventSink {
	return logSink{}
}

func (logSink) Emit(ctx context.Context, event *RunEvent) error {
	if event == nil {
		return nil
	}

	encoded, err := json.Marshal(event)
	if err != nil {
		logs.DebugContextf(ctx, "runtime event: type=%s run_id=%s seq=%d marshal_error=%v", event.Type, event.RunID, event.Seq, err)
		return nil
	}

	logs.DebugContextf(ctx, "runtime event: %s", string(encoded))
	return nil
}
