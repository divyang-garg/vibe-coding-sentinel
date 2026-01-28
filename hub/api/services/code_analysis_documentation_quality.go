// Package services provides documentation quality assessment functions
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package services

import (
	"regexp"
	"strings"
)

// assessDocumentationQuality assesses the quality of documentation
// Returns quality score (0-100) based on completeness, clarity, and examples
func (s *CodeAnalysisServiceImpl) assessDocumentationQuality(docs interface{}) float64 {
	if docs == nil {
		return 0.0
	}

	docsMap, ok := docs.(map[string]interface{})
	if !ok {
		return 0.0
	}

	// Get functions from documentation
	funcDocs, ok := docsMap["functions"].([]map[string]interface{})
	if !ok {
		// Try alternative format
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

	// Calculate quality score for each function and average
	totalScore := 0.0
	scoredCount := 0

	for _, doc := range funcDocs {
		score := s.scoreFunctionDocumentation(doc)
		if score > 0 {
			totalScore += score
			scoredCount++
		}
	}

	if scoredCount == 0 {
		return 0.0
	}

	return totalScore / float64(scoredCount)
}

// scoreFunctionDocumentation scores a single function's documentation quality
// Returns score from 0-100
func (s *CodeAnalysisServiceImpl) scoreFunctionDocumentation(doc map[string]interface{}) float64 {
	score := 0.0

	// Check for documentation string
	docStr, hasDoc := doc["documentation"].(string)
	if !hasDoc || strings.TrimSpace(docStr) == "" {
		return 0.0 // No documentation = 0 score
	}

	docStr = strings.TrimSpace(docStr)

	// Base score: Has documentation (50 points)
	score += 50.0

	// Completeness checks (+30 points)
	// Check for parameter documentation
	if params, ok := doc["parameters"].([]string); ok && len(params) > 0 {
		// Check if parameters are mentioned in documentation
		paramMentioned := false
		for _, param := range params {
			if strings.Contains(strings.ToLower(docStr), strings.ToLower(param)) {
				paramMentioned = true
				break
			}
		}
		if paramMentioned {
			score += 15.0
		}
	}

	// Check for return type documentation
	if returnType, ok := doc["returnType"].(string); ok && returnType != "" {
		if strings.Contains(strings.ToLower(docStr), "return") ||
			strings.Contains(strings.ToLower(docStr), strings.ToLower(returnType)) {
			score += 15.0
		}
	}

	// Clarity checks (+15 points)
	// Check documentation length (not too short, not too long)
	docLength := len(docStr)
	if docLength >= 20 && docLength <= 500 {
		score += 10.0
	} else if docLength > 500 {
		score += 5.0 // Too long, partial credit
	}

	// Check for proper formatting (newlines, structure)
	if strings.Contains(docStr, "\n") || strings.Contains(docStr, "  ") {
		score += 5.0
	}

	// Examples check (+5 points)
	if strings.Contains(strings.ToLower(docStr), "example") ||
		strings.Contains(strings.ToLower(docStr), "usage") ||
		strings.Contains(docStr, "```") ||
		strings.Contains(docStr, "Example:") {
		score += 5.0
	}

	// Language-specific quality checks
	language := "go"
	if lang, ok := doc["language"].(string); ok {
		language = lang
	}

	score += s.scoreLanguageSpecificQuality(docStr, language)

	// Ensure score is within bounds
	if score > 100.0 {
		score = 100.0
	}

	return score
}

// scoreLanguageSpecificQuality scores documentation based on language-specific standards
func (s *CodeAnalysisServiceImpl) scoreLanguageSpecificQuality(docStr, language string) float64 {
	score := 0.0

	switch language {
	case "go":
		// Go: Check for godoc format
		// Godoc typically starts with function name or description
		if len(docStr) > 0 && (strings.HasPrefix(strings.TrimSpace(docStr), strings.ToUpper(string(docStr[0]))) ||
			strings.Contains(docStr, "//")) {
			score += 5.0
		}
		// Check for parameter documentation format: "param: description"
		if matched, _ := regexp.MatchString(`\w+:\s+\w+`, docStr); matched {
			score += 5.0
		}

	case "javascript", "typescript":
		// JavaScript/TypeScript: Check for JSDoc format
		if strings.Contains(docStr, "/**") || strings.Contains(docStr, "@param") ||
			strings.Contains(docStr, "@returns") || strings.Contains(docStr, "@return") {
			score += 10.0
		}

	case "python":
		// Python: Check for docstring format
		if strings.Contains(docStr, "\"\"\"") || strings.Contains(docStr, "'''") {
			score += 5.0
		}
		// Check for Google/NumPy style docstrings
		if strings.Contains(docStr, "Args:") || strings.Contains(docStr, "Returns:") ||
			strings.Contains(docStr, "Parameters") {
			score += 5.0
		}
	}

	return score
}
