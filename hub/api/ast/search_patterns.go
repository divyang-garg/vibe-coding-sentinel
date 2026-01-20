// Package ast provides language-specific search patterns
// Complies with CODING_STANDARDS.md: Utility modules max 250 lines
package ast

import (
	"fmt"
	"regexp"
	"strings"
)

// LanguagePatterns contains language-specific search patterns
type LanguagePatterns struct {
	FunctionCall string // Pattern to find function calls
	FunctionRef  string // Pattern to find function references
	Import       string // Pattern to find imports
	Export       string // Pattern to find exports
}

// GetPatterns returns language-specific search patterns
func GetPatterns(language string) LanguagePatterns {
	switch language {
	case "go":
		return LanguagePatterns{
			FunctionCall: `\b%s\s*\(`,
			FunctionRef:  `\b%s\b`,
			Import:       `import\s+(?:"|')(.*%s.*)(?:"|')`,
			Export:       `^[A-Z]`, // Exported identifiers start with uppercase
		}
	case "javascript", "typescript":
		return LanguagePatterns{
			FunctionCall: `\b%s\s*\(`,
			FunctionRef:  `\b%s\b`,
			Import:       `import\s+.*\b%s\b|from\s+['"]%s['"]|require\(['"]%s['"]\)`,
			Export:       `export\s+(?:function|const|let|var|class)\s+%s|export\s+{\s*%s`,
		}
	case "python":
		return LanguagePatterns{
			FunctionCall: `\b%s\s*\(`,
			FunctionRef:  `\b%s\b`,
			Import:       `import\s+%s|from\s+%s\s+import`,
			Export:       `^def\s+%s|^class\s+%s`, // Python doesn't have explicit exports
		}
	default:
		// Default patterns (work for most languages)
		return LanguagePatterns{
			FunctionCall: `\b%s\s*\(`,
			FunctionRef:  `\b%s\b`,
			Import:       `import\s+.*%s`,
			Export:       `%s`,
		}
	}
}

// BuildFunctionPattern builds a regex pattern to find function calls
func BuildFunctionPattern(funcName, language string) string {
	patterns := GetPatterns(language)
	// Escape special regex characters in function name
	escaped := escapeRegex(funcName)
	return fmt.Sprintf(patterns.FunctionCall, escaped)
}

// BuildReferencePattern builds a regex pattern to find function/variable references
func BuildReferencePattern(identifier, language string) string {
	patterns := GetPatterns(language)
	// Escape special regex characters in identifier
	escaped := escapeRegex(identifier)
	return fmt.Sprintf(patterns.FunctionRef, escaped)
}

// BuildImportPattern builds a regex pattern to find imports
func BuildImportPattern(moduleName, language string) string {
	patterns := GetPatterns(language)
	escaped := escapeRegex(moduleName)
	return fmt.Sprintf(patterns.Import, escaped, escaped, escaped)
}

// BuildExportPattern builds a regex pattern to find exports
func BuildExportPattern(identifier, language string) string {
	patterns := GetPatterns(language)
	escaped := escapeRegex(identifier)
	return fmt.Sprintf(patterns.Export, escaped, escaped)
}

// escapeRegex escapes special regex characters in a string
func escapeRegex(s string) string {
	specialChars := []string{`\`, `^`, `$`, `.`, `|`, `?`, `*`, `+`, `(`, `)`, `[`, `]`, `{`, `}`}
	escaped := s
	for _, char := range specialChars {
		escaped = strings.ReplaceAll(escaped, char, `\`+char)
	}
	return escaped
}

// ValidatePattern validates that a regex pattern is valid
func ValidatePattern(pattern string) bool {
	_, err := regexp.Compile(pattern)
	return err == nil
}

// IsValidIdentifier checks if a string is a valid identifier for the given language
func IsValidIdentifier(name, language string) bool {
	if name == "" {
		return false
	}

	switch language {
	case "go":
		// Go identifiers: letter or _ followed by letters, digits, or _
		matched, _ := regexp.MatchString(`^[a-zA-Z_][a-zA-Z0-9_]*$`, name)
		return matched
	case "javascript", "typescript":
		// JS/TS identifiers: similar to Go but can start with $ or _
		matched, _ := regexp.MatchString(`^[a-zA-Z_$][a-zA-Z0-9_$]*$`, name)
		return matched
	case "python":
		// Python identifiers: letter or _ followed by letters, digits, or _
		matched, _ := regexp.MatchString(`^[a-zA-Z_][a-zA-Z0-9_]*$`, name)
		return matched
	default:
		// Default: allow alphanumeric and underscore
		matched, _ := regexp.MatchString(`^[a-zA-Z_][a-zA-Z0-9_]*$`, name)
		return matched
	}
}

// ExtractLanguageFromPath attempts to determine language from file path
func ExtractLanguageFromPath(filePath string) string {
	lowerPath := strings.ToLower(filePath)
	if strings.HasSuffix(lowerPath, ".go") {
		return "go"
	}
	if strings.HasSuffix(lowerPath, ".js") || strings.HasSuffix(lowerPath, ".jsx") {
		return "javascript"
	}
	if strings.HasSuffix(lowerPath, ".ts") || strings.HasSuffix(lowerPath, ".tsx") {
		return "typescript"
	}
	if strings.HasSuffix(lowerPath, ".py") {
		return "python"
	}
	return "unknown"
}
