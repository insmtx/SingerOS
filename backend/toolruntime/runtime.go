package toolruntime

import (
	"context"
	"fmt"

	auth "github.com/insmtx/SingerOS/backend/auth"
	githubprovider "github.com/insmtx/SingerOS/backend/pkg/providers/github"
	"github.com/insmtx/SingerOS/backend/tools"
)

// ExecuteRequest describes a single tool execution inside the runtime.
type ExecuteRequest struct {
	ToolName  string
	Selector  *auth.AuthSelector
	UserID    string
	AccountID string
	Input     map[string]interface{}
}

// ExecuteResult contains the tool output together with resolved runtime metadata.
type ExecuteResult struct {
	ToolName        string
	Output          map[string]interface{}
	ResolvedAccount *auth.AuthorizedAccount
	ResolvedBy      string
}

// Runtime is the minimal execution pipeline for SingerOS tools.
type Runtime struct {
	registry            *tools.Registry
	githubClientFactory *githubprovider.ClientFactory
}

// New creates a tool runtime backed by the shared registry and provider factories.
func New(registry *tools.Registry, githubClientFactory *githubprovider.ClientFactory) *Runtime {
	return &Runtime{
		registry:            registry,
		githubClientFactory: githubClientFactory,
	}
}

// Execute runs a tool through the unified runtime pipeline.
func (r *Runtime) Execute(ctx context.Context, req *ExecuteRequest) (*ExecuteResult, error) {
	if req == nil {
		return nil, fmt.Errorf("execute request is required")
	}
	if req.ToolName == "" {
		return nil, fmt.Errorf("tool name is required")
	}
	if r == nil || r.registry == nil {
		return nil, fmt.Errorf("tool registry is required")
	}

	tool, err := r.registry.Get(req.ToolName)
	if err != nil {
		return nil, err
	}

	input := cloneInput(req.Input)
	selector := mergedSelector(req)
	if selector.SubjectType == auth.SubjectTypeUser && selector.SubjectID != "" {
		input["user_id"] = selector.SubjectID
	} else if req.UserID != "" {
		input["user_id"] = req.UserID
	}
	if selector.ExplicitProfileID != "" {
		input["account_id"] = selector.ExplicitProfileID
	} else if req.AccountID != "" {
		input["account_id"] = req.AccountID
	}

	if err := tool.Validate(input); err != nil {
		return nil, fmt.Errorf("validate tool %s input: %w", req.ToolName, err)
	}

	execCtx, err := r.buildExecutionContext(ctx, tool.Info(), req)
	if err != nil {
		return nil, err
	}

	var output map[string]interface{}
	if runtimeTool, ok := tool.(tools.RuntimeTool); ok {
		output, err = runtimeTool.ExecuteWithContext(ctx, execCtx, input)
	} else {
		output, err = tool.Execute(ctx, input)
	}
	if err != nil {
		return nil, err
	}

	return &ExecuteResult{
		ToolName:        req.ToolName,
		Output:          output,
		ResolvedAccount: execCtx.ResolvedAccount,
		ResolvedBy:      execCtx.ResolvedBy,
	}, nil
}

func (r *Runtime) buildExecutionContext(ctx context.Context, info *tools.ToolInfo, req *ExecuteRequest) (*tools.ExecutionContext, error) {
	selector := mergedSelector(req)
	execCtx := &tools.ExecutionContext{
		UserID:    selector.SubjectID,
		AccountID: selector.ExplicitProfileID,
		Resources: make(map[string]interface{}),
		Selector:  selector,
	}

	if info == nil {
		return execCtx, nil
	}

	execCtx.Provider = info.Provider
	if execCtx.Selector != nil && execCtx.Selector.Provider == "" {
		execCtx.Selector.Provider = info.Provider
	}
	switch info.Provider {
	case auth.ProviderGitHub:
		if r.githubClientFactory == nil {
			return nil, fmt.Errorf("github client factory is required for tool %s", req.ToolName)
		}

		resolved, err := r.githubClientFactory.ResolveClient(ctx, &githubprovider.ResolveClientRequest{
			Selector:  selector,
			UserID:    selector.SubjectID,
			AccountID: selector.ExplicitProfileID,
		})
		if err != nil {
			return nil, err
		}

		execCtx.ResolvedAccount = resolved.Account
		execCtx.ResolvedBy = resolved.ResolvedBy
		execCtx.Resources[tools.ResourceGitHubResolvedClient] = resolved
	}

	return execCtx, nil
}

func mergedSelector(req *ExecuteRequest) *auth.AuthSelector {
	selector := &auth.AuthSelector{}
	if req != nil && req.Selector != nil {
		selector = cloneSelector(req.Selector)
	}
	if req == nil {
		return selector
	}
	if selector.ExplicitProfileID == "" {
		selector.ExplicitProfileID = req.AccountID
	}
	if selector.SubjectID == "" && req.UserID != "" {
		selector.SubjectID = req.UserID
	}
	if selector.SubjectType == "" && selector.SubjectID != "" {
		selector.SubjectType = auth.SubjectTypeUser
	}
	return selector
}

func cloneSelector(selector *auth.AuthSelector) *auth.AuthSelector {
	if selector == nil {
		return &auth.AuthSelector{}
	}
	cloned := *selector
	if selector.ExternalRefs != nil {
		cloned.ExternalRefs = make(map[string]string, len(selector.ExternalRefs))
		for key, value := range selector.ExternalRefs {
			cloned.ExternalRefs[key] = value
		}
	}
	return &cloned
}

func cloneInput(input map[string]interface{}) map[string]interface{} {
	if input == nil {
		return make(map[string]interface{})
	}

	cloned := make(map[string]interface{}, len(input))
	for key, value := range input {
		cloned[key] = value
	}

	return cloned
}
