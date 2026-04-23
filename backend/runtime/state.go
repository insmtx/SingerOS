package runtime

import (
	runtimeeino "github.com/insmtx/SingerOS/backend/runtime/eino"
	runtimeevents "github.com/insmtx/SingerOS/backend/runtime/events"
)

type runState struct {
	req          *RequestContext
	emitter      *runtimeevents.Emitter
	userInput    string
	systemPrompt string
	toolBinding  runtimeeino.ToolBinding
	maxStep      int
}
