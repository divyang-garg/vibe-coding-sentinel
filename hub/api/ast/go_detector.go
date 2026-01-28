// Package ast provides Go language-specific detection implementation
// Complies with CODING_STANDARDS.md: Language implementations max 400 lines
package ast

import (
	"fmt"

	sitter "github.com/smacker/go-tree-sitter"
)

// GoDetector implements LanguageDetector for Go
type GoDetector struct{}

// DetectSecurityMiddleware detects security middleware in Go code
func (d *GoDetector) DetectSecurityMiddleware(root *sitter.Node, code string) []ASTFinding {
	return detectSecurityMiddlewareGo(root, code)
}

// DetectUnused detects unused variables in Go code
func (d *GoDetector) DetectUnused(root *sitter.Node, code string) []ASTFinding {
	return detectUnusedVariablesGo(root, code)
}

// DetectDuplicates detects duplicate functions in Go code
func (d *GoDetector) DetectDuplicates(root *sitter.Node, code string) []ASTFinding {
	// Use the existing detectDuplicateFunctions with language parameter
	// Extract Go-specific logic from the switch statement
	findings := []ASTFinding{}
	functionMap := make(map[string][]*sitter.Node)

	TraverseAST(root, func(node *sitter.Node) bool {
		if node.Type() == "function_declaration" || node.Type() == "method_declaration" {
			var funcName string
			for i := 0; i < int(node.ChildCount()); i++ {
				child := node.Child(i)
				if child != nil {
					if child.Type() == "identifier" {
						funcName = safeSlice(code, child.StartByte(), child.EndByte())
						break
					} else if child.Type() == "field_identifier" {
						funcName = safeSlice(code, child.StartByte(), child.EndByte())
						break
					}
				}
			}
			if funcName != "" {
				functionMap[funcName] = append(functionMap[funcName], node)
			}
		}
		return true
	})

	// Find duplicates
	for funcName, nodes := range functionMap {
		if len(nodes) > 1 {
			for _, node := range nodes {
				findings = append(findings, ASTFinding{
					Type:       "duplicate_function",
					Severity:   "warning",
					Line:       int(node.StartPoint().Row) + 1,
					Column:     int(node.StartPoint().Column) + 1,
					Message:    fmt.Sprintf("Duplicate function '%s' found", funcName),
					Code:       safeSlice(code, node.StartByte(), node.EndByte()),
					Suggestion: fmt.Sprintf("Remove duplicate function '%s'", funcName),
					Confidence: 1.0,
				})
			}
		}
	}

	return findings
}

// DetectUnreachable detects unreachable code in Go
func (d *GoDetector) DetectUnreachable(root *sitter.Node, code string) []ASTFinding {
	return detectUnreachableCodeGo(root, code)
}

// DetectAsync detects missing await patterns (not applicable to Go)
func (d *GoDetector) DetectAsync(root *sitter.Node, code string) []ASTFinding {
	// Not applicable to Go
	return []ASTFinding{}
}

// DetectSQLInjection detects SQL injection vulnerabilities in Go code
func (d *GoDetector) DetectSQLInjection(root *sitter.Node, code string) []SecurityVulnerability {
	return detectSQLInjectionGo(root, code)
}

// DetectXSS detects XSS vulnerabilities in Go code
func (d *GoDetector) DetectXSS(root *sitter.Node, code string) []SecurityVulnerability {
	return detectXSSGo(root, code)
}

// DetectCommandInjection detects command injection vulnerabilities in Go code
func (d *GoDetector) DetectCommandInjection(root *sitter.Node, code string) []SecurityVulnerability {
	return detectCommandInjectionGo(root, code)
}

// DetectCrypto detects insecure crypto usage in Go code
func (d *GoDetector) DetectCrypto(root *sitter.Node, code string) []SecurityVulnerability {
	return detectInsecureCryptoGo(root, code)
}
