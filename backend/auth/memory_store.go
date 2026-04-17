package auth

import (
	"context"
	"fmt"
	"sync"
)

// InMemoryStore 是 Store 的内存实现。
type InMemoryStore struct {
	mu          sync.RWMutex
	oauthStates map[string]*OAuthState
	accounts    map[string]*AuthorizedAccount
	credentials map[string]*AccountCredential
	defaults    map[string]*UserProviderBinding
}

// NewInMemoryStore 创建一个新的内存授权存储。
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		oauthStates: make(map[string]*OAuthState),
		accounts:    make(map[string]*AuthorizedAccount),
		credentials: make(map[string]*AccountCredential),
		defaults:    make(map[string]*UserProviderBinding),
	}
}

// SaveOAuthState 保存一次 OAuth state。
func (s *InMemoryStore) SaveOAuthState(_ context.Context, state *OAuthState) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.oauthStates[s.oauthStateKey(state.Provider, state.State)] = cloneOAuthState(state)
	return nil
}

// ConsumeOAuthState 读取并删除一次 OAuth state。
func (s *InMemoryStore) ConsumeOAuthState(_ context.Context, provider, state string) (*OAuthState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := s.oauthStateKey(provider, state)
	saved, ok := s.oauthStates[key]
	if !ok {
		return nil, fmt.Errorf("oauth state not found")
	}

	delete(s.oauthStates, key)
	return cloneOAuthState(saved), nil
}

// UpsertAuthorizedAccount 保存或更新授权账户和凭证。
func (s *InMemoryStore) UpsertAuthorizedAccount(_ context.Context, account *AuthorizedAccount, credential *AccountCredential) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.accounts[account.ID] = cloneAuthorizedAccount(account)
	s.credentials[account.ID] = cloneCredential(credential)
	return nil
}

// GetAuthorizedAccount 返回指定账户。
func (s *InMemoryStore) GetAuthorizedAccount(_ context.Context, accountID string) (*AuthorizedAccount, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	account, ok := s.accounts[accountID]
	if !ok {
		return nil, fmt.Errorf("authorized account not found")
	}

	return cloneAuthorizedAccount(account), nil
}

// ListUserAccounts 返回某用户在某 provider 下的所有账户。
func (s *InMemoryStore) ListUserAccounts(_ context.Context, userID, provider string) ([]*AuthorizedAccount, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	accounts := make([]*AuthorizedAccount, 0)
	for _, account := range s.accounts {
		if account.UserID != userID || account.Provider != provider {
			continue
		}
		accounts = append(accounts, cloneAuthorizedAccount(account))
	}

	return accounts, nil
}

// GetCredential 返回指定账户的凭证。
func (s *InMemoryStore) GetCredential(_ context.Context, accountID string) (*AccountCredential, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	credential, ok := s.credentials[accountID]
	if !ok {
		return nil, fmt.Errorf("account credential not found")
	}

	return cloneCredential(credential), nil
}

// SetDefaultAccount 设置某用户在某 provider 下的默认账户。
func (s *InMemoryStore) SetDefaultAccount(_ context.Context, binding *UserProviderBinding) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.accounts[binding.AccountID]; !ok {
		return fmt.Errorf("authorized account not found for binding")
	}

	s.defaults[s.defaultBindingKey(binding.UserID, binding.Provider)] = cloneBinding(binding)
	return nil
}

// GetDefaultAccount 返回某用户在某 provider 下的默认账户。
func (s *InMemoryStore) GetDefaultAccount(_ context.Context, userID, provider string) (*AuthorizedAccount, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	binding, ok := s.defaults[s.defaultBindingKey(userID, provider)]
	if !ok {
		return nil, fmt.Errorf("default account not found")
	}

	account, ok := s.accounts[binding.AccountID]
	if !ok {
		return nil, fmt.Errorf("default account binding is stale")
	}

	return cloneAuthorizedAccount(account), nil
}

func (s *InMemoryStore) oauthStateKey(provider, state string) string {
	return provider + ":" + state
}

func (s *InMemoryStore) defaultBindingKey(userID, provider string) string {
	return userID + ":" + provider
}

func cloneAuthorizedAccount(account *AuthorizedAccount) *AuthorizedAccount {
	if account == nil {
		return nil
	}

	cloned := *account
	if account.Metadata != nil {
		cloned.Metadata = make(map[string]string, len(account.Metadata))
		for k, v := range account.Metadata {
			cloned.Metadata[k] = v
		}
	}
	if account.Scopes != nil {
		cloned.Scopes = append([]string(nil), account.Scopes...)
	}

	return &cloned
}

func cloneCredential(credential *AccountCredential) *AccountCredential {
	if credential == nil {
		return nil
	}

	cloned := *credential
	if credential.ExpiresAt != nil {
		expiresAt := *credential.ExpiresAt
		cloned.ExpiresAt = &expiresAt
	}
	if credential.Metadata != nil {
		cloned.Metadata = make(map[string]string, len(credential.Metadata))
		for k, v := range credential.Metadata {
			cloned.Metadata[k] = v
		}
	}

	return &cloned
}

func cloneBinding(binding *UserProviderBinding) *UserProviderBinding {
	if binding == nil {
		return nil
	}

	cloned := *binding
	return &cloned
}

func cloneOAuthState(state *OAuthState) *OAuthState {
	if state == nil {
		return nil
	}

	cloned := *state
	return &cloned
}
