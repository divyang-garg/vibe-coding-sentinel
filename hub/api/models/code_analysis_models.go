// Package models - Code analysis data models
// Complies with CODING_STANDARDS.md: Data Models max 200 lines
package models

// CodeAnalysisRequest represents a request for code analysis
type CodeAnalysisRequest struct {
	Code     string            `json:"code" validate:"required"`
	Language string            `json:"language" validate:"required"`
	Options  map[string]string `json:"options,omitempty"`
}

// CodeLintRequest represents a request for code linting
type CodeLintRequest struct {
	Code     string            `json:"code" validate:"required"`
	Language string            `json:"language" validate:"required"`
	Rules    []string          `json:"rules,omitempty"`
	Options  map[string]string `json:"options,omitempty"`
}

// CodeRefactorRequest represents a request for code refactoring
type CodeRefactorRequest struct {
	Code     string            `json:"code" validate:"required"`
	Language string            `json:"language" validate:"required"`
	Action   string            `json:"action" validate:"required"`
	Options  map[string]string `json:"options,omitempty"`
}

// DocumentationRequest represents a request for documentation generation
type DocumentationRequest struct {
	Code     string            `json:"code" validate:"required"`
	Language string            `json:"language" validate:"required"`
	Format   string            `json:"format,omitempty"`
	Options  map[string]string `json:"options,omitempty"`
}

// CodeValidationRequest represents a request for code validation
type CodeValidationRequest struct {
	Code     string            `json:"code" validate:"required"`
	Language string            `json:"language" validate:"required"`
	Schema   string            `json:"schema,omitempty"`
	Options  map[string]string `json:"options,omitempty"`
}

// ApplyFixRequest represents a request to apply fixes to code
type ApplyFixRequest struct {
	Code     string            `json:"code" validate:"required"`
	Language string            `json:"language" validate:"required"`
	FixType  string            `json:"fix_type" validate:"required"` // "security", "style", "performance"
	Options  map[string]string `json:"options,omitempty"`
}

// ApplyFixResponse represents the response from applying fixes
type ApplyFixResponse struct {
	FixedCode string                   `json:"fixed_code"`
	Changes   []map[string]interface{} `json:"changes"`
	Summary   string                   `json:"summary"`
}

// ValidateLLMConfigRequest represents a request to validate LLM configuration
type ValidateLLMConfigRequest struct {
	Config LLMConfig `json:"config" validate:"required"`
}

// ValidateLLMConfigResponse represents the response from LLM config validation
type ValidateLLMConfigResponse struct {
	Valid    bool     `json:"valid"`
	Errors   []string `json:"errors,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}

// CacheMetricsResponse represents cache metrics for a project
type CacheMetricsResponse struct {
	ProjectID     string  `json:"project_id"`
	HitRate       float64 `json:"hit_rate"`
	Hits          int64   `json:"hits"`
	Misses        int64   `json:"misses"`
	TotalRequests int64   `json:"total_requests"`
	CacheSize     int64   `json:"cache_size"`
}

// CostMetricsResponse represents cost metrics for a project
type CostMetricsResponse struct {
	ProjectID             string  `json:"project_id"`
	TotalCost             float64 `json:"total_cost"`
	TotalTokens           int64   `json:"total_tokens"`
	ModelSelectionSavings float64 `json:"model_selection_savings"`
	CheaperModelCount     int64   `json:"cheaper_model_count"`
	ExpensiveModelCount   int64   `json:"expensive_model_count"`
}
