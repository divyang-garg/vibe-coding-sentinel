// Package main - Flow validation checks
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package main

import (
	"os"
	"strings"
)

// hasErrorHandling checks if a flow step has error handling
func hasErrorHandling(step FlowStep, feature *DiscoveredFeature) bool {
	if step.File == "" {
		return false
	}

	// Read file content
	content, err := os.ReadFile(step.File)
	if err != nil {
		return false
	}

	codeContent := strings.ToLower(string(content))

	// Check for error handling patterns based on layer
	switch step.Layer {
	case "ui":
		// React: try-catch, error boundaries, .catch()
		return strings.Contains(codeContent, "catch") ||
			strings.Contains(codeContent, "errorboundary") ||
			strings.Contains(codeContent, ".catch(") ||
			strings.Contains(codeContent, "onerror")
	case "api":
		// Go: if err != nil, error return, panic recovery
		return strings.Contains(codeContent, "if err") ||
			strings.Contains(codeContent, "error") ||
			strings.Contains(codeContent, "recover()") ||
			strings.Contains(codeContent, "defer")
	case "logic":
		// Business logic: error handling, validation
		return strings.Contains(codeContent, "error") ||
			strings.Contains(codeContent, "err") ||
			strings.Contains(codeContent, "exception") ||
			strings.Contains(codeContent, "catch")
	case "database":
		// Database: transaction rollback, error handling
		return strings.Contains(codeContent, "rollback") ||
			strings.Contains(codeContent, "error") ||
			strings.Contains(codeContent, "catch")
	case "integration":
		// External API: error handling, retry logic
		return strings.Contains(codeContent, "error") ||
			strings.Contains(codeContent, "catch") ||
			strings.Contains(codeContent, "retry")
	}

	return false
}

// hasValidation checks if a flow step has input validation
func hasValidation(step FlowStep, feature *DiscoveredFeature) bool {
	if step.File == "" {
		return false
	}

	content, err := os.ReadFile(step.File)
	if err != nil {
		return false
	}

	codeContent := strings.ToLower(string(content))

	// Check for validation patterns
	validationPatterns := []string{
		"validate",
		"validation",
		"validator",
		"schema",
		"required",
		"check",
		"verify",
		"assert",
		"zod",       // TypeScript validation
		"yup",       // JavaScript validation
		"joi",       // JavaScript validation
		"validator", // Go validation
	}

	for _, pattern := range validationPatterns {
		if strings.Contains(codeContent, pattern) {
			return true
		}
	}

	return false
}

// hasRollback checks if a flow step has transaction rollback capability
func hasRollback(step FlowStep, feature *DiscoveredFeature) bool {
	if step.File == "" {
		return false
	}

	content, err := os.ReadFile(step.File)
	if err != nil {
		return false
	}

	codeContent := strings.ToLower(string(content))

	// Check for rollback/transaction patterns
	return strings.Contains(codeContent, "rollback") ||
		strings.Contains(codeContent, "transaction") ||
		strings.Contains(codeContent, "begin") ||
		strings.Contains(codeContent, "commit") ||
		strings.Contains(codeContent, "undo") ||
		strings.Contains(codeContent, "revert")
}

// hasTimeout checks if a flow step has timeout handling
func hasTimeout(step FlowStep, feature *DiscoveredFeature) bool {
	if step.File == "" {
		return false
	}

	content, err := os.ReadFile(step.File)
	if err != nil {
		return false
	}

	codeContent := strings.ToLower(string(content))

	// Check for timeout configuration
	return strings.Contains(codeContent, "timeout") ||
		strings.Contains(codeContent, "context.timeout") ||
		strings.Contains(codeContent, "deadline") ||
		strings.Contains(codeContent, "withtimeout")
}
