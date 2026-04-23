package githubtools

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	auth "github.com/insmtx/SingerOS/backend/auth"
	"github.com/insmtx/SingerOS/backend/config"
	githubprovider "github.com/insmtx/SingerOS/backend/providers/github"
)

func TestGitHubPublishReviewToolExecute(t *testing.T) {
	store := auth.NewInMemoryStore()
	resolver := auth.NewAccountResolver(store)
	authService := auth.NewService(store, resolver)
	factory := githubprovider.NewClientFactoryWithHTTPClient(config.GithubAppConfig{
		BaseURL: "https://api.github.test/",
	}, authService, &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.Method == http.MethodPost && req.URL.String() == "https://api.github.test/repos/insmtx/SingerOS/pulls/12/reviews" {
				return jsonResponse(req, `{
					"id": 9001,
					"node_id": "PRR_kwDOTest",
					"state": "COMMENTED",
					"body": "This change is safe to merge.",
					"html_url": "https://github.com/insmtx/SingerOS/pull/12#pullrequestreview-9001",
					"commit_id": "abc123",
					"submitted_at": "2026-04-15T08:00:00Z",
					"user": {"login":"octocat","id":1001}
				}`)
			}

			return &http.Response{
				StatusCode: http.StatusNotFound,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(`{"message":"not found"}`)),
				Request:    req,
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

	tool := NewPullRequestReviewPublishTool(factory)
	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"user_id":   "u1",
		"repo":      "insmtx/SingerOS",
		"pr_number": 12,
		"body":      "This change is safe to merge.",
		"event":     "COMMENT",
	})
	if err != nil {
		t.Fatalf("execute publish review tool: %v", err)
	}

	output := decodeGitHubToolOutput(t, result)
	review, ok := output["review"].(map[string]interface{})
	if !ok {
		t.Fatalf("unexpected publish review output: %+v", output)
	}
	if review["id"] != float64(9001) {
		t.Fatalf("unexpected review id: %+v", review)
	}
	if review["state"] != "COMMENTED" {
		t.Fatalf("unexpected review state: %+v", review)
	}
	if output["event"] != "COMMENT" {
		t.Fatalf("unexpected event: %+v", output)
	}
}
