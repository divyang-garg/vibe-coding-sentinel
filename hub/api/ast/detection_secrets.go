// Package ast provides secrets detection in code
// Complies with CODING_STANDARDS.md: Detection modules max 250 lines
package ast

import (
	"fmt"
	"regexp"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

// detectSecrets finds hardcoded secrets, API keys, and credentials
func detectSecrets(root *sitter.Node, code string, language string) []SecurityVulnerability {
	vulnerabilities := []SecurityVulnerability{}

	// Common secret patterns
	secretPatterns := []secretPattern{
		{
			name:     "api_key",
			pattern:  regexp.MustCompile(`(?i)(api[_-]?key|apikey)\s*[=:]\s*["']([A-Za-z0-9_\-]{20,})["']`),
			severity: "critical",
		},
		{
			name:     "aws_key",
			pattern:  regexp.MustCompile(`(?i)(aws[_-]?(access[_-]?key|secret[_-]?key))\s*[=:]\s*["']([A-Za-z0-9_\-+/=]{20,})["']`),
			severity: "critical",
		},
		{
			name:     "password",
			pattern:  regexp.MustCompile(`(?i)(password|passwd|pwd)\s*[=:]\s*["']([^"']{8,})["']`),
			severity: "critical",
		},
		{
			name:     "token",
			pattern:  regexp.MustCompile(`(?i)(token|bearer)\s*[=:]\s*["']([A-Za-z0-9_\-]{20,})["']`),
			severity: "high",
		},
		{
			name:     "private_key",
			pattern:  regexp.MustCompile(`(?i)(private[_-]?key|privkey)\s*[=:]\s*["']([A-Za-z0-9_\-+/=\s]{50,})["']`),
			severity: "critical",
		},
	}

	// Scan code for patterns
	lines := strings.Split(code, "\n")
	for lineNum, line := range lines {
		for _, pattern := range secretPatterns {
			matches := pattern.pattern.FindStringSubmatch(line)
			if len(matches) > 0 {
				// Found a potential secret
				col := strings.Index(line, matches[0]) + 1
				vuln := SecurityVulnerability{
					Type:        "hardcoded_secret",
					Severity:    pattern.severity,
					Line:        lineNum + 1,
					Column:      col,
					Message:     fmt.Sprintf("Potential hardcoded %s detected", pattern.name),
					Code:        strings.TrimSpace(line),
					Description: fmt.Sprintf("Hardcoded %s found in source code", pattern.name),
					Remediation: "Move secret to environment variable or secure secret management system",
					Confidence:  0.85,
				}
				vulnerabilities = append(vulnerabilities, vuln)
			}
		}
	}

	// Also check AST for variable assignments
	traverseAST(root, func(node *sitter.Node) bool {
		// Look for variable assignments with potential secrets
		if node.Type() == "assignment_expression" || node.Type() == "short_var_declaration" {
			codeSnippet := safeSlice(code, node.StartByte(), node.EndByte())
			if hasSecretInAssignment(codeSnippet) {
				line, col := getLineColumn(code, int(node.StartByte()))
				vuln := SecurityVulnerability{
					Type:        "hardcoded_secret",
					Severity:    "high",
					Line:        line,
					Column:      col,
					Message:     "Potential hardcoded secret in variable assignment",
					Code:        codeSnippet,
					Description: "Variable assignment contains potential secret value",
					Remediation: "Use environment variables or secure secret management",
					Confidence:  0.8,
				}
				vulnerabilities = append(vulnerabilities, vuln)
			}
		}
		return true
	})

	return vulnerabilities
}

// secretPattern represents a pattern for detecting secrets
type secretPattern struct {
	name     string
	pattern  *regexp.Regexp
	severity string
}

// hasSecretInAssignment checks if an assignment contains a potential secret
func hasSecretInAssignment(code string) bool {
	secretKeywords := []string{
		"password", "secret", "key", "token", "credential",
		"api_key", "apikey", "private_key", "access_key",
	}
	codeLower := strings.ToLower(code)
	for _, keyword := range secretKeywords {
		if strings.Contains(codeLower, keyword) {
			// Check if it's assigned to a string literal
			if strings.Contains(code, "=") && 
			   (strings.Contains(code, "\"") || strings.Contains(code, "'")) {
				// Check if value looks like a secret (long enough)
				if len(code) > 20 {
					return true
				}
			}
		}
	}
	return false
}
