// Package ast provides SQL injection detection
// Complies with CODING_STANDARDS.md: Detection modules max 250 lines
package ast

import (
	"fmt"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

// detectSQLInjection finds SQL injection vulnerabilities
func detectSQLInjection(root *sitter.Node, code string, language string) []SecurityVulnerability {
	vulnerabilities := []SecurityVulnerability{}

	// Language-specific detection
	switch language {
	case "go":
		vulnerabilities = append(vulnerabilities, detectSQLInjectionGo(root, code)...)
	case "javascript", "typescript":
		vulnerabilities = append(vulnerabilities, detectSQLInjectionJS(root, code)...)
	case "python":
		vulnerabilities = append(vulnerabilities, detectSQLInjectionPython(root, code)...)
	}

	return vulnerabilities
}

// detectSQLInjectionGo detects SQL injection in Go code
func detectSQLInjectionGo(root *sitter.Node, code string) []SecurityVulnerability {
	vulnerabilities := []SecurityVulnerability{}

	// Enhanced pattern-based detection for multi-line queries
	// Look for SQL query construction patterns across multiple lines
	codeLower := strings.ToLower(code)

	// Check for SQL keywords
	hasSQLKeyword := strings.Contains(codeLower, "select") || strings.Contains(codeLower, "insert") ||
		strings.Contains(codeLower, "update") || strings.Contains(codeLower, "delete") ||
		strings.Contains(codeLower, "where") || strings.Contains(codeLower, "from")

	if hasSQLKeyword {
		// Look for query execution functions
		queryFuncs := []string{"db.Query", "db.QueryRow", "db.Exec", "Query", "QueryRow", "Exec", "QueryContext", "ExecContext"}
		for _, funcName := range queryFuncs {
			if strings.Contains(code, funcName) {
				// Skip if it's a parameterized query (has ? or $ placeholders)
				if strings.Contains(code, "?") || strings.Contains(code, "$1") || strings.Contains(code, "$2") {
					continue // Safe parameterized query
				}

				// Check for string concatenation patterns
				// Pattern 1: Direct concatenation in query string
				if strings.Contains(code, "+") && strings.Contains(code, "\"") {
					// Check if concatenation is near SQL keywords
					lines := strings.Split(code, "\n")
					for i, line := range lines {
						lineLower := strings.ToLower(line)
						// Skip lines with parameterized query patterns
						if strings.Contains(line, "?") || strings.Contains(line, "$") {
							continue
						}
						if (strings.Contains(lineLower, "select") || strings.Contains(lineLower, "insert") ||
							strings.Contains(lineLower, "update") || strings.Contains(lineLower, "delete") ||
							strings.Contains(lineLower, "where")) && strings.Contains(line, "+") {
							vuln := SecurityVulnerability{
								Type:        "sql_injection",
								Severity:    "critical",
								Line:        i + 1,
								Column:      1,
								Message:     "Potential SQL injection: string concatenation in SQL query",
								Code:        strings.TrimSpace(line),
								Description: "SQL query constructed using string concatenation with user input",
								Remediation: "Use parameterized queries (e.g., db.Query with ? placeholders)",
								Confidence:  0.9,
							}
							vulnerabilities = append(vulnerabilities, vuln)
						}
					}
				}

				// Pattern 2: Query variable construction
				if strings.Contains(code, "query :=") || strings.Contains(code, "query =") ||
					strings.Contains(code, "var query") {
					// Check if query is built with concatenation (but not parameterized)
					if strings.Contains(code, "+") && !strings.Contains(code, "?") {
						lines := strings.Split(code, "\n")
						for i, line := range lines {
							// Skip parameterized queries
							if strings.Contains(line, "?") || strings.Contains(line, "$") {
								continue
							}
							if (strings.Contains(strings.ToLower(line), "query") ||
								strings.Contains(strings.ToLower(line), "sql")) &&
								strings.Contains(line, "+") &&
								(strings.Contains(line, "\"") || strings.Contains(line, "`")) {
								vuln := SecurityVulnerability{
									Type:        "sql_injection",
									Severity:    "critical",
									Line:        i + 1,
									Column:      1,
									Message:     "Potential SQL injection: query variable constructed with string concatenation",
									Code:        strings.TrimSpace(line),
									Description: "SQL query variable constructed using string concatenation",
									Remediation: "Use parameterized queries (e.g., db.Query with ? placeholders)",
									Confidence:  0.85,
								}
								vulnerabilities = append(vulnerabilities, vuln)
							}
						}
					}
				}
			}
		}
	}

	// AST-based detection for more complex patterns
	TraverseAST(root, func(node *sitter.Node) bool {
		// Look for database query calls
		if node.Type() == "call_expression" {
			codeSnippet := safeSlice(code, node.StartByte(), node.EndByte())
			funcName := getFunctionName(node, code)

			// Skip if it's a parameterized query (has ? or $ placeholders)
			if strings.Contains(codeSnippet, "?") || strings.Contains(codeSnippet, "$1") ||
				strings.Contains(codeSnippet, "$2") || strings.Contains(codeSnippet, "$") {
				return true // Safe parameterized query
			}

			// Check if it's a SQL function
			if isSQLFunction(funcName) || isSQLFunction(codeSnippet) {
				// Check if arguments contain string concatenation or user input patterns
				if hasStringConcatenation(node, code) || hasUserInputInSQL(codeSnippet) {
					line, col := getLineColumn(code, int(node.StartByte()))
					vuln := SecurityVulnerability{
						Type:        "sql_injection",
						Severity:    "critical",
						Line:        line,
						Column:      col,
						Message:     fmt.Sprintf("Potential SQL injection in %s: string concatenation in SQL query", funcName),
						Code:        codeSnippet,
						Description: "SQL query constructed using string concatenation with user input",
						Remediation: "Use parameterized queries (e.g., db.Query with ? placeholders)",
						Confidence:  0.9,
					}
					vulnerabilities = append(vulnerabilities, vuln)
				}
			}
		}

		// Also check for variable assignments that build SQL queries
		if node.Type() == "short_var_declaration" || node.Type() == "assignment_statement" {
			codeSnippet := safeSlice(code, node.StartByte(), node.EndByte())
			// Skip if it's a parameterized query
			if strings.Contains(codeSnippet, "?") || strings.Contains(codeSnippet, "$") {
				return true
			}
			codeLower := strings.ToLower(codeSnippet)
			if (strings.Contains(codeLower, "query") || strings.Contains(codeLower, "sql")) &&
				strings.Contains(codeSnippet, "+") &&
				(strings.Contains(codeSnippet, "SELECT") || strings.Contains(codeSnippet, "INSERT") ||
					strings.Contains(codeSnippet, "UPDATE") || strings.Contains(codeSnippet, "DELETE")) {
				line, col := getLineColumn(code, int(node.StartByte()))
				vuln := SecurityVulnerability{
					Type:        "sql_injection",
					Severity:    "critical",
					Line:        line,
					Column:      col,
					Message:     "Potential SQL injection: SQL query variable constructed with string concatenation",
					Code:        codeSnippet,
					Description: "SQL query variable built using string concatenation",
					Remediation: "Use parameterized queries (e.g., db.Query with ? placeholders)",
					Confidence:  0.88,
				}
				vulnerabilities = append(vulnerabilities, vuln)
			}
		}

		return true
	})

	return vulnerabilities
}

// detectSQLInjectionJS detects SQL injection in JavaScript/TypeScript code
func detectSQLInjectionJS(root *sitter.Node, code string) []SecurityVulnerability {
	vulnerabilities := []SecurityVulnerability{}

	TraverseAST(root, func(node *sitter.Node) bool {
		// Look for database query calls
		if node.Type() == "call_expression" {
			funcName := getFunctionName(node, code)
			if isSQLFunction(funcName) {
				// Check for template literals or string concatenation
				if hasStringConcatenation(node, code) || hasTemplateLiteral(node, code) {
					line, col := getLineColumn(code, int(node.StartByte()))
					vuln := SecurityVulnerability{
						Type:        "sql_injection",
						Severity:    "critical",
						Line:        line,
						Column:      col,
						Message:     fmt.Sprintf("Potential SQL injection in %s: string interpolation in SQL query", funcName),
						Code:        safeSlice(code, node.StartByte(), node.EndByte()),
						Description: "SQL query constructed using string interpolation with user input",
						Remediation: "Use parameterized queries or query builders (e.g., Prisma, TypeORM)",
						Confidence:  0.9,
					}
					vulnerabilities = append(vulnerabilities, vuln)
				}
			}
		}
		return true
	})

	return vulnerabilities
}

// detectSQLInjectionPython detects SQL injection in Python code
func detectSQLInjectionPython(root *sitter.Node, code string) []SecurityVulnerability {
	vulnerabilities := []SecurityVulnerability{}

	TraverseAST(root, func(node *sitter.Node) bool {
		// Look for database query calls
		if node.Type() == "call_expression" {
			funcName := getFunctionName(node, code)
			if isSQLFunction(funcName) {
				// Check for f-strings or % formatting
				if hasStringFormatting(node, code) {
					line, col := getLineColumn(code, int(node.StartByte()))
					vuln := SecurityVulnerability{
						Type:        "sql_injection",
						Severity:    "critical",
						Line:        line,
						Column:      col,
						Message:     fmt.Sprintf("Potential SQL injection in %s: string formatting in SQL query", funcName),
						Code:        safeSlice(code, node.StartByte(), node.EndByte()),
						Description: "SQL query constructed using string formatting with user input",
						Remediation: "Use parameterized queries (e.g., cursor.execute with %s placeholders)",
						Confidence:  0.9,
					}
					vulnerabilities = append(vulnerabilities, vuln)
				}
			}
		}
		return true
	})

	return vulnerabilities
}

// Helper functions
func isSQLFunction(name string) bool {
	sqlFuncs := []string{
		"Query", "QueryRow", "Exec", "QueryContext", "ExecContext",
		"query", "execute", "exec", "queryOne",
		"db.query", "db.execute", "db.exec",
		"cursor.execute", "cursor.executemany",
	}
	nameLower := strings.ToLower(name)
	for _, funcName := range sqlFuncs {
		if strings.Contains(nameLower, strings.ToLower(funcName)) {
			return true
		}
	}
	return false
}

func getFunctionName(node *sitter.Node, code string) string {
	if node.ChildCount() == 0 {
		return ""
	}
	firstChild := node.Child(0)
	if firstChild == nil {
		return ""
	}
	return safeSlice(code, firstChild.StartByte(), firstChild.EndByte())
}

func hasStringConcatenation(node *sitter.Node, code string) bool {
	// Look for + operator between strings
	codeSnippet := safeSlice(code, node.StartByte(), node.EndByte())
	return strings.Contains(codeSnippet, "+") &&
		(strings.Contains(codeSnippet, "\"") || strings.Contains(codeSnippet, "'"))
}

func hasTemplateLiteral(node *sitter.Node, code string) bool {
	codeSnippet := safeSlice(code, node.StartByte(), node.EndByte())
	return strings.Contains(codeSnippet, "${") || strings.Contains(codeSnippet, "`")
}

func hasStringFormatting(node *sitter.Node, code string) bool {
	codeSnippet := safeSlice(code, node.StartByte(), node.EndByte())
	return strings.Contains(codeSnippet, "%") || strings.Contains(codeSnippet, "f\"") ||
		strings.Contains(codeSnippet, "f'")
}

func hasUserInputInSQL(code string) bool {
	// Look for common user input patterns in SQL context
	userInputPatterns := []string{
		" + ", "+ ", "+", "fmt.Sprintf", "fmt.Sprintf",
		"id", "userID", "userId", "user_id",
	}
	codeLower := strings.ToLower(code)
	for _, pattern := range userInputPatterns {
		if strings.Contains(codeLower, strings.ToLower(pattern)) {
			// Check if it's in a SQL query string
			if strings.Contains(code, "SELECT") || strings.Contains(code, "INSERT") ||
				strings.Contains(code, "UPDATE") || strings.Contains(code, "DELETE") {
				return true
			}
		}
	}
	return false
}
