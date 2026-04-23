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
	githubprovider "github.com/insmtx/SingerOS/backend/pkg/providers/github"
	"github.com/insmtx/SingerOS/backend/toolruntime"
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
	if err := registry.Register(githubtools.NewAccountInfoTool(nil)); err != nil {
		t.Fatalf("register tool: %v", err)
	}

	adapter := NewToolAdapter(registry, toolruntime.New(registry, githubFactory))

	definitions := adapter.Definitions()
	if len(definitions) != 1 {
		t.Fatalf("expected 1 tool definition, got %d", len(definitions))
	}
	if definitions[0].Name != githubtools.ToolNameGetCurrentUser {
		t.Fatalf("unexpected tool definition: %+v", definitions[0])
	}
	if !definitions[0].ReadOnly {
		t.Fatalf("expected tool definition to be read-only")
	}

	einoTools, err := adapter.EinoTools(ToolBinding{UserID: "u1"})
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
		Name:   githubtools.ToolNameGetCurrentUser,
		UserID: "u1",
	})
	if err != nil {
		t.Fatalf("invoke tool adapter: %v", err)
	}
	if result.ResolvedAccountID != account.ID {
		t.Fatalf("expected resolved account %s, got %s", account.ID, result.ResolvedAccountID)
	}

	githubUser, ok := result.Output["github_user"].(map[string]interface{})
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

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
