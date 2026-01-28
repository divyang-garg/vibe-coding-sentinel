// Package ast provides helper functions for AST validator
// Complies with CODING_STANDARDS.md: Utility modules max 250 lines
package ast

import (
	"fmt"
	"path/filepath"
	"strings"
)

// validateDuplicateFunction validates if a function is truly a duplicate
func validateDuplicateFunction(funcName, filePath, projectRoot, language string) (ValidationResult, error) {
	// Search for other occurrences of the same function name
	pattern := BuildFunctionPattern(funcName, language)
	results, err := SearchCodebase(pattern, projectRoot, nil)
	if err != nil {
		return ValidationResult{}, fmt.Errorf("search failed: %w", err)
	}

	// Count occurrences in different files (excluding current file)
	fileSet := make(map[string]bool)
	for _, result := range results {
		relPath, _ := filepath.Rel(projectRoot, result.FilePath)
		// Exclude current file to check for duplicates in other files
		if relPath != filePath {
			fileSet[relPath] = true
		}
	}

	duplicateCount := len(fileSet)
	if duplicateCount > 0 {
		// Function exists in other files - confirmed duplicate
		return ValidationResult{
			FoundInCodebase: true,
			ReferenceCount:  duplicateCount,
			HasIntent:       false,
			IsExported:      isExportedIdentifier(funcName, language),
			Details:         fmt.Sprintf("Function found in %d other files", duplicateCount),
		}, nil
	}

	// Check for similar function names (fuzzy match)
	refPattern := BuildReferencePattern(funcName, language)
	refResults, err := SearchCodebase(refPattern, projectRoot, nil)
	if err != nil {
		return ValidationResult{}, fmt.Errorf("search failed: %w", err)
	}

	return ValidationResult{
		FoundInCodebase: len(refResults) > 1,
		ReferenceCount:  len(refResults),
		HasIntent:       false,
		IsExported:      isExportedIdentifier(funcName, language),
		Details:         fmt.Sprintf("Found %d references to function", len(refResults)),
	}, nil
}

// validateUnusedExport validates if an exported symbol is used externally
func validateUnusedExport(exportName, filePath, projectRoot, language string) (ValidationResult, error) {
	// Search for external usage of the export
	pattern := BuildReferencePattern(exportName, language)
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

	// Check if it's actually exported
	isExported := isExportedIdentifier(exportName, language)

	return ValidationResult{
		FoundInCodebase: externalRefs > 0,
		ReferenceCount:  externalRefs,
		HasIntent:       false,
		IsExported:      isExported,
		Details:         fmt.Sprintf("Found %d external references to export", externalRefs),
	}, nil
}

// validateUndefinedReference validates if a reference actually exists in the codebase
func validateUndefinedReference(refName, filePath, projectRoot, language string) (ValidationResult, error) {
	// Search for definition of the reference
	// Look for function definitions, variable declarations, etc.
	defPattern := BuildReferencePattern(refName, language)
	results, err := SearchCodebase(defPattern, projectRoot, nil)
	if err != nil {
		return ValidationResult{}, fmt.Errorf("search failed: %w", err)
	}

	// Count definitions (not just references) in other files
	defCount := 0
	for _, result := range results {
		relPath, _ := filepath.Rel(projectRoot, result.FilePath)
		// Exclude current file - we're checking if reference is defined elsewhere
		if relPath != filePath {
			// Check if this looks like a definition (function, var, const, etc.)
			content := strings.ToLower(result.Content)
			if strings.Contains(content, "func "+refName) ||
				strings.Contains(content, "var "+refName) ||
				strings.Contains(content, "const "+refName) ||
				strings.Contains(content, "def "+refName) ||
				strings.Contains(content, "function "+refName) {
				defCount++
			}
		}
	}

	return ValidationResult{
		FoundInCodebase: defCount > 0,
		ReferenceCount:  defCount,
		HasIntent:       false,
		IsExported:      isExportedIdentifier(refName, language),
		Details:         fmt.Sprintf("Found %d definitions for reference", defCount),
	}, nil
}

// validateCircularDependency validates circular dependencies
func validateCircularDependency(filePath, projectRoot, language string) (ValidationResult, error) {
	// Circular dependencies are structural issues that require cross-file analysis
	// This is a simplified validation - full validation requires dependency graph analysis
	// For now, we check if the file has imports that might create cycles
	importPattern := BuildImportPattern("", language)
	results, err := SearchCodebase(importPattern, projectRoot, nil)
	if err != nil {
		return ValidationResult{}, fmt.Errorf("search failed: %w", err)
	}

	// Count imports in this file
	importCount := 0
	for _, result := range results {
		relPath, _ := filepath.Rel(projectRoot, result.FilePath)
		if relPath == filePath {
			importCount++
		}
	}

	// Circular dependencies are confirmed by cross-file analysis
	// This validation confirms the file is part of a dependency chain
	return ValidationResult{
		FoundInCodebase: importCount > 0,
		ReferenceCount:  importCount,
		HasIntent:       false,
		IsExported:      false,
		Details:         fmt.Sprintf("File has %d imports, may be part of circular dependency", importCount),
	}, nil
}

// validateCrossFileDuplicate validates if a function is duplicated across files
func validateCrossFileDuplicate(funcName, filePath, projectRoot, language string) (ValidationResult, error) {
	// Search for function definitions across the codebase
	pattern := BuildFunctionPattern(funcName, language)
	results, err := SearchCodebase(pattern, projectRoot, nil)
	if err != nil {
		return ValidationResult{}, fmt.Errorf("search failed: %w", err)
	}

	// Count occurrences in different files (excluding current file)
	fileSet := make(map[string]bool)
	for _, result := range results {
		relPath, _ := filepath.Rel(projectRoot, result.FilePath)
		// Exclude current file - checking for cross-file duplicates
		if relPath != filePath {
			fileSet[relPath] = true
		}
	}

	duplicateCount := len(fileSet)
	isExported := isExportedIdentifier(funcName, language)

	return ValidationResult{
		FoundInCodebase: duplicateCount > 0,
		ReferenceCount:  duplicateCount,
		HasIntent:       false,
		IsExported:      isExported,
		Details:         fmt.Sprintf("Function found in %d other files", duplicateCount),
	}, nil
}

// extractExportNameFromFinding extracts export name from finding message
func extractExportNameFromFinding(finding *ASTFinding) string {
	// Message format: "Exported function 'exportName' is never used..."
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
		// Look for "export function", "export const", etc.
		if strings.Contains(firstLine, "export ") {
			parts := strings.Fields(firstLine)
			for i, part := range parts {
				if part == "export" && i+1 < len(parts) {
					// Next token might be function/const/let/var, skip to name
					if i+2 < len(parts) {
						return strings.TrimSuffix(parts[i+2], "(")
					}
				}
			}
		}
	}

	return ""
}

// extractReferenceNameFromFinding extracts reference name from finding message
func extractReferenceNameFromFinding(finding *ASTFinding) string {
	// Message format: "Reference to undefined symbol 'refName'"
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
		// Look for identifier usage
		parts := strings.Fields(firstLine)
		if len(parts) > 0 {
			// First identifier is likely the reference
			return strings.TrimSuffix(parts[0], "(")
		}
	}

	return ""
}
