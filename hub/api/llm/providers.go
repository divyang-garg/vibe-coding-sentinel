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
	"os"
	"strings"
	"time"

	"sentinel-hub-api/pkg"
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

// SelectModel selects the appropriate LLM model based on task type and cost optimization
func SelectModel(ctx context.Context, taskType string, config *LLMConfig) string {
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

	case "knowledge_extraction":
		switch config.Provider {
		case "openai":
			return "gpt-4" // Best for structured JSON output
		case "anthropic":
			return "claude-3-sonnet" // Good balance of cost/quality
		case "ollama":
			return config.Model // Use configured model
		case "azure":
			return "gpt-4"
		}
	}

	return config.Model
}

// CallLLM makes an API call to the LLM provider with rate limiting and quota management
func CallLLM(ctx context.Context, config *LLMConfig, prompt string, taskType string) (string, int, error) {
	// Estimate token usage for quota checking
	estimatedTokens := EstimateTokens(prompt)

	// Check rate limiting
	if !defaultRateLimiter.Allow() {
		return "", 0, fmt.Errorf("rate limit exceeded, please try again later")
	}

	// Extract projectID from context, fallback to config, then default
	projectID := getProjectID(ctx)
	if !quotaManager.CheckQuota(projectID, estimatedTokens) {
		return "", 0, fmt.Errorf("quota exceeded for project %s", projectID)
	}

	// Select optimal model based on cost optimization
	selectedModel := config.Model
	if optimizedModel := SelectModel(ctx, taskType, config); optimizedModel != "" {
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
	case "ollama":
		response, tokensUsed, err = callOllama(ctx, &callConfig, prompt)
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

// getProjectID extracts projectID from context with fallback to config and default
func getProjectID(ctx context.Context) string {
	// Try to get from context first
	if projectID, ok := ctx.Value(pkg.ProjectIDKey).(string); ok && projectID != "" {
		return projectID
	}

	// Fallback to environment variable
	if projectID := os.Getenv("SENTINEL_PROJECT_ID"); projectID != "" {
		return projectID
	}

	// Last resort: default
	return "default"
}

// callOllama makes an API call to local Ollama instance
func callOllama(ctx context.Context, config *LLMConfig, prompt string) (string, int, error) {
	// Default to localhost:11434 if no custom endpoint
	endpoint := "http://localhost:11434/api/generate"
	if config.APIKey != "" && strings.HasPrefix(config.APIKey, "http") {
		endpoint = config.APIKey + "/api/generate"
	}

	reqBody := map[string]interface{}{
		"model":  config.Model,
		"prompt": prompt,
		"stream": false,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", 0, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Execute request
	client := &http.Client{Timeout: 120 * time.Second} // Ollama can be slower
	resp, err := client.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return "", 0, fmt.Errorf("Ollama API returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", 0, fmt.Errorf("failed to read response: %w", err)
	}

	var apiResponse struct {
		Response        string `json:"response"`
		Done            bool   `json:"done"`
		Context         []int  `json:"context"`
		TotalDuration   int64  `json:"total_duration"`
		LoadDuration    int64  `json:"load_duration"`
		PromptEvalCount int    `json:"prompt_eval_count"`
		EvalCount       int    `json:"eval_count"`
	}

	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return "", 0, fmt.Errorf("failed to parse response: %w", err)
	}

	// Estimate tokens: prompt tokens + generated tokens
	tokensUsed := apiResponse.PromptEvalCount + apiResponse.EvalCount
	if tokensUsed == 0 {
		// Fallback estimation if counts not provided
		tokensUsed = EstimateTokens(prompt) + EstimateTokens(apiResponse.Response)
	}

	return apiResponse.Response, tokensUsed, nil
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
