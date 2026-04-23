package gitlab

import (
	"github.com/insmtx/SingerOS/backend/interaction"
)

type EventConverter struct{}

func NewEventConverter() *EventConverter {
	return &EventConverter{}
}

func (c *EventConverter) Convert(eventType string, payload []byte) (*interaction.Event, error) {
	return nil, nil
}
