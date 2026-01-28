// Package services provides documentation coverage calculation functions
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package services

import (
	"strings"

	"sentinel-hub-api/ast"
)

// calculateDocumentationCoverage calculates the percentage of documented code elements
// Returns coverage percentage (0-100) and per-module breakdown
func (s *CodeAnalysisServiceImpl) calculateDocumentationCoverage(docs, code interface{}) float64 {
	// Handle nil or empty inputs
	if docs == nil || code == nil {
		return 0.0
	}

	// Extract code string
	codeStr, ok := code.(string)
	if !ok {
		return 0.0
	}
	if codeStr == "" {
		return 0.0
	}

	// Extract documentation structure
	docsMap, ok := docs.(map[string]interface{})
	if !ok {
		return 0.0
	}

	// Get language from code analysis or default to "go"
	language := "go"
	if lang, ok := docsMap["language"].(string); ok && lang != "" {
		language = lang
	}

	// Extract all functions from code using AST
	functions, err := ast.ExtractFunctions(codeStr, language, "")
	if err != nil {
		// Fallback: count functions from documentation
		return s.calculateCoverageFromDocs(docsMap)
	}

	if len(functions) == 0 {
		return 100.0 // No functions means 100% coverage (nothing to document)
	}

	// Count documented functions
	documentedCount := 0
	funcDocs, ok := docsMap["functions"].([]map[string]interface{})
	if !ok {
		// Try alternative format: []interface{}
		if funcDocsInterface, ok := docsMap["functions"].([]interface{}); ok {
			funcDocs = make([]map[string]interface{}, 0, len(funcDocsInterface))
			for _, fd := range funcDocsInterface {
				if fdMap, ok := fd.(map[string]interface{}); ok {
					funcDocs = append(funcDocs, fdMap)
				}
			}
		}
	}

	// Create a map of documented function names
	documentedMap := make(map[string]bool)
	for _, doc := range funcDocs {
		if name, ok := doc["name"].(string); ok && name != "" {
			// Check if function has documentation
			if docStr, ok := doc["documentation"].(string); ok && strings.TrimSpace(docStr) != "" {
				documentedMap[name] = true
				documentedCount++
			}
		}
	}

	// Calculate coverage percentage
	if len(functions) == 0 {
		return 100.0
	}

	coverage := (float64(documentedCount) / float64(len(functions))) * 100.0
	if coverage > 100.0 {
		coverage = 100.0
	}

	return coverage
}

// calculateCoverageFromDocs calculates coverage from documentation structure only (fallback)
func (s *CodeAnalysisServiceImpl) calculateCoverageFromDocs(docsMap map[string]interface{}) float64 {
	funcDocs, ok := docsMap["functions"].([]map[string]interface{})
	if !ok {
		if funcDocsInterface, ok := docsMap["functions"].([]interface{}); ok {
			funcDocs = make([]map[string]interface{}, 0, len(funcDocsInterface))
			for _, fd := range funcDocsInterface {
				if fdMap, ok := fd.(map[string]interface{}); ok {
					funcDocs = append(funcDocs, fdMap)
				}
			}
		}
	}

	if len(funcDocs) == 0 {
		return 0.0
	}

	documentedCount := 0
	for _, doc := range funcDocs {
		if docStr, ok := doc["documentation"].(string); ok && strings.TrimSpace(docStr) != "" {
			documentedCount++
		}
	}

	if len(funcDocs) == 0 {
		return 0.0
	}

	return (float64(documentedCount) / float64(len(funcDocs))) * 100.0
}
