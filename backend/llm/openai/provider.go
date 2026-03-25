package openai

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	llm "github.com/insmtx/SingerOS/backend/llm"
)

var (
	ErrInvalidAPIKey     = errors.New("invalid API key")
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
	ErrQuotaExceeded     = errors.New("quota exceeded")
	ErrModelNotSupported = errors.New("model not supported")
	ErrNetworkError      = errors.New("network error")
	ErrAPITimeout        = errors.New("API timeout")
	ErrInvalidResponse   = errors.New("invalid response from API")
)

type APIError struct {
	Code       string
	Message    string
	HTTPStatus int
	Type       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("OpenAI API error: %s (code: %s, status: %d)", e.Message, e.Code, e.HTTPStatus)
}

func (e *APIError) Unwrap() error {
	switch e.Type {
	case "invalid_api_key":
		return ErrInvalidAPIKey
	case "rate_limit_exceeded":
		return ErrRateLimitExceeded
	case "insufficient_quota":
		return ErrQuotaExceeded
	}
	return nil
}

type Config struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
	MaxRetries int
	RetryDelay time.Duration
}

func DefaultConfig(apiKey string) *Config {
	return &Config{
		APIKey:     apiKey,
		BaseURL:    "https://api.openai.com/v1",
		MaxRetries: 3,
		RetryDelay: 1 * time.Second,
		HTTPClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

type Provider struct {
	config     *Config
	httpClient *http.Client
	tokenizer  *Tokenizer
}

func NewProvider(config *Config) *Provider {
	if config.HTTPClient == nil {
		config.HTTPClient = &http.Client{
			Timeout: 120 * time.Second,
		}
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = 1 * time.Second
	}
	return &Provider{
		config:     config,
		httpClient: config.HTTPClient,
		tokenizer:  NewTokenizer(),
	}
}

func (p *Provider) Name() string {
	return "openai"
}

func (p *Provider) Models() []string {
	return []string{
		"gpt-4",
		"gpt-4-turbo",
		"gpt-4-turbo-preview",
		"gpt-4o",
		"gpt-4o-mini",
		"gpt-3.5-turbo",
		"gpt-3.5-turbo-16k",
	}
}

func (p *Provider) CountTokens(text string) int {
	return p.tokenizer.Count(text)
}

func (p *Provider) Generate(ctx context.Context, req *llm.GenerateRequest) (*llm.GenerateResponse, error) {
	if err := p.validateRequest(req); err != nil {
		return nil, err
	}

	apiReq := &chatCompletionRequest{
		Model:       req.Model,
		Messages:    convertMessages(req.Messages),
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		Stop:        req.Stop,
		Stream:      false,
	}

	var resp *chatCompletionResponse
	var err error

	for attempt := 0; attempt <= p.config.MaxRetries; attempt++ {
		resp, err = p.doChatCompletion(ctx, apiReq)
		if err == nil {
			break
		}

		if !p.shouldRetry(err) {
			return nil, err
		}

		if attempt < p.config.MaxRetries {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(p.config.RetryDelay):
			}
		}
	}

	if err != nil {
		return nil, err
	}

	return &llm.GenerateResponse{
		Content:      resp.Choices[0].Message.Content,
		Usage:        convertUsage(resp.Usage),
		FinishReason: resp.Choices[0].FinishReason,
	}, nil
}

func (p *Provider) GenerateStream(ctx context.Context, req *llm.GenerateRequest) (<-chan llm.StreamChunk, error) {
	if err := p.validateRequest(req); err != nil {
		return nil, err
	}

	apiReq := &chatCompletionRequest{
		Model:       req.Model,
		Messages:    convertMessages(req.Messages),
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		Stop:        req.Stop,
		Stream:      true,
	}

	streamChan := make(chan llm.StreamChunk, 100)

	go func() {
		defer close(streamChan)
		if err := p.streamChatCompletion(ctx, apiReq, streamChan); err != nil {
			streamChan <- llm.StreamChunk{
				Content: err.Error(),
				Done:    true,
			}
		}
	}()

	return streamChan, nil
}

func (p *Provider) validateRequest(req *llm.GenerateRequest) error {
	if p.config.APIKey == "" {
		return ErrInvalidAPIKey
	}

	if req.Model == "" {
		return fmt.Errorf("model is required")
	}

	supported := false
	for _, m := range p.Models() {
		if m == req.Model || strings.HasPrefix(req.Model, m) {
			supported = true
			break
		}
	}
	if !supported {
		return fmt.Errorf("%w: %s", ErrModelNotSupported, req.Model)
	}

	return nil
}

type chatCompletionRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Temperature float64       `json:"temperature,omitempty"`
	Stop        []string      `json:"stop,omitempty"`
	Stream      bool          `json:"stream,omitempty"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatCompletionResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int         `json:"index"`
		Message      chatMessage `json:"message"`
		Delta        chatMessage `json:"delta"`
		FinishReason string      `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type streamChunk struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int         `json:"index"`
		Delta        chatMessage `json:"delta"`
		FinishReason string      `json:"finish_reason"`
	} `json:"choices"`
}

func convertMessages(msgs []llm.Message) []chatMessage {
	result := make([]chatMessage, len(msgs))
	for i, m := range msgs {
		result[i] = chatMessage{
			Role:    m.Role,
			Content: m.Content,
		}
	}
	return result
}

func convertUsage(usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}) llm.TokenUsage {
	return llm.TokenUsage{
		PromptTokens:     usage.PromptTokens,
		CompletionTokens: usage.CompletionTokens,
		TotalTokens:      usage.TotalTokens,
	}
}

func (p *Provider) doChatCompletion(ctx context.Context, req *chatCompletionRequest) (*chatCompletionResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.config.BaseURL+"/chat/completions", strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.config.APIKey)

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, ErrAPITimeout
		}
		return nil, fmt.Errorf("%w: %v", ErrNetworkError, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, p.parseAPIError(resp.StatusCode, respBody)
	}

	var result chatCompletionResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidResponse, err)
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("%w: no choices in response", ErrInvalidResponse)
	}

	return &result, nil
}

func (p *Provider) streamChatCompletion(ctx context.Context, req *chatCompletionRequest, streamChan chan<- llm.StreamChunk) error {
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.config.BaseURL+"/chat/completions", strings.NewReader(string(body)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.config.APIKey)

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return ErrAPITimeout
		}
		return fmt.Errorf("%w: %v", ErrNetworkError, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return p.parseAPIError(resp.StatusCode, respBody)
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			streamChan <- llm.StreamChunk{Content: "", Done: true}
			break
		}

		var chunk streamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}

		if len(chunk.Choices) > 0 {
			content := chunk.Choices[0].Delta.Content
			done := chunk.Choices[0].FinishReason == "stop"
			streamChan <- llm.StreamChunk{
				Content: content,
				Done:    done,
			}
			if done {
				break
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read stream: %w", err)
	}

	return nil
}

func (p *Provider) parseAPIError(status int, body []byte) error {
	var errResp struct {
		Error struct {
			Message string `json:"message"`
			Type    string `json:"type"`
			Code    string `json:"code"`
		} `json:"error"`
	}

	if err := json.Unmarshal(body, &errResp); err != nil {
		return &APIError{
			Message:    string(body),
			HTTPStatus: status,
		}
	}

	apiErr := &APIError{
		Code:       errResp.Error.Code,
		Message:    errResp.Error.Message,
		HTTPStatus: status,
		Type:       errResp.Error.Type,
	}
	return apiErr
}

func (p *Provider) shouldRetry(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.HTTPStatus == http.StatusTooManyRequests ||
			apiErr.HTTPStatus >= 500
	}

	if errors.Is(err, ErrNetworkError) || errors.Is(err, ErrAPITimeout) {
		return true
	}

	return false
}

type Tokenizer struct {
	avgCharsPerToken float64
	cache            sync.Map
}

func NewTokenizer() *Tokenizer {
	return &Tokenizer{
		avgCharsPerToken: 4.0,
	}
}

func (t *Tokenizer) Count(text string) int {
	if cached, ok := t.cache.Load(text); ok {
		return cached.(int)
	}

	approxTokens := len(text) / int(t.avgCharsPerToken)
	if approxTokens < 1 {
		approxTokens = 1
	}

	if len(text) < 10000 {
		t.cache.Store(text, approxTokens)
	}

	return approxTokens
}
