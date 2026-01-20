// Package extraction provides LLM-powered knowledge extraction
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package extraction

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// OllamaClient implements LLMClient for Ollama API
type OllamaClient struct {
	baseURL    string
	model      string
	httpClient *http.Client
}

// OllamaConfig configures the Ollama client
type OllamaConfig struct {
	BaseURL    string
	Model      string
	Timeout    time.Duration
}

// DefaultOllamaConfig returns default configuration
func DefaultOllamaConfig() OllamaConfig {
	return OllamaConfig{
		BaseURL: getEnvOrDefault("OLLAMA_HOST", "http://localhost:11434"),
		Model:   getEnvOrDefault("OLLAMA_MODEL", "llama3.2"),
		Timeout: 120 * time.Second,
	}
}

// NewOllamaClient creates a new Ollama LLM client
func NewOllamaClient(cfg OllamaConfig) *OllamaClient {
	return &OllamaClient{
		baseURL: cfg.BaseURL,
		model:   cfg.Model,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

// Call invokes the LLM with the given prompt
func (c *OllamaClient) Call(ctx context.Context, prompt string, taskType string) (string, int, error) {
	reqBody := map[string]interface{}{
		"model":  c.model,
		"prompt": prompt,
		"stream": false,
		"options": map[string]interface{}{
			"temperature": 0.2,
			"num_predict": 4096,
		},
	}
	
	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", 0, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/generate", bytes.NewReader(body))
	if err != nil {
		return "", 0, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("LLM request failed: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return "", 0, fmt.Errorf("LLM returned status %d", resp.StatusCode)
	}
	
	var result struct {
		Response string `json:"response"`
		Context  []int  `json:"context"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", 0, fmt.Errorf("failed to decode response: %w", err)
	}
	
	// Estimate tokens (rough approximation: 1 token ≈ 4 characters)
	tokens := len(prompt)/4 + len(result.Response)/4
	
	return result.Response, tokens, nil
}

func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
