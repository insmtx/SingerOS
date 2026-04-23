package githubtools

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
)

func TestAccountInfoToolExecute(t *testing.T) {
	responseBody, err := json.Marshal(map[string]interface{}{
		"id":           1001,
		"login":        "octocat",
		"name":         "The Octocat",
		"email":        "octocat@github.com",
		"avatar_url":   "https://github.com/images/error/octocat_happy.gif",
		"html_url":     "https://github.com/octocat",
		"company":      "@github",
		"location":     "San Francisco",
		"bio":          "Mona the Octocat",
		"public_repos": 8,
		"followers":    20,
		"following":    0,
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
	tool := NewAccountInfoTool(factory)

	now := time.Now().UTC()
	account := &auth.AuthorizedAccount{
		ID:                "github:u1:1001",
		UserID:            "u1",
		Provider:          auth.ProviderGitHub,
		OwnerType:         auth.AccountOwnerTypeUser,
		AccountType:       auth.AccountTypeUserOAuth,
		ExternalAccountID: "1001",
		DisplayName:       "octocat",
		Scopes:            []string{"read:user", "user:email"},
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

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"user_id": "u1",
	})
	if err != nil {
		t.Fatalf("execute tool: %v", err)
	}

	githubUser, ok := result["github_user"].(map[string]interface{})
	if !ok {
		t.Fatalf("github_user result missing")
	}
	if githubUser["login"] != "octocat" {
		t.Fatalf("expected login octocat, got %v", githubUser["login"])
	}

	authorizedAccount, ok := result["authorized_account"].(map[string]interface{})
	if !ok {
		t.Fatalf("authorized_account result missing")
	}
	if authorizedAccount["id"] != account.ID {
		t.Fatalf("expected account id %s, got %v", account.ID, authorizedAccount["id"])
	}
}

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
