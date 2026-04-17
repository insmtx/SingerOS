package auth

import "context"

// Store 定义授权账户的存储接口。
type Store interface {
	SaveOAuthState(ctx context.Context, state *OAuthState) error
	ConsumeOAuthState(ctx context.Context, provider, state string) (*OAuthState, error)

	UpsertAuthorizedAccount(ctx context.Context, account *AuthorizedAccount, credential *AccountCredential) error
	GetAuthorizedAccount(ctx context.Context, accountID string) (*AuthorizedAccount, error)
	ListUserAccounts(ctx context.Context, userID, provider string) ([]*AuthorizedAccount, error)

	GetCredential(ctx context.Context, accountID string) (*AccountCredential, error)

	SetDefaultAccount(ctx context.Context, binding *UserProviderBinding) error
	GetDefaultAccount(ctx context.Context, userID, provider string) (*AuthorizedAccount, error)
}
