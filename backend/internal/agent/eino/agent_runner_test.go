package eino

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	einomodel "github.com/cloudwego/eino/components/model"
	einoschema "github.com/cloudwego/eino/schema"
	auth "github.com/insmtx/SingerOS/backend/auth"
	"github.com/insmtx/SingerOS/backend/config"
	githubprovider "github.com/insmtx/SingerOS/backend/providers/github"
	runtimeprompt "github.com/insmtx/SingerOS/backend/internal/agent/prompt"
	"github.com/insmtx/SingerOS/backend/toolruntime"
	"github.com/insmtx/SingerOS/backend/tools"
	githubtools "github.com/insmtx/SingerOS/backend/tools/github"
)

func TestAgentRunnerGenerate(t *testing.T) {
	registry := tools.NewRegistry()
	if err := registry.Register(&mockTool{
		info: &tools.ToolInfo{
			Name:        "github.account.get_current_user",
			Description: "Read current GitHub account",
			ReadOnly:    true,
			InputSchema: &tools.Schema{
				Type: "object",
			},
		},
	}); err != nil {
		t.Fatalf("register mock tool: %v", err)
	}

	model := &fakeToolCallingModel{}
	adapter := NewToolAdapter(registry, toolruntime.New(registry, nil))
	runner, err := NewAgentRunner(context.Background(), &AgentRunnerConfig{
		Model:        model,
		ToolAdapter:  adapter,
		Binding:      ToolBinding{UserID: "u1"},
		SystemPrompt: "You are SingerOS.",
		Skills: &runtimeprompt.SkillsContext{
			SummarySection: "Available skills:\n- github-pr-review: Review pull requests.",
			AlwaysSections: []string{"## Skill: github-pr-review\nAlways inspect changed files first."},
		},
		Tools: &runtimeprompt.ToolsContext{
			SummarySection: "Available tools:\n- github.account.get_current_user: Read current GitHub account",
		},
	})
	if err != nil {
		t.Fatalf("new agent runner: %v", err)
	}

	message, err := runner.Generate(context.Background(), "who am I?")
	if err != nil {
		t.Fatalf("generate response: %v", err)
	}
	if message == nil {
		t.Fatalf("expected non-nil message")
	}
	if !strings.Contains(message.Content, "github.account.get_current_user") {
		t.Fatalf("unexpected final content: %s", message.Content)
	}
	if model.state == nil || len(model.state.calls) == 0 {
		t.Fatalf("expected model calls to be recorded")
	}
	foundSystemPrompt := false
	for _, call := range model.state.calls {
		if len(call) == 0 || call[0].Role != einoschema.System {
			continue
		}
		if strings.Contains(call[0].Content, "Available skills:") && strings.Contains(call[0].Content, "Available tools:") {
			foundSystemPrompt = true
			break
		}
	}
	if !foundSystemPrompt {
		t.Fatalf("expected system prompt with skills/tools summary to be injected")
	}
}

func TestAgentRunnerGenerateWithRealToolRuntime(t *testing.T) {
	store := auth.NewInMemoryStore()
	resolver := auth.NewAccountResolver(store)
	authService := auth.NewService(store, resolver)

	now := time.Now().UTC()
	account := &auth.AuthorizedAccount{
		ID:                "github:u1:1001",
		UserID:            "u1",
		Provider:          auth.ProviderGitHub,
		OwnerType:         auth.AccountOwnerTypeUser,
		AccountType:       auth.AccountTypeUserOAuth,
		ExternalAccountID: "1001",
		DisplayName:       "octocat",
		Scopes:            []string{"read:user"},
		Status:            auth.AccountStatusActive,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	credential := &auth.AccountCredential{
		AccountID:   account.ID,
		GrantType:   auth.GrantTypeOAuth2,
		AccessToken: "test-token",
	}
	if err := store.UpsertAuthorizedAccount(context.Background(), account, credential); err != nil {
		t.Fatalf("upsert account: %v", err)
	}
	if err := store.SetDefaultAccount(context.Background(), &auth.UserProviderBinding{
		UserID:    "u1",
		Provider:  auth.ProviderGitHub,
		AccountID: account.ID,
		IsDefault: true,
		Priority:  100,
	}); err != nil {
		t.Fatalf("set default account: %v", err)
	}

	responseBody, err := json.Marshal(map[string]interface{}{
		"id":    1001,
		"login": "octocat",
	})
	if err != nil {
		t.Fatalf("marshal response body: %v", err)
	}

	githubFactory := githubprovider.NewClientFactoryWithHTTPClient(config.GithubAppConfig{
		BaseURL: "https://api.github.test/",
	}, authService, &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(string(responseBody))),
				Request:    req,
			}, nil
		}),
	})

	registry := tools.NewRegistry()
	if err := registry.Register(githubtools.NewAccountInfoTool(nil)); err != nil {
		t.Fatalf("register github tool: %v", err)
	}

	model := &fakeToolCallingModel{}
	adapter := NewToolAdapter(registry, toolruntime.New(registry, githubFactory))
	runner, err := NewAgentRunner(context.Background(), &AgentRunnerConfig{
		Model:       model,
		ToolAdapter: adapter,
		Binding:     ToolBinding{UserID: "u1"},
		Skills:      &runtimeprompt.SkillsContext{},
		Tools:       runtimeprompt.BuildToolsContext(registry),
	})
	if err != nil {
		t.Fatalf("new agent runner: %v", err)
	}

	message, err := runner.Generate(context.Background(), "show my github account")
	if err != nil {
		t.Fatalf("generate response: %v", err)
	}
	if !strings.Contains(message.Content, "octocat") {
		t.Fatalf("expected real tool output in final content: %s", message.Content)
	}
}

type fakeToolCallingModel struct {
	state      *fakeToolCallingModelState
	boundTools []*einoschema.ToolInfo
}

var _ einomodel.ToolCallingChatModel = (*fakeToolCallingModel)(nil)

type fakeToolCallingModelState struct {
	calls [][]*einoschema.Message
}

func (m *fakeToolCallingModel) Generate(ctx context.Context, input []*einoschema.Message, opts ...einomodel.Option) (*einoschema.Message, error) {
	if m.state == nil {
		m.state = &fakeToolCallingModelState{}
	}
	copied := make([]*einoschema.Message, len(input))
	copy(copied, input)
	m.state.calls = append(m.state.calls, copied)

	last := input[len(input)-1]
	if last.Role == einoschema.Tool {
		return einoschema.AssistantMessage(fmt.Sprintf("final answer: %s", last.Content), nil), nil
	}

	toolName := "github.account.get_current_user"
	if len(m.boundTools) > 0 && m.boundTools[0] != nil && m.boundTools[0].Name != "" {
		toolName = m.boundTools[0].Name
	}

	return einoschema.AssistantMessage("", []einoschema.ToolCall{
		{
			ID:   "call_1",
			Type: "function",
			Function: einoschema.FunctionCall{
				Name:      toolName,
				Arguments: `{}`,
			},
		},
	}), nil
}

func (m *fakeToolCallingModel) Stream(ctx context.Context, input []*einoschema.Message, opts ...einomodel.Option) (*einoschema.StreamReader[*einoschema.Message], error) {
	return nil, fmt.Errorf("stream not implemented in test model")
}

func (m *fakeToolCallingModel) WithTools(tools []*einoschema.ToolInfo) (einomodel.ToolCallingChatModel, error) {
	state := m.state
	if state == nil {
		state = &fakeToolCallingModelState{}
		m.state = state
	}
	cloned := &fakeToolCallingModel{
		state:      state,
		boundTools: tools,
	}
	return cloned, nil
}

type mockTool struct {
	info *tools.ToolInfo
}

func (m *mockTool) Info() *tools.ToolInfo {
	return m.info
}

func (m *mockTool) Validate(input map[string]interface{}) error {
	return nil
}

func (m *mockTool) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	return map[string]interface{}{
		"tool": m.info.Name,
	}, nil
}
