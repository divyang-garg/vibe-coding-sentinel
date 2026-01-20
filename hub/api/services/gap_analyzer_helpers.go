// Gap Analysis Engine - Helper Functions
// Utility functions for gap analysis operations
// Complies with CODING_STANDARDS.md: Business Services max 400 lines

package services

import (
	"path/filepath"
	"strings"
)

func isCodeFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".go" || ext == ".js" || ext == ".ts" || ext == ".py" || ext == ".java"
}

func containsBusinessKeywords(line string) bool {
	keywords := []string{"order", "payment", "user", "account", "transaction", "rule", "validate", "check"}
	lineLower := strings.ToLower(line)
	for _, keyword := range keywords {
		if strings.Contains(lineLower, keyword) {
			return true
		}
	}
	return false
}

func extractFunctionNameGap(line string) string {
	// Simple extraction - full implementation would use AST
	parts := strings.Fields(line)
	for i, part := range parts {
		if part == "func" && i+1 < len(parts) {
			funcName := strings.TrimSpace(parts[i+1])
			// Remove parameters
			if idx := strings.Index(funcName, "("); idx > 0 {
				return funcName[:idx]
			}
			return funcName
		}
	}
	return ""
}

func extractKeyword(line string) string {
	keywords := []string{"order", "payment", "user", "account", "transaction"}
	lineLower := strings.ToLower(line)
	for _, keyword := range keywords {
		if strings.Contains(lineLower, keyword) {
			return keyword
		}
	}
	return ""
}
