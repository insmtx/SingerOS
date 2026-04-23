package agent

import (
	einoadapter "github.com/insmtx/SingerOS/backend/internal/agent/eino"
	agentevents "github.com/insmtx/SingerOS/backend/internal/agent/events"
)

type runState struct {
	req          *RequestContext
	emitter      *agentevents.Emitter
	userInput    string
	systemPrompt string
	toolBinding  einoadapter.ToolBinding
	maxStep      int
}
