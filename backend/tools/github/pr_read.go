package githubtools

import (
	"context"
	"fmt"

	gogithub "github.com/google/go-github/v78/github"
	auth "github.com/insmtx/SingerOS/backend/auth"
	githubprovider "github.com/insmtx/SingerOS/backend/providers/github"
	"github.com/insmtx/SingerOS/backend/tools"
)

const (
	ToolNameGetPullRequestMetadata = "github.pr.get_metadata"
	ToolNameGetPullRequestFiles    = "github.pr.get_files"
	ToolNameGetRepositoryFile      = "github.repo.get_file"
)

type PullRequestMetadataTool struct {
	factory *githubprovider.ClientFactory
}

func NewPullRequestMetadataTool(factory *githubprovider.ClientFactory) *PullRequestMetadataTool {
	return &PullRequestMetadataTool{factory: factory}
}

func (t *PullRequestMetadataTool) Info() *tools.ToolInfo {
	return &tools.ToolInfo{
		Name:        ToolNameGetPullRequestMetadata,
		Description: "Read GitHub pull request metadata including title, branches, author, and diff stats",
		Provider:    auth.ProviderGitHub,
		ReadOnly:    true,
		InputSchema: &tools.Schema{
			Type:     "object",
			Required: []string{"repo", "pr_number"},
			Properties: map[string]*tools.Property{
				"repo": {
					Type:        "string",
					Description: "Repository full name in owner/name format",
				},
				"pr_number": {
					Type:        "integer",
					Description: "Pull request number",
				},
			},
		},
	}
}

func (t *PullRequestMetadataTool) Validate(input map[string]interface{}) error {
	if input == nil {
		return fmt.Errorf("input is required")
	}
	repo, _ := input["repo"].(string)
	if repo == "" {
		return fmt.Errorf("repo is required")
	}
	if _, _, err := parseRepo(repo); err != nil {
		return err
	}
	if _, err := getIntValue(input["pr_number"]); err != nil {
		return fmt.Errorf("pr_number is required")
	}
	return nil
}

func (t *PullRequestMetadataTool) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	if err := t.Validate(input); err != nil {
		return nil, err
	}
	resolved, err := resolveClientDirect(ctx, t.factory, input)
	if err != nil {
		return nil, err
	}
	return t.buildResult(ctx, resolved, input)
}

func (t *PullRequestMetadataTool) ExecuteWithContext(ctx context.Context, execCtx *tools.ExecutionContext, input map[string]interface{}) (map[string]interface{}, error) {
	if err := t.Validate(input); err != nil {
		return nil, err
	}
	resolved, err := resolveClientFromContext(execCtx)
	if err != nil {
		return nil, err
	}
	return t.buildResult(ctx, resolved, input)
}

func (t *PullRequestMetadataTool) buildResult(ctx context.Context, resolved *githubprovider.ResolvedClient, input map[string]interface{}) (map[string]interface{}, error) {
	owner, repo, _ := parseRepo(input["repo"].(string))
	prNumber, _ := getIntValue(input["pr_number"])

	pr, _, err := resolved.Client.PullRequests.Get(ctx, owner, repo, prNumber)
	if err != nil {
		return nil, fmt.Errorf("get github pull request: %w", err)
	}

	return map[string]interface{}{
		"repo":      input["repo"],
		"pr_number": prNumber,
		"pull_request": map[string]interface{}{
			"number":        pr.GetNumber(),
			"title":         pr.GetTitle(),
			"body":          pr.GetBody(),
			"state":         pr.GetState(),
			"draft":         pr.GetDraft(),
			"html_url":      pr.GetHTMLURL(),
			"diff_url":      pr.GetDiffURL(),
			"patch_url":     pr.GetPatchURL(),
			"commits":       pr.GetCommits(),
			"changed_files": pr.GetChangedFiles(),
			"additions":     pr.GetAdditions(),
			"deletions":     pr.GetDeletions(),
			"mergeable":     pr.GetMergeable(),
			"head_ref":      pr.GetHead().GetRef(),
			"base_ref":      pr.GetBase().GetRef(),
			"author": map[string]interface{}{
				"login": pr.GetUser().GetLogin(),
				"id":    pr.GetUser().GetID(),
			},
		},
	}, nil
}

type PullRequestFilesTool struct {
	factory *githubprovider.ClientFactory
}

func NewPullRequestFilesTool(factory *githubprovider.ClientFactory) *PullRequestFilesTool {
	return &PullRequestFilesTool{factory: factory}
}

func (t *PullRequestFilesTool) Info() *tools.ToolInfo {
	return &tools.ToolInfo{
		Name:        ToolNameGetPullRequestFiles,
		Description: "List changed files in a GitHub pull request with per-file diff stats and patch snippets",
		Provider:    auth.ProviderGitHub,
		ReadOnly:    true,
		InputSchema: &tools.Schema{
			Type:     "object",
			Required: []string{"repo", "pr_number"},
			Properties: map[string]*tools.Property{
				"repo": {
					Type:        "string",
					Description: "Repository full name in owner/name format",
				},
				"pr_number": {
					Type:        "integer",
					Description: "Pull request number",
				},
			},
		},
	}
}

func (t *PullRequestFilesTool) Validate(input map[string]interface{}) error {
	return (&PullRequestMetadataTool{}).Validate(input)
}

func (t *PullRequestFilesTool) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	if err := t.Validate(input); err != nil {
		return nil, err
	}
	resolved, err := resolveClientDirect(ctx, t.factory, input)
	if err != nil {
		return nil, err
	}
	return t.buildResult(ctx, resolved, input)
}

func (t *PullRequestFilesTool) ExecuteWithContext(ctx context.Context, execCtx *tools.ExecutionContext, input map[string]interface{}) (map[string]interface{}, error) {
	if err := t.Validate(input); err != nil {
		return nil, err
	}
	resolved, err := resolveClientFromContext(execCtx)
	if err != nil {
		return nil, err
	}
	return t.buildResult(ctx, resolved, input)
}

func (t *PullRequestFilesTool) buildResult(ctx context.Context, resolved *githubprovider.ResolvedClient, input map[string]interface{}) (map[string]interface{}, error) {
	owner, repo, _ := parseRepo(input["repo"].(string))
	prNumber, _ := getIntValue(input["pr_number"])

	opts := &gogithub.ListOptions{PerPage: 100, Page: 1}
	files := make([]map[string]interface{}, 0)
	for {
		pageFiles, resp, err := resolved.Client.PullRequests.ListFiles(ctx, owner, repo, prNumber, opts)
		if err != nil {
			return nil, fmt.Errorf("list github pull request files: %w", err)
		}

		for _, file := range pageFiles {
			files = append(files, map[string]interface{}{
				"filename":     file.GetFilename(),
				"status":       file.GetStatus(),
				"sha":          file.GetSHA(),
				"blob_url":     file.GetBlobURL(),
				"raw_url":      file.GetRawURL(),
				"contents_url": file.GetContentsURL(),
				"patch":        file.GetPatch(),
				"additions":    file.GetAdditions(),
				"deletions":    file.GetDeletions(),
				"changes":      file.GetChanges(),
			})
		}

		if resp == nil || resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return map[string]interface{}{
		"repo":      input["repo"],
		"pr_number": prNumber,
		"files":     files,
	}, nil
}

type RepositoryFileTool struct {
	factory *githubprovider.ClientFactory
}

func NewRepositoryFileTool(factory *githubprovider.ClientFactory) *RepositoryFileTool {
	return &RepositoryFileTool{factory: factory}
}

func (t *RepositoryFileTool) Info() *tools.ToolInfo {
	return &tools.ToolInfo{
		Name:        ToolNameGetRepositoryFile,
		Description: "Read one file from a GitHub repository at an optional ref",
		Provider:    auth.ProviderGitHub,
		ReadOnly:    true,
		InputSchema: &tools.Schema{
			Type:     "object",
			Required: []string{"repo", "path"},
			Properties: map[string]*tools.Property{
				"repo": {
					Type:        "string",
					Description: "Repository full name in owner/name format",
				},
				"path": {
					Type:        "string",
					Description: "Path to the file inside the repository",
				},
				"ref": {
					Type:        "string",
					Description: "Optional branch, tag, or commit SHA",
				},
			},
		},
	}
}

func (t *RepositoryFileTool) Validate(input map[string]interface{}) error {
	if input == nil {
		return fmt.Errorf("input is required")
	}
	repo, _ := input["repo"].(string)
	if repo == "" {
		return fmt.Errorf("repo is required")
	}
	if _, _, err := parseRepo(repo); err != nil {
		return err
	}
	path, _ := input["path"].(string)
	if path == "" {
		return fmt.Errorf("path is required")
	}
	return nil
}

func (t *RepositoryFileTool) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	if err := t.Validate(input); err != nil {
		return nil, err
	}
	resolved, err := resolveClientDirect(ctx, t.factory, input)
	if err != nil {
		return nil, err
	}
	return t.buildResult(ctx, resolved, input)
}

func (t *RepositoryFileTool) ExecuteWithContext(ctx context.Context, execCtx *tools.ExecutionContext, input map[string]interface{}) (map[string]interface{}, error) {
	if err := t.Validate(input); err != nil {
		return nil, err
	}
	resolved, err := resolveClientFromContext(execCtx)
	if err != nil {
		return nil, err
	}
	return t.buildResult(ctx, resolved, input)
}

func (t *RepositoryFileTool) buildResult(ctx context.Context, resolved *githubprovider.ResolvedClient, input map[string]interface{}) (map[string]interface{}, error) {
	owner, repo, _ := parseRepo(input["repo"].(string))
	filePath := input["path"].(string)
	ref, _ := input["ref"].(string)

	fileContent, _, _, err := resolved.Client.Repositories.GetContents(ctx, owner, repo, filePath, &gogithub.RepositoryContentGetOptions{
		Ref: ref,
	})
	if err != nil {
		return nil, fmt.Errorf("get github repository file: %w", err)
	}
	if fileContent == nil {
		return nil, fmt.Errorf("path %s is not a file", filePath)
	}

	content, err := fileContent.GetContent()
	if err != nil {
		return nil, fmt.Errorf("decode github repository file: %w", err)
	}

	return map[string]interface{}{
		"repo":     input["repo"],
		"path":     filePath,
		"ref":      ref,
		"sha":      fileContent.GetSHA(),
		"encoding": fileContent.GetEncoding(),
		"size":     fileContent.GetSize(),
		"content":  content,
	}, nil
}

func getIntValue(value interface{}) (int, error) {
	switch typed := value.(type) {
	case int:
		return typed, nil
	case int32:
		return int(typed), nil
	case int64:
		return int(typed), nil
	case float64:
		return int(typed), nil
	default:
		return 0, fmt.Errorf("invalid integer value")
	}
}
