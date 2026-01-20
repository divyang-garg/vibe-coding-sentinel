// Package ast provides codebase validation for AST findings
// Complies with CODING_STANDARDS.md: Detection modules max 250 lines
package ast

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ValidationResult contains the result of validating a finding against the codebase
type ValidationResult struct {
	FoundInCodebase bool
	ReferenceCount  int
	HasIntent       bool
	IsExported      bool
	Details         string
}

// ValidateFinding validates a finding against the full codebase
func ValidateFinding(finding *ASTFinding, filePath, projectRoot, language string) error {
	if finding == nil {
		return fmt.Errorf("finding is nil")
	}

	// Auto-detect language from file path if not provided
	if language == "" {
		language = ExtractLanguageFromPath(filePath)
	}

	var result ValidationResult
	var err error

	switch finding.Type {
	case "orphaned_code":
		// Extract function name from message or code
		funcName := extractFunctionNameFromFinding(finding)
		if funcName != "" {
			result, err = validateOrphanedFunction(funcName, filePath, projectRoot, language)
		}
	case "unused_variable":
		// Extract variable name from message
		varName := extractVariableName(finding)
		if varName != "" {
			result, err = validateUnusedVariable(varName, filePath, projectRoot, language)
		}
	case "empty_catch":
		result, err = validateEmptyCatch(filePath, finding.Line, projectRoot)
	default:
		// For other types, set default validation
		result = ValidationResult{
			FoundInCodebase: false,
			ReferenceCount:  0,
			HasIntent:       false,
			IsExported:      false,
			Details:         "Validation not implemented for this finding type",
		}
	}

	if err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	// Calculate confidence and auto-fix safety
	confidence := CalculateConfidence(finding, result)
	autoFixSafe := DetermineAutoFixSafe(confidence, finding.Type)
	reasoning := GenerateReasoning(finding, result, confidence, autoFixSafe)
	fixType := DetermineFixType(finding.Type)

	// Update finding with validation results
	finding.Validated = true
	finding.Confidence = confidence
	finding.AutoFixSafe = autoFixSafe
	finding.Reasoning = reasoning
	finding.FixType = fixType

	return nil
}

// validateOrphanedFunction checks if a function is truly orphaned by searching the codebase
func validateOrphanedFunction(funcName, filePath, projectRoot, language string) (ValidationResult, error) {
	// Search for function calls: funcName(
	callPattern := BuildFunctionPattern(funcName, language)
	callResults, err := SearchCodebase(callPattern, projectRoot, nil)
	if err != nil {
		return ValidationResult{}, fmt.Errorf("search failed: %w", err)
	}

	// Search for function references: funcName (passed as value, assigned)
	refPattern := BuildReferencePattern(funcName, language)
	refResults, err := SearchCodebase(refPattern, projectRoot, nil)
	if err != nil {
		return ValidationResult{}, fmt.Errorf("search failed: %w", err)
	}

	// Combine and deduplicate results
	allResults := deduplicateResults(callResults, refResults)

	// Filter out matches in the same file (self-references don't count)
	externalRefs := 0
	for _, result := range allResults {
		relPath, _ := filepath.Rel(projectRoot, result.FilePath)
		if relPath != filePath {
			externalRefs++
		}
	}

	// Check if function is exported (language-specific)
	isExported := isExportedIdentifier(funcName, language)

	return ValidationResult{
		FoundInCodebase: externalRefs > 0,
		ReferenceCount:  externalRefs,
		HasIntent:       false,
		IsExported:      isExported,
		Details:         fmt.Sprintf("Found %d external references", externalRefs),
	}, nil
}

// validateUnusedVariable checks if a variable is truly unused
func validateUnusedVariable(varName, filePath, projectRoot, language string) (ValidationResult, error) {
	// Search for variable usage
	pattern := BuildReferencePattern(varName, language)
	results, err := SearchCodebase(pattern, projectRoot, nil)
	if err != nil {
		return ValidationResult{}, fmt.Errorf("search failed: %w", err)
	}

	// Count references outside the file
	externalRefs := 0
	for _, result := range results {
		relPath, _ := filepath.Rel(projectRoot, result.FilePath)
		if relPath != filePath {
			externalRefs++
		}
	}

	// Check if variable is exported (language-specific)
	isExported := isExportedIdentifier(varName, language)

	return ValidationResult{
		FoundInCodebase: externalRefs > 0,
		ReferenceCount:  externalRefs,
		HasIntent:       false,
		IsExported:      isExported,
		Details:         fmt.Sprintf("Found %d external references", externalRefs),
	}, nil
}

// validateEmptyCatch checks if an empty catch block has intent comments
func validateEmptyCatch(filePath string, line int, projectRoot string) (ValidationResult, error) {
	hasIntent := CheckIntentComment(filePath, line, projectRoot)

	return ValidationResult{
		FoundInCodebase: false,
		ReferenceCount:  0,
		HasIntent:       hasIntent,
		IsExported:      false,
		Details:         fmt.Sprintf("Intent comment found: %v", hasIntent),
	}, nil
}

// extractFunctionName extracts function name from finding message or code
func extractFunctionNameFromFinding(finding *ASTFinding) string {
	// Try to extract from message: "Potentially orphaned function: 'funcName' is defined..."
	if strings.Contains(finding.Message, "'") {
		parts := strings.Split(finding.Message, "'")
		if len(parts) >= 2 {
			return parts[1]
		}
	}

	// Try to extract from code snippet (first identifier)
	codeLines := strings.Split(finding.Code, "\n")
	if len(codeLines) > 0 {
		firstLine := strings.TrimSpace(codeLines[0])
		// Look for "func funcName" pattern
		if strings.HasPrefix(firstLine, "func ") {
			parts := strings.Fields(firstLine)
			if len(parts) >= 2 {
				// Remove parentheses if present: funcName() -> funcName
				funcName := strings.TrimSuffix(parts[1], "(")
				return strings.TrimSpace(funcName)
			}
		}
	}

	return ""
}

// extractVariableName extracts variable name from finding message
func extractVariableName(finding *ASTFinding) string {
	// Message format: "Unused variable: 'varName' is declared..."
	if strings.Contains(finding.Message, "'") {
		parts := strings.Split(finding.Message, "'")
		if len(parts) >= 2 {
			return parts[1]
		}
	}

	// Try to extract from code snippet
	codeLines := strings.Split(finding.Code, "\n")
	if len(codeLines) > 0 {
		firstLine := strings.TrimSpace(codeLines[0])
		// Look for "var varName" or "varName :=" patterns
		if strings.HasPrefix(firstLine, "var ") {
			parts := strings.Fields(firstLine)
			if len(parts) >= 2 {
				return parts[1]
			}
		} else if strings.Contains(firstLine, ":=") {
			parts := strings.Split(firstLine, ":=")
			if len(parts) >= 1 {
				vars := strings.Fields(parts[0])
				if len(vars) > 0 {
					return vars[len(vars)-1]
				}
			}
		}
	}

	return ""
}

// deduplicateResults removes duplicate search results
func deduplicateResults(results1, results2 []SearchResult) []SearchResult {
	seen := make(map[string]bool)
	combined := []SearchResult{}

	for _, r := range results1 {
		key := fmt.Sprintf("%s:%d", r.FilePath, r.Line)
		if !seen[key] {
			seen[key] = true
			combined = append(combined, r)
		}
	}

	for _, r := range results2 {
		key := fmt.Sprintf("%s:%d", r.FilePath, r.Line)
		if !seen[key] {
			seen[key] = true
			combined = append(combined, r)
		}
	}

	return combined
}

// isExportedIdentifier checks if an identifier is exported based on language rules
func isExportedIdentifier(name, language string) bool {
	if name == "" {
		return false
	}
	// Python: underscore prefix means private
	if language == "python" {
		return !strings.HasPrefix(name, "_")
	}
	// Go/JS/TS: uppercase first letter means exported
	return len(name) > 0 && strings.ToUpper(string(name[0])) == string(name[0])
}
