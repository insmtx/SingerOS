package toolruntime

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
	"github.com/insmtx/SingerOS/backend/tools"
	githubtools "github.com/insmtx/SingerOS/backend/tools/github"
)

func TestRuntimeExecuteGitHubTool(t *testing.T) {
	responseBody, err := json.Marshal(map[string]interface{}{
		"id":    1001,
		"login": "octocat",
		"name":  "The Octocat",
	})
	if err != nil {
		t.Fatalf("marshal mock response: %v", err)
	}

	store := auth.NewInMemoryStore()
	resolver := auth.NewAccountResolver(store)
	authService := auth.NewService(store, resolver)
	factory := githubprovider.NewClientFactoryWithHTTPClient(config.GithubAppConfig{
		BaseURL: "https://api.github.test/",
	}, authService, &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.URL.String() != "https://api.github.test/user" {
				return &http.Response{
					StatusCode: http.StatusNotFound,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader(`{"message":"not found"}`)),
					Request:    req,
				}, nil
			}

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

	registry := tools.NewRegistry()
	if err := registry.Register(githubtools.NewAccountInfoTool(nil)); err != nil {
		t.Fatalf("register tool: %v", err)
	}

	runtime := New(registry, factory)
	result, err := runtime.Execute(context.Background(), &ExecuteRequest{
		ToolName: githubtools.ToolNameGetCurrentUser,
		UserID:   "u1",
	})
	if err != nil {
		t.Fatalf("execute tool runtime: %v", err)
	}

	if result.ResolvedAccount == nil || result.ResolvedAccount.ID != account.ID {
		t.Fatalf("unexpected resolved account: %+v", result.ResolvedAccount)
	}
	if result.ResolvedBy != "subject_default" {
		t.Fatalf("expected resolved by subject_default, got %s", result.ResolvedBy)
	}

	githubUser, ok := result.Output["github_user"].(map[string]interface{})
	if !ok {
		t.Fatalf("github_user result missing")
	}
	if githubUser["login"] != "octocat" {
		t.Fatalf("expected login octocat, got %v", githubUser["login"])
	}
}

func TestRuntimeExecuteGitHubToolWithSelector(t *testing.T) {
	responseBody, err := json.Marshal(map[string]interface{}{
		"id":    1001,
		"login": "octocat",
		"name":  "The Octocat",
	})
	if err != nil {
		t.Fatalf("marshal mock response: %v", err)
	}

	store := auth.NewInMemoryStore()
	resolver := auth.NewAccountResolver(store)
	authService := auth.NewService(store, resolver)
	factory := githubprovider.NewClientFactoryWithHTTPClient(config.GithubAppConfig{
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

	now := time.Now().UTC()
	account := &auth.AuthorizedAccount{
		ID:                "github:u1:1001",
		UserID:            "u1",
		Provider:          auth.ProviderGitHub,
		OwnerType:         auth.AccountOwnerTypeUser,
		AccountType:       auth.AccountTypeUserOAuth,
		ExternalAccountID: "1001",
		DisplayName:       "octocat",
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

	registry := tools.NewRegistry()
	if err := registry.Register(githubtools.NewAccountInfoTool(nil)); err != nil {
		t.Fatalf("register tool: %v", err)
	}

	runtime := New(registry, factory)
	result, err := runtime.Execute(context.Background(), &ExecuteRequest{
		ToolName: githubtools.ToolNameGetCurrentUser,
		Selector: &auth.AuthSelector{
			Provider:    auth.ProviderGitHub,
			SubjectType: auth.SubjectTypeUser,
			SubjectID:   "u1",
		},
	})
	if err != nil {
		t.Fatalf("execute tool runtime with selector: %v", err)
	}
	if result.ResolvedAccount == nil || result.ResolvedAccount.ID != account.ID {
		t.Fatalf("unexpected resolved account: %+v", result.ResolvedAccount)
	}
	if result.ResolvedBy != "subject_default" {
		t.Fatalf("expected resolved by subject_default, got %s", result.ResolvedBy)
	}
}

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
