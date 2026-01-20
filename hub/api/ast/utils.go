// Package ast provides AST utility functions
// Complies with CODING_STANDARDS.md: Utility modules max 250 lines
package ast

import (
	"strings"
	"unicode"

	sitter "github.com/smacker/go-tree-sitter"
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
	codeLen := len(code)
	if byteOffset > codeLen {
		byteOffset = codeLen
	}
	if byteOffset < 0 {
		byteOffset = 0
	}
	lines := strings.Split(code[:byteOffset], "\n")
	line = len(lines)
	if line == 0 {
		line = 1
		column = 1
		return line, column
	}
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

// safeSlice safely extracts a substring with bounds checking
func safeSlice(code string, start, end uint32) string {
	codeLen := uint32(len(code))
	if start > codeLen {
		start = codeLen
	}
	if end > codeLen {
		end = codeLen
	}
	if start > end {
		return ""
	}
	return code[start:end]
}

// shouldExcludeFunction checks if a function should be excluded from orphaned detection
func shouldExcludeFunction(funcName string, config DetectionConfig) bool {
	// Check exact matches
	for _, excluded := range config.ExcludedFunctions {
		if funcName == excluded {
			return true
		}
	}
	// Check prefixes
	for _, prefix := range config.ExcludedPrefixes {
		if strings.HasPrefix(funcName, prefix) {
			return true
		}
	}
	// Check exported functions (uppercase first letter)
	if config.TrustExported && len(funcName) > 0 && unicode.IsUpper(rune(funcName[0])) {
		return true
	}
	return false
}
