// Phase 14A: Business Logic Analyzer
// Analyzes business logic functions for correctness, error handling, and semantic issues

package services

import (
	"context"
	"fmt"
	"os"
	"strings"
)

// LogicLayerFinding represents a finding from business logic analysis
type LogicLayerFinding struct {
	Type     string `json:"type"`     // "semantic_error", "missing_error_handling", "signature_mismatch"
	Location string `json:"location"` // File path and line number
	Issue    string `json:"issue"`
	Severity string `json:"severity"` // "critical", "high", "medium", "low"
}

// analyzeBusinessLogic analyzes business logic functions
func analyzeBusinessLogic(ctx context.Context, projectID string, feature *DiscoveredFeature) ([]LogicLayerFinding, error) {
	return analyzeBusinessLogicWithDepth(ctx, projectID, feature, "medium")
}

// analyzeBusinessLogicWithDepth analyzes business logic functions with specified depth
// Phase 14D: Added depth parameter to control LLM usage
func analyzeBusinessLogicWithDepth(ctx context.Context, projectID string, feature *DiscoveredFeature, depth string) ([]LogicLayerFinding, error) {
	findings := []LogicLayerFinding{}

	if feature.LogicLayer == nil {
		return findings, nil
	}

	// Use AST analyzer to analyze functions
	for _, function := range feature.LogicLayer.Functions {
		// Read function file
		data, err := os.ReadFile(function.File)
		if err != nil {
			LogWarn(ctx, "Failed to read function file %s: %v", function.File, err)
			continue
		}

		// Analyze error handling (always runs, no LLM)
		errorHandlingFindings := analyzeErrorHandling(string(data), function)
		findings = append(findings, errorHandlingFindings...)

		// Phase 14D: Skip LLM semantic analysis for surface depth
		if depth == "surface" {
			// Use pattern-based checks only
			semanticFindings := checkSemanticIssues(string(data), function)
			findings = append(findings, semanticFindings...)
			continue
		}

		// Perform semantic analysis with LLM (if configured)
		// Get business rules for context
		var businessRule interface{} = nil
		// In production, would fetch relevant business rules for this function

		// Try semantic analysis with LLM (respects depth parameter)
		semanticFindings, err := semanticAnalysis(ctx, projectID, function, businessRule, depth)
		if err != nil {
			LogWarn(ctx, "Semantic analysis failed for function %s: %v", function.Name, err)
			// Fall back to pattern-based checks
			semanticFindings = checkSemanticIssues(string(data), function)
		}
		findings = append(findings, semanticFindings...)
	}

	return findings, nil
}

// analyzeErrorHandling analyzes error handling in business logic functions
func analyzeErrorHandling(code string, function BusinessLogicFunctionInfo) []LogicLayerFinding {
	findings := []LogicLayerFinding{}

	// Detect error handling patterns based on language
	// This is simplified - would use AST in production

	// Check for try-catch (JavaScript/TypeScript/Python)
	hasTryCatch := strings.Contains(code, "try") && strings.Contains(code, "catch")

	// Check for error returns (Go)
	hasErrorReturn := strings.Contains(code, "error") && strings.Contains(code, "return")

	// Check for error handling in function
	if !hasTryCatch && !hasErrorReturn {
		findings = append(findings, LogicLayerFinding{
			Type:     "missing_error_handling",
			Location: fmt.Sprintf("%s:%d", function.File, function.LineNumber),
			Issue:    fmt.Sprintf("Function %s may be missing error handling", function.Name),
			Severity: "high",
		})
	}

	return findings
}

// checkSemanticIssues checks for semantic issues in business logic
func checkSemanticIssues(code string, function BusinessLogicFunctionInfo) []LogicLayerFinding {
	findings := []LogicLayerFinding{}

	// Simplified semantic checks
	// In production, would use LLM for semantic analysis

	// Check for potential null/undefined issues
	if strings.Contains(code, ".") && !strings.Contains(code, "?.") && !strings.Contains(code, "if") {
		// Potential null reference (simplified check)
		findings = append(findings, LogicLayerFinding{
			Type:     "semantic_error",
			Location: fmt.Sprintf("%s:%d", function.File, function.LineNumber),
			Issue:    fmt.Sprintf("Function %s may have potential null reference issues", function.Name),
			Severity: "medium",
		})
	}

	return findings
}
