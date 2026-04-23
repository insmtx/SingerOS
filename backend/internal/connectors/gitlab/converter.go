package gitlab

import (
	"github.com/insmtx/SingerOS/backend/pkg/event"
)

type EventConverter struct{}

func NewEventConverter() *EventConverter {
	return &EventConverter{}
}

func (c *EventConverter) Convert(eventType string, payload []byte) (*event.Event, error) {
	return nil, nil
}
