// Package ast provides duplicate detection functionality
// Complies with CODING_STANDARDS.md: Detection modules max 250 lines
package ast

import (
	sitter "github.com/smacker/go-tree-sitter"
)

// detectUnreachableCode finds unreachable code blocks.
// Uses registry when a detector is registered; otherwise falls back to switch.
func detectUnreachableCode(root *sitter.Node, code string, language string) []ASTFinding {
	if d := GetLanguageDetector(language); d != nil {
		return d.DetectUnreachable(root, code)
	}
	findings := []ASTFinding{}
	switch language {
	case "go":
		findings = detectUnreachableCodeGo(root, code)
	case "javascript", "typescript":
		findings = detectUnreachableCodeJS(root, code)
	case "python":
		findings = detectUnreachableCodePython(root, code)
	}
	return findings
}

// detectUnreachableCodeGo finds unreachable code in Go functions
func detectUnreachableCodeGo(root *sitter.Node, code string) []ASTFinding {
	findings := []ASTFinding{}

	// Simple implementation: look for statements after return statements in functions
	TraverseAST(root, func(node *sitter.Node) bool {
		if node.Type() == "function_declaration" || node.Type() == "method_declaration" {
			// Check function body for return statements followed by more statements
			for i := 0; i < int(node.ChildCount()); i++ {
				child := node.Child(i)
				if child != nil && child.Type() == "block" {
					// Check block statements
					foundReturn := false
					for j := 0; j < int(child.ChildCount()); j++ {
						stmt := child.Child(j)
						if stmt != nil {
							if stmt.Type() == "return_statement" {
								foundReturn = true
							} else if foundReturn && isStatementNode(stmt, "go") {
								// Found unreachable code
								startLine, startCol := getLineColumn(code, int(stmt.StartByte()))
								endLine, endCol := getLineColumn(code, int(stmt.EndByte()))

								findings = append(findings, ASTFinding{
									Type:       "unreachable_code",
									Severity:   "warning",
									Line:       startLine,
									Column:     startCol,
									EndLine:    endLine,
									EndColumn:  endCol,
									Message:    "Unreachable code detected after return statement",
									Code:       safeSlice(code, stmt.StartByte(), stmt.EndByte()),
									Suggestion: "Remove unreachable code or move it before the return statement",
								})
								break
							}
						}
					}
					break
				}
			}
		}
		return true
	})

	return findings
}
