// Package ast provides duplicate function detection
// Complies with CODING_STANDARDS.md: Detection modules max 250 lines
package ast

import (
	"fmt"

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
							funcName = safeSlice(code, child.StartByte(), child.EndByte())
							isFunction = true
							break
						} else if child.Type() == "parameter_list" {
							// This is a method receiver - get the method name after it
							continue
						} else if child.Type() == "field_identifier" {
							// Method name in method_declaration
							funcName = safeSlice(code, child.StartByte(), child.EndByte())
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
						funcName = safeSlice(code, child.StartByte(), child.EndByte())
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
						funcName = safeSlice(code, child.StartByte(), child.EndByte())
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
					Code:       safeSlice(code, node.StartByte(), node.EndByte()),
					Suggestion: fmt.Sprintf("Remove duplicate definition of '%s' or rename one of them", funcName),
				})
			}
		}
	}

	return findings
}
