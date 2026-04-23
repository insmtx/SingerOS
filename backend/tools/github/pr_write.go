package githubtools

import (
	"context"
	"fmt"
	"strings"

	gogithub "github.com/google/go-github/v78/github"
	githubprovider "github.com/insmtx/SingerOS/backend/providers/github"
	"github.com/insmtx/SingerOS/backend/tools"
)

const ToolNamePublishPullRequestReview = "github.pr.publish_review"

var allowedReviewEvents = map[string]struct{}{
	"COMMENT":         {},
	"APPROVE":         {},
	"REQUEST_CHANGES": {},
}

type PullRequestReviewPublishTool struct {
	tools.BaseTool
	factory *githubprovider.ClientFactory
}

func NewPullRequestReviewPublishTool(factory *githubprovider.ClientFactory) *PullRequestReviewPublishTool {
	return &PullRequestReviewPublishTool{
		BaseTool: tools.NewBaseTool(
			ToolNamePublishPullRequestReview,
			"Publish a GitHub pull request review with a summary body and optional inline review comments",
			tools.Schema{
				Type:     "object",
				Required: []string{"repo", "pr_number", "body"},
				Properties: map[string]*tools.Property{
					"repo": {
						Type:        "string",
						Description: "Repository full name in owner/name format",
					},
					"pr_number": {
						Type:        "integer",
						Description: "Pull request number",
					},
					"body": {
						Type:        "string",
						Description: "Review summary body to publish",
					},
					"event": {
						Type:        "string",
						Description: "Optional review event: COMMENT, APPROVE, or REQUEST_CHANGES",
						Enum:        []string{"COMMENT", "APPROVE", "REQUEST_CHANGES"},
					},
					"commit_id": {
						Type:        "string",
						Description: "Optional commit SHA used to anchor inline review comments",
					},
					"comments": {
						Type:        "array",
						Description: "Optional inline review comments",
						Items: &tools.Property{
							Type: "object",
						},
					},
				},
			},
		),
		factory: factory,
	}
}

func (t *PullRequestReviewPublishTool) Validate(input map[string]interface{}) error {
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

	body, _ := input["body"].(string)
	if strings.TrimSpace(body) == "" {
		return fmt.Errorf("body is required")
	}

	if event, ok := input["event"].(string); ok && event != "" {
		event = strings.ToUpper(strings.TrimSpace(event))
		if _, exists := allowedReviewEvents[event]; !exists {
			return fmt.Errorf("event must be one of COMMENT, APPROVE, REQUEST_CHANGES")
		}
	}

	if _, err := buildDraftReviewComments(input["comments"]); err != nil {
		return err
	}

	return nil
}

func (t *PullRequestReviewPublishTool) Execute(ctx context.Context, input map[string]interface{}) (string, error) {
	if err := t.Validate(input); err != nil {
		return "", err
	}
	resolved, err := resolveClientDirect(ctx, t.factory, input)
	if err != nil {
		return "", err
	}
	return t.buildResult(ctx, resolved, input)
}

func (t *PullRequestReviewPublishTool) buildResult(ctx context.Context, resolved *githubprovider.ResolvedClient, input map[string]interface{}) (string, error) {
	owner, repo, _ := parseRepo(input["repo"].(string))
	prNumber, _ := getIntValue(input["pr_number"])
	body := strings.TrimSpace(input["body"].(string))

	reviewRequest := &gogithub.PullRequestReviewRequest{
		Body: gogithub.Ptr(body),
	}

	if event, ok := input["event"].(string); ok && strings.TrimSpace(event) != "" {
		reviewEvent := strings.ToUpper(strings.TrimSpace(event))
		reviewRequest.Event = gogithub.Ptr(reviewEvent)
	} else {
		reviewRequest.Event = gogithub.Ptr("COMMENT")
	}

	if commitID, ok := input["commit_id"].(string); ok && strings.TrimSpace(commitID) != "" {
		reviewRequest.CommitID = gogithub.Ptr(strings.TrimSpace(commitID))
	}

	comments, err := buildDraftReviewComments(input["comments"])
	if err != nil {
		return "", err
	}
	if len(comments) > 0 {
		reviewRequest.Comments = comments
	}

	review, _, err := resolved.Client.PullRequests.CreateReview(ctx, owner, repo, prNumber, reviewRequest)
	if err != nil {
		return "", fmt.Errorf("create github pull request review: %w", err)
	}

	result := map[string]interface{}{
		"repo":      input["repo"],
		"pr_number": prNumber,
		"review": map[string]interface{}{
			"id":       review.GetID(),
			"node_id":  review.GetNodeID(),
			"state":    review.GetState(),
			"body":     review.GetBody(),
			"html_url": review.GetHTMLURL(),
			"commit_id": func() string {
				if review.CommitID == nil {
					return ""
				}
				return *review.CommitID
			}(),
			"submitted_at": review.GetSubmittedAt().String(),
			"user": map[string]interface{}{
				"login": review.GetUser().GetLogin(),
				"id":    review.GetUser().GetID(),
			},
		},
	}

	if reviewRequest.Event != nil {
		result["event"] = *reviewRequest.Event
	}
	if len(comments) > 0 {
		result["comment_count"] = len(comments)
	}

	return tools.JSONString(result)
}

func buildDraftReviewComments(value interface{}) ([]*gogithub.DraftReviewComment, error) {
	if value == nil {
		return nil, nil
	}

	rawComments, ok := value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("comments must be an array")
	}

	comments := make([]*gogithub.DraftReviewComment, 0, len(rawComments))
	for index, raw := range rawComments {
		commentInput, ok := raw.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("comments[%d] must be an object", index)
		}

		path, _ := commentInput["path"].(string)
		if strings.TrimSpace(path) == "" {
			return nil, fmt.Errorf("comments[%d].path is required", index)
		}

		body, _ := commentInput["body"].(string)
		if strings.TrimSpace(body) == "" {
			return nil, fmt.Errorf("comments[%d].body is required", index)
		}

		comment := &gogithub.DraftReviewComment{
			Path: gogithub.Ptr(strings.TrimSpace(path)),
			Body: gogithub.Ptr(strings.TrimSpace(body)),
		}

		if position, err := getOptionalIntValue(commentInput["position"]); err != nil {
			return nil, fmt.Errorf("comments[%d].position: %w", index, err)
		} else if position > 0 {
			comment.Position = gogithub.Ptr(position)
		}

		if line, err := getOptionalIntValue(commentInput["line"]); err != nil {
			return nil, fmt.Errorf("comments[%d].line: %w", index, err)
		} else if line > 0 {
			comment.Line = gogithub.Ptr(line)
		}

		if startLine, err := getOptionalIntValue(commentInput["start_line"]); err != nil {
			return nil, fmt.Errorf("comments[%d].start_line: %w", index, err)
		} else if startLine > 0 {
			comment.StartLine = gogithub.Ptr(startLine)
		}

		if side, ok := commentInput["side"].(string); ok && strings.TrimSpace(side) != "" {
			comment.Side = gogithub.Ptr(strings.ToUpper(strings.TrimSpace(side)))
		}
		if startSide, ok := commentInput["start_side"].(string); ok && strings.TrimSpace(startSide) != "" {
			comment.StartSide = gogithub.Ptr(strings.ToUpper(strings.TrimSpace(startSide)))
		}

		hasPositionStyle := comment.Position != nil
		hasLineStyle := comment.Line != nil || comment.StartLine != nil || comment.Side != nil || comment.StartSide != nil
		switch {
		case hasPositionStyle && hasLineStyle:
			return nil, fmt.Errorf("comments[%d] cannot mix position and line-based fields", index)
		case !hasPositionStyle && !hasLineStyle:
			return nil, fmt.Errorf("comments[%d] requires either position or line-based fields", index)
		}

		comments = append(comments, comment)
	}

	return comments, nil
}

func getOptionalIntValue(value interface{}) (int, error) {
	if value == nil {
		return 0, nil
	}
	return getIntValue(value)
}
