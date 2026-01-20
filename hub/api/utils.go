// Phase 12: Utility Functions
// Common helper functions to reduce code duplication

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
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

// getParser returns a parser for a language (stub - requires tree-sitter)
func getParser(language string) (interface{}, error) {
	return nil, fmt.Errorf("getParser not implemented (tree-sitter integration required)")
}

// traverseAST traverses an AST tree (stub - requires tree-sitter)
func traverseAST(node interface{}, callback interface{}) {
	// Stub - tree-sitter integration required
}

// ASTFinding represents a finding from AST analysis
type ASTFinding struct {
	Type    string `json:"type"`
	Line    int    `json:"line"`
	Column  int    `json:"column"`
	Message string `json:"message"`
}

// analyzeAST analyzes code using AST (stub - requires tree-sitter)
func analyzeAST(code, language string, options []string) (interface{}, []ASTFinding, error) {
	// Stub - tree-sitter integration required
	return nil, []ASTFinding{}, fmt.Errorf("analyzeAST not implemented (tree-sitter integration required)")
}

// ImplementationEvidence represents evidence of rule implementation
type ImplementationEvidence struct {
	Feature     string   `json:"feature"`
	Files       []string `json:"files"`
	Functions   []string `json:"functions"`
	Endpoints   []string `json:"endpoints"`
	Tests       []string `json:"tests"`
	Confidence  float64  `json:"confidence"`
	LineNumbers []int    `json:"line_numbers"`
}

// detectBusinessRuleImplementation detects business rule implementations (stub)
func detectBusinessRuleImplementation(rule KnowledgeItem, codebasePath string) ImplementationEvidence {
	// Stub - would analyze codebase for business rule implementation
	return ImplementationEvidence{
		Feature:     "",
		Files:       []string{},
		Functions:   []string{},
		Endpoints:   []string{},
		Tests:       []string{},
		Confidence:  0.0,
		LineNumbers: []int{},
	}
}

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

// extractFunctionSignature extracts function signature from code (stub)
func extractFunctionSignature(node interface{}, code string, language string) string {
	return ""
}

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

// trackUsage tracks LLM usage (stub - in production save to database)
func trackUsage(ctx context.Context, usage *LLMUsage) error {
	// TODO: Implement database persistence
	return nil
}

// getLLMConfig retrieves LLM configuration for a project
func getLLMConfig(ctx context.Context, projectID string) (*LLMConfig, error) {
	// Return a default config for now - in production query database
	return &LLMConfig{
		Provider: "openai",
		Model:    "gpt-3.5-turbo",
	}, nil
}

// selectModelWithDepth selects LLM model based on depth (stub)
func selectModelWithDepth(ctx context.Context, projectID string, config *LLMConfig, mode string, depth int, feature string) (string, error) {
	if config != nil {
		return config.Model, nil
	}
	return "", fmt.Errorf("no LLM config provided")
}

// callLLMWithDepth calls LLM with depth settings (stub)
// Note: LLM functions are implemented in services package
func callLLMWithDepth(ctx context.Context, config *LLMConfig, prompt string, taskType string, model string, mode string) (string, int, error) {
	return "", 0, fmt.Errorf("callLLMWithDepth not implemented - use services package")
}

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

// extractFeatureKeywords extracts keywords from feature name
func extractFeatureKeywords(featureName string) []string {
	var keywords []string
	words := []rune(featureName)
	var current []rune
	for _, r := range words {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			current = append(current, r)
		} else {
			if len(current) > 0 {
				keywords = append(keywords, string(current))
				current = nil
			}
		}
	}
	if len(current) > 0 {
		keywords = append(keywords, string(current))
	}
	return keywords
}

// extractKeywords extracts meaningful keywords from text
// Phase 14E: Shared function for task verification and dependency detection
func extractKeywords(text string) []string {
	// Simple keyword extraction - split by common separators
	words := strings.FieldsFunc(text, func(r rune) bool {
		return r == ' ' || r == '-' || r == '_' || r == '(' || r == ')' || r == '[' || r == ']'
	})

	keywords := []string{}
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true, "but": true,
		"in": true, "on": true, "at": true, "to": true, "for": true, "of": true,
		"with": true, "by": true, "is": true, "are": true, "was": true, "were": true,
		"be": true, "been": true, "have": true, "has": true, "had": true,
		"do": true, "does": true, "did": true, "will": true, "would": true,
		"should": true, "could": true, "may": true, "might": true,
	}

	for _, word := range words {
		word = strings.ToLower(strings.TrimSpace(word))
		if len(word) > 2 && !stopWords[word] {
			keywords = append(keywords, word)
		}
	}

	return keywords
}
