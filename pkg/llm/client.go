package llm

import (
	"context"
	"fmt"
)

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

// Client defines the interface for LLM providers
type Client interface {
	GenerateResponse(ctx context.Context, prompt string, model string) (string, error)
	GenerateResponseFromHistory(ctx context.Context, messages []Message, model string) (string, error)
	GetAvailableModels() []string
	GetProviderName() string
}

// Config holds configuration for LLM clients
type Config struct {
	APIKey    string
	BaseURL   string // Optional, for custom endpoints
	MaxTokens int    // Optional, for response length limits
}

// NewClient creates a new LLM client based on the provider
func NewClient(provider string, config Config) (Client, error) {
	switch provider {
	case "openai", "gpt-3.5-turbo", "gpt-4", "gpt-4-turbo":
		return NewOpenAIClient(config)
	case "anthropic", "claude-3-sonnet", "claude-3-haiku", "claude-3-opus", "claude-3-5-sonnet":
		return NewAnthropicClient(config)
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s", provider)
	}
}

// DetectProviderFromModel detects the provider from a model name
func DetectProviderFromModel(model string) string {
	switch {
	case contains(model, []string{"gpt-3.5", "gpt-4", "gpt-4o"}):
		return "openai"
	case contains(model, []string{"claude-3", "claude-3.5"}):
		return "anthropic"
	default:
		return "openai" // Default fallback
	}
}

// contains checks if a string contains any of the substrings
func contains(s string, substrings []string) bool {
	for _, sub := range substrings {
		if len(sub) <= len(s) {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
		}
	}
	return false
}

// NodeToMessage converts a database node to an LLM message
// Maps node types: "user" -> "user", "llm" -> "assistant"
func NodeToMessage(nodeType, content string) Message {
	role := "user"
	if nodeType == "llm" {
		role = "assistant"
	}
	return Message{
		Role:    role,
		Content: content,
	}
}
