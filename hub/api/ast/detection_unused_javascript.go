package ast

import (
	"fmt"

	sitter "github.com/smacker/go-tree-sitter"
)

func detectUnusedVariablesJS(root *sitter.Node, code string) []ASTFinding {
	findings := []ASTFinding{}
	variableDeclarations := make(map[string]*sitter.Node)
	variableUsages := make(map[string]bool)
	importedNames := make(map[string]bool)

	traverseAST(root, func(node *sitter.Node) bool {
		nodeType := node.Type()

		if nodeType == "import_statement" || nodeType == "import_declaration" {
			for i := 0; i < int(node.ChildCount()); i++ {
				child := node.Child(i)
				if child != nil {
					if child.Type() == "import_specifier" || child.Type() == "identifier" {
						for j := 0; j < int(child.ChildCount()); j++ {
							grandchild := child.Child(j)
							if grandchild != nil && (grandchild.Type() == "identifier" || grandchild.Type() == "shorthand_property_identifier") {
								importName := safeSlice(code, grandchild.StartByte(), grandchild.EndByte())
								importedNames[importName] = true
							}
						}
					}
				}
			}
		}

		if nodeType == "lexical_declaration" || nodeType == "variable_declaration" {
			for i := 0; i < int(node.ChildCount()); i++ {
				child := node.Child(i)
				if child != nil {
					if child.Type() == "variable_declarator" {
						for j := 0; j < int(child.ChildCount()); j++ {
							grandchild := child.Child(j)
							if grandchild != nil {
								if grandchild.Type() == "identifier" {
									varName := safeSlice(code, grandchild.StartByte(), grandchild.EndByte())
									variableDeclarations[varName] = node
								} else if grandchild.Type() == "array_pattern" || grandchild.Type() == "object_pattern" {
									traverseAST(grandchild, func(destNode *sitter.Node) bool {
										if destNode.Type() == "identifier" || destNode.Type() == "shorthand_property_identifier" {
											varName := safeSlice(code, destNode.StartByte(), destNode.EndByte())
											variableDeclarations[varName] = node
										}
										return true
									})
								}
							}
						}
					}
				}
			}
		}

		return true
	})

	traverseAST(root, func(node *sitter.Node) bool {
		if node.Type() == "identifier" || node.Type() == "shorthand_property_identifier" {
			parent := node.Parent()
			inDeclaration := false
			if parent != nil {
				parentType := parent.Type()
				if parentType == "variable_declarator" || parentType == "lexical_declaration" ||
					parentType == "variable_declaration" || parentType == "array_pattern" ||
					parentType == "object_pattern" || parentType == "import_specifier" {
					if parentType == "variable_declarator" {
						if node == parent.Child(0) {
							inDeclaration = true
						}
					} else {
						inDeclaration = true
					}
				}
			}

			if !inDeclaration {
				varName := safeSlice(code, node.StartByte(), node.EndByte())
				variableUsages[varName] = true
			}
		}

		return true
	})

	for varName, node := range variableDeclarations {
		if !variableUsages[varName] && !importedNames[varName] {
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
