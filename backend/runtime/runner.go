// Package runtime defines the unified task runner boundary for SingerOS.
//
// The orchestrator should depend on this package instead of binding directly
// to a specific agent implementation. This keeps the event ingress layer stable
// while the runtime evolves behind a single Eino-based execution boundary.
package runtime

import (
	"context"

	"github.com/insmtx/SingerOS/backend/interaction"
)

// Runner handles a normalized event inside the agent runtime.
type Runner interface {
	HandleEvent(ctx context.Context, event *interaction.Event) error
}
