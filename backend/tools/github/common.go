package githubtools

import (
	"context"
	"fmt"
	"strings"

	githubprovider "github.com/insmtx/SingerOS/backend/providers/github"
)

func parseRepo(fullName string) (string, string, error) {
	parts := strings.Split(fullName, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("repo must be in owner/name format")
	}

	return parts[0], parts[1], nil
}

func resolveClientDirect(ctx context.Context, factory *githubprovider.ClientFactory, input map[string]interface{}) (*githubprovider.ResolvedClient, error) {
	if factory == nil {
		return nil, fmt.Errorf("github client factory is required")
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
