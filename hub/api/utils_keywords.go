// Package main keyword extraction utilities
// Keyword extraction and text processing functions
// Complies with CODING_STANDARDS.md: Utilities max 250 lines

package main

import (
	"strings"
)

// extractKeywords extracts meaningful keywords from text
// Phase 14E: Shared function for task verification and dependency detection
func extractKeywords(text string) []string {
	// Simple keyword extraction - split by common separators
	words := strings.FieldsFunc(text, func(r rune) bool {
		return r == ' ' || r == '-' || r == '_' || r == '(' || r == ')' || r == '[' || r == ']'
	})

	keywords := []string{}
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true, "but": true,
		"in": true, "on": true, "at": true, "to": true, "for": true, "of": true,
		"with": true, "by": true, "is": true, "are": true, "was": true, "were": true,
		"be": true, "been": true, "have": true, "has": true, "had": true,
		"do": true, "does": true, "did": true, "will": true, "would": true,
		"should": true, "could": true, "may": true, "might": true,
	}

	for _, word := range words {
		word = strings.ToLower(strings.TrimSpace(word))
		if len(word) > 2 && !stopWords[word] {
			keywords = append(keywords, word)
		}
	}

	return keywords
}

// extractFeatureKeywords extracts keywords from feature name
func extractFeatureKeywordsFromUtils(featureName string) []string {
	var keywords []string
	words := []rune(featureName)
	var current []rune
	for _, r := range words {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			current = append(current, r)
		} else {
			if len(current) > 0 {
				keywords = append(keywords, string(current))
				current = nil
			}
		}
	}
	if len(current) > 0 {
		keywords = append(keywords, string(current))
	}
	return keywords
}
