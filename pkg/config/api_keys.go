package config

import (
	"os"

	"github.com/aarose/bonsai/pkg/llm"
)

// GetAPIKey retrieves the API key for the specified LLM provider
func GetAPIKey(provider string) string {
	envVar := GetAPIKeyEnvVar(provider)
	return os.Getenv(envVar)
}

// GetAPIKeyEnvVar returns the environment variable name for the API key
func GetAPIKeyEnvVar(provider string) string {
	switch provider {
	case "openai", "gpt-3.5-turbo", "gpt-4", "gpt-4-turbo", "gpt-4o":
		return "OPENAI_API_KEY"
	case "anthropic", "claude-3-sonnet", "claude-3-haiku", "claude-3-opus", "claude-3-5-sonnet":
		return "ANTHROPIC_API_KEY"
	default:
		// Try to detect provider from model name
		detectedProvider := llm.DetectProviderFromModel(provider)
		return GetAPIKeyEnvVar(detectedProvider)
	}
}

// GetDefaultModel returns a cost-effective default model for the provider
func GetDefaultModel(provider string) string {
	switch provider {
	case "openai":
		return "gpt-3.5-turbo"
	case "anthropic":
		return "claude-3-haiku"
	default:
		// Try to detect provider from model name
		detectedProvider := llm.DetectProviderFromModel(provider)
		return GetDefaultModel(detectedProvider)
	}
}
