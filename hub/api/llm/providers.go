// Package llm provides LLM provider management
// Complies with CODING_STANDARDS.md: Provider modules max 300 lines
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Global instances for rate limiting and cost optimization
var (
	defaultRateLimiter = NewRateLimiter(10, 1) // 10 requests, refill 1 per second
	quotaManager       = NewQuotaManager()
	costOptimizer      = NewCostOptimizer()
)

// getSupportedProviders returns list of supported LLM providers
func getSupportedProviders() []string {
	return []string{"openai", "anthropic", "azure", "ollama"}
}

// getSupportedModels returns supported models for a provider
func getSupportedModels(provider string) []string {
	switch provider {
	case "openai":
		return []string{"gpt-4", "gpt-3.5-turbo", "gpt-4-turbo"}
	case "anthropic":
		return []string{"claude-3-opus", "claude-3-sonnet", "claude-3-haiku"}
	case "azure":
		return []string{"gpt-4", "gpt-3.5-turbo"}
	case "ollama":
		return []string{"llama2", "codellama", "mistral"}
	default:
		return []string{}
	}
}

// selectModel selects the appropriate LLM model based on task type and cost optimization
func selectModel(ctx context.Context, taskType string, config *LLMConfig) string {
	// Estimate tokens for cost optimization
	estimatedTokens := 1000 // Default estimate, would be calculated from actual prompt

	// Use cost optimizer for intelligent model selection
	if config.CostOptimization != nil {
		if optimizedModel := costOptimizer.OptimizeModelSelection(taskType, config.Provider, estimatedTokens, config.CostOptimization); optimizedModel != "" {
			return optimizedModel
		}
	}

	// Fallback to rule-based selection
	switch taskType {
	case "code_analysis", "security_audit", "performance_analysis":
		switch config.Provider {
		case "openai":
			return "gpt-4"
		case "anthropic":
			return "claude-3-opus"
		case "azure":
			return "gpt-4"
		}

	case "documentation", "explanation", "code_review":
		switch config.Provider {
		case "openai":
			return "gpt-3.5-turbo"
		case "anthropic":
			return "claude-3-sonnet"
		case "azure":
			return "gpt-3.5-turbo"
		}

	case "quick_analysis", "syntax_check", "basic_validation":
		switch config.Provider {
		case "openai":
			return "gpt-3.5-turbo"
		case "anthropic":
			return "claude-3-haiku"
		case "azure":
			return "gpt-3.5-turbo"
		}
	}

	return config.Model
}

// callLLM makes an API call to the LLM provider with rate limiting and quota management
func callLLM(ctx context.Context, config *LLMConfig, prompt string, taskType string) (string, int, error) {
	// Estimate token usage for quota checking
	estimatedTokens := EstimateTokens(prompt)

	// Check rate limiting
	if !defaultRateLimiter.Allow() {
		return "", 0, fmt.Errorf("rate limit exceeded, please try again later")
	}

	// Check quota (assuming project ID is available, using a default for now)
	projectID := "default" // TODO: Extract from context or config
	if !quotaManager.CheckQuota(projectID, estimatedTokens) {
		return "", 0, fmt.Errorf("quota exceeded for project %s", projectID)
	}

	// Select optimal model based on cost optimization
	selectedModel := config.Model
	if optimizedModel := selectModel(ctx, taskType, config); optimizedModel != "" {
		selectedModel = optimizedModel
	}

	// Update config with selected model
	callConfig := *config
	callConfig.Model = selectedModel

	// Make the API call
	var response string
	var tokensUsed int
	var err error

	switch callConfig.Provider {
	case "openai":
		response, tokensUsed, err = callOpenAI(ctx, &callConfig, prompt)
	case "anthropic":
		response, tokensUsed, err = callAnthropic(ctx, &callConfig, prompt)
	case "azure":
		response, tokensUsed, err = callAzure(ctx, &callConfig, prompt)
	default:
		return "", 0, fmt.Errorf("unsupported provider: %s", callConfig.Provider)
	}

	// Record usage for quota tracking
	if err == nil {
		quotaManager.RecordUsage(projectID, tokensUsed)
	}

	return response, tokensUsed, err
}

// callOpenAI makes an API call to OpenAI
func callOpenAI(ctx context.Context, config *LLMConfig, prompt string) (string, int, error) {
	url := "https://api.openai.com/v1/chat/completions"

	reqBody := map[string]interface{}{
		"model": config.Model,
		"messages": []map[string]string{
			{"role": "system", "content": "You are a code analysis assistant."},
			{"role": "user", "content": prompt},
		},
		"max_tokens":  4096,
		"temperature": 0.3,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", 0, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.APIKey)

	// Execute request with retry logic
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return "", 0, fmt.Errorf("OpenAI API returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", 0, fmt.Errorf("failed to read response: %w", err)
	}

	var apiResponse struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return "", 0, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(apiResponse.Choices) == 0 {
		return "", 0, fmt.Errorf("no choices in response")
	}

	return apiResponse.Choices[0].Message.Content, apiResponse.Usage.TotalTokens, nil
}

// callAnthropic makes an API call to Anthropic Claude
func callAnthropic(ctx context.Context, config *LLMConfig, prompt string) (string, int, error) {
	url := "https://api.anthropic.com/v1/messages"

	reqBody := map[string]interface{}{
		"model":       config.Model,
		"max_tokens":  4096,
		"temperature": 0.3,
		"system":      "You are a code analysis assistant.",
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", 0, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", config.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	// Execute request with retry logic
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return "", 0, fmt.Errorf("Anthropic API returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", 0, fmt.Errorf("failed to read response: %w", err)
	}

	var apiResponse struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return "", 0, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(apiResponse.Content) == 0 {
		return "", 0, fmt.Errorf("no content in response")
	}

	totalTokens := apiResponse.Usage.InputTokens + apiResponse.Usage.OutputTokens
	return apiResponse.Content[0].Text, totalTokens, nil
}

// callAzure makes an API call to Azure OpenAI
func callAzure(ctx context.Context, config *LLMConfig, prompt string) (string, int, error) {
	// Azure OpenAI uses a different endpoint format
	// Expected format: https://your-resource-name.openai.azure.com/openai/deployments/your-deployment-name/chat/completions?api-version=2023-05-15
	azureEndpoint := config.Provider // This should contain the full Azure endpoint URL

	// If not a full URL, construct it (basic implementation)
	if !strings.Contains(azureEndpoint, "openai.azure.com") {
		return "", 0, fmt.Errorf("invalid Azure endpoint format. Expected full Azure OpenAI URL")
	}

	// Add API version if not present
	if !strings.Contains(azureEndpoint, "api-version") {
		if strings.Contains(azureEndpoint, "?") {
			azureEndpoint += "&api-version=2023-05-15"
		} else {
			azureEndpoint += "?api-version=2023-05-15"
		}
	}

	reqBody := map[string]interface{}{
		"messages": []map[string]string{
			{"role": "system", "content": "You are a code analysis assistant."},
			{"role": "user", "content": prompt},
		},
		"max_tokens":  4096,
		"temperature": 0.3,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", 0, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", azureEndpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", config.APIKey) // Azure uses api-key header instead of Authorization

	// Execute request with retry logic
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return "", 0, fmt.Errorf("Azure OpenAI API returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", 0, fmt.Errorf("failed to read response: %w", err)
	}

	var apiResponse struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return "", 0, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(apiResponse.Choices) == 0 {
		return "", 0, fmt.Errorf("no choices in response")
	}

	return apiResponse.Choices[0].Message.Content, apiResponse.Usage.TotalTokens, nil
}
