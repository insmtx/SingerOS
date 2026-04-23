package githubtools

import (
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	auth "github.com/insmtx/SingerOS/backend/auth"
	"github.com/insmtx/SingerOS/backend/config"
	githubprovider "github.com/insmtx/SingerOS/backend/providers/github"
)

func TestGitHubReadToolsExecute(t *testing.T) {
	store := auth.NewInMemoryStore()
	resolver := auth.NewAccountResolver(store)
	authService := auth.NewService(store, resolver)
	factory := githubprovider.NewClientFactoryWithHTTPClient(config.GithubAppConfig{
		BaseURL: "https://api.github.test/",
	}, authService, &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch {
			case req.URL.String() == "https://api.github.test/repos/insmtx/SingerOS/pulls/12":
				return jsonResponse(req, `{
					"number": 12,
					"title": "Add Eino runtime",
					"body": "Introduce Eino runtime runner",
					"state": "open",
					"draft": false,
					"html_url": "https://github.com/insmtx/SingerOS/pull/12",
					"diff_url": "https://github.com/insmtx/SingerOS/pull/12.diff",
					"patch_url": "https://github.com/insmtx/SingerOS/pull/12.patch",
					"commits": 3,
					"changed_files": 5,
					"additions": 120,
					"deletions": 15,
					"mergeable": true,
					"user": {"login":"octocat","id":1001},
					"head": {"ref":"feature/eino"},
					"base": {"ref":"main"}
				}`)
			case strings.HasPrefix(req.URL.String(), "https://api.github.test/repos/insmtx/SingerOS/pulls/12/files"):
				return jsonResponse(req, `[
					{
						"filename":"backend/runtime/eino_runner.go",
						"status":"added",
						"sha":"abc123",
						"blob_url":"https://github.com/insmtx/SingerOS/blob/abc123/backend/runtime/eino_runner.go",
						"raw_url":"https://raw.githubusercontent.com/insmtx/SingerOS/abc123/backend/runtime/eino_runner.go",
						"contents_url":"https://api.github.com/repos/insmtx/SingerOS/contents/backend/runtime/eino_runner.go?ref=abc123",
						"patch":"@@ -0,0 +1,10 @@",
						"additions":10,
						"deletions":0,
						"changes":10
					}
				]`)
			case strings.HasPrefix(req.URL.String(), "https://api.github.test/repos/insmtx/SingerOS/contents/backend/runtime/eino_runner.go"):
				content := base64.StdEncoding.EncodeToString([]byte("package runtime\n"))
				return jsonResponse(req, `{
					"type":"file",
					"encoding":"base64",
					"size":16,
					"sha":"file123",
					"content":"`+content+`"
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

	metadataTool := NewPullRequestMetadataTool(factory)
	metadataOutputRaw, err := metadataTool.Execute(context.Background(), map[string]interface{}{
		"user_id":   "u1",
		"repo":      "insmtx/SingerOS",
		"pr_number": 12,
	})
	if err != nil {
		t.Fatalf("execute metadata tool: %v", err)
	}
	metadataOutput := decodeGitHubToolOutput(t, metadataOutputRaw)
	pr, ok := metadataOutput["pull_request"].(map[string]interface{})
	if !ok || pr["title"] != "Add Eino runtime" {
		t.Fatalf("unexpected metadata output: %+v", metadataOutput)
	}

	filesTool := NewPullRequestFilesTool(factory)
	filesOutputRaw, err := filesTool.Execute(context.Background(), map[string]interface{}{
		"user_id":   "u1",
		"repo":      "insmtx/SingerOS",
		"pr_number": 12,
	})
	if err != nil {
		t.Fatalf("execute files tool: %v", err)
	}
	filesOutput := decodeGitHubToolOutput(t, filesOutputRaw)
	fileList, ok := filesOutput["files"].([]map[string]interface{})
	if !ok {
		rawList, rawOK := filesOutput["files"].([]interface{})
		if !rawOK || len(rawList) != 1 {
			t.Fatalf("unexpected files output: %+v", filesOutput)
		}
		firstFile, _ := rawList[0].(map[string]interface{})
		if firstFile["filename"] != "backend/runtime/eino_runner.go" {
			t.Fatalf("unexpected file entry: %+v", firstFile)
		}
	} else if len(fileList) != 1 || fileList[0]["filename"] != "backend/runtime/eino_runner.go" {
		t.Fatalf("unexpected file list: %+v", fileList)
	}

	fileTool := NewRepositoryFileTool(factory)
	fileContentOutputRaw, err := fileTool.Execute(context.Background(), map[string]interface{}{
		"user_id": "u1",
		"repo":    "insmtx/SingerOS",
		"path":    "backend/runtime/eino_runner.go",
		"ref":     "main",
	})
	if err != nil {
		t.Fatalf("execute repository file tool: %v", err)
	}
	fileContentOutput := decodeGitHubToolOutput(t, fileContentOutputRaw)
	if fileContentOutput["content"] != "package runtime\n" {
		t.Fatalf("unexpected file content: %+v", fileContentOutput)
	}
}

func jsonResponse(req *http.Request, body string) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header: http.Header{
			"Content-Type": []string{"application/json"},
		},
		Body:    io.NopCloser(strings.NewReader(body)),
		Request: req,
	}, nil
}
