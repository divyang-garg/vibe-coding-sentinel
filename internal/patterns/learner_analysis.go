// Pattern Learning - Analysis Functions
// Codebase analysis and pattern detection functions
// Complies with CODING_STANDARDS.md: Utilities max 250 lines

package patterns

import (
	"os"
	"path/filepath"
	"strings"
)

// analyzeCodebase walks through the codebase and analyzes patterns
func analyzeCodebase(codebasePath string, patterns *PatternData) error {
	return filepath.Walk(codebasePath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		// Skip common directories
		if shouldSkipPath(path) {
			return nil
		}

		ext := filepath.Ext(path)
		if ext != "" {
			patterns.FileExtensions[ext]++
		}

		// Read file content
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		contentStr := string(content)

		// Detect languages and frameworks
		detectLanguageAndFramework(path, contentStr, patterns)

		// Analyze naming patterns
		analyzeNamingPatterns(path, contentStr, patterns)

		// Analyze import patterns (for supported languages)
		analyzeImportPatterns(path, contentStr, patterns)

		// Analyze code style
		analyzeCodeStyle(path, contentStr, patterns)

		return nil
	})
}

// shouldSkipPath determines if a path should be skipped
func shouldSkipPath(path string) bool {
	skipDirs := []string{
		"/node_modules/", "/.git/", "/build/", "/dist/",
		"/__pycache__/", "/vendor/", "/.next/", "/target/",
		"/bin/", "/obj/", "/.vscode/", "/.idea/",
	}

	pathLower := strings.ToLower(path)
	for _, skipDir := range skipDirs {
		if strings.Contains(pathLower, skipDir) {
			return true
		}
	}
	return false
}

// detectLanguageAndFramework detects programming language and framework
func detectLanguageAndFramework(path, content string, patterns *PatternData) {
	ext := filepath.Ext(path)
	filename := strings.ToLower(filepath.Base(path))

	switch ext {
	case ".js", ".jsx", ".ts", ".tsx":
		patterns.Languages["JavaScript/TypeScript"]++
		if strings.Contains(content, "react") {
			patterns.Frameworks["React"]++
		}
		if strings.Contains(content, "vue") {
			patterns.Frameworks["Vue.js"]++
		}
		if strings.Contains(content, "angular") {
			patterns.Frameworks["Angular"]++
		}
	case ".py":
		patterns.Languages["Python"]++
		if strings.Contains(content, "django") {
			patterns.Frameworks["Django"]++
		}
		if strings.Contains(content, "flask") {
			patterns.Frameworks["Flask"]++
		}
	case ".go":
		patterns.Languages["Go"]++
	case ".java":
		patterns.Languages["Java"]++
		if strings.Contains(content, "spring") {
			patterns.Frameworks["Spring"]++
		}
	case ".cs":
		patterns.Languages["C#"]++
		if strings.Contains(content, "aspnet") {
			patterns.Frameworks["ASP.NET"]++
		}
	case ".rb":
		patterns.Languages["Ruby"]++
		if strings.Contains(content, "rails") {
			patterns.Frameworks["Ruby on Rails"]++
		}
	}

	// Detect project type from config files
	if filename == "package.json" {
		patterns.Frameworks["Node.js"]++
	} else if filename == "go.mod" {
		patterns.Frameworks["Go Modules"]++
	} else if filename == "requirements.txt" || filename == "pyproject.toml" {
		patterns.Frameworks["Python"]++
	}
}

// analyzeNamingPatterns detects naming conventions
func analyzeNamingPatterns(path, content string, patterns *PatternData) {
	// Simple heuristics for naming patterns
	if strings.Contains(content, "camelCase") || containsCamelCase(content) {
		patterns.NamingPatterns["camelCase"]++
	}
	if strings.Contains(content, "PascalCase") || containsPascalCase(content) {
		patterns.NamingPatterns["PascalCase"]++
	}
	if strings.Contains(content, "snake_case") || containsSnakeCase(content) {
		patterns.NamingPatterns["snake_case"]++
	}
}

// containsCamelCase checks if content contains camelCase identifiers
func containsCamelCase(content string) bool {
	// Simple check for lowercase start with uppercase in middle
	for i := 1; i < len(content)-1; i++ {
		if content[i] >= 'A' && content[i] <= 'Z' &&
			content[i-1] >= 'a' && content[i-1] <= 'z' {
			return true
		}
	}
	return false
}

// containsPascalCase checks if content contains PascalCase identifiers
func containsPascalCase(content string) bool {
	// Look for patterns like "class ClassName" or "type TypeName"
	return strings.Contains(content, "class ") || strings.Contains(content, "type ")
}

// containsSnakeCase checks if content contains snake_case identifiers
func containsSnakeCase(content string) bool {
	return strings.Contains(content, "_")
}

// findPrimaryLanguage finds the most used language
func findPrimaryLanguage(patterns *PatternData) string {
	var primaryLang string
	var maxCount int

	// Prioritize TypeScript/JavaScript
	if count := patterns.Languages["JavaScript/TypeScript"]; count > 0 {
		if patterns.FileExtensions[".ts"] > patterns.FileExtensions[".js"] {
			primaryLang = "TypeScript"
		} else {
			primaryLang = "JavaScript"
		}
		maxCount = count
	}

	// Find other primary languages
	for lang, count := range patterns.Languages {
		if count > maxCount {
			maxCount = count
			primaryLang = lang
		}
	}

	return primaryLang
}

// contains checks if a string slice contains a value
func contains(slice []string, value string) bool {
	for _, s := range slice {
		if s == value {
			return true
		}
	}
	return false
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
