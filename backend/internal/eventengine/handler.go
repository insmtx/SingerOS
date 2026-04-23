package eventengine

import (
	"context"

	"github.com/insmtx/SingerOS/backend/interaction"
	"github.com/ygpkg/yg-go/logs"
)

func (e *EventEngine) registerDefaultHandlers() {
	e.handlers[interaction.TopicGithubIssueComment] = e.handleIssueComment
	e.handlers[interaction.TopicGithubPullRequest] = e.handlePullRequest
	e.handlers[interaction.TopicGithubPush] = e.handlePush
}

func (e *EventEngine) Start(ctx context.Context) error {
	for topic, handler := range e.handlers {
		go func(t string, h EventHandler) {
			logs.InfoContextf(ctx, "Starting subscription for topic: %s", t)
			err := e.subscriber.Subscribe(ctx, t, func(event any) {
				interactionEvent, ok := event.(*interaction.Event)
				if !ok {
					logs.ErrorContextf(ctx, "Invalid event type received")
					return
				}

				logs.DebugContextf(ctx, "Received event on topic %s: %+v", t, interactionEvent)

				if err := h(ctx, interactionEvent); err != nil {
					logs.ErrorContextf(ctx, "Error handling event on topic %s: %v", t, err)
				}
			})

			if err != nil {
				logs.ErrorContextf(ctx, "Failed to subscribe to topic %s: %v", t, err)
			}
		}(topic, handler)
	}

	return nil
}

func (e *EventEngine) RegisterHandler(topic string, handler EventHandler) {
	e.handlers[topic] = handler
}

func (e *EventEngine) GetHandler(topic string) (EventHandler, error) {
	handler, exists := e.handlers[topic]
	if !exists {
		return nil, nil
	}
	return handler, nil
}

func (e *EventEngine) handleIssueComment(ctx context.Context, event *interaction.Event) error {
	logs.InfoContextf(ctx, "Processing GitHub issue comment event: %+v", event)
	return nil
}

func (e *EventEngine) handlePullRequest(ctx context.Context, event *interaction.Event) error {
	logs.InfoContextf(ctx, "Processing GitHub pull request event: %+v", event)
	return nil
}

func (e *EventEngine) handlePush(ctx context.Context, event *interaction.Event) error {
	logs.InfoContextf(ctx, "Processing GitHub push event: %+v", event)
	return nil
}
