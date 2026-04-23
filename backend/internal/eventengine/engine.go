package eventengine

import (
	"context"

	"github.com/insmtx/SingerOS/backend/interaction"
	"github.com/insmtx/SingerOS/backend/interaction/eventbus"
	"github.com/insmtx/SingerOS/backend/internal/execution"
)

type EventHandler func(ctx context.Context, event *interaction.Event) error

type EventEngine struct {
	subscriber      eventbus.Subscriber
	executionEngine execution.Engine
	handlers        map[string]EventHandler
}

func NewEventEngine(subscriber eventbus.Subscriber, execEngine execution.Engine) *EventEngine {
	engine := &EventEngine{
		subscriber:      subscriber,
		executionEngine: execEngine,
		handlers:        make(map[string]EventHandler),
	}

	engine.registerDefaultHandlers()

	return engine
}
