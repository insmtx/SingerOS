package agent

import (
	"context"
	"encoding/base64"
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
	"github.com/insmtx/SingerOS/backend/interaction"
	githubprovider "github.com/insmtx/SingerOS/backend/providers/github"
	runtimeeino "github.com/insmtx/SingerOS/backend/internal/agent/eino"
	runtimeprompt "github.com/insmtx/SingerOS/backend/internal/agent/prompt"
	"github.com/insmtx/SingerOS/backend/toolruntime"
	"github.com/insmtx/SingerOS/backend/tools"
	githubtools "github.com/insmtx/SingerOS/backend/tools/github"
)

func TestEinoRunnerSystemPromptForPullRequestEvent(t *testing.T) {
	runner := &EinoRunner{
		systemPrompt: "base prompt",
	}

	prompt := runner.systemPromptForEvent(&interaction.Event{
		EventType: "pull_request",
	})

	if prompt == "base prompt" {
		t.Fatalf("expected pull request prompt extension")
	}
	if !containsSubstring(prompt, "base prompt") {
		t.Fatalf("expected base prompt to be preserved: %s", prompt)
	}
	if !containsSubstring(prompt, "use GitHub tools to inspect metadata, changed files") {
		t.Fatalf("expected pull request guidance, got: %s", prompt)
	}
}

func TestEinoRunnerSystemPromptForGenericEvent(t *testing.T) {
	runner := &EinoRunner{
		systemPrompt: "base prompt",
	}

	prompt := runner.systemPromptForEvent(&interaction.Event{
		EventType: "issue_comment",
	})

	if prompt != "base prompt" {
		t.Fatalf("expected base prompt for generic events, got: %s", prompt)
	}
}

func TestFormatLLMResultForLog(t *testing.T) {
	if got := formatLLMResultForLog(nil); got != "<nil>" {
		t.Fatalf("expected <nil>, got %s", got)
	}

	message := einoschema.AssistantMessage("final answer", nil)
	if got := formatLLMResultForLog(message); !containsSubstring(got, "final answer") {
		t.Fatalf("expected formatted message content, got %s", got)
	}

	longText := strings.Repeat("a", 2100)
	longMessage := einoschema.AssistantMessage(longText, nil)
	if got := formatLLMResultForLog(longMessage); !containsSubstring(got, "...(truncated)") {
		t.Fatalf("expected truncated output, got %s", got)
	}
}

func TestEinoRunnerSystemPromptForPushEvent(t *testing.T) {
	runner := &EinoRunner{
		systemPrompt: "base prompt",
	}

	prompt := runner.systemPromptForEvent(&interaction.Event{
		EventType: "push",
	})

	if !containsSubstring(prompt, "base prompt") {
		t.Fatalf("expected base prompt to be preserved: %s", prompt)
	}
	if !containsSubstring(prompt, "For GitHub push events") {
		t.Fatalf("expected push event guidance, got: %s", prompt)
	}
}

func TestAuthSelectorFromEventPrefersSenderAndInstallationRefs(t *testing.T) {
	selector := authSelectorFromEvent(&interaction.Event{
		EventID:    "evt_1",
		Channel:    "github",
		EventType:  "pull_request",
		Actor:      "display-only-actor",
		Repository: "insmtx/SingerOS",
		Context: map[string]interface{}{
			"provider": "github",
			"action":   "synchronize",
		},
		Payload: map[string]interface{}{
			"installation": map[string]interface{}{
				"id": 99,
			},
			"sender": map[string]interface{}{
				"id":    1001,
				"login": "octocat",
			},
		},
	})

	if selector.Provider != auth.ProviderGitHub {
		t.Fatalf("expected github provider, got %s", selector.Provider)
	}
	if selector.ScopeType != auth.ScopeTypeEvent || selector.ScopeID != "evt_1" {
		t.Fatalf("unexpected selector scope: %+v", selector)
	}
	if selector.SubjectID != "octocat" {
		t.Fatalf("expected sender login as subject, got %s", selector.SubjectID)
	}
	if selector.ExternalRefs["github.installation_id"] != "99" {
		t.Fatalf("expected installation ref, got %+v", selector.ExternalRefs)
	}
	if selector.ExternalRefs["github.sender_id"] != "1001" {
		t.Fatalf("expected sender id ref, got %+v", selector.ExternalRefs)
	}
}

func TestEinoRunnerHandlePullRequestEventEndToEnd(t *testing.T) {
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
		Scopes:            []string{"repo"},
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

	var publishRequestBody string
	factory := githubprovider.NewClientFactoryWithHTTPClient(config.GithubAppConfig{
		BaseURL: "https://api.github.test/",
	}, authService, &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch {
			case req.Method == http.MethodGet && req.URL.String() == "https://api.github.test/repos/insmtx/SingerOS/pulls/12":
				return jsonHTTPResponse(req, `{
					"number": 12,
					"title": "Add Eino runtime",
					"body": "Introduce Eino runtime runner",
					"state": "open",
					"draft": false,
					"html_url": "https://github.com/insmtx/SingerOS/pull/12",
					"changed_files": 1,
					"commits": 1,
					"additions": 42,
					"deletions": 5,
					"user": {"login":"octocat","id":1001},
					"head": {"ref":"feature/eino","sha":"abc123"},
					"base": {"ref":"main","sha":"def456"}
				}`)
			case req.Method == http.MethodGet && strings.HasPrefix(req.URL.String(), "https://api.github.test/repos/insmtx/SingerOS/pulls/12/files"):
				return jsonHTTPResponse(req, `[
					{
						"filename":"backend/runtime/eino_runner.go",
						"status":"modified",
						"sha":"abc123",
						"patch":"@@ -1,3 +1,8 @@",
						"additions":5,
						"deletions":1,
						"changes":6
					}
				]`)
			case req.Method == http.MethodGet && strings.HasPrefix(req.URL.String(), "https://api.github.test/repos/insmtx/SingerOS/contents/backend/runtime/eino_runner.go"):
				content := base64.StdEncoding.EncodeToString([]byte("package runtime\n\nfunc example() {}\n"))
				return jsonHTTPResponse(req, `{
					"type":"file",
					"encoding":"base64",
					"size":35,
					"sha":"file123",
					"content":"`+content+`"
				}`)
			case req.Method == http.MethodPost && req.URL.String() == "https://api.github.test/repos/insmtx/SingerOS/pulls/12/reviews":
				body, err := io.ReadAll(req.Body)
				if err != nil {
					return nil, err
				}
				publishRequestBody = string(body)
				return jsonHTTPResponse(req, `{
					"id": 9001,
					"node_id": "PRR_kwDOTest",
					"state": "COMMENTED",
					"body": "I checked the runtime wiring and found no blocking issues.",
					"html_url": "https://github.com/insmtx/SingerOS/pull/12#pullrequestreview-9001",
					"commit_id": "abc123",
					"submitted_at": "2026-04-15T08:00:00Z",
					"user": {"login":"octocat","id":1001}
				}`)
			default:
				return &http.Response{
					StatusCode: http.StatusNotFound,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader(`{"message":"not found"}`)),
					Request:    req,
				}, nil
			}
		}),
	})

	registry := tools.NewRegistry()
	registerRuntimeTool(t, registry, githubtools.NewPullRequestMetadataTool(nil))
	registerRuntimeTool(t, registry, githubtools.NewPullRequestFilesTool(nil))
	registerRuntimeTool(t, registry, githubtools.NewRepositoryFileTool(nil))
	registerRuntimeTool(t, registry, githubtools.NewPullRequestReviewPublishTool(nil))

	model := &scriptedPRReviewModel{}
	runner := &EinoRunner{
		chatModel:   model,
		toolAdapter: runtimeeino.NewToolAdapter(registry, toolruntime.New(registry, factory)),
		skills: &runtimeprompt.SkillsContext{
			SummarySection: "Available skills:\n- github-pr-review: Review GitHub pull requests.",
			AlwaysSections: []string{"## Skill: github-pr-review\nDo not auto-approve. Use COMMENT unless there is concrete merge-blocking evidence."},
		},
		tools:        runtimeprompt.BuildToolsContext(registry),
		systemPrompt: defaultEinoSystemPrompt,
	}

	err := runner.HandleEvent(context.Background(), &interaction.Event{
		Channel:    "github",
		EventType:  "pull_request",
		Actor:      "u1",
		Repository: "insmtx/SingerOS",
		Context: map[string]interface{}{
			"provider": "github",
			"action":   "synchronize",
		},
		Payload: map[string]interface{}{
			"action": "synchronize",
			"pull_request": map[string]interface{}{
				"number": 12,
				"title":  "Add Eino runtime",
				"head": map[string]interface{}{
					"ref": "feature/eino",
				},
				"base": map[string]interface{}{
					"ref": "main",
				},
			},
			"installation": map[string]interface{}{
				"id": 99,
			},
		},
	})
	if err != nil {
		t.Fatalf("handle event: %v", err)
	}

	expectedTools := []string{
		githubtools.ToolNameGetPullRequestMetadata,
		githubtools.ToolNameGetPullRequestFiles,
		githubtools.ToolNameGetRepositoryFile,
		githubtools.ToolNamePublishPullRequestReview,
	}
	if len(model.state.requestedTools) != len(expectedTools) {
		t.Fatalf("unexpected tool call count: %+v", model.state.requestedTools)
	}
	for index, expected := range expectedTools {
		if model.state.requestedTools[index] != expected {
			t.Fatalf("unexpected tool sequence: %+v", model.state.requestedTools)
		}
	}
	if !containsSubstring(publishRequestBody, `"event":"COMMENT"`) {
		t.Fatalf("expected default COMMENT review event, got body: %s", publishRequestBody)
	}
	if !containsSubstring(publishRequestBody, `"body":"I checked the runtime wiring and found no blocking issues."`) {
		t.Fatalf("unexpected publish review body: %s", publishRequestBody)
	}
	if len(model.state.calls) == 0 || len(model.state.calls[0]) == 0 {
		t.Fatalf("expected model calls to be recorded")
	}
	systemPrompt := model.state.calls[0][0].Content
	if !containsSubstring(systemPrompt, "Do not auto-approve") {
		t.Fatalf("expected skill guidance in system prompt: %s", systemPrompt)
	}
	if !containsSubstring(systemPrompt, "Prefer COMMENT by default") {
		t.Fatalf("expected pull request prompt guidance in system prompt: %s", systemPrompt)
	}
	if !containsSubstring(model.state.calls[0][1].Content, `"installation":`) {
		t.Fatalf("expected raw event payload in user input: %s", model.state.calls[0][1].Content)
	}
}

type scriptedPRReviewModel struct {
	state      *scriptedPRReviewState
	boundTools []*einoschema.ToolInfo
}

type scriptedPRReviewState struct {
	calls          [][]*einoschema.Message
	requestedTools []string
}

var _ einomodel.ToolCallingChatModel = (*scriptedPRReviewModel)(nil)

func (m *scriptedPRReviewModel) Generate(ctx context.Context, input []*einoschema.Message, opts ...einomodel.Option) (*einoschema.Message, error) {
	if m.state == nil {
		m.state = &scriptedPRReviewState{}
	}
	copied := make([]*einoschema.Message, len(input))
	copy(copied, input)
	m.state.calls = append(m.state.calls, copied)

	toolResponses := 0
	for _, message := range input {
		if message.Role == einoschema.Tool {
			toolResponses++
		}
	}

	switch toolResponses {
	case 0:
		return m.toolCall(githubtools.ToolNameGetPullRequestMetadata, `{"repo":"insmtx/SingerOS","pr_number":12}`)
	case 1:
		return m.toolCall(githubtools.ToolNameGetPullRequestFiles, `{"repo":"insmtx/SingerOS","pr_number":12}`)
	case 2:
		return m.toolCall(githubtools.ToolNameGetRepositoryFile, `{"repo":"insmtx/SingerOS","path":"backend/runtime/eino_runner.go","ref":"feature/eino"}`)
	case 3:
		return m.toolCall(githubtools.ToolNamePublishPullRequestReview, `{"repo":"insmtx/SingerOS","pr_number":12,"body":"I checked the runtime wiring and found no blocking issues."}`)
	default:
		return einoschema.AssistantMessage("Review completed and published.", nil), nil
	}
}

func (m *scriptedPRReviewModel) Stream(ctx context.Context, input []*einoschema.Message, opts ...einomodel.Option) (*einoschema.StreamReader[*einoschema.Message], error) {
	return nil, fmt.Errorf("stream not implemented in scripted test model")
}

func (m *scriptedPRReviewModel) WithTools(tools []*einoschema.ToolInfo) (einomodel.ToolCallingChatModel, error) {
	state := m.state
	if state == nil {
		state = &scriptedPRReviewState{}
		m.state = state
	}
	return &scriptedPRReviewModel{
		state:      state,
		boundTools: tools,
	}, nil
}

func (m *scriptedPRReviewModel) toolCall(name string, arguments string) (*einoschema.Message, error) {
	m.state.requestedTools = append(m.state.requestedTools, name)
	return einoschema.AssistantMessage("", []einoschema.ToolCall{
		{
			ID:   fmt.Sprintf("call_%d", len(m.state.requestedTools)),
			Type: "function",
			Function: einoschema.FunctionCall{
				Name:      name,
				Arguments: arguments,
			},
		},
	}), nil
}

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func registerRuntimeTool(t *testing.T, registry *tools.Registry, tool tools.Tool) {
	t.Helper()
	if err := registry.Register(tool); err != nil {
		t.Fatalf("register tool %s: %v", tool.Info().Name, err)
	}
}

func jsonHTTPResponse(req *http.Request, body string) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header: http.Header{
			"Content-Type": []string{"application/json"},
		},
		Body:    io.NopCloser(strings.NewReader(body)),
		Request: req,
	}, nil
}

func containsSubstring(value string, sub string) bool {
	return len(sub) == 0 || (len(value) >= len(sub) && (value == sub || stringContains(value, sub)))
}

func stringContains(value string, sub string) bool {
	for idx := 0; idx+len(sub) <= len(value); idx++ {
		if value[idx:idx+len(sub)] == sub {
			return true
		}
	}
	return false
}
