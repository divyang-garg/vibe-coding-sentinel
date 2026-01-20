// Package repository provides LLM-powered extraction for hub API
// Complies with CODING_STANDARDS.md: Repositories max 350 lines
package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// LLMExtractor provides LLM-based knowledge extraction
type LLMExtractor struct {
	baseURL    string
	model      string
	httpClient *http.Client
	enabled    bool
}

// NewLLMExtractor creates a new LLM extractor
func NewLLMExtractor() *LLMExtractor {
	baseURL := os.Getenv("OLLAMA_HOST")
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	model := os.Getenv("OLLAMA_MODEL")
	if model == "" {
		model = "llama3.2"
	}
	
	return &LLMExtractor{
		baseURL: baseURL,
		model:   model,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
		enabled: baseURL != "",
	}
}

// ExtractWithLLM extracts knowledge using LLM
func (e *LLMExtractor) ExtractWithLLM(ctx context.Context, text string, docID string) ([]map[string]interface{}, error) {
	if !e.enabled {
		return nil, fmt.Errorf("LLM extraction not enabled")
	}
	
	prompt := fmt.Sprintf(`Extract business rules from this text in JSON format:
{
  "business_rules": [
    {
      "id": "BR-XXX",
      "title": "Rule title",
      "description": "Rule description",
      "confidence": 0.8
    }
  ]
}

Text: %s`, text)
	
	reqBody := map[string]interface{}{
		"model":  e.model,
		"prompt": prompt,
		"stream": false,
		"options": map[string]interface{}{
			"temperature": 0.2,
			"num_predict": 4096,
		},
	}
	
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", e.baseURL+"/api/generate", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := e.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("LLM request failed: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("LLM returned status %d", resp.StatusCode)
	}
	
	var result struct {
		Response string `json:"response"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	// Parse JSON from response
	var parsed struct {
		BusinessRules []map[string]interface{} `json:"business_rules"`
	}
	if err := json.Unmarshal([]byte(result.Response), &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}
	
	return parsed.BusinessRules, nil
}
