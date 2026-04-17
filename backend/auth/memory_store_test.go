package auth

import (
	"context"
	"testing"
	"time"
)

func TestInMemoryStoreSavesDefaultAccount(t *testing.T) {
	store := NewInMemoryStore()
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
	credential := &AccountCredential{
		AccountID:   account.ID,
		GrantType:   GrantTypeOAuth2,
		AccessToken: "token-1",
	}

	if err := store.UpsertAuthorizedAccount(context.Background(), account, credential); err != nil {
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

	defaultAccount, err := store.GetDefaultAccount(context.Background(), "u1", ProviderGitHub)
	if err != nil {
		t.Fatalf("get default account: %v", err)
	}
	if defaultAccount.ID != account.ID {
		t.Fatalf("expected default account %s, got %s", account.ID, defaultAccount.ID)
	}
}
