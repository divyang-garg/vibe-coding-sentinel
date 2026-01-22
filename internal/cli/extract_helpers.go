// Package cli provides helper implementations for knowledge extraction
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package cli

import (
	"context"
	"fmt"
	"os"
	"sync"
)

// simpleCache implements extraction.Cache
type simpleCache struct {
	mu    sync.RWMutex
	store map[string]string
}

func newSimpleCache() *simpleCache {
	return &simpleCache{
		store: make(map[string]string),
	}
}

func (c *simpleCache) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.store[key]
	return val, ok
}

func (c *simpleCache) Set(key string, value string, tokensUsed int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store[key] = value
}

// cliLogger implements extraction.Logger
type cliLogger struct{}

func newCLILogger() *cliLogger {
	return &cliLogger{}
}

func (l *cliLogger) Debug(msg string, args ...interface{}) {
	// Silent in CLI mode
}

func (l *cliLogger) Info(msg string, args ...interface{}) {
	fmt.Printf("[INFO] "+msg+"\n", args...)
}

func (l *cliLogger) Warn(msg string, args ...interface{}) {
	fmt.Printf("[WARN] "+msg+"\n", args...)
}

func (l *cliLogger) Error(msg string, args ...interface{}) {
	fmt.Printf("[ERROR] "+msg+"\n", args...)
}

// llmClientAdapter implements extraction.LLMClient
// Note: This is a stub implementation that returns an error
// Full LLM integration requires Hub API dependency which is not available in standalone CLI
type llmClientAdapter struct {
	provider string
	apiKey   string
	model    string
}

func newLLMClientAdapter() (*llmClientAdapter, error) {
	provider := os.Getenv("LLM_PROVIDER")
	if provider == "" {
		provider = "openai" // Default
	}

	apiKey := os.Getenv("LLM_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("LLM_API_KEY environment variable is required")
	}

	model := os.Getenv("LLM_MODEL")
	if model == "" {
		// Default models by provider
		switch provider {
		case "openai":
			model = "gpt-3.5-turbo"
		case "anthropic":
			model = "claude-3-haiku"
		case "ollama":
			model = "llama2"
		default:
			model = "gpt-3.5-turbo" // Fallback default
		}
	}

	return &llmClientAdapter{
		provider: provider,
		apiKey:   apiKey,
		model:    model,
	}, nil
}

func (a *llmClientAdapter) Call(ctx context.Context, prompt string, taskType string) (string, int, error) {
	// LLM calls require Hub API integration
	// For standalone CLI, return error indicating Hub API is required
	return "", 0, fmt.Errorf("LLM functionality requires Hub API integration. Please use 'sentinel knowledge extract --fallback' for pattern-based extraction")
}

// noOpLLMClient implements extraction.LLMClient for when LLM is disabled
type noOpLLMClient struct{}

func (a *noOpLLMClient) Call(ctx context.Context, prompt string, taskType string) (string, int, error) {
	return "", 0, fmt.Errorf("LLM client is disabled")
}
