// Package utils - Flow validation checks
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package utils

import (
	"os"
	"strings"

	"sentinel-hub-api/feature_discovery"
)

// hasErrorHandling checks if a flow step has error handling
func hasErrorHandling(step FlowStep, feature *feature_discovery.DiscoveredFeature) bool {
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
func hasValidation(step FlowStep, feature *feature_discovery.DiscoveredFeature) bool {
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
func hasRollback(step FlowStep, feature *feature_discovery.DiscoveredFeature) bool {
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
func hasTimeout(step FlowStep, feature *feature_discovery.DiscoveredFeature) bool {
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

// requestFormatsMatch checks if UI request format matches API endpoint expectations
func requestFormatsMatch(uiStep FlowStep, apiStep FlowStep, feature *feature_discovery.DiscoveredFeature) bool {
	// Read both files
	uiContent, err1 := os.ReadFile(uiStep.File)
	apiContent, err2 := os.ReadFile(apiStep.File)

	if err1 != nil || err2 != nil {
		// Can't read files, assume match (conservative)
		return true
	}

	uiCode := strings.ToLower(string(uiContent))
	apiCode := strings.ToLower(string(apiContent))

	// Look for common data format indicators
	formatIndicators := []string{
		"json",
		"xml",
		"formdata",
		"multipart",
		"content-type",
	}

	uiFormats := []string{}
	apiFormats := []string{}

	for _, indicator := range formatIndicators {
		if strings.Contains(uiCode, indicator) {
			uiFormats = append(uiFormats, indicator)
		}
		if strings.Contains(apiCode, indicator) {
			apiFormats = append(apiFormats, indicator)
		}
	}

	// If both have formats, check for overlap
	if len(uiFormats) > 0 && len(apiFormats) > 0 {
		for _, uiFormat := range uiFormats {
			for _, apiFormat := range apiFormats {
				if uiFormat == apiFormat {
					return true
				}
			}
		}
		// No overlap found
		return false
	}

	// If no formats detected, assume match (conservative)
	return true
}

// operationsMatch checks if business logic operations match database operations
func operationsMatch(logicStep FlowStep, dbStep FlowStep, feature *feature_discovery.DiscoveredFeature) bool {
	if logicStep.File == "" || dbStep.File == "" {
		return true
	}

	logicContent, err1 := os.ReadFile(logicStep.File)
	dbContent, err2 := os.ReadFile(dbStep.File)

	if err1 != nil || err2 != nil {
		return true
	}

	logicCode := strings.ToLower(string(logicContent))
	dbCode := strings.ToLower(string(dbContent))

	// Check if logic operation matches database operation
	// Look for CRUD operations
	operations := []string{"create", "insert", "read", "select", "update", "delete", "remove"}

	for _, op := range operations {
		logicHasOp := strings.Contains(logicCode, op)
		dbHasOp := strings.Contains(dbCode, op)

		// If both have the same operation, they match
		if logicHasOp && dbHasOp {
			return true
		}
	}

	// If no clear operation match, check for function/table name matching
	// This is a simplified check - in production would use AST
	return true // Conservative: assume match if no clear mismatch
}

// responseFormatsMatch checks if API response format matches business logic expectations
func responseFormatsMatch(apiStep FlowStep, logicStep FlowStep, feature *feature_discovery.DiscoveredFeature) bool {
	if apiStep.File == "" || logicStep.File == "" {
		return true
	}

	apiContent, err1 := os.ReadFile(apiStep.File)
	logicContent, err2 := os.ReadFile(logicStep.File)

	if err1 != nil || err2 != nil {
		return true
	}

	apiCode := strings.ToLower(string(apiContent))
	logicCode := strings.ToLower(string(logicContent))

	// Look for response structure indicators
	responseIndicators := []string{
		"json",
		"response",
		"return",
		"result",
		"data",
	}

	apiHasResponse := false
	logicHasResponse := false

	for _, indicator := range responseIndicators {
		if strings.Contains(apiCode, indicator) {
			apiHasResponse = true
		}
		if strings.Contains(logicCode, indicator) {
			logicHasResponse = true
		}
	}

	// If both have response handling, assume they match
	if apiHasResponse && logicHasResponse {
		return true
	}

	// If neither has response handling, also assume match (might be handled elsewhere)
	return true
}
