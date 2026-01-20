// Package ast provides Python unused variable detection
// Complies with CODING_STANDARDS.md: Detection modules max 250 lines
package ast

import (
	"fmt"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

// detectUnusedVariablesPython finds unused variables in Python code
func detectUnusedVariablesPython(root *sitter.Node, code string) []ASTFinding {
	findings := []ASTFinding{}
	variableDeclarations := make(map[string]*sitter.Node)
	declarationPositions := make(map[uint32]bool) // Track byte offsets of declared variable/parameter names
	variableUsages := make(map[string]bool)
	importedNames := make(map[string]bool)

	// First pass: collect all variable declarations and imports
	traverseAST(root, func(node *sitter.Node) bool {
		nodeType := node.Type()

		// Track imports (imported names should not be flagged as unused)
		if nodeType == "import_statement" || nodeType == "import_from_statement" {
			for i := 0; i < int(node.ChildCount()); i++ {
				child := node.Child(i)
				if child != nil {
					if child.Type() == "dotted_as_name" || child.Type() == "dotted_name" || child.Type() == "import_from_statement" {
						for j := 0; j < int(child.ChildCount()); j++ {
							grandchild := child.Child(j)
							if grandchild != nil && (grandchild.Type() == "identifier" || grandchild.Type() == "aliased_import") {
								varName := safeSlice(code, grandchild.StartByte(), grandchild.EndByte())
								importedNames[varName] = true
							}
						}
					}
				}
			}
		}

		// Track variable assignments at module level and in functions
		if nodeType == "assignment" {
			for i := 0; i < int(node.ChildCount()); i++ {
				child := node.Child(i)
				if child != nil {
					if child.Type() == "left_hand_side" || child.Type() == "identifier" {
						// Get variable name
						if child.Type() == "identifier" {
							varName := safeSlice(code, child.StartByte(), child.EndByte())
							// Skip special Python names like __all__
							if !strings.HasPrefix(varName, "__") || varName == "__all__" {
								variableDeclarations[varName] = node
								declarationPositions[child.StartByte()] = true
							}
						} else {
							// Handle tuple/list unpacking: a, b = ...
							traverseAST(child, func(destNode *sitter.Node) bool {
								if destNode.Type() == "identifier" && destNode.Parent() != nil && destNode.Parent().Type() == "tuple" {
									varName := safeSlice(code, destNode.StartByte(), destNode.EndByte())
									if !strings.HasPrefix(varName, "_") || varName == "_" {
										variableDeclarations[varName] = node
										declarationPositions[destNode.StartByte()] = true
									}
								}
								return true
							})
						}
					}
				}
			}
		}

		// Track function parameters (may be unused)
		if nodeType == "function_definition" {
			for i := 0; i < int(node.ChildCount()); i++ {
				child := node.Child(i)
				if child != nil && child.Type() == "parameters" {
					traverseAST(child, func(paramNode *sitter.Node) bool {
						if paramNode.Type() == "identifier" {
							paramName := safeSlice(code, paramNode.StartByte(), paramNode.EndByte())
							// Skip self/cls and single underscore
							if paramName != "self" && paramName != "cls" && paramName != "_" {
								variableDeclarations[paramName] = node
								declarationPositions[paramNode.StartByte()] = true
							}
						}
						return true
					})
				}
			}
		}

		return true
	})

	// Second pass: collect all variable usages (excluding declaration positions)
	traverseAST(root, func(node *sitter.Node) bool {
		if node.Type() == "identifier" {
			// Skip if this identifier is at a declaration position
			if !declarationPositions[node.StartByte()] {
				// Also check if this is in an assignment (declaration context)
				parent := node.Parent()
				inAssignment := false
				if parent != nil && parent.Type() == "assignment" {
					// Check if this is on the left side (declaration)
					for i := 0; i < int(parent.ChildCount()); i++ {
						if parent.Child(i) == node {
							inAssignment = true
							break
						}
					}
				}

				if !inAssignment {
					varName := safeSlice(code, node.StartByte(), node.EndByte())
					variableUsages[varName] = true
				}
			}
		}

		return true
	})

	// Check for unused variables (excluding imports and _ convention)
	for varName, node := range variableDeclarations {
		if !variableUsages[varName] && !importedNames[varName] && varName != "_" {
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
				Code:       safeSlice(code, node.StartByte(), node.EndByte()),
				Suggestion: fmt.Sprintf("Remove unused variable '%s' or use it in an expression", varName),
			})
		}
	}

	return findings
}
