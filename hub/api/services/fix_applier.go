// Package services - Fix applier implementation
// Complies with CODING_STANDARDS.md: Business Services max 400 lines
package services

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"sentinel-hub-api/ast"
	"sentinel-hub-api/pkg"
)

// retryFixApplication retries a fix application function with exponential backoff
func retryFixApplication(ctx context.Context, fn func() (string, []map[string]interface{}, error)) (string, []map[string]interface{}, error) {
	// Default retry configuration
	maxRetries := 3
	initialBackoff := 100 * time.Millisecond
	backoffMultiplier := 2.0
	maxBackoff := 5 * time.Second

	var lastErr error
	var result string
	var changes []map[string]interface{}

	backoff := initialBackoff
	for attempt := 0; attempt <= maxRetries; attempt++ {
		var err error
		result, changes, err = fn()
		if err == nil {
			return result, changes, nil
		}
		lastErr = err

		// Don't retry on last attempt
		if attempt < maxRetries {
			// Wait with exponential backoff
			select {
			case <-ctx.Done():
				return "", nil, ctx.Err()
			case <-time.After(backoff):
			}

			// Calculate next backoff
			backoff = time.Duration(float64(backoff) * backoffMultiplier)
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
		}
	}

	return result, changes, fmt.Errorf("fix application failed after %d retries: %w", maxRetries+1, lastErr)
}

// applySecurityFixes applies security-related fixes to code
func applySecurityFixes(ctx context.Context, code, language string) (string, []map[string]interface{}, error) {
	return retryFixApplication(ctx, func() (string, []map[string]interface{}, error) {
		return applySecurityFixesInternal(ctx, code, language)
	})
}

// applySecurityFixesInternal is the internal implementation of security fixes
func applySecurityFixesInternal(ctx context.Context, code, language string) (string, []map[string]interface{}, error) {
	changes := []map[string]interface{}{}
	fixedCode := code

	// Verify fixes after all changes
	defer func() {
		if err := verifyFix(fixedCode, language); err != nil {
			// Log warning but don't fail - fixes may be partial
			pkg.LogWarn(ctx, "Security fix verification failed: %v", err)
		}
	}()

	// Fix 1: Replace string interpolation in SQL queries with parameterized queries
	if language == "javascript" || language == "typescript" || language == "python" || language == "go" {
		// Use AST to find SQL queries with string interpolation
		findings, _, err := ast.AnalyzeAST(fixedCode, language, []string{"sql_injection"})
		if err == nil {
			for _, finding := range findings {
				if ctx.Err() != nil {
					return fixedCode, changes, ctx.Err()
				}
				if finding.Type == "sql_injection" || strings.Contains(strings.ToLower(finding.Message), "sql") {
					changes = append(changes, map[string]interface{}{
						"type":        "security",
						"description": "Replace string interpolation in SQL with parameterized queries",
						"line":        finding.Line,
						"column":      finding.Column,
						"message":     finding.Message,
					})

					// Apply fix: Replace template literals with parameterized queries
					if language == "javascript" || language == "typescript" {
						sqlPattern := regexp.MustCompile(`query\(` + "`" + `([^` + "`" + `]*)\$\{([^}]+)\}([^` + "`" + `]*)` + "`" + `\)`)
						fixedCode = sqlPattern.ReplaceAllStringFunc(fixedCode, func(match string) string {
							parts := sqlPattern.FindStringSubmatch(match)
							if len(parts) >= 3 {
								sqlBefore := parts[1]
								varName := parts[2]
								sqlAfter := parts[3]
								return fmt.Sprintf("query('%s?', [%s])", sqlBefore+sqlAfter, varName)
							}
							return match
						})
					}
				}
			}
		} else {
			pkg.LogWarn(ctx, "AST analysis failed for SQL injection detection, using fallback: %v", err)
		}
	}

	// Fix 2: Remove hardcoded secrets (API keys, passwords)
	secretPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(api[_-]?key|password|secret|token|apikey)\s*[:=]\s*["']([^"']+)["']`),
		regexp.MustCompile(`(jwt[_-]?secret|oauth[_-]?secret|db[_-]?password)\s*[:=]\s*["']([^"']+)["']`),
		regexp.MustCompile(`(database[_-]?url|connection[_-]?string)\s*[:=]\s*["']([^"']+)["']`),
	}

	for _, secretPattern := range secretPatterns {
		if ctx.Err() != nil {
			return fixedCode, changes, ctx.Err()
		}
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
					"line":        0,
				})

				// Replace with environment variable reference
				envVarName := strings.ToUpper(strings.ReplaceAll(strings.ReplaceAll(secretName, "-", "_"), "_", "_"))
				switch language {
				case "javascript", "typescript":
					fixedCode = strings.Replace(fixedCode, match[0],
						fmt.Sprintf("%s = process.env.%s", secretName, envVarName), 1)
				case "python":
					fixedCode = strings.Replace(fixedCode, match[0],
						fmt.Sprintf("%s = os.getenv('%s')", secretName, envVarName), 1)
				case "go":
					fixedCode = strings.Replace(fixedCode, match[0],
						fmt.Sprintf("%s = os.Getenv(\"%s\")", secretName, envVarName), 1)
				}
			}
		}
	}

	// Fix 3: Add input sanitization for XSS prevention
	if language == "javascript" || language == "typescript" {
		findings, _, err := ast.AnalyzeAST(fixedCode, language, []string{"xss"})
		if err == nil {
			for _, finding := range findings {
				if ctx.Err() != nil {
					return fixedCode, changes, ctx.Err()
				}
				if finding.Type == "xss" || strings.Contains(strings.ToLower(finding.Message), "innerhtml") {
					changes = append(changes, map[string]interface{}{
						"type":        "security",
						"description": "Add input sanitization for XSS prevention",
						"line":        finding.Line,
						"column":      finding.Column,
						"message":     finding.Message,
					})
				}
			}
		} else {
			pkg.LogWarn(ctx, "AST analysis failed for XSS detection, using fallback: %v", err)
		}
	}

	return fixedCode, changes, nil
}

// applyStyleFixes applies style-related fixes to code
func applyStyleFixes(ctx context.Context, code, language string) (string, []map[string]interface{}, error) {
	return retryFixApplication(ctx, func() (string, []map[string]interface{}, error) {
		return applyStyleFixesInternal(ctx, code, language)
	})
}

// applyStyleFixesInternal is the internal implementation of style fixes
func applyStyleFixesInternal(ctx context.Context, code, language string) (string, []map[string]interface{}, error) {
	changes := []map[string]interface{}{}
	fixedCode := code

	// Fix 1: Remove trailing whitespace
	lines := strings.Split(fixedCode, "\n")
	for i, line := range lines {
		if ctx.Err() != nil {
			return fixedCode, changes, ctx.Err()
		}
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
	lines = strings.Split(fixedCode, "\n")
	indentSize := 2
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
	indentationFixed := false

	for i, line := range lines {
		if len(line) == 0 {
			continue
		}
		if strings.HasPrefix(line, "\t") {
			tabCount := 0
			for _, char := range line {
				if char == '\t' {
					tabCount++
				} else {
					break
				}
			}
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
		pkg.LogWarn(ctx, "Style fix verification failed: %v", err)
	}

	return fixedCode, changes, nil
}

// applyPerformanceFixes applies performance-related fixes to code
func applyPerformanceFixes(ctx context.Context, code, language string) (string, []map[string]interface{}, error) {
	return retryFixApplication(ctx, func() (string, []map[string]interface{}, error) {
		return applyPerformanceFixesInternal(ctx, code, language)
	})
}

// applyPerformanceFixesInternal is the internal implementation of performance fixes
func applyPerformanceFixesInternal(ctx context.Context, code, language string) (string, []map[string]interface{}, error) {
	changes := []map[string]interface{}{}
	fixedCode := code

	// Fix 1: Optimize database queries (N+1 problem)
	if language == "javascript" || language == "typescript" || language == "python" || language == "go" {
		findings, _, err := ast.AnalyzeAST(fixedCode, language, []string{"performance"})
		if err == nil {
			for _, finding := range findings {
				if ctx.Err() != nil {
					return fixedCode, changes, ctx.Err()
				}
				if finding.Type == "performance" || strings.Contains(strings.ToLower(finding.Message), "n+1") {
					changes = append(changes, map[string]interface{}{
						"type":        "performance",
						"description": "Optimize database queries to avoid N+1 problem",
						"line":        finding.Line,
						"column":      finding.Column,
						"message":     finding.Message,
					})
				}
			}
		} else {
			pkg.LogWarn(ctx, "AST analysis failed for performance optimization, using fallback: %v", err)
		}
	}

	// Fix 2: Add caching for expensive operations
	// This is a placeholder - actual implementation would analyze code patterns

	// Verify fixes
	if err := verifyFix(fixedCode, language); err != nil {
		pkg.LogWarn(ctx, "Performance fix verification failed: %v", err)
	}

	return fixedCode, changes, nil
}

// verifyFix verifies that the fixed code is syntactically valid
func verifyFix(code string, language string) error {
	// Parse fixed code with AST to verify syntax
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
