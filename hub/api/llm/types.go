// Package llm provides LLM types
// Complies with CODING_STANDARDS.md: Type definitions max 200 lines
package llm

// LLMConfig contains LLM provider configuration
type LLMConfig struct {
	ID               string                  `json:"id,omitempty"`
	Provider         string                  `json:"provider"`
	APIKey           string                  `json:"api_key"` // Decrypted for use
	Model            string                  `json:"model"`
	KeyType          string                  `json:"key_type"`
	CostOptimization *CostOptimizationConfig `json:"cost_optimization,omitempty"`
}

// LLMResponse represents the response from an LLM API call
type LLMResponse struct {
	Content          string
	TokensUsed       int
	PromptTokens     int
	CompletionTokens int
}

// CostOptimizationConfig contains cost optimization settings
type CostOptimizationConfig struct {
	UseCache          bool    `json:"use_cache"`
	CacheTTLHours     int     `json:"cache_ttl_hours"`
	ProgressiveDepth  bool    `json:"progressive_depth"`
	MaxCostPerRequest float64 `json:"max_cost_per_request,omitempty"`
}

// LLMUsage tracks token usage and costs
type LLMUsage struct {
	ID            string  `json:"id"`
	ProjectID     string  `json:"project_id"`
	ValidationID  string  `json:"validation_id,omitempty"`
	Provider      string  `json:"provider"`
	Model         string  `json:"model"`
	TokensUsed    int     `json:"tokens_used"`
	EstimatedCost float64 `json:"estimated_cost"`
	CreatedAt     string  `json:"created_at"`
}
