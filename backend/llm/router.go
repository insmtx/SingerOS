package llm

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

var (
	ErrNoProviderRegistered = errors.New("no provider registered")
	ErrProviderNotFound     = errors.New("provider not found")
	ErrModelNotFound        = errors.New("model not found")
	ErrNoFallbackAvailable  = errors.New("no fallback provider available")
	ErrQuotaExceeded        = errors.New("token quota exceeded")
)

type ProviderInfo struct {
	Provider  Provider
	Priority  int
	Models    []string
	IsEnabled bool
}

type QuotaChecker interface {
	CheckQuota(provider string, model string, tokens int) error
}

type RouterConfig struct {
	EnableFallback bool
	MaxRetries     int
	QuotaChecker   QuotaChecker
}

type Router struct {
	mu           sync.RWMutex
	providers    map[string]*ProviderInfo
	modelMap     map[string][]string
	config       RouterConfig
	quotaChecker QuotaChecker
}

func NewRouter(config RouterConfig) *Router {
	return &Router{
		providers:    make(map[string]*ProviderInfo),
		modelMap:     make(map[string][]string),
		config:       config,
		quotaChecker: config.QuotaChecker,
	}
}

func (r *Router) RegisterProvider(provider Provider, priority int, enabled bool) error {
	if provider == nil {
		return errors.New("provider cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	name := provider.Name()
	r.providers[name] = &ProviderInfo{
		Provider:  provider,
		Priority:  priority,
		Models:    provider.Models(),
		IsEnabled: enabled,
	}

	for _, model := range provider.Models() {
		r.modelMap[model] = append(r.modelMap[model], name)
	}

	return nil
}

func (r *Router) UnregisterProvider(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	info, exists := r.providers[name]
	if !exists {
		return
	}

	for _, model := range info.Models {
		providers := r.modelMap[model]
		for i, p := range providers {
			if p == name {
				r.modelMap[model] = append(providers[:i], providers[i+1:]...)
				break
			}
		}
		if len(r.modelMap[model]) == 0 {
			delete(r.modelMap, model)
		}
	}

	delete(r.providers, name)
}

func (r *Router) GetProvider(name string) (Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	info, exists := r.providers[name]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrProviderNotFound, name)
	}
	if !info.IsEnabled {
		return nil, fmt.Errorf("%w: %s (disabled)", ErrProviderNotFound, name)
	}

	return info.Provider, nil
}

func (r *Router) Route(model string) (Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	providerNames, exists := r.modelMap[model]
	if !exists || len(providerNames) == 0 {
		return nil, fmt.Errorf("%w: %s", ErrModelNotFound, model)
	}

	for _, name := range providerNames {
		info := r.providers[name]
		if info.IsEnabled {
			return info.Provider, nil
		}
	}

	return nil, fmt.Errorf("%w: no enabled provider for model %s", ErrModelNotFound, model)
}

func (r *Router) RouteWithFallback(model string) (Provider, string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	providerNames, exists := r.modelMap[model]
	if !exists || len(providerNames) == 0 {
		return nil, "", fmt.Errorf("%w: %s", ErrModelNotFound, model)
	}

	for _, name := range providerNames {
		info := r.providers[name]
		if info.IsEnabled {
			return info.Provider, name, nil
		}
	}

	return nil, "", fmt.Errorf("%w: no enabled provider for model %s", ErrModelNotFound, model)
}

func (r *Router) Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error) {
	if req == nil {
		return nil, errors.New("request cannot be nil")
	}

	provider, name, err := r.RouteWithFallback(req.Model)
	if err != nil {
		return nil, err
	}

	if r.quotaChecker != nil {
		estimatedTokens := 0
		if err := r.quotaChecker.CheckQuota(name, req.Model, estimatedTokens); err != nil {
			if !r.config.EnableFallback {
				return nil, err
			}
		}
	}

	resp, err := provider.Generate(ctx, req)
	if err != nil {
		if r.config.EnableFallback {
			return r.tryFallback(ctx, req, name)
		}
		return nil, err
	}

	return resp, nil
}

func (r *Router) tryFallback(ctx context.Context, req *GenerateRequest, failedProvider string) (*GenerateResponse, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	providerNames, exists := r.modelMap[req.Model]
	if !exists {
		return nil, ErrNoFallbackAvailable
	}

	for _, name := range providerNames {
		if name == failedProvider {
			continue
		}

		info := r.providers[name]
		if !info.IsEnabled {
			continue
		}

		resp, err := info.Provider.Generate(ctx, req)
		if err == nil {
			return resp, nil
		}
	}

	return nil, ErrNoFallbackAvailable
}

func (r *Router) GenerateStream(ctx context.Context, req *GenerateRequest) (<-chan StreamChunk, error) {
	if req == nil {
		return nil, errors.New("request cannot be nil")
	}

	provider, err := r.Route(req.Model)
	if err != nil {
		return nil, err
	}

	return provider.GenerateStream(ctx, req)
}

func (r *Router) ListProviders() []ProviderInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]ProviderInfo, 0, len(r.providers))
	for _, info := range r.providers {
		result = append(result, *info)
	}

	return result
}

func (r *Router) ListModels() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	models := make([]string, 0, len(r.modelMap))
	for model := range r.modelMap {
		models = append(models, model)
	}

	return models
}

func (r *Router) EnableProvider(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	info, exists := r.providers[name]
	if !exists {
		return fmt.Errorf("%w: %s", ErrProviderNotFound, name)
	}

	info.IsEnabled = true
	return nil
}

func (r *Router) DisableProvider(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	info, exists := r.providers[name]
	if !exists {
		return fmt.Errorf("%w: %s", ErrProviderNotFound, name)
	}

	info.IsEnabled = false
	return nil
}

func (r *Router) SetQuotaChecker(checker QuotaChecker) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.quotaChecker = checker
}
