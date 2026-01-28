// Package services - Task Verification Helper Functions
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package services

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// detectLanguageFromFileForVerifier detects language from file extension for task verifier
// Note: This is a task-verifier-specific version to avoid conflict with existing functions
func detectLanguageFromFileForVerifier(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".go":
		return "go"
	case ".js", ".jsx":
		return "javascript"
	case ".ts", ".tsx":
		return "typescript"
	case ".py":
		return "python"
	case ".java":
		return "java"
	default:
		return ""
	}
}

// extractFunctionCallsRegex is fallback regex-based implementation
func extractFunctionCallsRegex(codebasePath string, keywords []string, sourceFiles []string) ([]string, error) {
	var callSites []string

	// Build a map of keywords for faster lookup
	keywordMap := make(map[string]bool)
	for _, kw := range keywords {
		keywordMap[strings.ToLower(kw)] = true
	}

	// Patterns for function calls in different languages
	callPatterns := map[string]*regexp.Regexp{
		".go":   regexp.MustCompile(`\b([A-Za-z_][A-Za-z0-9_]*)\s*\(`),
		".js":   regexp.MustCompile(`\b([A-Za-z_][A-Za-z0-9_]*)\s*\(`),
		".ts":   regexp.MustCompile(`\b([A-Za-z_][A-Za-z0-9_]*)\s*\(`),
		".py":   regexp.MustCompile(`\b([A-Za-z_][A-Za-z0-9_]*)\s*\(`),
		".java": regexp.MustCompile(`\b([A-Za-z_][A-Za-z0-9_]*)\s*\(`),
	}

	// Process only files in sourceFiles list
	for _, file := range sourceFiles {
		fullPath := filepath.Join(codebasePath, file)

		// Skip test files
		if strings.Contains(fullPath, "_test.") || strings.Contains(fullPath, ".test.") {
			continue
		}

		ext := filepath.Ext(fullPath)
		pattern, ok := callPatterns[ext]
		if !ok {
			continue
		}

		content, err := os.ReadFile(fullPath)
		if err != nil {
			continue
		}

		contentStr := string(content)
		lines := strings.Split(contentStr, "\n")

		// Find function calls matching keywords
		for lineNum, line := range lines {
			matches := pattern.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				if len(match) > 1 {
					funcName := strings.ToLower(match[1])
					// Check if function name matches any keyword
					for keyword := range keywordMap {
						if strings.Contains(funcName, keyword) || strings.Contains(keyword, funcName) {
							callSite := fmt.Sprintf("%s:%d", file, lineNum+1)
							callSites = append(callSites, callSite)
							break
						}
					}
				}
			}
		}
	}

	return callSites, nil
}

// readFile is a helper function to read file content
// This allows for easier testing by mocking file operations
func readFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}
