// Package agent defines the Agent Runtime for SingerOS.
//
// The Agent Runtime is responsible for:
// - Managing Agent lifecycle
// - Calling LLM models
// - Managing memory and context
// - Coordinating tool/skill calls
package agent

import (
	"context"

	"github.com/insmtx/SingerOS/backend/interaction"
)

// Runtime handles agent execution within the SingerOS system.
type Runtime interface {
	HandleEvent(ctx context.Context, event *interaction.Event) error
}
