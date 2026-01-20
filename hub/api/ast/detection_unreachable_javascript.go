package ast

import (
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

func detectUnreachableCodeJS(root *sitter.Node, code string) []ASTFinding {
	findings := []ASTFinding{}

	traverseAST(root, func(node *sitter.Node) bool {
		var bodyNode *sitter.Node

		if node.Type() == "function_declaration" || node.Type() == "function" || node.Type() == "arrow_function" || node.Type() == "function_expression" {
			for i := 0; i < int(node.ChildCount()); i++ {
				child := node.Child(i)
				if child != nil && (child.Type() == "statement_block" || child.Type() == "expression") {
					bodyNode = child
					break
				}
			}

			if bodyNode != nil {
				foundTerminator := false
				for i := 0; i < int(bodyNode.ChildCount()); i++ {
					stmt := bodyNode.Child(i)
					if stmt != nil {
						stmtType := stmt.Type()

						if foundTerminator && isStatementNode(stmt, "javascript") {
							startLine, startCol := getLineColumn(code, int(stmt.StartByte()))
							endLine, endCol := getLineColumn(code, int(stmt.EndByte()))

							findings = append(findings, ASTFinding{
								Type:       "unreachable_code",
								Severity:   "warning",
								Line:       startLine,
								Column:     startCol,
								EndLine:    endLine,
								EndColumn:  endCol,
								Message:    "Unreachable code detected after return/throw/break/continue",
								Code:       safeSlice(code, stmt.StartByte(), stmt.EndByte()),
								Suggestion: "Remove unreachable code or move it before the terminating statement",
							})
							break
						}

						if stmtType == "return_statement" || stmtType == "throw_statement" ||
							stmtType == "break_statement" || stmtType == "continue_statement" {
							foundTerminator = true
						} else if stmtType == "if_statement" {
							for j := 0; j < int(stmt.ChildCount()); j++ {
								child := stmt.Child(j)
								// JavaScript uses parenthesized_expression for if conditions
								if child != nil && (child.Type() == "condition" || child.Type() == "parenthesized_expression") {
									condCode := safeSlice(code, child.StartByte(), child.EndByte())
									if strings.Contains(strings.ToLower(condCode), "true") {
										for k := 0; k < int(stmt.ChildCount()); k++ {
											bodyChild := stmt.Child(k)
											if bodyChild != nil && bodyChild.Type() == "statement_block" {
												if hasTerminatingStatement(bodyChild, code) {
													foundTerminator = true
												}
											}
										}
									}
								}
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

func hasTerminatingStatement(node *sitter.Node, code string) bool {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child != nil {
			childType := child.Type()
			if childType == "return_statement" || childType == "throw_statement" ||
				childType == "break_statement" || childType == "continue_statement" {
				return true
			}
		}
	}
	return false
}
