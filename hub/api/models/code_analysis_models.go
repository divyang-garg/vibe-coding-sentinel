// Package models - Code analysis data models
// Complies with CODING_STANDARDS.md: Data Models max 200 lines
package models

// ASTAnalysisRequest represents a request for AST analysis
type ASTAnalysisRequest struct {
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
