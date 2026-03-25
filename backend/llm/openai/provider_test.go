package openai

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	llm "github.com/insmtx/SingerOS/backend/llm"
)

func TestProvider_Name(t *testing.T) {
	p := NewProvider(DefaultConfig("test-key"))
	if p.Name() != "openai" {
		t.Errorf("expected name 'openai', got '%s'", p.Name())
	}
}

func TestProvider_Models(t *testing.T) {
	p := NewProvider(DefaultConfig("test-key"))
	models := p.Models()

	expectedModels := []string{
		"gpt-4",
		"gpt-4-turbo",
		"gpt-4-turbo-preview",
		"gpt-4o",
		"gpt-4o-mini",
		"gpt-3.5-turbo",
		"gpt-3.5-turbo-16k",
	}

	if len(models) != len(expectedModels) {
		t.Errorf("expected %d models, got %d", len(expectedModels), len(models))
	}

	for i, expected := range expectedModels {
		if models[i] != expected {
			t.Errorf("expected model '%s', got '%s'", expected, models[i])
		}
	}
}

func TestProvider_CountTokens(t *testing.T) {
	p := NewProvider(DefaultConfig("test-key"))

	tests := []struct {
		name     string
		text     string
		minCount int
	}{
		{
			name:     "short text",
			text:     "Hello, world!",
			minCount: 1,
		},
		{
			name:     "longer text",
			text:     "This is a longer piece of text that should result in more tokens.",
			minCount: 5,
		},
		{
			name:     "empty text",
			text:     "",
			minCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := p.CountTokens(tt.text)
			if count < tt.minCount {
				t.Errorf("CountTokens(%q) = %d, want at least %d", tt.text, count, tt.minCount)
			}
		})
	}
}

func TestProvider_Generate_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		if r.Header.Get("Authorization") != "Bearer test-api-key" {
			t.Errorf("missing or incorrect Authorization header")
		}

		var req chatCompletionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}

		if req.Model != "gpt-4" {
			t.Errorf("expected model 'gpt-4', got '%s'", req.Model)
		}

		resp := chatCompletionResponse{
			ID:      "test-id",
			Object:  "chat.completion",
			Created: time.Now().Unix(),
			Model:   "gpt-4",
			Choices: []struct {
				Index        int         `json:"index"`
				Message      chatMessage `json:"message"`
				Delta        chatMessage `json:"delta"`
				FinishReason string      `json:"finish_reason"`
			}{
				{
					Index: 0,
					Message: chatMessage{
						Role:    "assistant",
						Content: "Hello! How can I assist you today?",
					},
					FinishReason: "stop",
				},
			},
		}
		resp.Usage.PromptTokens = 10
		resp.Usage.CompletionTokens = 8
		resp.Usage.TotalTokens = 18

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	config := DefaultConfig("test-api-key")
	config.BaseURL = server.URL
	p := NewProvider(config)

	ctx := context.Background()
	req := &llm.GenerateRequest{
		Model: "gpt-4",
		Messages: []llm.Message{
			{Role: "user", Content: "Hello!"},
		},
	}

	resp, err := p.Generate(ctx, req)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if resp.Content != "Hello! How can I assist you today?" {
		t.Errorf("expected content 'Hello! How can I assist you today?', got '%s'", resp.Content)
	}

	if resp.FinishReason != "stop" {
		t.Errorf("expected finish reason 'stop', got '%s'", resp.FinishReason)
	}

	if resp.Usage.PromptTokens != 10 {
		t.Errorf("expected 10 prompt tokens, got %d", resp.Usage.PromptTokens)
	}
}

func TestProvider_Generate_WithRetry(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount < 3 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}

		resp := chatCompletionResponse{
			ID:    "test-id",
			Model: "gpt-4",
			Choices: []struct {
				Index        int         `json:"index"`
				Message      chatMessage `json:"message"`
				Delta        chatMessage `json:"delta"`
				FinishReason string      `json:"finish_reason"`
			}{
				{
					Index: 0,
					Message: chatMessage{
						Role:    "assistant",
						Content: "Success after retry",
					},
					FinishReason: "stop",
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	config := DefaultConfig("test-api-key")
	config.BaseURL = server.URL
	config.MaxRetries = 3
	config.RetryDelay = 10 * time.Millisecond
	p := NewProvider(config)

	ctx := context.Background()
	req := &llm.GenerateRequest{
		Model:    "gpt-4",
		Messages: []llm.Message{{Role: "user", Content: "test"}},
	}

	resp, err := p.Generate(ctx, req)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if callCount != 3 {
		t.Errorf("expected 3 calls, got %d", callCount)
	}

	if resp.Content != "Success after retry" {
		t.Errorf("expected 'Success after retry', got '%s'", resp.Content)
	}
}

func TestProvider_Generate_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
		resp := chatCompletionResponse{
			ID:    "test-id",
			Model: "gpt-4",
			Choices: []struct {
				Index        int         `json:"index"`
				Message      chatMessage `json:"message"`
				Delta        chatMessage `json:"delta"`
				FinishReason string      `json:"finish_reason"`
			}{
				{
					Index:        0,
					Message:      chatMessage{Role: "assistant", Content: "response"},
					FinishReason: "stop",
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	config := DefaultConfig("test-api-key")
	config.BaseURL = server.URL
	p := NewProvider(config)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	req := &llm.GenerateRequest{
		Model:    "gpt-4",
		Messages: []llm.Message{{Role: "user", Content: "test"}},
	}

	_, err := p.Generate(ctx, req)
	if err == nil {
		t.Error("expected error due to context cancellation")
	}
}

func TestProvider_Generate_ErrorResponses(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		responseBody   string
		expectedErr    error
		expectedErrMsg string
	}{
		{
			name:         "invalid API key",
			statusCode:   http.StatusUnauthorized,
			responseBody: `{"error": {"message": "Invalid API key", "type": "invalid_api_key", "code": "invalid_api_key"}}`,
			expectedErr:  ErrInvalidAPIKey,
		},
		{
			name:         "rate limit exceeded",
			statusCode:   http.StatusTooManyRequests,
			responseBody: `{"error": {"message": "Rate limit exceeded", "type": "rate_limit_exceeded", "code": "rate_limit_exceeded"}}`,
			expectedErr:  ErrRateLimitExceeded,
		},
		{
			name:         "quota exceeded",
			statusCode:   http.StatusForbidden,
			responseBody: `{"error": {"message": "Quota exceeded", "type": "insufficient_quota", "code": "insufficient_quota"}}`,
			expectedErr:  ErrQuotaExceeded,
		},
		{
			name:         "server error",
			statusCode:   http.StatusInternalServerError,
			responseBody: `{"error": {"message": "Internal server error", "type": "server_error", "code": "internal_error"}}`,
		},
		{
			name:         "bad request",
			statusCode:   http.StatusBadRequest,
			responseBody: `{"error": {"message": "Bad request", "type": "invalid_request_error", "code": "invalid_request"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			config := DefaultConfig("test-api-key")
			config.BaseURL = server.URL
			config.MaxRetries = 0
			p := NewProvider(config)

			ctx := context.Background()
			req := &llm.GenerateRequest{
				Model:    "gpt-4",
				Messages: []llm.Message{{Role: "user", Content: "test"}},
			}

			_, err := p.Generate(ctx, req)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}

			if tt.expectedErr != nil && !isError(err, tt.expectedErr) {
				t.Errorf("expected error to wrap %v, got %v", tt.expectedErr, err)
			}
		})
	}
}

func TestProvider_Generate_InvalidRequest(t *testing.T) {
	p := NewProvider(DefaultConfig("test-api-key"))

	tests := []struct {
		name        string
		req         *llm.GenerateRequest
		expectedErr error
	}{
		{
			name:        "empty model",
			req:         &llm.GenerateRequest{Messages: []llm.Message{{Role: "user", Content: "test"}}},
			expectedErr: nil,
		},
		{
			name:        "unsupported model",
			req:         &llm.GenerateRequest{Model: "unsupported-model", Messages: []llm.Message{{Role: "user", Content: "test"}}},
			expectedErr: ErrModelNotSupported,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := p.Generate(context.Background(), tt.req)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}

			if tt.expectedErr != nil && !isError(err, tt.expectedErr) {
				t.Errorf("expected error to wrap %v, got %v", tt.expectedErr, err)
			}
		})
	}
}

func TestProvider_Generate_NoAPIKey(t *testing.T) {
	p := NewProvider(DefaultConfig(""))

	req := &llm.GenerateRequest{
		Model:    "gpt-4",
		Messages: []llm.Message{{Role: "user", Content: "test"}},
	}

	_, err := p.Generate(context.Background(), req)
	if !isError(err, ErrInvalidAPIKey) {
		t.Errorf("expected ErrInvalidAPIKey, got %v", err)
	}
}

func TestProvider_GenerateStream_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-api-key" {
			t.Errorf("missing Authorization header")
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		flusher := w.(http.Flusher)

		chunks := []string{
			`data: {"id":"test-id","object":"chat.completion.chunk","created":1234567890,"model":"gpt-4","choices":[{"index":0,"delta":{"role":"assistant"},"finish_reason":null}]}`,
			`data: {"id":"test-id","object":"chat.completion.chunk","created":1234567890,"model":"gpt-4","choices":[{"index":0,"delta":{"content":"Hello"},"finish_reason":null}]}`,
			`data: {"id":"test-id","object":"chat.completion.chunk","created":1234567890,"model":"gpt-4","choices":[{"index":0,"delta":{"content":" World"},"finish_reason":null}]}`,
			`data: {"id":"test-id","object":"chat.completion.chunk","created":1234567890,"model":"gpt-4","choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}`,
			`data: [DONE]`,
		}

		for _, chunk := range chunks {
			w.Write([]byte(chunk + "\n\n"))
			flusher.Flush()
			time.Sleep(10 * time.Millisecond)
		}
	}))
	defer server.Close()

	config := DefaultConfig("test-api-key")
	config.BaseURL = server.URL
	p := NewProvider(config)

	ctx := context.Background()
	req := &llm.GenerateRequest{
		Model:    "gpt-4",
		Messages: []llm.Message{{Role: "user", Content: "test"}},
	}

	stream, err := p.GenerateStream(ctx, req)
	if err != nil {
		t.Fatalf("GenerateStream failed: %v", err)
	}

	var chunks []string
	var doneCount int

	for chunk := range stream {
		if chunk.Done {
			doneCount++
		} else if chunk.Content != "" {
			chunks = append(chunks, chunk.Content)
		}
	}

	if len(chunks) < 2 {
		t.Errorf("expected at least 2 content chunks, got %d", len(chunks))
	}

	if doneCount < 1 {
		t.Errorf("expected at least 1 done chunk, got %d", doneCount)
	}

	content := strings.Join(chunks, "")
	if !strings.Contains(content, "Hello") || !strings.Contains(content, "World") {
		t.Errorf("expected content to contain 'Hello World', got '%s'", content)
	}
}

func TestProvider_GenerateStream_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		flusher := w.(http.Flusher)
		for i := 0; i < 100; i++ {
			w.Write([]byte(`data: {"id":"test-id","model":"gpt-4","choices":[{"index":0,"delta":{"content":"test"},"finish_reason":null}]}` + "\n\n"))
			flusher.Flush()
			time.Sleep(50 * time.Millisecond)
		}
	}))
	defer server.Close()

	config := DefaultConfig("test-api-key")
	config.BaseURL = server.URL
	p := NewProvider(config)

	ctx, cancel := context.WithCancel(context.Background())

	stream, err := p.GenerateStream(ctx, req)
	if err != nil {
		t.Fatalf("GenerateStream failed: %v", err)
	}

	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	for range stream {
	}
}

func TestProvider_GenerateStream_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"error": {"message": "Invalid API key", "type": "invalid_api_key", "code": "invalid_api_key"}}`))
	}))
	defer server.Close()

	config := DefaultConfig("test-api-key")
	config.BaseURL = server.URL
	p := NewProvider(config)

	ctx := context.Background()
	req := &llm.GenerateRequest{
		Model:    "gpt-4",
		Messages: []llm.Message{{Role: "user", Content: "test"}},
	}

	stream, err := p.GenerateStream(ctx, req)
	if err != nil {
		t.Fatalf("GenerateStream failed: %v", err)
	}

	for chunk := range stream {
		if chunk.Done && strings.Contains(chunk.Content, "error") {
			return
		}
	}
}

func TestProvider_GenerateStream_InvalidRequest(t *testing.T) {
	p := NewProvider(DefaultConfig("test-api-key"))

	req := &llm.GenerateRequest{
		Model: "",
	}

	_, err := p.GenerateStream(context.Background(), req)
	if err == nil {
		t.Error("expected error for invalid request")
	}
}

func TestProvider_Generate_NetworkError(t *testing.T) {
	p := NewProvider(&Config{
		APIKey:     "test-key",
		BaseURL:    "http://localhost:99999",
		MaxRetries: 1,
		RetryDelay: 1 * time.Millisecond,
	})

	ctx := context.Background()
	req := &llm.GenerateRequest{
		Model:    "gpt-4",
		Messages: []llm.Message{{Role: "user", Content: "test"}},
	}

	_, err := p.Generate(ctx, req)
	if err == nil {
		t.Error("expected error for network failure")
	}

	if !isError(err, ErrNetworkError) && !isError(err, ErrAPITimeout) {
		t.Errorf("expected network or timeout error, got %v", err)
	}
}

func TestAPIError_Error(t *testing.T) {
	err := &APIError{
		Code:       "test_code",
		Message:    "test message",
		HTTPStatus: 400,
		Type:       "invalid_request_error",
	}

	expected := "OpenAI API error: test message (code: test_code, status: 400)"
	if err.Error() != expected {
		t.Errorf("expected error string '%s', got '%s'", expected, err.Error())
	}
}

func TestAPIError_Unwrap(t *testing.T) {
	tests := []struct {
		name           string
		errType        string
		expectedUnwrap error
	}{
		{"invalid_api_key", "invalid_api_key", ErrInvalidAPIKey},
		{"rate_limit_exceeded", "rate_limit_exceeded", ErrRateLimitExceeded},
		{"insufficient_quota", "insufficient_quota", ErrQuotaExceeded},
		{"unknown_type", "unknown_type", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &APIError{Type: tt.errType}
			unwrapped := err.Unwrap()
			if unwrapped != tt.expectedUnwrap {
				t.Errorf("expected unwrap to return %v, got %v", tt.expectedUnwrap, unwrapped)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig("test-key")

	if config.APIKey != "test-key" {
		t.Errorf("expected API key 'test-key', got '%s'", config.APIKey)
	}

	if config.BaseURL != "https://api.openai.com/v1" {
		t.Errorf("expected base URL 'https://api.openai.com/v1', got '%s'", config.BaseURL)
	}

	if config.MaxRetries != 3 {
		t.Errorf("expected 3 max retries, got %d", config.MaxRetries)
	}

	if config.HTTPClient == nil {
		t.Error("expected HTTPClient to be non-nil")
	}
}

func TestNewProvider(t *testing.T) {
	t.Run("with all fields set", func(t *testing.T) {
		config := &Config{
			APIKey:     "test-key",
			BaseURL:    "https://custom.api.com",
			MaxRetries: 5,
			RetryDelay: 2 * time.Second,
		}
		p := NewProvider(config)

		if p.Name() != "openai" {
			t.Errorf("expected provider name 'openai', got '%s'", p.Name())
		}
	})

	t.Run("with nil HTTPClient", func(t *testing.T) {
		config := &Config{
			APIKey:     "test-key",
			BaseURL:    "https://api.openai.com/v1",
			MaxRetries: 3,
		}
		p := NewProvider(config)

		if p.httpClient == nil {
			t.Error("expected HTTPClient to be initialized")
		}
	})

	t.Run("with zero MaxRetries", func(t *testing.T) {
		config := &Config{
			APIKey:  "test-key",
			BaseURL: "https://api.openai.com/v1",
		}
		p := NewProvider(config)

		if p.config.MaxRetries != 3 {
			t.Errorf("expected MaxRetries to be 3, got %d", p.config.MaxRetries)
		}
	})
}

func TestTokenizer_Count(t *testing.T) {
	tokenizer := NewTokenizer()

	t.Run("short text", func(t *testing.T) {
		count := tokenizer.Count("Hello")
		if count < 1 {
			t.Errorf("expected at least 1 token, got %d", count)
		}
	})

	t.Run("empty text", func(t *testing.T) {
		count := tokenizer.Count("")
		if count != 1 {
			t.Errorf("expected 1 token for empty text, got %d", count)
		}
	})

	t.Run("caching", func(t *testing.T) {
		text := "This is a test string for caching"
		count1 := tokenizer.Count(text)
		count2 := tokenizer.Count(text)
		if count1 != count2 {
			t.Errorf("cached value should be consistent: %d != %d", count1, count2)
		}
	})
}

func TestProvider_ValidateRequest(t *testing.T) {
	p := NewProvider(DefaultConfig("test-key"))

	tests := []struct {
		name    string
		req     *llm.GenerateRequest
		wantErr bool
	}{
		{
			name:    "valid classic GPT-4",
			req:     &llm.GenerateRequest{Model: "gpt-4", Messages: []llm.Message{{Role: "user", Content: "Hello"}}},
			wantErr: false,
		},
		{
			name:    "valid GPT-4o",
			req:     &llm.GenerateRequest{Model: "gpt-4o", Messages: []llm.Message{{Role: "user", Content: "test"}}},
			wantErr: false,
		},
		{
			name:    "valid GPT-4o-mini",
			req:     &llm.GenerateRequest{Model: "gpt-4o-mini", Messages: []llm.Message{{Role: "user", Content: "test"}}},
			wantErr: false,
		},
		{
			name:    "valid with version suffix",
			req:     &llm.GenerateRequest{Model: "gpt-4-0613", Messages: []llm.Message{{Role: "user", Content: "test"}}},
			wantErr: false,
		},
		{
			name:    "valid GPT-3.5 with version",
			req:     &llm.GenerateRequest{Model: "gpt-3.5-turbo-0301", Messages: []llm.Message{{Role: "user", Content: "test"}}},
			wantErr: false,
		},
		{
			name:    "invalid model",
			req:     &llm.GenerateRequest{Model: "unsupported-model", Messages: []llm.Message{{Role: "user", Content: "test"}}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := p.validateRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestProvider_ShouldRetry(t *testing.T) {
	p := NewProvider(DefaultConfig("test-key"))

	tests := []struct {
		name      string
		err       error
		shouldTry bool
	}{
		{
			name:      "rate limit error",
			err:       &APIError{HTTPStatus: http.StatusTooManyRequests},
			shouldTry: true,
		},
		{
			name:      "server error",
			err:       &APIError{HTTPStatus: http.StatusInternalServerError},
			shouldTry: true,
		},
		{
			name:      "bad gateway",
			err:       &APIError{HTTPStatus: http.StatusBadGateway},
			shouldTry: true,
		},
		{
			name:      "client error",
			err:       &APIError{HTTPStatus: http.StatusBadRequest},
			shouldTry: false,
		},
		{
			name:      "unauthorized",
			err:       &APIError{HTTPStatus: http.StatusUnauthorized, Type: "invalid_api_key"},
			shouldTry: false,
		},
		{
			name:      "network error",
			err:       ErrNetworkError,
			shouldTry: true,
		},
		{
			name:      "timeout error",
			err:       ErrAPITimeout,
			shouldTry: true,
		},
		{
			name:      "other error",
			err:       errors.New("some error"),
			shouldTry: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := p.shouldRetry(tt.err)
			if result != tt.shouldTry {
				t.Errorf("shouldRetry() = %v, want %v", result, tt.shouldTry)
			}
		})
	}
}

func TestConvertMessages(t *testing.T) {
	msgs := []llm.Message{
		{Role: "system", Content: "You are a helpful assistant."},
		{Role: "user", Content: "Hello!"},
		{Role: "assistant", Content: "Hi! How can I help?"},
	}

	result := convertMessages(msgs)

	if len(result) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(result))
	}

	for i, msg := range result {
		if msg.Role != msgs[i].Role {
			t.Errorf("message %d: expected role '%s', got '%s'", i, msgs[i].Role, msg.Role)
		}
		if msg.Content != msgs[i].Content {
			t.Errorf("message %d: expected content '%s', got '%s'", i, msgs[i].Content, msg.Content)
		}
	}
}

func TestConvertUsage(t *testing.T) {
	usage := struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	}{
		PromptTokens:     100,
		CompletionTokens: 50,
		TotalTokens:      150,
	}

	result := convertUsage(usage)

	if result.PromptTokens != 100 {
		t.Errorf("expected 100 prompt tokens, got %d", result.PromptTokens)
	}
	if result.CompletionTokens != 50 {
		t.Errorf("expected 50 completion tokens, got %d", result.CompletionTokens)
	}
	if result.TotalTokens != 150 {
		t.Errorf("expected 150 total tokens, got %d", result.TotalTokens)
	}
}

func TestProvider_Generate_WithMaxTokensAndTemperature(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req chatCompletionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}

		if req.MaxTokens != 100 {
			t.Errorf("expected MaxTokens 100, got %d", req.MaxTokens)
		}

		if req.Temperature != 0.7 {
			t.Errorf("expected Temperature 0.7, got %f", req.Temperature)
		}

		if len(req.Stop) != 1 || req.Stop[0] != "END" {
			t.Errorf("expected Stop ['END'], got %v", req.Stop)
		}

		resp := chatCompletionResponse{
			ID:    "test-id",
			Model: "gpt-4",
			Choices: []struct {
				Index        int         `json:"index"`
				Message      chatMessage `json:"message"`
				Delta        chatMessage `json:"delta"`
				FinishReason string      `json:"finish_reason"`
			}{
				{Index: 0, Message: chatMessage{Role: "assistant", Content: "Response"}, FinishReason: "stop"},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	config := DefaultConfig("test-key")
	config.BaseURL = server.URL
	p := NewProvider(config)

	ctx := context.Background()
	req := &llm.GenerateRequest{
		Model:       "gpt-4",
		Messages:    []llm.Message{{Role: "user", Content: "test"}},
		MaxTokens:   100,
		Temperature: 0.7,
		Stop:        []string{"END"},
	}

	resp, err := p.Generate(ctx, req)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if resp.Content != "Response" {
		t.Errorf("expected 'Response', got '%s'", resp.Content)
	}
}

func isError(err error, target error) bool {
	for err != nil {
		if err == target {
			return true
		}
		if unwrapper, ok := err.(interface{ Unwrap() error }); ok {
			err = unwrapper.Unwrap()
		} else {
			break
		}
	}
	return false
}

var req = &llm.GenerateRequest{
	Model:    "gpt-4",
	Messages: []llm.Message{{Role: "user", Content: "test"}},
}

func TestProvider_GenerateStream_InvalidAPIKey(t *testing.T) {
	p := NewProvider(DefaultConfig(""))

	_, err := p.GenerateStream(context.Background(), req)
	if !isError(err, ErrInvalidAPIKey) {
		t.Errorf("expected ErrInvalidAPIKey, got %v", err)
	}
}

func TestProvider_CountTokens_Consistency(t *testing.T) {
	p := NewProvider(DefaultConfig("test-key"))

	text := "This is a test string that should produce consistent token counts"
	count1 := p.CountTokens(text)
	count2 := p.CountTokens(text)

	if count1 != count2 {
		t.Errorf("token count should be consistent: %d != %d", count1, count2)
	}
}

func dummyServer(jsonStr string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.Copy(io.Discard, r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(jsonStr))
	}))
}
