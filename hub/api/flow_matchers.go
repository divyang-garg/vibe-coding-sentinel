// Package main - Flow matching utilities
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package main

import (
	"strings"
)

// extractFeatureKeywords extracts meaningful keywords from a string
func extractFeatureKeywords(text string) []string {
	// Split by common separators and filter out common words
	words := strings.FieldsFunc(text, func(r rune) bool {
		return r == '_' || r == '-' || r == ' ' || r == '.' || r == '/'
	})

	keywords := []string{}
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true,
		"but": true, "in": true, "on": true, "at": true, "to": true,
		"for": true, "of": true, "with": true, "by": true, "from": true,
		"get": true, "set": true, "is": true, "are": true, "was": true,
		"component": true, "function": true, "handler": true,
	}

	for _, word := range words {
		wordLower := strings.ToLower(word)
		if len(wordLower) > 2 && !stopWords[wordLower] {
			keywords = append(keywords, wordLower)
		}
	}

	return keywords
}

// matchesComponentToEndpoint checks if a UI component matches an API endpoint
func matchesComponentToEndpoint(componentName string, endpointPath string) bool {
	// Simplified matching - would use semantic analysis in production
	componentLower := strings.ToLower(componentName)
	pathLower := strings.ToLower(endpointPath)

	// Extract keywords from component name
	keywords := extractFeatureKeywords(componentLower)

	for _, keyword := range keywords {
		if strings.Contains(pathLower, keyword) {
			return true
		}
	}

	return false
}

// matchesEndpointToFunction checks if an API endpoint handler matches a function name
func matchesEndpointToFunction(handler string, functionName string) bool {
	// Simplified matching
	handlerLower := strings.ToLower(handler)
	functionLower := strings.ToLower(functionName)

	return strings.Contains(functionLower, handlerLower) || strings.Contains(handlerLower, functionLower)
}

// matchesFunctionToTable checks if a function name matches a database table name
func matchesFunctionToTable(functionName string, tableName string) bool {
	// Simplified matching
	functionLower := strings.ToLower(functionName)
	tableLower := strings.ToLower(tableName)

	// Extract keywords from function name
	keywords := extractFeatureKeywords(functionLower)

	for _, keyword := range keywords {
		if strings.Contains(tableLower, keyword) {
			return true
		}
	}

	return false
}
