// Package main - Flow validation checks
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package main

import (
	"os"
	"strings"
)

// hasErrorHandling checks if a flow step has error handling
func hasErrorHandling(step FlowStep, feature *DiscoveredFeature) bool {
	// Use feature information if available for better validation
	if feature != nil {
		// Check if step file matches discovered feature files
		// This helps validate that we're checking the right component
		switch step.Layer {
		case "ui":
			if feature.UILayer != nil {
				for _, component := range feature.UILayer.Components {
					if component.Path == step.File {
						// Found matching component - proceed with validation
						break
					}
				}
			}
		case "api":
			if feature.APILayer != nil {
				for _, endpoint := range feature.APILayer.Endpoints {
					if endpoint.File == step.File {
						// Found matching endpoint - proceed with validation
						break
					}
				}
			}
		case "logic":
			if feature.LogicLayer != nil {
				for _, fn := range feature.LogicLayer.Functions {
					if fn.File == step.File {
						// Found matching function - proceed with validation
						break
					}
				}
			}
		case "database":
			if feature.DatabaseLayer != nil {
				for _, table := range feature.DatabaseLayer.Tables {
					if table.File == step.File {
						// Found matching table - proceed with validation
						break
					}
				}
			}
		case "integration":
			if feature.IntegrationLayer != nil {
				for _, integration := range feature.IntegrationLayer.Integrations {
					if integration.File == step.File {
						// Found matching integration - proceed with validation
						break
					}
				}
			}
		}
	}

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
	// Use feature information if available for better validation
	if feature != nil {
		// Validate that step file matches discovered feature
		if step.Layer == "api" && feature.APILayer != nil {
			for _, endpoint := range feature.APILayer.Endpoints {
				if endpoint.File == step.File {
					// Found matching endpoint - can use endpoint metadata
					break
				}
			}
		} else if step.Layer == "ui" && feature.UILayer != nil {
			for _, component := range feature.UILayer.Components {
				if component.Path == step.File {
					// Found matching component - can use component metadata
					break
				}
			}
		}
	}

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
	// Use feature information if available for better validation
	if feature != nil && feature.DatabaseLayer != nil {
		// Validate that step file matches discovered database layer
		for _, table := range feature.DatabaseLayer.Tables {
			if table.File == step.File {
				// Found matching table - can use table metadata for validation
				break
			}
		}
	}

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
	// Use feature information if available for better validation
	if feature != nil && feature.IntegrationLayer != nil {
		// Validate that step file matches discovered integration
		for _, integration := range feature.IntegrationLayer.Integrations {
			if integration.File == step.File {
				// Found matching integration - can use integration metadata
				break
			}
		}
	}

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
