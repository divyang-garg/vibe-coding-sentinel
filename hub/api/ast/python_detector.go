// Package ast provides Python language-specific detection
// Complies with CODING_STANDARDS.md: Language implementations max 400 lines
package ast

import (
	"fmt"

	sitter "github.com/smacker/go-tree-sitter"
)

// PythonDetector implements LanguageDetector for Python
type PythonDetector struct{}

// DetectSecurityMiddleware detects security middleware in Python code
func (d *PythonDetector) DetectSecurityMiddleware(root *sitter.Node, code string) []ASTFinding {
	return detectSecurityMiddlewarePython(root, code)
}

// DetectUnused detects unused variables in Python code
func (d *PythonDetector) DetectUnused(root *sitter.Node, code string) []ASTFinding {
	return detectUnusedVariablesPython(root, code)
}

// DetectDuplicates detects duplicate function definitions in Python code
func (d *PythonDetector) DetectDuplicates(root *sitter.Node, code string) []ASTFinding {
	findings := []ASTFinding{}
	functionMap := make(map[string][]*sitter.Node)

	TraverseAST(root, func(node *sitter.Node) bool {
		if node.Type() == "function_definition" {
			var funcName string
			for i := 0; i < int(node.ChildCount()); i++ {
				child := node.Child(i)
				if child != nil && child.Type() == "identifier" {
					funcName = safeSlice(code, child.StartByte(), child.EndByte())
					break
				}
			}
			if funcName != "" {
				functionMap[funcName] = append(functionMap[funcName], node)
			}
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

// DetectUnreachable detects unreachable code in Python
func (d *PythonDetector) DetectUnreachable(root *sitter.Node, code string) []ASTFinding {
	return detectUnreachableCodePython(root, code)
}

// DetectAsync detects missing await (not applicable to Python in same way as JS)
func (d *PythonDetector) DetectAsync(root *sitter.Node, code string) []ASTFinding {
	return []ASTFinding{}
}

// DetectSQLInjection detects SQL injection in Python code
func (d *PythonDetector) DetectSQLInjection(root *sitter.Node, code string) []SecurityVulnerability {
	return detectSQLInjectionPython(root, code)
}

// DetectXSS detects XSS vulnerabilities in Python code
func (d *PythonDetector) DetectXSS(root *sitter.Node, code string) []SecurityVulnerability {
	return detectXSSPython(root, code)
}

// DetectCommandInjection detects command injection in Python code
func (d *PythonDetector) DetectCommandInjection(root *sitter.Node, code string) []SecurityVulnerability {
	return detectCommandInjectionPython(root, code)
}

// DetectCrypto detects insecure crypto usage in Python code
func (d *PythonDetector) DetectCrypto(root *sitter.Node, code string) []SecurityVulnerability {
	return detectInsecureCryptoPython(root, code)
}
