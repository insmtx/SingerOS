package auth

import (
	"context"
	"testing"
	"time"
)

func TestAccountResolverUsesDefaultAccountFirst(t *testing.T) {
	store := NewInMemoryStore()
	resolver := NewAccountResolver(store)
	now := time.Now().UTC()

	first := &AuthorizedAccount{
		ID:          "github:u1:100",
		UserID:      "u1",
		Provider:    ProviderGitHub,
		OwnerType:   AccountOwnerTypeUser,
		AccountType: AccountTypeUserOAuth,
		Status:      AccountStatusActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	second := &AuthorizedAccount{
		ID:          "github:u1:200",
		UserID:      "u1",
		Provider:    ProviderGitHub,
		OwnerType:   AccountOwnerTypeUser,
		AccountType: AccountTypeUserOAuth,
		Status:      AccountStatusActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	for _, account := range []*AuthorizedAccount{first, second} {
		if err := store.UpsertAuthorizedAccount(context.Background(), account, &AccountCredential{
			AccountID:   account.ID,
			GrantType:   GrantTypeOAuth2,
			AccessToken: account.ID + "-token",
		}); err != nil {
			t.Fatalf("upsert account: %v", err)
		}
	}

	if err := store.SetDefaultAccount(context.Background(), &UserProviderBinding{
		UserID:    "u1",
		Provider:  ProviderGitHub,
		AccountID: second.ID,
		IsDefault: true,
		Priority:  100,
	}); err != nil {
		t.Fatalf("set default account: %v", err)
	}

	resolved, err := resolver.Resolve(context.Background(), &ResolveAccountRequest{
		UserID:   "u1",
		Provider: ProviderGitHub,
	})
	if err != nil {
		t.Fatalf("resolve account: %v", err)
	}
	if resolved.Account.ID != second.ID {
		t.Fatalf("expected default account %s, got %s", second.ID, resolved.Account.ID)
	}
	if resolved.ResolvedBy != "subject_default" {
		t.Fatalf("expected resolved by subject_default, got %s", resolved.ResolvedBy)
	}
}

func TestAccountResolverUsesSelectorSubject(t *testing.T) {
	store := NewInMemoryStore()
	resolver := NewAccountResolver(store)
	now := time.Now().UTC()

	account := &AuthorizedAccount{
		ID:          "github:u1:100",
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
		AccessToken: "token-1",
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

	resolved, err := resolver.Resolve(context.Background(), &ResolveAccountRequest{
		Selector: &AuthSelector{
			Provider:    ProviderGitHub,
			SubjectType: SubjectTypeUser,
			SubjectID:   "u1",
		},
	})
	if err != nil {
		t.Fatalf("resolve account with selector: %v", err)
	}
	if resolved.Account.ID != account.ID {
		t.Fatalf("expected account %s, got %s", account.ID, resolved.Account.ID)
	}
	if resolved.ResolvedBy != "subject_default" {
		t.Fatalf("expected resolved by subject_default, got %s", resolved.ResolvedBy)
	}
}
