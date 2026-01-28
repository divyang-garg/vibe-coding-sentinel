// Phase 12: Utility Functions
// Common helper functions to reduce code duplication

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"sentinel-hub-api/llm"
	"sentinel-hub-api/models"
)

// marshalJSONB marshals a value to JSON string for JSONB storage
func marshalJSONB(v interface{}) (string, error) {
	if v == nil {
		return "null", nil
	}
	data, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return string(data), nil
}

// unmarshalJSONB unmarshals a JSON string from JSONB storage
func unmarshalJSONB(data string, v interface{}) error {
	if data == "" || data == "null" {
		return nil // Empty or null JSONB
	}
	if err := json.Unmarshal([]byte(data), v); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	return nil
}

// Phase 14E: Security Functions

// sanitizePath sanitizes a file path to prevent directory traversal attacks
func sanitizePath(p string) string {
	// Remove any ".." to prevent directory traversal
	return filepath.Clean(p)
}

// isValidPath validates that a path is safe to use
func isValidPath(p string) bool {
	// Check if path is absolute or relative and does not contain ".." after cleaning
	cleanPath := filepath.Clean(p)
	if strings.Contains(cleanPath, "..") {
		return false
	}

	// Prevent access to sensitive system directories
	sensitiveDirs := []string{
		"/etc", "/proc", "/sys", "/dev", "/boot", "/root", "/home",
		"C:\\Windows", "C:\\Program Files", "C:\\Users",
	}

	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return false
	}

	for _, sensitive := range sensitiveDirs {
		if strings.HasPrefix(absPath, sensitive) {
			return false
		}
	}

	return true
}

// NOTE: Deprecated AST functions (getParser, traverseAST, analyzeAST, ASTFinding) have been removed.
// All code should use the AST package directly: github.com/divyang-garg/sentinel-hub-api/hub/api/ast

// ImplementationEvidence represents evidence of rule implementation
// NOTE: This type is kept for backward compatibility with main package files.
// The actual implementation is in hub/api/services/doc_sync_business.go
type ImplementationEvidence struct {
	Feature     string           `json:"feature"`
	Files       []string         `json:"files"`
	Functions   []string         `json:"functions"`
	Endpoints   []string         `json:"endpoints"`
	Tests       []string         `json:"tests"`
	Confidence  float64          `json:"confidence"`
	LineNumbers map[string][]int `json:"line_numbers"` // function name or file path -> line numbers
}

// NOTE: Business rule detection functions have been moved to utils_business_rule.go
// to comply with CODING_STANDARDS.md file size limits (utilities max 250 lines)

// getLineColumn gets line and column from byte offset
func getLineColumn(code string, offset int) (int, int) {
	if offset >= len(code) {
		offset = len(code) - 1
	}
	if offset < 0 {
		offset = 0
	}
	lines := strings.Count(code[:offset], "\n")
	return lines + 1, 1
}

// Note: detectLanguageFromFile is defined in task_verifier.go
// Note: extractFunctionNameFromFile is defined in impact_analyzer.go

// getProjectFromContext extracts project from context
// This is a main package version that returns *Project
func getProjectFromContext(ctx context.Context) (*Project, error) {
	if project, ok := ctx.Value("project").(*Project); ok && project != nil {
		return project, nil
	}
	return nil, fmt.Errorf("project not found in context")
}

// calculateEstimatedCost calculates estimated LLM cost based on tokens
func calculateEstimatedCost(provider, model string, tokens int) float64 {
	// Pricing per 1K tokens (approximations)
	rates := map[string]map[string]float64{
		"openai": {
			"gpt-4": 0.03, "gpt-3.5-turbo": 0.002, "gpt-4-turbo": 0.03,
		},
		"anthropic": {
			"claude-3-opus": 0.015, "claude-3-sonnet": 0.003, "claude-3-haiku": 0.00025,
		},
	}
	if providerRates, ok := rates[provider]; ok {
		if rate, ok := providerRates[model]; ok {
			return (float64(tokens) / 1000.0) * rate
		}
	}
	return 0.0
}

// servicesTrackUsage is a bridge function to delegate to services.TrackUsage
// This allows main package to use the services package implementation with database persistence
// Uses *models.LLMUsage to ensure type compatibility across packages
var servicesTrackUsage func(ctx context.Context, usage *models.LLMUsage) error

// SetServicesTrackUsage sets the bridge function for trackUsage
// This should be called during application startup in main_minimal.go
func SetServicesTrackUsage(f func(ctx context.Context, usage *models.LLMUsage) error) {
	servicesTrackUsage = f
}

// trackUsage tracks LLM usage
// This function delegates to services.TrackUsage for database persistence
// Falls back to no-op if bridge is not initialized (backward compatibility)
func trackUsage(ctx context.Context, usage *LLMUsage) error {
	if servicesTrackUsage != nil {
		// Convert main.LLMUsage (alias to models.LLMUsage) to *models.LLMUsage
		// Since both are aliases, this is a safe conversion
		modelsUsage := (*models.LLMUsage)(usage)
		return servicesTrackUsage(ctx, modelsUsage)
	}
	// Fallback: return nil if bridge not initialized (maintains backward compatibility)
	// In production, this should be initialized during app startup
	return nil
}

// getLLMConfig retrieves LLM configuration for a project
func getLLMConfig(ctx context.Context, projectID string) (*LLMConfig, error) {
	// Query database for project-specific LLM configuration
	llmConfigs, err := llm.ListLLMConfigs(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list LLM configs: %w", err)
	}

	if len(llmConfigs) == 0 {
		// Fallback to default config if no configuration found
		return &LLMConfig{
			Provider: "openai",
			Model:    "gpt-3.5-turbo",
		}, nil
	}

	// Convert llm.LLMConfig to models.LLMConfig (return first/most recent)
	llmCfg := llmConfigs[0]
	config := &LLMConfig{
		ID:       llmCfg.ID,
		Provider: llmCfg.Provider,
		APIKey:   llmCfg.APIKey,
		Model:    llmCfg.Model,
		KeyType:  llmCfg.KeyType,
	}

	// Convert cost optimization config if present
	if llmCfg.CostOptimization != nil {
		config.CostOptimization = models.CostOptimizationConfig{
			UseCache:          llmCfg.CostOptimization.UseCache,
			CacheTTLHours:     llmCfg.CostOptimization.CacheTTLHours,
			ProgressiveDepth:  llmCfg.CostOptimization.ProgressiveDepth,
			MaxCostPerRequest: llmCfg.CostOptimization.MaxCostPerRequest,
		}
	}

	return config, nil
}

// NOTE: selectModelWithDepth and callLLMWithDepth have been moved to llm_cache_analysis.go
// to support Phase 14D cost optimization features. They are no longer in this file.

// requestIDKey is the context key for request ID
type ctxKey string

const requestIDKey ctxKey = "requestID"

// appendIfNotExists appends a string to a slice if not already present
func appendIfNotExists(slice []string, s string) []string {
	for _, v := range slice {
		if v == s {
			return slice
		}
	}
	return append(slice, s)
}

// detectTestFramework detects the test framework from file path
func detectTestFramework(filePath string) string {
	lower := strings.ToLower(filePath)
	switch {
	case strings.HasSuffix(lower, "_test.go"):
		return "go-testing"
	case strings.Contains(lower, ".test.js") || strings.Contains(lower, ".spec.js"):
		return "jest"
	case strings.Contains(lower, ".test.ts") || strings.Contains(lower, ".spec.ts"):
		return "jest"
	case strings.Contains(lower, "test_") && strings.HasSuffix(lower, ".py"):
		return "pytest"
	default:
		return "unknown"
	}
}

// contextKeyMain type for context keys in main package
type contextKeyMain string

// projectKey is the context key for project in main package
const projectKey contextKeyMain = "project"

// Note: countTestCases is defined in test_analyzer.go

// sanitizeString sanitizes a string for safe storage
func sanitizeString(s string, maxLen int) string {
	// Remove null bytes and control characters
	s = strings.Map(func(r rune) rune {
		if r < 32 && r != '\n' && r != '\r' && r != '\t' {
			return -1
		}
		return r
	}, s)
	// Truncate if too long
	if len(s) > maxLen {
		s = s[:maxLen]
	}
	return s
}

// NOTE: Keyword extraction functions have been moved to utils_keywords.go
// to comply with CODING_STANDARDS.md file size limits (utilities max 250 lines)
