package agent

import "context"

type AgentRuntime interface {
	Run(ctx context.Context, req *RequestContext) (*RunResult, error)
}
