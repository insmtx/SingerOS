package simplechat

import (
	"context"
	"fmt"

	"github.com/insmtx/SingerOS/backend/internal/agent"
)

type ConsoleSink struct{}

func NewConsoleSink() *ConsoleSink {
	return &ConsoleSink{}
}

func (s *ConsoleSink) Emit(_ context.Context, event *agent.RunEvent) error {
	switch event.Type {
	case agent.RunEventMessageDelta:
		fmt.Print(event.Content)
	case agent.RunEventCompleted:
		fmt.Println()
	case agent.RunEventToolCallStarted:
		fmt.Printf("\n[Tool Call Started] %s\n", event.Content)
	case agent.RunEventToolCallOutput:
		fmt.Printf("\n[Tool Output] %s\n", event.Content)
	case agent.RunEventFailed:
		fmt.Printf("\n[Error] %s\n", event.Content)
	}
	return nil
}
