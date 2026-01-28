// Package ast provides confidence scoring for AST findings
// Complies with CODING_STANDARDS.md: Utility modules max 250 lines
package ast

import (
	"fmt"
	"strings"
)

// CalculateConfidence calculates confidence score based on finding type and validation result
func CalculateConfidence(finding *ASTFinding, validationResult ValidationResult) float64 {
	if finding == nil {
		return 0.0
	}

	switch finding.Type {
	case "orphaned_code":
		return calculateOrphanedConfidence(finding, validationResult)
	case "unused_variable":
		return calculateUnusedVariableConfidence(finding, validationResult)
	case "empty_catch":
		return calculateEmptyCatchConfidence(finding, validationResult)
	case "duplicate_function":
		return 0.80 // Needs human choice, not auto-fixable
	default:
		return 0.50 // Default moderate confidence
	}
}

// CalculateConfidenceWithEdgeCases calculates confidence with edge case penalties
func CalculateConfidenceWithEdgeCases(finding *ASTFinding, validationResult ValidationResult, edgeCases EdgeCaseResult) float64 {
	base := CalculateConfidence(finding, validationResult)

	// Apply edge case penalty
	adjusted := base - edgeCases.ConfidencePenalty

	// Ensure confidence doesn't go below 0
	if adjusted < 0.0 {
		adjusted = 0.0
	}

	return adjusted
}

// calculateOrphanedConfidence calculates confidence for orphaned code findings
func calculateOrphanedConfidence(_ *ASTFinding, result ValidationResult) float64 {
	// If references found in codebase, confidence is 0 (not orphaned)
	if result.FoundInCodebase {
		return 0.0
	}

	// If exported, never auto-fix (might be public API)
	if result.IsExported {
		return 0.0
	}

	// No references found, high confidence it's orphaned
	return 0.95
}

// calculateUnusedVariableConfidence calculates confidence for unused variable findings
func calculateUnusedVariableConfidence(_ *ASTFinding, result ValidationResult) float64 {
	// If exported, never auto-fix
	if result.IsExported {
		return 0.0
	}

	// If found in other files, might be used via reflection or other mechanisms
	if result.FoundInCodebase {
		return 0.0
	}

	// Local scope only, high confidence (safe to delete)
	return 0.95
}

// calculateEmptyCatchConfidence calculates confidence for empty catch block findings
func calculateEmptyCatchConfidence(_ *ASTFinding, result ValidationResult) float64 {
	// If intent comment found, don't auto-fix
	if result.HasIntent {
		return 0.0
	}

	// No intent comment, high confidence it's safe to fix
	return 0.85
}

// DetermineAutoFixSafe determines if a finding is safe for automated refactoring
func DetermineAutoFixSafe(confidence float64, findingType string) bool {
	// Require 95%+ confidence for auto-fix
	if confidence < 0.95 {
		return false
	}

	// Some finding types should never be auto-fixed
	neverAutoFix := []string{"duplicate_function", "syntax_error"}
	for _, neverType := range neverAutoFix {
		if findingType == neverType {
			return false
		}
	}

	return true
}

// GenerateReasoning generates a human-readable explanation for the confidence and auto-fix decision
func GenerateReasoning(finding *ASTFinding, validationResult ValidationResult, confidence float64, autoFixSafe bool) string {
	var parts []string

	switch finding.Type {
	case "orphaned_code":
		if validationResult.FoundInCodebase {
			parts = append(parts, fmt.Sprintf("Function is referenced %d times in other files", validationResult.ReferenceCount))
			parts = append(parts, "Not safe to remove")
		} else if validationResult.IsExported {
			parts = append(parts, "Function is exported (public API)")
			parts = append(parts, "May be used by external code")
		} else {
			parts = append(parts, "No references found in codebase")
			parts = append(parts, "High confidence function is orphaned")
		}

	case "unused_variable":
		if validationResult.IsExported {
			parts = append(parts, "Variable is exported (public API)")
			parts = append(parts, "May be used by external code")
		} else if validationResult.FoundInCodebase {
			parts = append(parts, fmt.Sprintf("Variable referenced %d times in other files", validationResult.ReferenceCount))
			parts = append(parts, "May be used via reflection or other mechanisms")
		} else {
			parts = append(parts, "Variable is local scope only")
			parts = append(parts, "No external references found")
		}

	case "empty_catch":
		if validationResult.HasIntent {
			parts = append(parts, "Intent comment (TODO/FIXME) found nearby")
			parts = append(parts, "Empty catch block may be intentional")
		} else {
			parts = append(parts, "No intent comments found")
			parts = append(parts, "Empty catch block likely unintentional")
		}

	default:
		parts = append(parts, "Standard validation applied")
	}

	// Add confidence and auto-fix status
	parts = append(parts, fmt.Sprintf("Confidence: %.0f%%", confidence*100))
	if autoFixSafe {
		parts = append(parts, "Safe for automated refactoring")
	} else {
		parts = append(parts, "Requires human review")
	}

	return strings.Join(parts, ". ")
}

// DetermineFixType determines the appropriate fix type for a finding
func DetermineFixType(findingType string) string {
	switch findingType {
	case "orphaned_code":
		return "delete"
	case "unused_variable":
		return "delete"
	case "empty_catch":
		return "refactor"
	case "duplicate_function":
		return "refactor"
	default:
		return "comment"
	}
}
