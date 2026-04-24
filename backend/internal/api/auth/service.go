package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

const oauthStateBytes = 16

// ThirdPartyAuthService 提供统一的第三方平台授权账户接入与运行时解析能力。
type ThirdPartyAuthService struct {
	store           Store
	resolver        *AccountResolver
	providers       map[string]AuthorizationProvider
	authResolvers   map[string][]ProviderAuthResolver
	defaultResolver ProviderAuthResolver
}

// NewThirdPartyAuthService 创建一个新的第三方平台授权服务。
func NewThirdPartyAuthService(store Store, resolver *AccountResolver) *ThirdPartyAuthService {
	service := &ThirdPartyAuthService{
		store:         store,
		resolver:      resolver,
		providers:     make(map[string]AuthorizationProvider),
		authResolvers: make(map[string][]ProviderAuthResolver),
	}
	if resolver != nil {
		service.defaultResolver = resolver
	}
	return service
}

// RegisterProvider 注册一个授权 provider。
func (s *ThirdPartyAuthService) RegisterProvider(provider AuthorizationProvider) {
	if provider == nil {
		return
	}
	s.providers[provider.ProviderCode()] = provider
}

// RegisterAuthResolver registers a provider-specific runtime authorization resolver.
func (s *ThirdPartyAuthService) RegisterAuthResolver(provider string, resolver ProviderAuthResolver) {
	if s == nil || provider == "" || resolver == nil {
		return
	}
	s.authResolvers[provider] = append(s.authResolvers[provider], resolver)
}

// StartAuthorization 发起某个 provider 的用户授权。
func (s *ThirdPartyAuthService) StartAuthorization(ctx context.Context, req *StartAuthorizationRequest) (string, error) {
	if req == nil {
		return "", fmt.Errorf("authorization request is required")
	}
	if req.UserID == "" {
		return "", fmt.Errorf("user id is required")
	}

	provider, ok := s.providers[req.Provider]
	if !ok {
		return "", fmt.Errorf("authorization provider %s is not registered", req.Provider)
	}

	stateValue, err := generateOAuthState()
	if err != nil {
		return "", err
	}

	state := &OAuthState{
		State:       stateValue,
		UserID:      req.UserID,
		Provider:    req.Provider,
		RedirectURI: req.RedirectURI,
		CreatedAt:   time.Now().UTC(),
	}

	if err := s.store.SaveOAuthState(ctx, state); err != nil {
		return "", err
	}

	return provider.BuildAuthorizationURL(req, state)
}

// HandleAuthorizationCallback 处理 provider 回调并保存授权账户。
func (s *ThirdPartyAuthService) HandleAuthorizationCallback(ctx context.Context, req *AuthorizationCallbackRequest) (*AuthorizationResult, error) {
	if req == nil {
		return nil, fmt.Errorf("authorization callback request is required")
	}
	if req.Code == "" {
		return nil, fmt.Errorf("authorization code is required")
	}
	if req.State == "" {
		return nil, fmt.Errorf("oauth state is required")
	}

	provider, ok := s.providers[req.Provider]
	if !ok {
		return nil, fmt.Errorf("authorization provider %s is not registered", req.Provider)
	}

	state, err := s.store.ConsumeOAuthState(ctx, req.Provider, req.State)
	if err != nil {
		return nil, err
	}

	result, err := provider.CompleteAuthorization(&CompleteAuthorizationRequest{
		State: state,
		Code:  req.Code,
	})
	if err != nil {
		return nil, err
	}

	if result == nil || result.Account == nil || result.Credential == nil {
		return nil, fmt.Errorf("authorization provider returned incomplete result")
	}

	if err := s.store.UpsertAuthorizedAccount(ctx, result.Account, result.Credential); err != nil {
		return nil, err
	}

	if _, err := s.store.GetDefaultAccount(ctx, result.Account.UserID, result.Account.Provider); err != nil {
		if err := s.store.SetDefaultAccount(ctx, &UserProviderBinding{
			UserID:    result.Account.UserID,
			Provider:  result.Account.Provider,
			AccountID: result.Account.ID,
			IsDefault: true,
			Priority:  100,
		}); err != nil {
			return nil, err
		}
	}

	return result, nil
}

// ResolveAccount 解析运行时可用账户。
func (s *ThirdPartyAuthService) ResolveAccount(ctx context.Context, req *ResolveAccountRequest) (*ResolvedAccount, error) {
	return s.resolver.Resolve(ctx, req)
}

// ResolveAuthorization resolves runtime authorization through provider-specific resolvers first.
func (s *ThirdPartyAuthService) ResolveAuthorization(ctx context.Context, req *ResolveAuthorizationRequest) (*ResolvedAuthorization, error) {
	if req == nil {
		return nil, fmt.Errorf("resolve authorization request is required")
	}

	selector := mergeSelector(&ResolveAccountRequest{
		Selector:  req.Selector,
		UserID:    req.UserID,
		Provider:  req.Provider,
		AccountID: req.AccountID,
	})
	if selector.Provider == "" {
		return nil, fmt.Errorf("provider is required")
	}

	request := &ResolveAuthorizationRequest{
		Selector:  selector,
		UserID:    req.UserID,
		Provider:  selector.Provider,
		AccountID: req.AccountID,
	}

	for _, resolver := range s.authResolvers[selector.Provider] {
		resolved, handled, err := resolver.ResolveAuthorization(ctx, request)
		if err != nil {
			return nil, err
		}
		if handled {
			return resolved, nil
		}
	}

	if s.defaultResolver == nil {
		return nil, fmt.Errorf("no authorization resolver registered for provider %s", selector.Provider)
	}

	resolved, handled, err := s.defaultResolver.ResolveAuthorization(ctx, request)
	if err != nil {
		return nil, err
	}
	if !handled {
		return nil, fmt.Errorf("no authorization resolver matched for provider %s", selector.Provider)
	}
	return resolved, nil
}

func generateOAuthState() (string, error) {
	b := make([]byte, oauthStateBytes)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate oauth state: %w", err)
	}
	return hex.EncodeToString(b), nil
}
