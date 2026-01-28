// Package ast provides JavaScript/TypeScript language-specific detection
// Complies with CODING_STANDARDS.md: Language implementations max 400 lines
package ast

import (
	"fmt"

	sitter "github.com/smacker/go-tree-sitter"
)

// JsDetector implements LanguageDetector for JavaScript and TypeScript
type JsDetector struct{}

// DetectSecurityMiddleware detects security middleware in JS/TS code
func (d *JsDetector) DetectSecurityMiddleware(root *sitter.Node, code string) []ASTFinding {
	return detectSecurityMiddlewareJS(root, code)
}

// DetectUnused detects unused variables in JS/TS code
func (d *JsDetector) DetectUnused(root *sitter.Node, code string) []ASTFinding {
	return detectUnusedVariablesJS(root, code)
}

// DetectDuplicates detects duplicate function definitions in JS/TS code
func (d *JsDetector) DetectDuplicates(root *sitter.Node, code string) []ASTFinding {
	findings := []ASTFinding{}
	functionMap := make(map[string][]*sitter.Node)

	TraverseAST(root, func(node *sitter.Node) bool {
		var funcName string
		if node.Type() == "function_declaration" || node.Type() == "function" {
			for i := 0; i < int(node.ChildCount()); i++ {
				child := node.Child(i)
				if child != nil && child.Type() == "identifier" {
					funcName = safeSlice(code, child.StartByte(), child.EndByte())
					break
				}
			}
		}
		if funcName != "" {
			functionMap[funcName] = append(functionMap[funcName], node)
		}
		return true
	})

	for funcName, nodes := range functionMap {
		if len(nodes) > 1 {
			for _, node := range nodes {
				startLine, startCol := getLineColumn(code, int(node.StartByte()))
				endLine, endCol := getLineColumn(code, int(node.EndByte()))
				findings = append(findings, ASTFinding{
					Type:       "duplicate_function",
					Severity:   "error",
					Line:       startLine,
					Column:     startCol,
					EndLine:    endLine,
					EndColumn:  endCol,
					Message:    fmt.Sprintf("Duplicate function definition: '%s' is defined %d times", funcName, len(nodes)),
					Code:       safeSlice(code, node.StartByte(), node.EndByte()),
					Suggestion: fmt.Sprintf("Remove duplicate definition of '%s' or rename one of them", funcName),
				})
			}
		}
	}
	return findings
}

// DetectUnreachable detects unreachable code in JS/TS
func (d *JsDetector) DetectUnreachable(root *sitter.Node, code string) []ASTFinding {
	return detectUnreachableCodeJS(root, code)
}

// DetectAsync detects missing await in JS/TS async functions
func (d *JsDetector) DetectAsync(root *sitter.Node, code string) []ASTFinding {
	return detectMissingAwaitJS(root, code)
}

// DetectSQLInjection detects SQL injection in JS/TS code
func (d *JsDetector) DetectSQLInjection(root *sitter.Node, code string) []SecurityVulnerability {
	return detectSQLInjectionJS(root, code)
}

// DetectXSS detects XSS vulnerabilities in JS/TS code
func (d *JsDetector) DetectXSS(root *sitter.Node, code string) []SecurityVulnerability {
	return detectXSSJS(root, code)
}

// DetectCommandInjection detects command injection in JS/TS code
func (d *JsDetector) DetectCommandInjection(root *sitter.Node, code string) []SecurityVulnerability {
	return detectCommandInjectionJS(root, code)
}

// DetectCrypto detects insecure crypto usage in JS/TS code
func (d *JsDetector) DetectCrypto(root *sitter.Node, code string) []SecurityVulnerability {
	return detectInsecureCryptoJS(root, code)
}
