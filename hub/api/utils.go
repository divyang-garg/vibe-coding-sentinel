// Phase 12: Utility Functions
// Common helper functions to reduce code duplication

package main

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
)

// marshalJSONB marshals a value to JSON string for JSONB storage
func marshalJSONB(v interface{}) (string, error) {
	if v == nil {
		return "null", nil
	}
	data, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return string(data), nil
}

// unmarshalJSONB unmarshals a JSON string from JSONB storage
func unmarshalJSONB(data string, v interface{}) error {
	if data == "" || data == "null" {
		return nil // Empty or null JSONB
	}
	if err := json.Unmarshal([]byte(data), v); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	return nil
}

// Phase 14E: Security Functions

// sanitizePath sanitizes a file path to prevent directory traversal attacks
func sanitizePath(p string) string {
	// Remove any ".." to prevent directory traversal
	return filepath.Clean(p)
}

// isValidPath validates that a path is safe to use
func isValidPath(p string) bool {
	// Check if path is absolute or relative and does not contain ".." after cleaning
	cleanPath := filepath.Clean(p)
	if strings.Contains(cleanPath, "..") {
		return false
	}

	// Prevent access to sensitive system directories
	sensitiveDirs := []string{
		"/etc", "/proc", "/sys", "/dev", "/boot", "/root", "/home",
		"C:\\Windows", "C:\\Program Files", "C:\\Users",
	}

	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return false
	}

	for _, sensitive := range sensitiveDirs {
		if strings.HasPrefix(absPath, sensitive) {
			return false
		}
	}

	return true
}

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
