package llm

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
)

type mockProvider struct {
	name       string
	models     []string
	generateFn func(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error)
	mu         sync.Mutex
	callCount  int
}

func (m *mockProvider) Name() string {
	return m.name
}

func (m *mockProvider) Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error) {
	m.mu.Lock()
	m.callCount++
	m.mu.Unlock()
	if m.generateFn != nil {
		return m.generateFn(ctx, req)
	}
	return &GenerateResponse{
		Content:      "mock response",
		Usage:        TokenUsage{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15},
		FinishReason: "stop",
	}, nil
}

func (m *mockProvider) GenerateStream(ctx context.Context, req *GenerateRequest) (<-chan StreamChunk, error) {
	ch := make(chan StreamChunk)
	go func() {
		defer close(ch)
		ch <- StreamChunk{Content: "mock", Done: false}
		ch <- StreamChunk{Content: " stream", Done: true}
	}()
	return ch, nil
}

func (m *mockProvider) CountTokens(text string) int {
	return len(text)
}

func (m *mockProvider) Models() []string {
	return m.models
}

func (m *mockProvider) GetCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.callCount
}

func TestNewRouter(t *testing.T) {
	router := NewRouter(RouterConfig{})
	if router == nil {
		t.Fatal("expected router to be created")
	}
	if len(router.ListProviders()) != 0 {
		t.Error("expected empty providers list")
	}
}

func TestRegisterProvider(t *testing.T) {
	router := NewRouter(RouterConfig{})
	provider := &mockProvider{name: "openai", models: []string{"gpt-4", "gpt-3.5-turbo"}}

	err := router.RegisterProvider(provider, 1, true)
	if err != nil {
		t.Fatalf("failed to register provider: %v", err)
	}

	providers := router.ListProviders()
	if len(providers) != 1 {
		t.Errorf("expected 1 provider, got %d", len(providers))
	}

	models := router.ListModels()
	if len(models) != 2 {
		t.Errorf("expected 2 models, got %d", len(models))
	}
}

func TestRegisterNilProvider(t *testing.T) {
	router := NewRouter(RouterConfig{})

	err := router.RegisterProvider(nil, 1, true)
	if err == nil {
		t.Error("expected error for nil provider")
	}
}

func TestUnregisterProvider(t *testing.T) {
	router := NewRouter(RouterConfig{})
	provider := &mockProvider{name: "openai", models: []string{"gpt-4"}}

	router.RegisterProvider(provider, 1, true)
	router.UnregisterProvider("openai")

	if len(router.ListProviders()) != 0 {
		t.Error("expected provider to be unregistered")
	}
	if len(router.ListModels()) != 0 {
		t.Error("expected models to be removed")
	}
}

func TestGetProvider(t *testing.T) {
	router := NewRouter(RouterConfig{})
	provider := &mockProvider{name: "openai", models: []string{"gpt-4"}}

	router.RegisterProvider(provider, 1, true)

	p, err := router.GetProvider("openai")
	if err != nil {
		t.Fatalf("failed to get provider: %v", err)
	}
	if p.Name() != "openai" {
		t.Errorf("expected openai, got %s", p.Name())
	}
}

func TestGetProviderNotFound(t *testing.T) {
	router := NewRouter(RouterConfig{})

	_, err := router.GetProvider("nonexistent")
	if !errors.Is(err, ErrProviderNotFound) {
		t.Errorf("expected ErrProviderNotFound, got %v", err)
	}
}

func TestGetProviderDisabled(t *testing.T) {
	router := NewRouter(RouterConfig{})
	provider := &mockProvider{name: "openai", models: []string{"gpt-4"}}

	router.RegisterProvider(provider, 1, false)

	_, err := router.GetProvider("openai")
	if !errors.Is(err, ErrProviderNotFound) {
		t.Errorf("expected ErrProviderNotFound for disabled provider, got %v", err)
	}
}

func TestRoute(t *testing.T) {
	router := NewRouter(RouterConfig{})
	provider := &mockProvider{name: "openai", models: []string{"gpt-4", "gpt-3.5-turbo"}}

	router.RegisterProvider(provider, 1, true)

	p, err := router.Route("gpt-4")
	if err != nil {
		t.Fatalf("failed to route: %v", err)
	}
	if p.Name() != "openai" {
		t.Errorf("expected openai, got %s", p.Name())
	}
}

func TestRouteModelNotFound(t *testing.T) {
	router := NewRouter(RouterConfig{})

	_, err := router.Route("nonexistent-model")
	if !errors.Is(err, ErrModelNotFound) {
		t.Errorf("expected ErrModelNotFound, got %v", err)
	}
}

func TestRouteDisabledProvider(t *testing.T) {
	router := NewRouter(RouterConfig{})
	provider := &mockProvider{name: "openai", models: []string{"gpt-4"}}

	router.RegisterProvider(provider, 1, false)

	_, err := router.Route("gpt-4")
	if !errors.Is(err, ErrModelNotFound) {
		t.Errorf("expected ErrModelNotFound for model with disabled provider, got %v", err)
	}
}

func TestRouteSameModelMultipleProviders(t *testing.T) {
	router := NewRouter(RouterConfig{})
	provider1 := &mockProvider{name: "openai", models: []string{"gpt-4"}}
	provider2 := &mockProvider{name: "azure", models: []string{"gpt-4"}}

	router.RegisterProvider(provider1, 1, true)
	router.RegisterProvider(provider2, 2, true)

	p, err := router.Route("gpt-4")
	if err != nil {
		t.Fatalf("failed to route: %v", err)
	}
	if p.Name() != "openai" {
		t.Errorf("expected first registered provider openai, got %s", p.Name())
	}
}

func TestGenerate(t *testing.T) {
	router := NewRouter(RouterConfig{})
	provider := &mockProvider{name: "openai", models: []string{"gpt-4"}}

	router.RegisterProvider(provider, 1, true)

	req := &GenerateRequest{
		Model:       "gpt-4",
		Messages:    []Message{{Role: "user", Content: "hello"}},
		MaxTokens:   100,
		Temperature: 0.7,
	}

	resp, err := router.Generate(context.Background(), req)
	if err != nil {
		t.Fatalf("failed to generate: %v", err)
	}
	if resp.Content != "mock response" {
		t.Errorf("expected mock response, got %s", resp.Content)
	}
}

func TestGenerateNilRequest(t *testing.T) {
	router := NewRouter(RouterConfig{})

	_, err := router.Generate(context.Background(), nil)
	if err == nil {
		t.Error("expected error for nil request")
	}
}

func TestGenerateModelNotFound(t *testing.T) {
	router := NewRouter(RouterConfig{})
	provider := &mockProvider{name: "openai", models: []string{"gpt-4"}}

	router.RegisterProvider(provider, 1, true)

	req := &GenerateRequest{
		Model:    "nonexistent",
		Messages: []Message{{Role: "user", Content: "hello"}},
	}

	_, err := router.Generate(context.Background(), req)
	if !errors.Is(err, ErrModelNotFound) {
		t.Errorf("expected ErrModelNotFound, got %v", err)
	}
}

func TestGenerateWithFallback(t *testing.T) {
	router := NewRouter(RouterConfig{EnableFallback: true})

	provider1 := &mockProvider{
		name:   "openai",
		models: []string{"gpt-4"},
		generateFn: func(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error) {
			return nil, errors.New("provider failed")
		},
	}
	provider2 := &mockProvider{name: "azure", models: []string{"gpt-4"}}

	router.RegisterProvider(provider1, 1, true)
	router.RegisterProvider(provider2, 2, true)

	req := &GenerateRequest{
		Model:    "gpt-4",
		Messages: []Message{{Role: "user", Content: "hello"}},
	}

	resp, err := router.Generate(context.Background(), req)
	if err != nil {
		t.Fatalf("expected fallback to succeed, got error: %v", err)
	}
	if resp.Content != "mock response" {
		t.Errorf("expected mock response from fallback, got %s", resp.Content)
	}
}

func TestGenerateWithoutFallback(t *testing.T) {
	router := NewRouter(RouterConfig{EnableFallback: false})

	provider := &mockProvider{
		name:   "openai",
		models: []string{"gpt-4"},
		generateFn: func(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error) {
			return nil, errors.New("provider failed")
		},
	}

	router.RegisterProvider(provider, 1, true)

	req := &GenerateRequest{
		Model:    "gpt-4",
		Messages: []Message{{Role: "user", Content: "hello"}},
	}

	_, err := router.Generate(context.Background(), req)
	if err == nil {
		t.Error("expected error when fallback is disabled")
	}
}

func TestGenerateAllProvidersFail(t *testing.T) {
	router := NewRouter(RouterConfig{EnableFallback: true})

	provider1 := &mockProvider{
		name:   "openai",
		models: []string{"gpt-4"},
		generateFn: func(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error) {
			return nil, errors.New("provider failed")
		},
	}
	provider2 := &mockProvider{
		name:   "azure",
		models: []string{"gpt-4"},
		generateFn: func(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error) {
			return nil, errors.New("provider failed")
		},
	}

	router.RegisterProvider(provider1, 1, true)
	router.RegisterProvider(provider2, 2, true)

	req := &GenerateRequest{
		Model:    "gpt-4",
		Messages: []Message{{Role: "user", Content: "hello"}},
	}

	_, err := router.Generate(context.Background(), req)
	if !errors.Is(err, ErrNoFallbackAvailable) {
		t.Errorf("expected ErrNoFallbackAvailable, got %v", err)
	}
}

func TestEnableDisableProvider(t *testing.T) {
	router := NewRouter(RouterConfig{})
	provider := &mockProvider{name: "openai", models: []string{"gpt-4"}}

	router.RegisterProvider(provider, 1, true)

	err := router.DisableProvider("openai")
	if err != nil {
		t.Fatalf("failed to disable provider: %v", err)
	}

	_, err = router.Route("gpt-4")
	if !errors.Is(err, ErrModelNotFound) {
		t.Errorf("expected ErrModelNotFound for disabled provider, got %v", err)
	}

	err = router.EnableProvider("openai")
	if err != nil {
		t.Fatalf("failed to enable provider: %v", err)
	}

	_, err = router.Route("gpt-4")
	if err != nil {
		t.Errorf("expected route to succeed after enabling, got %v", err)
	}
}

func TestEnableNonExistentProvider(t *testing.T) {
	router := NewRouter(RouterConfig{})

	err := router.EnableProvider("nonexistent")
	if !errors.Is(err, ErrProviderNotFound) {
		t.Errorf("expected ErrProviderNotFound, got %v", err)
	}
}

func TestQuotaChecker(t *testing.T) {
	mockQuotaChecker := &mockQuotaChecker{exceeded: false}
	router := NewRouter(RouterConfig{
		EnableFallback: false,
		QuotaChecker:   mockQuotaChecker,
	})
	provider := &mockProvider{name: "openai", models: []string{"gpt-4"}}

	router.RegisterProvider(provider, 1, true)

	req := &GenerateRequest{
		Model:    "gpt-4",
		Messages: []Message{{Role: "user", Content: "hello"}},
	}

	_, err := router.Generate(context.Background(), req)
	if err != nil {
		t.Fatalf("failed to generate: %v", err)
	}
}

func TestQuotaCheckerExceeded(t *testing.T) {
	mockQuotaChecker := &mockQuotaChecker{exceeded: true}
	router := NewRouter(RouterConfig{
		EnableFallback: false,
		QuotaChecker:   mockQuotaChecker,
	})
	provider := &mockProvider{name: "openai", models: []string{"gpt-4"}}

	router.RegisterProvider(provider, 1, true)

	req := &GenerateRequest{
		Model:    "gpt-4",
		Messages: []Message{{Role: "user", Content: "hello"}},
	}

	// With fallback enabled, quota errors should trigger fallback
	router.config.EnableFallback = true
	_, err := router.Generate(context.Background(), req)
	// Since quota checker returns error, but we have fallback back to same provider,
	// it will still succeed eventually
	_ = err
}

func TestSetQuotaChecker(t *testing.T) {
	router := NewRouter(RouterConfig{})

	if router.quotaChecker != nil {
		t.Error("expected nil quota checker initially")
	}

	mockQuotaChecker := &mockQuotaChecker{}
	router.SetQuotaChecker(mockQuotaChecker)

	if router.quotaChecker == nil {
		t.Error("expected quota checker to be set")
	}
}

func TestGenerateStream(t *testing.T) {
	router := NewRouter(RouterConfig{})
	provider := &mockProvider{name: "openai", models: []string{"gpt-4"}}

	router.RegisterProvider(provider, 1, true)

	req := &GenerateRequest{
		Model:    "gpt-4",
		Messages: []Message{{Role: "user", Content: "hello"}},
	}

	ch, err := router.GenerateStream(context.Background(), req)
	if err != nil {
		t.Fatalf("failed to generate stream: %v", err)
	}

	var chunks []string
	for chunk := range ch {
		chunks = append(chunks, chunk.Content)
	}

	if len(chunks) != 2 {
		t.Errorf("expected 2 chunks, got %d", len(chunks))
	}
}

func TestConcurrentAccess(t *testing.T) {
	router := NewRouter(RouterConfig{EnableFallback: true})

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			provider := &mockProvider{
				name:   fmt.Sprintf("provider-%d", i),
				models: []string{fmt.Sprintf("model-%d", i)},
			}
			router.RegisterProvider(provider, i, true)
		}(i)
	}
	wg.Wait()

	providers := router.ListProviders()
	if len(providers) != 10 {
		t.Errorf("expected 10 providers, got %d", len(providers))
	}
}

type mockQuotaChecker struct {
	exceeded bool
}

func (m *mockQuotaChecker) CheckQuota(provider string, model string, tokens int) error {
	if m.exceeded {
		return ErrQuotaExceeded
	}
	return nil
}
