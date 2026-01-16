// Package ast provides duplicate detection functionality
// Complies with CODING_STANDARDS.md: Detection modules max 250 lines
package ast

import (
	"fmt"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

// detectDuplicateFunctions finds duplicate function definitions
func detectDuplicateFunctions(root *sitter.Node, code string, language string) []ASTFinding {
	findings := []ASTFinding{}
	functionMap := make(map[string][]*sitter.Node)

	// Traverse AST to find all function definitions
	traverseAST(root, func(node *sitter.Node) bool {
		var funcName string
		var isFunction bool

		switch language {
		case "go":
			if node.Type() == "function_declaration" || node.Type() == "method_declaration" {
				// For method_declaration, format is: receiver method_name
				// For function_declaration, format is: func_name
				for i := 0; i < int(node.ChildCount()); i++ {
					child := node.Child(i)
					if child != nil {
						if child.Type() == "identifier" {
							funcName = code[child.StartByte():child.EndByte()]
							isFunction = true
							break
						} else if child.Type() == "parameter_list" {
							// This is a method receiver - get the method name after it
							continue
						} else if child.Type() == "field_identifier" {
							// Method name in method_declaration
							funcName = code[child.StartByte():child.EndByte()]
							isFunction = true
							break
						}
					}
				}
			}
		case "javascript", "typescript":
			if node.Type() == "function_declaration" || node.Type() == "function" {
				// Find the function name identifier
				for i := 0; i < int(node.ChildCount()); i++ {
					child := node.Child(i)
					if child != nil && child.Type() == "identifier" {
						funcName = code[child.StartByte():child.EndByte()]
						isFunction = true
						break
					}
				}
			}
		case "python":
			if node.Type() == "function_definition" {
				// Find the function name identifier
				for i := 0; i < int(node.ChildCount()); i++ {
					child := node.Child(i)
					if child != nil && child.Type() == "identifier" {
						funcName = code[child.StartByte():child.EndByte()]
						isFunction = true
						break
					}
				}
			}
		}

		if isFunction && funcName != "" {
			functionMap[funcName] = append(functionMap[funcName], node)
		}

		return true // Continue traversal
	})

	// Check for duplicates
	for funcName, nodes := range functionMap {
		if len(nodes) > 1 {
			// Found duplicate - report all occurrences
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
					Code:       code[node.StartByte():node.EndByte()],
					Suggestion: fmt.Sprintf("Remove duplicate definition of '%s' or rename one of them", funcName),
				})
			}
		}
	}

	return findings
}

// detectUnusedVariables finds unused variable declarations
func detectUnusedVariables(root *sitter.Node, code string, language string) []ASTFinding {
	findings := []ASTFinding{}

	switch language {
	case "go":
		findings = detectUnusedVariablesGo(root, code)
	case "javascript", "typescript":
		findings = detectUnusedVariablesJS(root, code)
	case "python":
		findings = detectUnusedVariablesPython(root, code)
	}

	return findings
}

// detectUnusedVariablesGo finds unused variables in Go code
func detectUnusedVariablesGo(root *sitter.Node, code string) []ASTFinding {
	findings := []ASTFinding{}
	variableDeclarations := make(map[string]*sitter.Node)
	variableUsages := make(map[string]bool)

	// First pass: collect all variable declarations
	traverseAST(root, func(node *sitter.Node) bool {
		if node.Type() == "short_var_declaration" || node.Type() == "var_declaration" {
			// Get variable name
			for i := 0; i < int(node.ChildCount()); i++ {
				child := node.Child(i)
				if child != nil && child.Type() == "identifier" {
					varName := code[child.StartByte():child.EndByte()]
					variableDeclarations[varName] = node
					break
				}
			}
		}
		return true
	})

	// Second pass: collect all variable usages
	traverseAST(root, func(node *sitter.Node) bool {
		if node.Type() == "identifier" {
			varName := code[node.StartByte():node.EndByte()]
			variableUsages[varName] = true
		}
		return true
	})

	// Check for unused variables
	for varName, node := range variableDeclarations {
		if !variableUsages[varName] {
			startLine, startCol := getLineColumn(code, int(node.StartByte()))
			endLine, endCol := getLineColumn(code, int(node.EndByte()))

			findings = append(findings, ASTFinding{
				Type:       "unused_variable",
				Severity:   "warning",
				Line:       startLine,
				Column:     startCol,
				EndLine:    endLine,
				EndColumn:  endCol,
				Message:    fmt.Sprintf("Unused variable: '%s' is declared but never used", varName),
				Code:       code[node.StartByte():node.EndByte()],
				Suggestion: fmt.Sprintf("Remove unused variable '%s' or use it in an expression", varName),
			})
		}
	}

	return findings
}

// detectUnusedVariablesJS finds unused variables in JavaScript/TypeScript code
func detectUnusedVariablesJS(root *sitter.Node, code string) []ASTFinding {
	// Basic implementation - can be enhanced later
	findings := []ASTFinding{}
	return findings
}

// detectUnusedVariablesPython finds unused variables in Python code
func detectUnusedVariablesPython(root *sitter.Node, code string) []ASTFinding {
	// Basic implementation - can be enhanced later
	findings := []ASTFinding{}
	return findings
}

// detectUnreachableCode finds unreachable code blocks
func detectUnreachableCode(root *sitter.Node, code string, language string) []ASTFinding {
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
	traverseAST(root, func(node *sitter.Node) bool {
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
									Code:       code[stmt.StartByte():stmt.EndByte()],
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

// detectUnreachableCodeJS finds unreachable code in JavaScript/TypeScript functions
func detectUnreachableCodeJS(root *sitter.Node, code string) []ASTFinding {
	// Basic implementation - can be enhanced later
	findings := []ASTFinding{}
	return findings
}

// detectUnreachableCodePython finds unreachable code in Python functions
func detectUnreachableCodePython(root *sitter.Node, code string) []ASTFinding {
	// Basic implementation - can be enhanced later
	findings := []ASTFinding{}
	return findings
}

// detectOrphanedCode finds code that is never executed
func detectOrphanedCode(root *sitter.Node, code string, language string) []ASTFinding {
	findings := []ASTFinding{}

	// Basic implementation: look for functions that are defined but never called
	// This is a simplified version - full implementation would require call graph analysis

	switch language {
	case "go":
		findings = detectOrphanedCodeGo(root, code)
	}

	return findings
}

// detectOrphanedCodeGo finds orphaned functions in Go code
func detectOrphanedCodeGo(root *sitter.Node, code string) []ASTFinding {
	findings := []ASTFinding{}
	functionNames := make(map[string]*sitter.Node)
	functionCalls := make(map[string]bool)

	// First pass: collect all function definitions
	traverseAST(root, func(node *sitter.Node) bool {
		if node.Type() == "function_declaration" || node.Type() == "method_declaration" {
			// Get function name
			for i := 0; i < int(node.ChildCount()); i++ {
				child := node.Child(i)
				if child != nil && child.Type() == "identifier" {
					funcName := code[child.StartByte():child.EndByte()]
					functionNames[funcName] = node
					break
				}
			}
		}
		return true
	})

	// Second pass: collect all function calls
	traverseAST(root, func(node *sitter.Node) bool {
		if node.Type() == "call_expression" {
			// Get function name being called
			for i := 0; i < int(node.ChildCount()); i++ {
				child := node.Child(i)
				if child != nil && child.Type() == "identifier" {
					funcName := code[child.StartByte():child.EndByte()]
					functionCalls[funcName] = true
					break
				}
			}
		}
		return true
	})

	// Check for orphaned functions (simplified - doesn't handle method calls, etc.)
	for funcName, node := range functionNames {
		if !functionCalls[funcName] && funcName != "main" && !strings.HasPrefix(funcName, "Test") {
			startLine, startCol := getLineColumn(code, int(node.StartByte()))
			endLine, endCol := getLineColumn(code, int(node.EndByte()))

			findings = append(findings, ASTFinding{
				Type:       "orphaned_code",
				Severity:   "info",
				Line:       startLine,
				Column:     startCol,
				EndLine:    endLine,
				EndColumn:  endCol,
				Message:    fmt.Sprintf("Potentially orphaned function: '%s' is defined but never called", funcName),
				Code:       code[node.StartByte():node.EndByte()],
				Suggestion: fmt.Sprintf("Consider removing unused function '%s' or ensure it is called", funcName),
			})
		}
	}

	return findings
}

// detectEmptyCatchBlocks finds empty catch blocks
func detectEmptyCatchBlocks(root *sitter.Node, code string, language string) []ASTFinding {
	// Implementation will be extracted from main ast_analyzer.go
	findings := []ASTFinding{}

	// Placeholder implementation
	return findings
}

// detectMissingAwait finds missing await keywords in async functions
func detectMissingAwait(root *sitter.Node, code string, language string) []ASTFinding {
	// Implementation will be extracted from main ast_analyzer.go
	findings := []ASTFinding{}

	// Placeholder implementation
	return findings
}

// detectBraceMismatch finds mismatched braces/brackets
func detectBraceMismatch(tree *sitter.Tree, code string, language string) []ASTFinding {
	// Implementation will be extracted from main ast_analyzer.go
	findings := []ASTFinding{}

	// Placeholder implementation
	return findings
}
