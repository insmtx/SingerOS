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
	runtimeevents "github.com/insmtx/SingerOS/backend/runtime/events"
	"github.com/insmtx/SingerOS/backend/tools"
	githubtools "github.com/insmtx/SingerOS/backend/tools/github"
)

func TestFlowGenerate(t *testing.T) {
	registry := tools.NewRegistry()
	if err := registry.Register(&mockTool{
		BaseTool: tools.NewBaseTool(
			"test.account.get_current_user",
			"Read current test account",
			tools.Schema{
				Type: "object",
			},
		),
	}); err != nil {
		t.Fatalf("register mock tool: %v", err)
	}

	model := &fakeToolCallingModel{}
	adapter := NewToolAdapter(registry)
	flow, err := NewFlow(context.Background(), &FlowConfig{
		Model:        model,
		ToolAdapter:  adapter,
		Binding:      ToolBinding{ToolContext: tools.ToolContext{UserID: "u1"}},
		SystemPrompt: "You are SingerOS.\n\nAvailable skills:\n- github-pr-review: Review pull requests.",
	})
	if err != nil {
		t.Fatalf("new flow: %v", err)
	}

	message, err := flow.Generate(context.Background(), "who am I?")
	if err != nil {
		t.Fatalf("generate response: %v", err)
	}
	if message == nil {
		t.Fatalf("expected non-nil message")
	}
	if !strings.Contains(message.Content, "test.account.get_current_user") {
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
		if strings.Contains(call[0].Content, "Available tools:") {
			t.Fatalf("tool summary should not be injected into system prompt: %s", call[0].Content)
		}
		if strings.Contains(call[0].Content, "You are SingerOS.") && strings.Contains(call[0].Content, "Available skills:") {
			foundSystemPrompt = true
			break
		}
	}
	if !foundSystemPrompt {
		t.Fatalf("expected system prompt with skills summary to be injected")
	}
}

func TestFlowGenerateWithRealToolContext(t *testing.T) {
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
	if err := registry.Register(githubtools.NewAccountInfoTool(githubFactory)); err != nil {
		t.Fatalf("register github tool: %v", err)
	}

	model := &fakeToolCallingModel{}
	adapter := NewToolAdapter(registry)
	flow, err := NewFlow(context.Background(), &FlowConfig{
		Model:       model,
		ToolAdapter: adapter,
		Binding:     ToolBinding{ToolContext: tools.ToolContext{UserID: "u1"}},
	})
	if err != nil {
		t.Fatalf("new flow: %v", err)
	}

	message, err := flow.Generate(context.Background(), "show my github account")
	if err != nil {
		t.Fatalf("generate response: %v", err)
	}
	if !strings.Contains(message.Content, "octocat") {
		t.Fatalf("expected real tool output in final content: %s", message.Content)
	}
}

func TestFlowStreamEmitsMessageEvents(t *testing.T) {
	registry := tools.NewRegistry()
	flow, err := NewFlow(context.Background(), &FlowConfig{
		Model:       &streamingTextModel{},
		ToolAdapter: NewToolAdapter(registry),
		Binding:     ToolBinding{ToolContext: tools.ToolContext{UserID: "u1"}},
	})
	if err != nil {
		t.Fatalf("new flow: %v", err)
	}

	var emitted []*runtimeevents.RunEvent
	emitter := runtimeevents.NewEmitter("run_stream", "trace_stream", runtimeevents.SinkFunc(func(ctx context.Context, event *runtimeevents.RunEvent) error {
		emitted = append(emitted, event)
		return nil
	}))
	message, err := flow.Stream(context.Background(), "say hello", emitter)
	if err != nil {
		t.Fatalf("stream response: %v", err)
	}
	if message == nil || strings.TrimSpace(message.Content) != "hello world" {
		t.Fatalf("unexpected streamed message: %+v", message)
	}

	var deltaCount int
	for _, event := range emitted {
		switch event.Type {
		case runtimeevents.RunEventMessageDelta:
			deltaCount++
		}
	}
	if deltaCount != 2 {
		t.Fatalf("expected two delta events, got %d: %+v", deltaCount, emitted)
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

	toolName := "test.account.get_current_user"
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

type streamingTextModel struct{}

var _ einomodel.ToolCallingChatModel = (*streamingTextModel)(nil)

func (m *streamingTextModel) Generate(ctx context.Context, input []*einoschema.Message, opts ...einomodel.Option) (*einoschema.Message, error) {
	return einoschema.AssistantMessage("hello world", nil), nil
}

func (m *streamingTextModel) Stream(ctx context.Context, input []*einoschema.Message, opts ...einomodel.Option) (*einoschema.StreamReader[*einoschema.Message], error) {
	return einoschema.StreamReaderFromArray([]*einoschema.Message{
		einoschema.AssistantMessage("hello ", nil),
		einoschema.AssistantMessage("world", nil),
	}), nil
}

func (m *streamingTextModel) WithTools(tools []*einoschema.ToolInfo) (einomodel.ToolCallingChatModel, error) {
	return m, nil
}

type mockTool struct {
	tools.BaseTool
}

func (m *mockTool) Validate(input map[string]interface{}) error {
	return nil
}

func (m *mockTool) Execute(ctx context.Context, input map[string]interface{}) (string, error) {
	return tools.JSONString(map[string]interface{}{
		"tool": m.Name(),
	})
}
