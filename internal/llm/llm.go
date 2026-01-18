package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"MusicDev33/econsim/internal/config"
)

type Provider string

const (
	ProviderClaude Provider = "claude"
	ProviderOpenAI Provider = "openai"
	ProviderGrok   Provider = "grok"
)

// NewClient returns an LLM client based on the specified provider.
// If no provider is specified (empty string), it defaults to Claude if available,
// then falls back to OpenAI, then Grok.
// Panics if the requested provider has no API key configured.
func NewClient(provider Provider) LLMProvider {
	cfg := config.Get()

	if provider == "" {
		return newDefaultClient(cfg)
	}

	return newSpecificClient(cfg, provider)
}

func newDefaultClient(cfg *config.Config) LLMProvider {
	// Priority: Claude > OpenAI > Grok
	if cfg.AkClaude != "" {
		return NewClaudeClient(cfg.AkClaude)
	}
	if cfg.AkOpenAI != "" {
		return NewOpenAIClient(cfg.AkOpenAI)
	}
	if cfg.AkGrok != "" {
		return NewGrokClient(cfg.AkGrok)
	}

	// Config validation ensures at least one key exists, so this shouldn't happen
	panic("no LLM provider configured")
}

func newSpecificClient(cfg *config.Config, provider Provider) LLMProvider {
	switch provider {
	case ProviderClaude:
		if cfg.AkClaude == "" {
			panic("claude provider requested but akClaude not set in config")
		}
		return NewClaudeClient(cfg.AkClaude)

	case ProviderOpenAI:
		if cfg.AkOpenAI == "" {
			panic("openai provider requested but akOpenAI not set in config")
		}
		return NewOpenAIClient(cfg.AkOpenAI)

	case ProviderGrok:
		if cfg.AkGrok == "" {
			panic("grok provider requested but akGrok not set in config")
		}
		return NewGrokClient(cfg.AkGrok)

	default:
		panic(fmt.Sprintf("unknown provider: %s", provider))
	}
}

type LLMProvider interface {
	SendChat(msg string) (*ChatResponse, error)
}

type LLMHandler struct {
	Client LLMProvider
}

func NewLLMHandler(client LLMProvider) *LLMHandler {
	return &LLMHandler{
		Client: client,
	}
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Messages    []Message `json:"messages"`
	Model       string    `json:"model"`
	Stream      bool      `json:"stream"`
	Temperature float64   `json:"temperature"`
}

type ChatRes struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type ChatResponse struct {
	// Add fields based on the actual API response structure
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int     `json:"index"`
		Message      Message `json:"message"`
		FinishReason string  `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type GrokClient struct {
	APIKey string
	Client *http.Client
}

func NewGrokClient(apiKey string) *GrokClient {
	return &GrokClient{
		APIKey: apiKey,
		Client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *GrokClient) SendChat(msg string) (*ChatResponse, error) {
	chatReq := ChatRequest{
		Messages: []Message{
			{
				Role:    "user",
				Content: msg,
			},
		},
		Model:       "grok-4-1-fast-non-reasoning",
		Stream:      false,
		Temperature: 0.7,
	}

	return sendOpenAICompatibleRequest(c.Client, "https://api.x.ai/v1/chat/completions", c.APIKey, chatReq)
}

// OpenAIClient implements LLMProvider for OpenAI's ChatGPT
type OpenAIClient struct {
	APIKey string
	Model  string
	Client *http.Client
}

func NewOpenAIClient(apiKey string) *OpenAIClient {
	return &OpenAIClient{
		APIKey: apiKey,
		Model:  "gpt-4o",
		Client: &http.Client{Timeout: 60 * time.Second},
	}
}

func (c *OpenAIClient) SendChat(msg string) (*ChatResponse, error) {
	chatReq := ChatRequest{
		Messages: []Message{
			{
				Role:    "user",
				Content: msg,
			},
		},
		Model:       c.Model,
		Stream:      false,
		Temperature: 0.7,
	}

	return sendOpenAICompatibleRequest(c.Client, "https://api.openai.com/v1/chat/completions", c.APIKey, chatReq)
}

// ClaudeClient implements LLMProvider for Anthropic's Claude
type ClaudeClient struct {
	APIKey string
	Model  string
	Client *http.Client
}

func NewClaudeClient(apiKey string) *ClaudeClient {
	return &ClaudeClient{
		APIKey: apiKey,
		Model:  "claude-sonnet-4-20250514",
		Client: &http.Client{Timeout: 60 * time.Second},
	}
}

// Claude-specific request/response types
type claudeRequest struct {
	Model       string          `json:"model"`
	MaxTokens   int             `json:"max_tokens"`
	Messages    []claudeMessage `json:"messages"`
	Temperature float64         `json:"temperature,omitempty"`
}

type claudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type claudeResponse struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Role    string `json:"role"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Model        string `json:"model"`
	StopReason   string `json:"stop_reason"`
	StopSequence string `json:"stop_sequence"`
	Usage        struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

func (c *ClaudeClient) SendChat(msg string) (*ChatResponse, error) {
	claudeReq := claudeRequest{
		Model:     c.Model,
		MaxTokens: 4096,
		Messages: []claudeMessage{
			{
				Role:    "user",
				Content: msg,
			},
		},
		Temperature: 0.7,
	}

	jsonData, err := json.Marshal(claudeReq)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, body)
	}

	var claudeResp claudeResponse
	if err := json.Unmarshal(body, &claudeResp); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	// Convert Claude response to standard ChatResponse format
	content := ""
	if len(claudeResp.Content) > 0 {
		content = claudeResp.Content[0].Text
	}

	return &ChatResponse{
		ID:      claudeResp.ID,
		Object:  claudeResp.Type,
		Created: time.Now().Unix(),
		Model:   claudeResp.Model,
		Choices: []struct {
			Index        int     `json:"index"`
			Message      Message `json:"message"`
			FinishReason string  `json:"finish_reason"`
		}{
			{
				Index:        0,
				Message:      Message{Role: claudeResp.Role, Content: content},
				FinishReason: claudeResp.StopReason,
			},
		},
		Usage: struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		}{
			PromptTokens:     claudeResp.Usage.InputTokens,
			CompletionTokens: claudeResp.Usage.OutputTokens,
			TotalTokens:      claudeResp.Usage.InputTokens + claudeResp.Usage.OutputTokens,
		},
	}, nil
}

// sendOpenAICompatibleRequest handles requests for OpenAI-compatible APIs (OpenAI, Grok)
func sendOpenAICompatibleRequest(client *http.Client, url, apiKey string, chatReq ChatRequest) (*ChatResponse, error) {
	jsonData, err := json.Marshal(chatReq)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, body)
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	return &chatResp, nil
}
