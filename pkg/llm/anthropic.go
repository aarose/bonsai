package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// AnthropicClient implements the Client interface for Anthropic
type AnthropicClient struct {
	config     Config
	httpClient *http.Client
}

// AnthropicRequest represents the request structure for Anthropic API
type AnthropicRequest struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	Messages  []Message `json:"messages"`
}

// AnthropicResponse represents the response structure from Anthropic API
type AnthropicResponse struct {
	Content []ContentBlock `json:"content"`
	Error   *APIError      `json:"error,omitempty"`
}

// ContentBlock represents a content block in Anthropic response
type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// NewAnthropicClient creates a new Anthropic client
func NewAnthropicClient(config Config) (*AnthropicClient, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("Anthropic API key is required")
	}

	return &AnthropicClient{
		config: config,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}, nil
}

// GenerateResponse generates a response using Anthropic API
func (c *AnthropicClient) GenerateResponse(ctx context.Context, prompt string, model string) (string, error) {
	// Use the provided model or default to claude-3-haiku
	if model == "" {
		model = "claude-3-haiku-20240307"
	}

	// Normalize model name for Anthropic
	model = normalizeAnthropicModel(model)

	maxTokens := 1000 // Default max tokens
	if c.config.MaxTokens > 0 {
		maxTokens = c.config.MaxTokens
	}

	request := AnthropicRequest{
		Model:     model,
		MaxTokens: maxTokens,
		Messages: []Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.config.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var apiErr APIError
		if err := json.Unmarshal(body, &apiErr); err != nil {
			return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
		}
		return "", fmt.Errorf("API error: %s", apiErr.Message)
	}

	var response AnthropicResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(response.Content) == 0 {
		return "", fmt.Errorf("no response content received")
	}

	// Combine all text content blocks
	var result string
	for _, block := range response.Content {
		if block.Type == "text" {
			result += block.Text
		}
	}

	return result, nil
}

// GetAvailableModels returns the list of available Anthropic models
func (c *AnthropicClient) GetAvailableModels() []string {
	return []string{
		"claude-3-haiku-20240307",
		"claude-3-sonnet-20240229",
		"claude-3-opus-20240229",
		"claude-3-5-sonnet-20241022",
	}
}

// GetProviderName returns the provider name
func (c *AnthropicClient) GetProviderName() string {
	return "anthropic"
}

// normalizeAnthropicModel normalizes model names for Anthropic API
func normalizeAnthropicModel(model string) string {
	switch model {
	case "claude-3-haiku", "claude-3.5-haiku":
		return "claude-3-haiku-20240307"
	case "claude-3-sonnet", "claude-3.5-sonnet":
		return "claude-3-5-sonnet-20241022"
	case "claude-3-opus":
		return "claude-3-opus-20240229"
	default:
		// If it's already a full model name, use it as-is
		if contains(model, []string{"claude-3", "claude-3.5"}) {
			return model
		}
		// Default to haiku if unrecognized
		return "claude-3-haiku-20240307"
	}
}
