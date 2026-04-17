package auth

import (
	"context"
	"testing"
	"time"
)

func TestServiceResolveAuthorizationUsesProviderResolverFirst(t *testing.T) {
	store := NewInMemoryStore()
	accountResolver := NewAccountResolver(store)
	service := NewService(store, accountResolver)

	now := time.Now().UTC()
	defaultAccount := &AuthorizedAccount{
		ID:          "github:u1:default",
		UserID:      "u1",
		Provider:    ProviderGitHub,
		OwnerType:   AccountOwnerTypeUser,
		AccountType: AccountTypeUserOAuth,
		Status:      AccountStatusActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := store.UpsertAuthorizedAccount(context.Background(), defaultAccount, &AccountCredential{
		AccountID:   defaultAccount.ID,
		GrantType:   GrantTypeOAuth2,
		AccessToken: "default-token",
	}); err != nil {
		t.Fatalf("upsert default account: %v", err)
	}
	if err := store.SetDefaultAccount(context.Background(), &UserProviderBinding{
		UserID:    "u1",
		Provider:  ProviderGitHub,
		AccountID: defaultAccount.ID,
		IsDefault: true,
		Priority:  100,
	}); err != nil {
		t.Fatalf("set default account: %v", err)
	}

	service.RegisterAuthResolver(ProviderGitHub, staticAuthResolver{
		resolved: &ResolvedAuthorization{
			Provider:   ProviderGitHub,
			ProfileID:  "github:custom:profile",
			ResolvedBy: "custom_provider_resolver",
			Account: &AuthorizedAccount{
				ID:          "github:custom:profile",
				Provider:    ProviderGitHub,
				OwnerType:   AccountOwnerTypeSystem,
				AccountType: AccountTypeAppInstallation,
				Status:      AccountStatusActive,
			},
			Credential: &AccountCredential{
				AccountID:   "github:custom:profile",
				GrantType:   "test",
				AccessToken: "custom-token",
			},
		},
	})

	resolved, err := service.ResolveAuthorization(context.Background(), &ResolveAuthorizationRequest{
		Selector: &AuthSelector{
			Provider:    ProviderGitHub,
			SubjectType: SubjectTypeUser,
			SubjectID:   "u1",
		},
	})
	if err != nil {
		t.Fatalf("resolve authorization: %v", err)
	}
	if resolved.ProfileID != "github:custom:profile" {
		t.Fatalf("expected provider resolver profile, got %s", resolved.ProfileID)
	}
	if resolved.ResolvedBy != "custom_provider_resolver" {
		t.Fatalf("expected custom resolver, got %s", resolved.ResolvedBy)
	}
}

func TestServiceResolveAuthorizationFallsBackToAccountResolver(t *testing.T) {
	store := NewInMemoryStore()
	accountResolver := NewAccountResolver(store)
	service := NewService(store, accountResolver)

	now := time.Now().UTC()
	account := &AuthorizedAccount{
		ID:          "github:u1:default",
		UserID:      "u1",
		Provider:    ProviderGitHub,
		OwnerType:   AccountOwnerTypeUser,
		AccountType: AccountTypeUserOAuth,
		Status:      AccountStatusActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := store.UpsertAuthorizedAccount(context.Background(), account, &AccountCredential{
		AccountID:   account.ID,
		GrantType:   GrantTypeOAuth2,
		AccessToken: "default-token",
	}); err != nil {
		t.Fatalf("upsert account: %v", err)
	}
	if err := store.SetDefaultAccount(context.Background(), &UserProviderBinding{
		UserID:    "u1",
		Provider:  ProviderGitHub,
		AccountID: account.ID,
		IsDefault: true,
		Priority:  100,
	}); err != nil {
		t.Fatalf("set default account: %v", err)
	}

	service.RegisterAuthResolver(ProviderGitHub, staticAuthResolver{})

	resolved, err := service.ResolveAuthorization(context.Background(), &ResolveAuthorizationRequest{
		Selector: &AuthSelector{
			Provider:    ProviderGitHub,
			SubjectType: SubjectTypeUser,
			SubjectID:   "u1",
		},
	})
	if err != nil {
		t.Fatalf("resolve authorization: %v", err)
	}
	if resolved.ProfileID != account.ID {
		t.Fatalf("expected default account profile %s, got %s", account.ID, resolved.ProfileID)
	}
	if resolved.ResolvedBy != "subject_default" {
		t.Fatalf("expected fallback account resolver, got %s", resolved.ResolvedBy)
	}
}

type staticAuthResolver struct {
	resolved *ResolvedAuthorization
	err      error
}

func (r staticAuthResolver) ResolveAuthorization(context.Context, *ResolveAuthorizationRequest) (*ResolvedAuthorization, bool, error) {
	if r.err != nil {
		return nil, true, r.err
	}
	if r.resolved == nil {
		return nil, false, nil
	}
	return r.resolved, true, nil
}
