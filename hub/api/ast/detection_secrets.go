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

	// Get language-specific patterns and remediation
	secretPatterns := getSecretPatterns(language)
	remediation := getLanguageSpecificRemediation(language)
	assignmentNodeTypes := getAssignmentNodeTypes(language)

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
					Message:     fmt.Sprintf("Potential hardcoded %s detected in %s code", pattern.name, language),
					Code:        strings.TrimSpace(line),
					Description: fmt.Sprintf("Hardcoded %s found in %s source code", pattern.name, language),
					Remediation: remediation,
					Confidence:  0.85,
				}
				vulnerabilities = append(vulnerabilities, vuln)
			}
		}
	}

	// Also check AST for variable assignments (language-specific node types)
	TraverseAST(root, func(node *sitter.Node) bool {
		// Look for variable assignments with potential secrets (language-specific)
		nodeType := node.Type()
		isAssignment := false
		for _, assignType := range assignmentNodeTypes {
			if nodeType == assignType {
				isAssignment = true
				break
			}
		}
		if isAssignment {
			codeSnippet := safeSlice(code, node.StartByte(), node.EndByte())
			if hasSecretInAssignment(codeSnippet) {
				line, col := getLineColumn(code, int(node.StartByte()))
				vuln := SecurityVulnerability{
					Type:        "hardcoded_secret",
					Severity:    "high",
					Line:        line,
					Column:      col,
					Message:     fmt.Sprintf("Potential hardcoded secret in %s variable assignment", language),
					Code:        codeSnippet,
					Description: fmt.Sprintf("Variable assignment in %s contains potential secret value", language),
					Remediation: remediation,
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

// getSecretPatterns returns language-specific secret detection patterns
func getSecretPatterns(language string) []secretPattern {
	// Common patterns that work across languages
	commonPatterns := []secretPattern{
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

	// Add language-specific patterns
	switch language {
	case "go":
		// Go-specific: check for const declarations with secrets
		commonPatterns = append(commonPatterns, secretPattern{
			name:     "go_const_secret",
			pattern:  regexp.MustCompile(`(?i)const\s+\w+\s*=\s*["']([A-Za-z0-9_\-]{20,})["']`),
			severity: "critical",
		})
	case "javascript", "typescript":
		// JS/TS-specific: check for process.env overrides
		commonPatterns = append(commonPatterns, secretPattern{
			name:     "env_override",
			pattern:  regexp.MustCompile(`(?i)process\.env\.\w+\s*=\s*["']([A-Za-z0-9_\-]{20,})["']`),
			severity: "critical",
		})
	case "python":
		// Python-specific: check for os.environ overrides
		commonPatterns = append(commonPatterns, secretPattern{
			name:     "os_environ_override",
			pattern:  regexp.MustCompile(`(?i)os\.environ\[["']\w+["']\]\s*=\s*["']([A-Za-z0-9_\-]{20,})["']`),
			severity: "critical",
		})
	}

	return commonPatterns
}

// getLanguageSpecificRemediation returns language-specific remediation suggestions
func getLanguageSpecificRemediation(language string) string {
	switch language {
	case "go":
		return "Use os.Getenv() or a secrets management library like HashiCorp Vault"
	case "javascript", "typescript":
		return "Use process.env with environment variables or a secrets management service like AWS Secrets Manager"
	case "python":
		return "Use os.getenv() or os.environ with environment variables, or a secrets management library"
	case "java":
		return "Use System.getenv() or a configuration management library like Spring Cloud Config"
	default:
		return "Move secret to environment variable or secure secret management system"
	}
}

// getAssignmentNodeTypes returns language-specific AST node types for variable assignments
func getAssignmentNodeTypes(language string) []string {
	switch language {
	case "go":
		return []string{"assignment_expression", "short_var_declaration", "var_declaration", "const_declaration"}
	case "javascript", "typescript":
		return []string{"assignment_expression", "variable_declarator", "lexical_declaration"}
	case "python":
		return []string{"assignment", "augmented_assignment"}
	case "java":
		return []string{"assignment_expression", "variable_declarator"}
	default:
		// Default to common assignment types
		return []string{"assignment_expression", "variable_declarator"}
	}
}
