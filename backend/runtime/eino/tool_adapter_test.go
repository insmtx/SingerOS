package eino

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	auth "github.com/insmtx/SingerOS/backend/auth"
	"github.com/insmtx/SingerOS/backend/config"
	githubprovider "github.com/insmtx/SingerOS/backend/providers/github"
	"github.com/insmtx/SingerOS/backend/tools"
	githubtools "github.com/insmtx/SingerOS/backend/tools/github"
)

func TestToolAdapterDefinitionsAndInvoke(t *testing.T) {
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
		"name":  "The Octocat",
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
				Header: http.Header{
					"Content-Type": []string{"application/json"},
				},
				Body:    io.NopCloser(strings.NewReader(string(responseBody))),
				Request: req,
			}, nil
		}),
	})

	registry := tools.NewRegistry()
	if err := registry.Register(githubtools.NewAccountInfoTool(githubFactory)); err != nil {
		t.Fatalf("register tool: %v", err)
	}

	adapter := NewToolAdapter(registry)

	definitions := adapter.Definitions()
	if len(definitions) != 1 {
		t.Fatalf("expected 1 tool definition, got %d", len(definitions))
	}
	if definitions[0].Name != githubtools.ToolNameGetCurrentUser {
		t.Fatalf("unexpected tool definition: %+v", definitions[0])
	}

	einoTools, err := adapter.EinoTools(ToolBinding{ToolContext: tools.ToolContext{UserID: "u1"}})
	if err != nil {
		t.Fatalf("build eino tools: %v", err)
	}
	if len(einoTools) != 1 {
		t.Fatalf("expected 1 eino tool, got %d", len(einoTools))
	}
	einoToolInfo, err := einoTools[0].Info(context.Background())
	if err != nil {
		t.Fatalf("get eino tool info: %v", err)
	}
	if einoToolInfo.Name != githubtools.ToolNameGetCurrentUser {
		t.Fatalf("unexpected eino tool name: %s", einoToolInfo.Name)
	}
	if einoToolInfo.ParamsOneOf == nil {
		t.Fatalf("expected eino tool params to be populated")
	}

	result, err := adapter.Invoke(context.Background(), &ToolCallRequest{
		Name:        githubtools.ToolNameGetCurrentUser,
		ToolContext: tools.ToolContext{UserID: "u1"},
	})
	if err != nil {
		t.Fatalf("invoke tool adapter: %v", err)
	}

	output := decodeToolOutput(t, result.Output)
	authorizedAccount, ok := output["authorized_account"].(map[string]interface{})
	if !ok {
		t.Fatalf("authorized_account output missing")
	}
	if authorizedAccount["id"] != account.ID {
		t.Fatalf("expected resolved account %s, got %v", account.ID, authorizedAccount["id"])
	}
	githubUser, ok := output["github_user"].(map[string]interface{})
	if !ok {
		t.Fatalf("github_user output missing")
	}
	if githubUser["login"] != "octocat" {
		t.Fatalf("expected login octocat, got %v", githubUser["login"])
	}

	invokableTool, ok := einoTools[0].(interface {
		InvokableRun(ctx context.Context, argumentsInJSON string, opts ...interface{}) (string, error)
	})
	if ok {
		_ = invokableTool
	}
}

func TestToolAdapterEinoToolsAllowedTools(t *testing.T) {
	registry := tools.NewRegistry()
	if err := registry.Register(&mockTool{
		BaseTool: tools.NewBaseTool("node_shell", "Execute shell command", tools.Schema{}),
	}); err != nil {
		t.Fatalf("register node shell tool: %v", err)
	}
	if err := registry.Register(&mockTool{
		BaseTool: tools.NewBaseTool("node_file_read", "Read file", tools.Schema{}),
	}); err != nil {
		t.Fatalf("register node file read tool: %v", err)
	}

	adapter := NewToolAdapter(registry)
	einoTools, err := adapter.EinoTools(ToolBinding{
		AllowedTools: []string{"node_file_read"},
	})
	if err != nil {
		t.Fatalf("build filtered eino tools: %v", err)
	}
	if len(einoTools) != 1 {
		t.Fatalf("expected 1 filtered tool, got %d", len(einoTools))
	}

	info, err := einoTools[0].Info(context.Background())
	if err != nil {
		t.Fatalf("get filtered tool info: %v", err)
	}
	if info.Name != "node_file_read" {
		t.Fatalf("unexpected filtered tool: %s", info.Name)
	}
}

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func decodeToolOutput(t *testing.T, output string) map[string]interface{} {
	t.Helper()

	var decoded map[string]interface{}
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("decode tool output: %v\n%s", err, output)
	}
	return decoded
}
