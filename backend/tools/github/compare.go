package githubtools

import (
	"context"
	"fmt"

	gogithub "github.com/google/go-github/v78/github"
	auth "github.com/insmtx/SingerOS/backend/auth"
	githubprovider "github.com/insmtx/SingerOS/backend/providers/github"
	"github.com/insmtx/SingerOS/backend/tools"
)

const ToolNameCompareCommits = "github.repo.compare_commits"

type CompareCommitsTool struct {
	factory *githubprovider.ClientFactory
}

func NewCompareCommitsTool(factory *githubprovider.ClientFactory) *CompareCommitsTool {
	return &CompareCommitsTool{factory: factory}
}

func (t *CompareCommitsTool) Info() *tools.ToolInfo {
	return &tools.ToolInfo{
		Name:        ToolNameCompareCommits,
		Description: "Compare two Git references in a GitHub repository and return commit and changed-file details",
		Provider:    auth.ProviderGitHub,
		ReadOnly:    true,
		InputSchema: &tools.Schema{
			Type:     "object",
			Required: []string{"repo", "base", "head"},
			Properties: map[string]*tools.Property{
				"repo": {
					Type:        "string",
					Description: "Repository full name in owner/name format",
				},
				"base": {
					Type:        "string",
					Description: "Base branch or commit SHA",
				},
				"head": {
					Type:        "string",
					Description: "Head branch or commit SHA",
				},
			},
		},
	}
}

func (t *CompareCommitsTool) Validate(input map[string]interface{}) error {
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
	base, _ := input["base"].(string)
	if base == "" {
		return fmt.Errorf("base is required")
	}
	head, _ := input["head"].(string)
	if head == "" {
		return fmt.Errorf("head is required")
	}
	return nil
}

func (t *CompareCommitsTool) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	if err := t.Validate(input); err != nil {
		return nil, err
	}
	resolved, err := resolveClientDirect(ctx, t.factory, input)
	if err != nil {
		return nil, err
	}
	return t.buildResult(ctx, resolved, input)
}

func (t *CompareCommitsTool) ExecuteWithContext(ctx context.Context, execCtx *tools.ExecutionContext, input map[string]interface{}) (map[string]interface{}, error) {
	if err := t.Validate(input); err != nil {
		return nil, err
	}
	resolved, err := resolveClientFromContext(execCtx)
	if err != nil {
		return nil, err
	}
	return t.buildResult(ctx, resolved, input)
}

func (t *CompareCommitsTool) buildResult(ctx context.Context, resolved *githubprovider.ResolvedClient, input map[string]interface{}) (map[string]interface{}, error) {
	owner, repo, _ := parseRepo(input["repo"].(string))
	base := input["base"].(string)
	head := input["head"].(string)

	comparison, _, err := resolved.Client.Repositories.CompareCommits(ctx, owner, repo, base, head, &gogithub.ListOptions{PerPage: 100})
	if err != nil {
		return nil, fmt.Errorf("compare github commits: %w", err)
	}

	files := make([]map[string]interface{}, 0, len(comparison.Files))
	for _, file := range comparison.Files {
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

	commits := make([]map[string]interface{}, 0, len(comparison.Commits))
	for _, commit := range comparison.Commits {
		commits = append(commits, map[string]interface{}{
			"sha":      commit.GetSHA(),
			"message":  commit.GetCommit().GetMessage(),
			"html_url": commit.GetHTMLURL(),
			"author": map[string]interface{}{
				"login": commit.GetAuthor().GetLogin(),
				"id":    commit.GetAuthor().GetID(),
			},
		})
	}

	return map[string]interface{}{
		"repo": input["repo"],
		"base": base,
		"head": head,
		"comparison": map[string]interface{}{
			"status":        comparison.GetStatus(),
			"ahead_by":      comparison.GetAheadBy(),
			"behind_by":     comparison.GetBehindBy(),
			"total_commits": comparison.GetTotalCommits(),
			"html_url":      comparison.GetHTMLURL(),
			"diff_url":      comparison.GetDiffURL(),
			"patch_url":     comparison.GetPatchURL(),
			"base_commit":   comparison.GetBaseCommit().GetSHA(),
			"merge_base":    comparison.GetMergeBaseCommit().GetSHA(),
		},
		"commits": commits,
		"files":   files,
	}, nil
}
