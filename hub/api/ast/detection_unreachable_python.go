// Package ast provides Python unreachable code detection
// Complies with CODING_STANDARDS.md: Detection modules max 250 lines
package ast

import (
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

// detectUnreachableCodePython finds unreachable code in Python functions
func detectUnreachableCodePython(root *sitter.Node, code string) []ASTFinding {
	findings := []ASTFinding{}

	// Traverse AST to find functions
	traverseAST(root, func(node *sitter.Node) bool {
		var bodyNode *sitter.Node

		// Find function body
		if node.Type() == "function_definition" {
			// Find the function body (block or suite)
			for i := 0; i < int(node.ChildCount()); i++ {
				child := node.Child(i)
				if child != nil && (child.Type() == "block" || child.Type() == "suite") {
					bodyNode = child
					break
				}
			}

			if bodyNode != nil {
				// Check for unreachable code after return/raise/break/continue
				foundTerminator := false
				for i := 0; i < int(bodyNode.ChildCount()); i++ {
					stmt := bodyNode.Child(i)
					if stmt != nil {
						stmtType := stmt.Type()

						if foundTerminator && isStatementNode(stmt, "python") {
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
								Message:    "Unreachable code detected after return/raise/break/continue",
								Code:       safeSlice(code, stmt.StartByte(), stmt.EndByte()),
								Suggestion: "Remove unreachable code or move it before the terminating statement",
							})
							break
						}

						// Check if this is a terminating statement
						if stmtType == "return_statement" || stmtType == "raise_statement" ||
							stmtType == "break_statement" || stmtType == "continue_statement" {
							foundTerminator = true
						} else if stmtType == "expression_statement" {
							// Check for sys.exit() calls
							stmtCode := safeSlice(code, stmt.StartByte(), stmt.EndByte())
							if strings.Contains(stmtCode, "sys.exit") || strings.Contains(stmtCode, "exit()") {
								foundTerminator = true
							}
						}
					}
				}
			}
		}

		return true
	})

	return findings
}
