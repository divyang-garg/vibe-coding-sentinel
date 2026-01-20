// Package ast provides XSS vulnerability detection
// Complies with CODING_STANDARDS.md: Detection modules max 250 lines
package ast

import (
	"fmt"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

// detectXSS finds XSS vulnerabilities
func detectXSS(root *sitter.Node, code string, language string) []SecurityVulnerability {
	vulnerabilities := []SecurityVulnerability{}

	// Language-specific detection
	switch language {
	case "go":
		vulnerabilities = append(vulnerabilities, detectXSSGo(root, code)...)
	case "javascript", "typescript":
		vulnerabilities = append(vulnerabilities, detectXSSJS(root, code)...)
	case "python":
		vulnerabilities = append(vulnerabilities, detectXSSPython(root, code)...)
	}

	return vulnerabilities
}

// detectXSSGo detects XSS in Go code
func detectXSSGo(root *sitter.Node, code string) []SecurityVulnerability {
	vulnerabilities := []SecurityVulnerability{}

	traverseAST(root, func(node *sitter.Node) bool {
		// Look for HTML template rendering
		if node.Type() == "call_expression" {
			funcName := getFunctionName(node, code)
			if isHTMLRenderFunction(funcName) {
				// Check if user input is used without escaping
				if hasUnescapedUserInput(node, code) {
					line, col := getLineColumn(code, int(node.StartByte()))
					vuln := SecurityVulnerability{
						Type:        "xss",
						Severity:    "high",
						Line:        line,
						Column:      col,
						Message:     fmt.Sprintf("Potential XSS in %s: unescaped user input in HTML output", funcName),
						Code:        safeSlice(code, node.StartByte(), node.EndByte()),
						Description: "User input rendered in HTML without proper escaping",
						Remediation: "Use html/template package with automatic escaping or html.EscapeString()",
						Confidence:  0.85,
					}
					vulnerabilities = append(vulnerabilities, vuln)
				}
			}
		}
		return true
	})

	return vulnerabilities
}

// detectXSSJS detects XSS in JavaScript/TypeScript code
func detectXSSJS(root *sitter.Node, code string) []SecurityVulnerability {
	vulnerabilities := []SecurityVulnerability{}

		// Pattern-based detection for simple cases
	lines := strings.Split(code, "\n")
	for lineNum, lineText := range lines {
		lineLower := strings.ToLower(lineText)
		if (strings.Contains(lineLower, "innerhtml") || strings.Contains(lineLower, "outerhtml") || 
		    strings.Contains(lineLower, "document.write")) &&
		   (strings.Contains(lineText, "=") || strings.Contains(lineText, "(")) {
			// Check if it's an assignment or function call with potential user input
			if strings.Contains(lineText, ".") {
				_, col := getLineColumn(code, strings.Index(code, lineText))
				vuln := SecurityVulnerability{
					Type:        "xss",
					Severity:    "high",
					Line:        lineNum + 1,
					Column:      col,
					Message:     "Potential XSS: user input assigned to innerHTML/outerHTML or used in document.write",
					Code:        strings.TrimSpace(lineText),
					Description: "User input directly inserted into DOM without sanitization",
					Remediation: "Use textContent instead of innerHTML, or sanitize with DOMPurify",
					Confidence:  0.85,
				}
				vulnerabilities = append(vulnerabilities, vuln)
			}
		}
	}

	// AST-based detection
	traverseAST(root, func(node *sitter.Node) bool {
		// Look for innerHTML, outerHTML, document.write
		if node.Type() == "assignment_expression" || node.Type() == "call_expression" {
			codeSnippet := safeSlice(code, node.StartByte(), node.EndByte())
			if isDangerousDOMOperation(codeSnippet) {
				// Check if user input is used
				if hasUserInput(node, code) {
					line, col := getLineColumn(code, int(node.StartByte()))
					vuln := SecurityVulnerability{
						Type:        "xss",
						Severity:    "high",
						Line:        line,
						Column:      col,
						Message:     "Potential XSS: user input assigned to innerHTML/outerHTML or used in document.write",
						Code:        codeSnippet,
						Description: "User input directly inserted into DOM without sanitization",
						Remediation: "Use textContent instead of innerHTML, or sanitize with DOMPurify",
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

// detectXSSPython detects XSS in Python code
func detectXSSPython(root *sitter.Node, code string) []SecurityVulnerability {
	vulnerabilities := []SecurityVulnerability{}

	traverseAST(root, func(node *sitter.Node) bool {
		// Look for template rendering (Flask, Django)
		if node.Type() == "call_expression" {
			funcName := getFunctionName(node, code)
			if isHTMLRenderFunction(funcName) {
				// Check if user input is used without escaping
				if hasUnescapedUserInput(node, code) {
					line, col := getLineColumn(code, int(node.StartByte()))
					vuln := SecurityVulnerability{
						Type:        "xss",
						Severity:    "high",
						Line:        line,
						Column:      col,
						Message:     fmt.Sprintf("Potential XSS in %s: unescaped user input in template", funcName),
						Code:        safeSlice(code, node.StartByte(), node.EndByte()),
						Description: "User input rendered in template without auto-escaping",
						Remediation: "Use template auto-escaping or mark_safe() only for trusted content",
						Confidence:  0.85,
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
func isHTMLRenderFunction(name string) bool {
	htmlFuncs := []string{
		"Execute", "ExecuteTemplate", "Render", "render",
		"render_template", "render_to_string", "Template",
	}
	nameLower := strings.ToLower(name)
	for _, funcName := range htmlFuncs {
		if strings.Contains(nameLower, strings.ToLower(funcName)) {
			return true
		}
	}
	return false
}

func isDangerousDOMOperation(code string) bool {
	dangerousOps := []string{
		"innerHTML", "outerHTML", "document.write", "document.writeln",
		"eval(", "Function(", "setTimeout(", "setInterval(",
	}
	codeLower := strings.ToLower(code)
	for _, op := range dangerousOps {
		if strings.Contains(codeLower, strings.ToLower(op)) {
			return true
		}
	}
	return false
}

func hasUnescapedUserInput(node *sitter.Node, code string) bool {
	// Simplified check - look for common user input patterns
	codeSnippet := safeSlice(code, node.StartByte(), node.EndByte())
	userInputPatterns := []string{
		"req.", "request.", "c.", "ctx.", "params.", "query.",
		"form.", "body.", "input.", "user.", "data.",
	}
	codeLower := strings.ToLower(codeSnippet)
	for _, pattern := range userInputPatterns {
		if strings.Contains(codeLower, pattern) {
			// Check if it's escaped
			if !strings.Contains(codeLower, "escape") && 
			   !strings.Contains(codeLower, "sanitize") &&
			   !strings.Contains(codeLower, "html/template") {
				return true
			}
		}
	}
	return false
}

func hasUserInput(node *sitter.Node, code string) bool {
	codeSnippet := safeSlice(code, node.StartByte(), node.EndByte())
	userInputPatterns := []string{
		"req.", "request.", "params.", "query.", "body.",
		"form.", "input.", "user.", "data.", "value.",
	}
	codeLower := strings.ToLower(codeSnippet)
	for _, pattern := range userInputPatterns {
		if strings.Contains(codeLower, pattern) {
			return true
		}
	}
	return false
}
