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
	"github.com/insmtx/SingerOS/backend/toolruntime"
	"github.com/insmtx/SingerOS/backend/tools"
)

func TestCompareCommitsToolThroughRuntime(t *testing.T) {
	store := auth.NewInMemoryStore()
	resolver := auth.NewAccountResolver(store)
	authService := auth.NewService(store, resolver)
	factory := githubprovider.NewClientFactoryWithHTTPClient(config.GithubAppConfig{
		BaseURL: "https://api.github.test/",
	}, authService, &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.URL.String() == "https://api.github.test/repos/insmtx/SingerOS/compare/abc123...def456?per_page=100" {
				return jsonResponse(req, `{
					"status":"ahead",
					"ahead_by":1,
					"behind_by":0,
					"total_commits":1,
					"html_url":"https://github.com/insmtx/SingerOS/compare/abc123...def456",
					"diff_url":"https://github.com/insmtx/SingerOS/compare/abc123...def456.diff",
					"patch_url":"https://github.com/insmtx/SingerOS/compare/abc123...def456.patch",
					"base_commit":{"sha":"abc123"},
					"merge_base_commit":{"sha":"abc123"},
					"commits":[
						{
							"sha":"def456",
							"html_url":"https://github.com/insmtx/SingerOS/commit/def456",
							"commit":{"message":"feat: add runtime"},
							"author":{"login":"octocat","id":1001}
						}
					],
					"files":[
						{
							"filename":"backend/runtime/eino_runner.go",
							"status":"modified",
							"sha":"file123",
							"patch":"@@ -1,3 +1,8 @@",
							"additions":5,
							"deletions":1,
							"changes":6
						}
					]
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

	registry := tools.NewRegistry()
	mustRegisterTool(t, registry, NewCompareCommitsTool(nil))

	runtime := toolruntime.New(registry, factory)
	result, err := runtime.Execute(context.Background(), &toolruntime.ExecuteRequest{
		ToolName: ToolNameCompareCommits,
		UserID:   "u1",
		Input: map[string]interface{}{
			"repo": "insmtx/SingerOS",
			"base": "abc123",
			"head": "def456",
		},
	})
	if err != nil {
		t.Fatalf("execute compare commits tool: %v", err)
	}

	comparison, ok := result.Output["comparison"].(map[string]interface{})
	if !ok || comparison["status"] != "ahead" {
		t.Fatalf("unexpected comparison output: %+v", result.Output)
	}
	files, ok := result.Output["files"].([]map[string]interface{})
	if !ok {
		rawFiles, rawOK := result.Output["files"].([]interface{})
		if !rawOK || len(rawFiles) != 1 {
			t.Fatalf("unexpected files output: %+v", result.Output)
		}
		file, _ := rawFiles[0].(map[string]interface{})
		if file["filename"] != "backend/runtime/eino_runner.go" {
			t.Fatalf("unexpected compare file: %+v", file)
		}
	} else if len(files) != 1 || files[0]["filename"] != "backend/runtime/eino_runner.go" {
		t.Fatalf("unexpected compare files: %+v", files)
	}
}
