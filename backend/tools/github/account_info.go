package githubtools

import (
	"context"
	"fmt"

	auth "github.com/insmtx/SingerOS/backend/auth"
	"github.com/insmtx/SingerOS/backend/providers/github"
	"github.com/insmtx/SingerOS/backend/tools"
)

const (
	// ToolNameGetCurrentUser 是读取当前 GitHub 授权账户信息的 Tool 名称。
	ToolNameGetCurrentUser = "github.account.get_current_user"
)

// AccountInfoTool 读取当前 GitHub 授权账户信息。
type AccountInfoTool struct {
	factory *githubprovider.ClientFactory
}

// NewAccountInfoTool 创建一个新的 GitHub 账户信息 Tool。
func NewAccountInfoTool(factory *githubprovider.ClientFactory) *AccountInfoTool {
	return &AccountInfoTool{factory: factory}
}

// Info 返回 Tool 元信息。
func (t *AccountInfoTool) Info() *tools.ToolInfo {
	return &tools.ToolInfo{
		Name:        ToolNameGetCurrentUser,
		Description: "读取当前已授权 GitHub 账户的用户信息",
		Provider:    auth.ProviderGitHub,
		ReadOnly:    true,
		InputSchema: &tools.Schema{
			Type:     "object",
			Required: []string{"user_id"},
			Properties: map[string]*tools.Property{
				"user_id": {
					Type:        "string",
					Description: "SingerOS user id used to resolve the default GitHub account",
				},
				"account_id": {
					Type:        "string",
					Description: "Optional explicit authorized GitHub account id",
				},
			},
		},
	}
}

// Validate 校验 Tool 输入参数。
func (t *AccountInfoTool) Validate(input map[string]interface{}) error {
	if input == nil {
		return fmt.Errorf("input is required")
	}

	userID, ok := input["user_id"].(string)
	if !ok || userID == "" {
		return fmt.Errorf("user_id is required")
	}

	if accountID, ok := input["account_id"]; ok {
		accountIDStr, valid := accountID.(string)
		if !valid || accountIDStr == "" {
			return fmt.Errorf("account_id must be a non-empty string")
		}
	}

	return nil
}

// Execute 执行 GitHub 账户信息读取。
func (t *AccountInfoTool) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	if t.factory == nil {
		return nil, fmt.Errorf("tool runtime execution context is required")
	}
	if err := t.Validate(input); err != nil {
		return nil, err
	}

	userID := input["user_id"].(string)
	accountID, _ := input["account_id"].(string)

	resolved, err := t.factory.ResolveClient(ctx, &githubprovider.ResolveClientRequest{
		UserID:    userID,
		AccountID: accountID,
	})
	if err != nil {
		return nil, err
	}

	return t.buildResult(ctx, resolved)
}

// ExecuteWithContext consumes a runtime-injected GitHub client instead of resolving credentials inside the tool.
func (t *AccountInfoTool) ExecuteWithContext(ctx context.Context, execCtx *tools.ExecutionContext, input map[string]interface{}) (map[string]interface{}, error) {
	if execCtx == nil {
		return nil, fmt.Errorf("execution context is required")
	}
	if err := t.Validate(input); err != nil {
		return nil, err
	}
	if execCtx.Resources == nil {
		return nil, fmt.Errorf("execution resources are required")
	}

	resolvedAny, ok := execCtx.Resources[tools.ResourceGitHubResolvedClient]
	if !ok || resolvedAny == nil {
		return nil, fmt.Errorf("resolved github client is required")
	}

	resolved, ok := resolvedAny.(*githubprovider.ResolvedClient)
	if !ok {
		return nil, fmt.Errorf("invalid resolved github client resource")
	}

	return t.buildResult(ctx, resolved)
}

func (t *AccountInfoTool) buildResult(ctx context.Context, resolved *githubprovider.ResolvedClient) (map[string]interface{}, error) {
	user, _, err := resolved.Client.Users.Get(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("get github current user: %w", err)
	}

	return map[string]interface{}{
		"provider":    auth.ProviderGitHub,
		"resolved_by": resolved.ResolvedBy,
		"authorized_account": map[string]interface{}{
			"id":                  resolved.Account.ID,
			"user_id":             resolved.Account.UserID,
			"provider":            resolved.Account.Provider,
			"account_type":        resolved.Account.AccountType,
			"external_account_id": resolved.Account.ExternalAccountID,
			"display_name":        resolved.Account.DisplayName,
			"scopes":              resolved.Account.Scopes,
			"status":              resolved.Account.Status,
		},
		"github_user": map[string]interface{}{
			"id":           user.GetID(),
			"login":        user.GetLogin(),
			"name":         user.GetName(),
			"email":        user.GetEmail(),
			"avatar_url":   user.GetAvatarURL(),
			"html_url":     user.GetHTMLURL(),
			"company":      user.GetCompany(),
			"location":     user.GetLocation(),
			"bio":          user.GetBio(),
			"public_repos": user.GetPublicRepos(),
			"followers":    user.GetFollowers(),
			"following":    user.GetFollowing(),
		},
	}, nil
}
