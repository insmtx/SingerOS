package githubtools

import (
	"context"
	"fmt"
	"strings"

	githubprovider "github.com/insmtx/SingerOS/backend/pkg/providers/github"
	"github.com/insmtx/SingerOS/backend/tools"
)

func parseRepo(fullName string) (string, string, error) {
	parts := strings.Split(fullName, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("repo must be in owner/name format")
	}

	return parts[0], parts[1], nil
}

func resolveClientFromContext(execCtx *tools.ExecutionContext) (*githubprovider.ResolvedClient, error) {
	if execCtx == nil {
		return nil, fmt.Errorf("execution context is required")
	}
	if execCtx.Resources == nil {
		return nil, fmt.Errorf("execution resources are required")
	}

	resource, ok := execCtx.Resources[tools.ResourceGitHubResolvedClient]
	if !ok || resource == nil {
		return nil, fmt.Errorf("resolved github client is required")
	}

	resolved, ok := resource.(*githubprovider.ResolvedClient)
	if !ok {
		return nil, fmt.Errorf("invalid resolved github client resource")
	}

	return resolved, nil
}

func resolveClientDirect(ctx context.Context, factory *githubprovider.ClientFactory, input map[string]interface{}) (*githubprovider.ResolvedClient, error) {
	if factory == nil {
		return nil, fmt.Errorf("tool runtime execution context is required")
	}

	userID, _ := input["user_id"].(string)
	if userID == "" {
		return nil, fmt.Errorf("user_id is required")
	}
	accountID, _ := input["account_id"].(string)

	return factory.ResolveClient(ctx, &githubprovider.ResolveClientRequest{
		UserID:    userID,
		AccountID: accountID,
	})
}
