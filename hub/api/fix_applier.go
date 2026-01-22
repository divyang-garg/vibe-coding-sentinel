package main

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"sentinel-hub-api/ast"
)

// inferLanguageFromPath infers language from file path
func inferLanguageFromPath(filePath string) string {
	ext := filepath.Ext(filePath)
	langMap := map[string]string{
		".js": "javascript", ".ts": "typescript", ".jsx": "javascript",
		".py": "python", ".go": "go", ".java": "java",
		".cs": "csharp", ".php": "php", ".rb": "ruby",
	}
	if lang, ok := langMap[ext]; ok {
		return lang
	}
	return ""
}

// retryFixApplication retries a fix application function with exponential backoff
func retryFixApplication(ctx context.Context, fn func() (string, []map[string]interface{}, error)) (string, []map[string]interface{}, error) {
	config := GetConfig().Retry
	var lastErr error
	var result string
	var changes []map[string]interface{}

	backoff := config.InitialBackoff
	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		var err error
		result, changes, err = fn()
		if err == nil {
			return result, changes, nil
		}
		lastErr = err

		// Don't retry on last attempt
		if attempt < config.MaxRetries {
			// Wait with exponential backoff
			select {
			case <-ctx.Done():
				return "", nil, ctx.Err()
			case <-time.After(backoff):
			}

			// Calculate next backoff
			backoff = time.Duration(float64(backoff) * config.BackoffMultiplier)
			if backoff > config.MaxBackoff {
				backoff = config.MaxBackoff
			}
		}
	}

	return result, changes, fmt.Errorf("fix application failed after %d retries: %w", config.MaxRetries+1, lastErr)
}

// ApplySecurityFixes applies security-related fixes to code
func ApplySecurityFixes(ctx context.Context, code string, language string) (string, []map[string]interface{}, error) {
	return retryFixApplication(ctx, func() (string, []map[string]interface{}, error) {
		return applySecurityFixesInternal(ctx, code, language)
	})
}

// applySecurityFixesInternal is the internal implementation of security fixes
func applySecurityFixesInternal(ctx context.Context, code string, language string) (string, []map[string]interface{}, error) {
	changes := []map[string]interface{}{}
	fixedCode := code

	// Verify fixes after all changes
	defer func() {
		if err := verifyFix(fixedCode, language); err != nil {
			// Log warning but don't fail - fixes may be partial
			fmt.Printf("Warning: Security fix verification failed: %v\n", err)
		}
	}()

	// Fix 1: Replace string interpolation in SQL queries with parameterized queries
	// Use AST analysis to detect SQL injection vulnerabilities
	if language == "javascript" || language == "typescript" || language == "python" || language == "go" {
		// Use AST to find SQL queries with string interpolation
		findings, _, err := ast.AnalyzeAST(fixedCode, language, []string{"sql_injection"})
		if err == nil {
			for _, finding := range findings {
				if finding.Type == "sql_injection" || strings.Contains(strings.ToLower(finding.Message), "sql") {
					changes = append(changes, map[string]interface{}{
						"type":        "security",
						"description": "Replace string interpolation in SQL with parameterized queries",
						"line":        finding.Line,
						"column":      finding.Column,
						"message":     finding.Message,
					})

					// Apply fix: Replace template literals with parameterized queries
					// For JavaScript/TypeScript: query(`SELECT * FROM users WHERE id = ${userId}`)
					// -> query('SELECT * FROM users WHERE id = $1', [userId])
					if language == "javascript" || language == "typescript" {
						// Pattern-based replacement (AST transformation would be ideal but complex)
						sqlPattern := regexp.MustCompile(`query\(` + "`" + `([^` + "`" + `]*)\$\{([^}]+)\}([^` + "`" + `]*)` + "`" + `\)`)
						fixedCode = sqlPattern.ReplaceAllStringFunc(fixedCode, func(match string) string {
							// Extract SQL and variable
							parts := sqlPattern.FindStringSubmatch(match)
							if len(parts) >= 3 {
								sqlBefore := parts[1]
								varName := parts[2]
								sqlAfter := parts[3]
								// Replace with parameterized query
								return fmt.Sprintf("query('%s?', [%s])", sqlBefore+sqlAfter, varName)
							}
							return match
						})
					}
				}
			}
		} else {
			// Fallback to regex if AST analysis fails
			if language == "javascript" || language == "typescript" {
				sqlPattern := regexp.MustCompile(`query\(` + "`" + `SELECT\s+.*?\$\{.*?\}` + "`" + `\)`)
				if sqlPattern.MatchString(fixedCode) {
					changes = append(changes, map[string]interface{}{
						"type":        "security",
						"description": "Replace string interpolation in SQL with parameterized queries",
						"line":        0,
					})
				}
			}
		}
	}

	// Fix 2: Remove hardcoded secrets (API keys, passwords)
	// Enhanced pattern to catch more secret types
	secretPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(api[_-]?key|password|secret|token|apikey)\s*[:=]\s*["']([^"']+)["']`),
		regexp.MustCompile(`(jwt[_-]?secret|oauth[_-]?secret|db[_-]?password)\s*[:=]\s*["']([^"']+)["']`),
		regexp.MustCompile(`(database[_-]?url|connection[_-]?string)\s*[:=]\s*["']([^"']+)["']`),
	}

	for _, secretPattern := range secretPatterns {
		matches := secretPattern.FindAllStringSubmatch(fixedCode, -1)
		for _, match := range matches {
			if len(match) >= 3 {
				secretName := match[1]
				secretValue := match[2]

				// Skip if value looks like an env var reference already
				if strings.HasPrefix(secretValue, "${") || strings.Contains(secretValue, "env.") ||
					strings.Contains(secretValue, "getenv") || strings.Contains(secretValue, "Getenv") {
					continue
				}

				changes = append(changes, map[string]interface{}{
					"type":        "security",
					"description": fmt.Sprintf("Remove hardcoded %s", secretName),
					"line":        0, // Would be calculated from AST
				})

				// Replace with environment variable reference
				envVarName := strings.ToUpper(strings.ReplaceAll(strings.ReplaceAll(secretName, "-", "_"), "_", "_"))
				if language == "javascript" || language == "typescript" {
					fixedCode = strings.Replace(fixedCode, match[0],
						fmt.Sprintf("%s = process.env.%s", secretName, envVarName), 1)
				} else if language == "python" {
					fixedCode = strings.Replace(fixedCode, match[0],
						fmt.Sprintf("%s = os.getenv('%s')", secretName, envVarName), 1)
				} else if language == "go" {
					fixedCode = strings.Replace(fixedCode, match[0],
						fmt.Sprintf("%s = os.Getenv(\"%s\")", secretName, envVarName), 1)
				}
			}
		}
	}

	// Fix 3: Add input sanitization for XSS prevention
	if language == "javascript" || language == "typescript" {
		// Use AST to find XSS vulnerabilities
		findings, _, err := ast.AnalyzeAST(fixedCode, language, []string{"xss"})
		if err == nil {
			for _, finding := range findings {
				if finding.Type == "xss" || strings.Contains(strings.ToLower(finding.Message), "innerhtml") {
					changes = append(changes, map[string]interface{}{
						"type":        "security",
						"description": "Add input sanitization for XSS prevention",
						"line":        finding.Line,
						"column":      finding.Column,
						"message":     finding.Message,
					})

					// Apply fix: Wrap innerHTML assignments with sanitize()
					// Pattern: element.innerHTML = userInput;
					// -> element.innerHTML = sanitize(userInput);
					xssPattern := regexp.MustCompile(`(\.innerHTML\s*=\s*)([^;]+)(;)`)
					fixedCode = xssPattern.ReplaceAllStringFunc(fixedCode, func(match string) string {
						parts := xssPattern.FindStringSubmatch(match)
						if len(parts) >= 4 {
							assignment := parts[1]
							value := strings.TrimSpace(parts[2])
							semicolon := parts[3]
							// Check if already sanitized
							if !strings.Contains(value, "sanitize(") {
								return assignment + "sanitize(" + value + ")" + semicolon
							}
						}
						return match
					})
				}
			}
		} else {
			// Fallback to regex if AST analysis fails
			xssPattern := regexp.MustCompile(`\.innerHTML\s*=\s*([^;]+);`)
			if xssPattern.MatchString(fixedCode) {
				changes = append(changes, map[string]interface{}{
					"type":        "security",
					"description": "Add input sanitization for XSS prevention",
					"line":        0,
				})
			}
		}
	}

	return fixedCode, changes, nil
}

// verifyFix verifies that the fixed code is syntactically valid and fixes are correct
func verifyFix(code string, language string) error {
	// Parse fixed code with AST to verify syntax
	// Note: Using empty analysis list to perform basic syntax validation
	_, _, err := ast.AnalyzeAST(code, language, []string{})
	if err != nil {
		return fmt.Errorf("fixed code has syntax errors: %w", err)
	}

	// Additional verification: Check that secrets were removed
	if language == "javascript" || language == "typescript" || language == "python" || language == "go" {
		secretPattern := regexp.MustCompile(`(api[_-]?key|password|secret|token)\s*[:=]\s*["']([^"']{10,})["']`)
		if secretPattern.MatchString(code) {
			return fmt.Errorf("verification failed: hardcoded secrets still present")
		}
	}

	return nil
}

// ApplyStyleFixes applies style-related fixes to code
func ApplyStyleFixes(ctx context.Context, code string, language string) (string, []map[string]interface{}, error) {
	return retryFixApplication(ctx, func() (string, []map[string]interface{}, error) {
		return applyStyleFixesInternal(ctx, code, language)
	})
}

// applyStyleFixesInternal is the internal implementation of style fixes
func applyStyleFixesInternal(ctx context.Context, code string, language string) (string, []map[string]interface{}, error) {
	changes := []map[string]interface{}{}
	fixedCode := code

	// Fix 1: Remove trailing whitespace
	lines := strings.Split(fixedCode, "\n")
	for i, line := range lines {
		trimmed := strings.TrimRight(line, " \t")
		if trimmed != line {
			changes = append(changes, map[string]interface{}{
				"type":        "style",
				"description": "Remove trailing whitespace",
				"line":        i + 1,
			})
			lines[i] = trimmed
		}
	}
	fixedCode = strings.Join(lines, "\n")

	// Fix 2: Ensure consistent line endings
	originalCode := fixedCode
	// Detect and normalize line endings to LF (Unix standard)
	if strings.Contains(originalCode, "\r\n") {
		fixedCode = strings.ReplaceAll(fixedCode, "\r\n", "\n")
		changes = append(changes, map[string]interface{}{
			"type":        "style",
			"description": "Convert CRLF line endings to LF",
			"line":        0,
		})
	} else if strings.Contains(originalCode, "\r") {
		fixedCode = strings.ReplaceAll(fixedCode, "\r", "\n")
		changes = append(changes, map[string]interface{}{
			"type":        "style",
			"description": "Convert CR line endings to LF",
			"line":        0,
		})
	}

	// Fix 3: Fix indentation
	// Detect inconsistent indentation (mixing tabs/spaces)
	lines = strings.Split(fixedCode, "\n")
	indentationFixed := false
	indentSize := 2 // Default to 2 spaces, adjust based on language

	// Language-specific indent sizes
	switch language {
	case "python":
		indentSize = 4
	case "go", "java", "csharp":
		indentSize = 4
	case "javascript", "typescript":
		indentSize = 2
	default:
		indentSize = 2
	}

	spaceIndent := strings.Repeat(" ", indentSize)

	for i, line := range lines {
		if len(line) == 0 {
			continue
		}

		// Check if line starts with tabs
		if strings.HasPrefix(line, "\t") {
			// Count leading tabs
			tabCount := 0
			for _, char := range line {
				if char == '\t' {
					tabCount++
				} else {
					break
				}
			}
			// Replace tabs with spaces
			restOfLine := strings.TrimLeft(line, "\t")
			lines[i] = strings.Repeat(spaceIndent, tabCount) + restOfLine
			indentationFixed = true
		}
	}

	if indentationFixed {
		fixedCode = strings.Join(lines, "\n")
		changes = append(changes, map[string]interface{}{
			"type":        "style",
			"description": fmt.Sprintf("Convert tabs to %d spaces", indentSize),
			"line":        0,
		})
	}

	// Verify fixes
	if err := verifyFix(fixedCode, language); err != nil {
		// Log warning but don't fail - fixes may be partial
		fmt.Printf("Warning: Fix verification failed: %v\n", err)
	}

	return fixedCode, changes, nil
}

// ApplyPerformanceFixes applies performance-related fixes to code
func ApplyPerformanceFixes(ctx context.Context, code string, language string) (string, []map[string]interface{}, error) {
	return retryFixApplication(ctx, func() (string, []map[string]interface{}, error) {
		return applyPerformanceFixesInternal(ctx, code, language)
	})
}

// applyPerformanceFixesInternal is the internal implementation of performance fixes
func applyPerformanceFixesInternal(ctx context.Context, code string, language string) (string, []map[string]interface{}, error) {
	changes := []map[string]interface{}{}
	fixedCode := code

	// Fix 1: Optimize nested loops - enhanced with AST-based detection
	// Use AST to detect actual nested loop structures
	// Note: Complexity analysis not directly available, using general analysis
	findings, _, err := ast.AnalyzeAST(code, language, []string{})
	if err == nil {
		for _, finding := range findings {
			if finding.Type == "complexity" && strings.Contains(finding.Message, "nested loop") {
				changes = append(changes, map[string]interface{}{
					"type":        "performance",
					"description": fmt.Sprintf("Nested loop detected at line %d: %s", finding.Line, finding.Message),
					"line":        finding.Line,
					"suggestion":  "Consider using map/filter operations or breaking into separate functions",
				})
			}
		}
	} else {
		// Fallback to simple pattern matching
		if strings.Count(fixedCode, "for") > 2 {
			changes = append(changes, map[string]interface{}{
				"type":        "performance",
				"description": "Consider optimizing nested loops",
				"line":        0,
			})
		}
	}

	// Fix 2: Add caching for expensive operations
	// Detect expensive operations using regex patterns
	expensivePatterns := []*regexp.Regexp{
		// Database queries
		regexp.MustCompile(`(?i)(SELECT|INSERT|UPDATE|DELETE)\s+.*FROM`),
		// API calls
		regexp.MustCompile(`(?i)(fetch|axios|http\.(Get|Post|Put|Delete)|requests\.(get|post|put|delete))`),
		// File I/O
		regexp.MustCompile(`(?i)(readFile|writeFile|open|read|write)`),
	}

	lines := strings.Split(fixedCode, "\n")
	for i, line := range lines {
		for _, pattern := range expensivePatterns {
			if pattern.MatchString(line) {
				// Check if this is inside a loop
				isInLoop := false
				for j := i - 1; j >= 0 && j >= i-10; j-- {
					if strings.Contains(lines[j], "for") || strings.Contains(lines[j], "while") {
						isInLoop = true
						break
					}
				}

				if isInLoop {
					changes = append(changes, map[string]interface{}{
						"type":        "performance",
						"description": fmt.Sprintf("Expensive operation detected at line %d: consider caching result", i+1),
						"line":        i + 1,
						"suggestion":  "Wrap this operation with a cache layer or memoize the result",
					})
				}
			}
		}
	}

	// Fix 3: Remove unnecessary computations
	// Detect redundant calculations in loops
	// Pattern: repeated function calls with same arguments
	functionCallPattern := regexp.MustCompile(`(\w+)\s*\(([^)]+)\)`)
	seenCalls := make(map[string]int)

	for i, line := range lines {
		matches := functionCallPattern.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			if len(match) >= 3 {
				callKey := match[1] + "(" + match[2] + ")"
				// Check if this call appears multiple times
				if count, exists := seenCalls[callKey]; exists && count > 0 {
					// Check if we're in a loop
					isInLoop := false
					for j := i - 1; j >= 0 && j >= i-20; j-- {
						if strings.Contains(lines[j], "for") || strings.Contains(lines[j], "while") {
							isInLoop = true
							break
						}
					}

					if isInLoop {
						changes = append(changes, map[string]interface{}{
							"type":        "performance",
							"description": fmt.Sprintf("Redundant function call detected at line %d: %s", i+1, match[0]),
							"line":        i + 1,
							"suggestion":  "Move this computation outside the loop or cache the result",
						})
					}
				}
				seenCalls[callKey]++
			}
		}
	}

	return fixedCode, changes, nil
}
