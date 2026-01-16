// Package ast provides AST utility functions
// Complies with CODING_STANDARDS.md: Utility modules max 250 lines
package ast

import (
	sitter "github.com/smacker/go-tree-sitter"
	"strings"
)

// traverseAST traverses the AST tree with a visitor function
func traverseAST(node *sitter.Node, visitor func(*sitter.Node) bool) {
	if node == nil {
		return
	}

	if !visitor(node) {
		return
	}

	for i := 0; i < int(node.ChildCount()); i++ {
		traverseAST(node.Child(i), visitor)
	}
}

// countNodes counts all nodes in the AST
func countNodes(node *sitter.Node) int {
	count := 1
	for i := 0; i < int(node.ChildCount()); i++ {
		count += countNodes(node.Child(i))
	}
	return count
}

// getLineColumn converts byte offset to line and column numbers
func getLineColumn(code string, byteOffset int) (line, column int) {
	lines := strings.Split(code[:byteOffset], "\n")
	line = len(lines)
	column = len(lines[len(lines)-1]) + 1
	return line, column
}

// isStatementNode checks if a node represents a statement
func isStatementNode(node *sitter.Node, language string) bool {
	if node == nil {
		return false
	}

	nodeType := node.Type()
	switch language {
	case "go":
		return nodeType == "expression_statement" ||
			nodeType == "return_statement" ||
			nodeType == "if_statement" ||
			nodeType == "for_statement"
	case "javascript", "typescript":
		return strings.Contains(nodeType, "statement")
	case "python":
		return strings.Contains(nodeType, "statement")
	default:
		return false
	}
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
