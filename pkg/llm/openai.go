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

// OpenAIClient implements the Client interface for OpenAI
type OpenAIClient struct {
	config     Config
	httpClient *http.Client
}

// OpenAIRequest represents the request structure for OpenAI API
type OpenAIRequest struct {
	Model     string    `json:"model"`
	Messages  []Message `json:"messages"`
	MaxTokens int       `json:"max_tokens,omitempty"`
}

// OpenAIResponse represents the response structure from OpenAI API
type OpenAIResponse struct {
	Choices []Choice  `json:"choices"`
	Error   *APIError `json:"error,omitempty"`
}

// Choice represents a choice in the OpenAI response
type Choice struct {
	Message Message `json:"message"`
}

// Message represents a message in the conversation
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// APIError represents an API error
type APIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

// NewOpenAIClient creates a new OpenAI client
func NewOpenAIClient(config Config) (*OpenAIClient, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required")
	}

	return &OpenAIClient{
		config: config,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}, nil
}

// GenerateResponse generates a response using OpenAI API
func (c *OpenAIClient) GenerateResponse(ctx context.Context, prompt string, model string) (string, error) {
	// Use the provided model or default to gpt-3.5-turbo
	if model == "" {
		model = "gpt-3.5-turbo"
	}

	// Normalize model name for OpenAI
	model = normalizeOpenAIModel(model)

	request := OpenAIRequest{
		Model: model,
		Messages: []Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	if c.config.MaxTokens > 0 {
		request.MaxTokens = c.config.MaxTokens
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)

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

	var response OpenAIResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no response choices received")
	}

	return response.Choices[0].Message.Content, nil
}

// GetAvailableModels returns the list of available OpenAI models
func (c *OpenAIClient) GetAvailableModels() []string {
	return []string{
		"gpt-3.5-turbo",
		"gpt-4",
		"gpt-4-turbo",
		"gpt-4o",
		"gpt-4o-mini",
	}
}

// GetProviderName returns the provider name
func (c *OpenAIClient) GetProviderName() string {
	return "openai"
}

// normalizeOpenAIModel normalizes model names for OpenAI API
func normalizeOpenAIModel(model string) string {
	switch model {
	case "gpt-4", "gpt-4-turbo":
		return "gpt-4"
	case "gpt-3.5", "gpt-3.5-turbo":
		return "gpt-3.5-turbo"
	case "gpt-4o", "gpt-4o-mini":
		return model
	default:
		// If it's not a recognized model, try to use it as-is
		return model
	}
}
