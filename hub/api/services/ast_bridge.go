// Package services AST bridge
// Bridge to connect services layer with AST package
// Complies with CODING_STANDARDS.md: Business Services max 400 lines
package services

import (
	"fmt"
	"sentinel-hub-api/ast"
	"strings"
	"sync"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"
	"github.com/smacker/go-tree-sitter/javascript"
	"github.com/smacker/go-tree-sitter/python"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
)

var parserCache = make(map[string]*sitter.Parser)
var parserMutex sync.RWMutex
var parserOnce sync.Once

// getParser returns a parser for a language
// This mirrors the implementation in ast/parsers.go since getParser is not exported
func getParser(language string) (*sitter.Parser, error) {
	// Normalize language name
	lang := normalizeLanguage(language)

	// Read lock for cache check
	parserMutex.RLock()
	if parser, ok := parserCache[lang]; ok {
		parserMutex.RUnlock()
		return parser, nil
	}
	parserMutex.RUnlock()

	// Write lock for cache miss
	parserMutex.Lock()
	defer parserMutex.Unlock()

	// Double-check after acquiring write lock
	if parser, ok := parserCache[lang]; ok {
		return parser, nil
	}

	// Create and cache parser
	var parser *sitter.Parser
	switch lang {
	case "go", "golang":
		p := sitter.NewParser()
		p.SetLanguage(golang.GetLanguage())
		parser = p
	case "javascript", "js", "jsx":
		p := sitter.NewParser()
		p.SetLanguage(javascript.GetLanguage())
		parser = p
	case "typescript", "ts", "tsx":
		p := sitter.NewParser()
		p.SetLanguage(typescript.GetLanguage())
		parser = p
	case "python", "py":
		p := sitter.NewParser()
		p.SetLanguage(python.GetLanguage())
		parser = p
	default:
		return nil, fmt.Errorf("unsupported language: %s (supported: go, javascript, typescript, python)", language)
	}

	parserCache[lang] = parser
	return parser, nil
}

// normalizeLanguage normalizes language names
func normalizeLanguage(lang string) string {
	lang = strings.ToLower(lang)
	switch lang {
	case "js", "javascript", "jsx":
		return "javascript"
	case "ts", "typescript", "tsx":
		return "typescript"
	case "py", "python":
		return "python"
	case "go", "golang":
		return "go"
	default:
		return lang
	}
}

// AnalyzeCode performs AST analysis on code using the AST package
func AnalyzeCode(code, language string, analyses []string) ([]ast.ASTFinding, ast.AnalysisStats, error) {
	return ast.AnalyzeAST(code, language, analyses)
}

// traverseAST wraps the AST package's traverseAST function for services layer
// This allows logic_analyzer_helpers.go to use the real implementation
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
