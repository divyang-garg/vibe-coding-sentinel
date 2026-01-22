// Package services AST bridge
// Bridge to connect services layer with AST package
// Complies with CODING_STANDARDS.md: Business Services max 400 lines
package services

import (
	"strings"

	"sentinel-hub-api/ast"

	sitter "github.com/smacker/go-tree-sitter"
)

// GetParser returns a parser for a language
// Wraps the AST package's GetParser function
func GetParser(language string) (*sitter.Parser, error) {
	return ast.GetParser(language)
}

// getParser returns a parser for a language (deprecated - use GetParser)
// Kept for backward compatibility
func getParser(language string) (*sitter.Parser, error) {
	return GetParser(language)
}

// AnalyzeCode performs AST analysis on code using the AST package
func AnalyzeCode(code, language string, analyses []string) ([]ast.ASTFinding, ast.AnalysisStats, error) {
	return ast.AnalyzeAST(code, language, analyses)
}

// TraverseAST wraps the AST package's TraverseAST function for services layer
// This allows logic_analyzer_helpers.go to use the real implementation
func TraverseAST(node *sitter.Node, visitor func(*sitter.Node) bool) {
	ast.TraverseAST(node, visitor)
}

// getLineColumn converts byte offset to line and column numbers
// Wraps the AST package utility (same implementation as ast/utils.go)
func getLineColumn(code string, byteOffset int) (int, int) {
	if byteOffset >= len(code) {
		byteOffset = len(code)
	}
	if byteOffset < 0 {
		byteOffset = 0
	}

	lines := strings.Split(code[:byteOffset], "\n")
	line := len(lines)
	if line == 0 {
		line = 1
	}
	column := len(lines[len(lines)-1]) + 1
	return line, column
}
