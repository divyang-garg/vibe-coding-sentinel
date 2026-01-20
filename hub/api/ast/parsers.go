// Package ast provides AST parsing and analysis capabilities
// Complies with CODING_STANDARDS.md: AST modules max 300 lines
package ast

import (
	"fmt"
	"strings"
	"sync"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"
	"github.com/smacker/go-tree-sitter/javascript"
	"github.com/smacker/go-tree-sitter/python"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
)

var parsers = make(map[string]*sitter.Parser)
var parsersMutex sync.RWMutex
var parsersOnce sync.Once

// initParsers initializes the parser pool
func initParsers() {
	// Go parser
	goParser := sitter.NewParser()
	goParser.SetLanguage(golang.GetLanguage())
	parsers["go"] = goParser
	parsers["golang"] = goParser

	// JavaScript parser
	jsParser := sitter.NewParser()
	jsParser.SetLanguage(javascript.GetLanguage())
	parsers["javascript"] = jsParser
	parsers["js"] = jsParser
	parsers["jsx"] = jsParser

	// TypeScript parser
	tsParser := sitter.NewParser()
	tsParser.SetLanguage(typescript.GetLanguage())
	parsers["typescript"] = tsParser
	parsers["ts"] = tsParser
	parsers["tsx"] = tsParser

	// Python parser
	pyParser := sitter.NewParser()
	pyParser.SetLanguage(python.GetLanguage())
	parsers["python"] = pyParser
	parsers["py"] = pyParser
}

// getParser gets a parser for the specified language
func getParser(language string) (*sitter.Parser, error) {
	// Normalize language name
	lang := normalizeLanguage(language)

	// Initialize parsers once with sync.Once
	parsersOnce.Do(initParsers)

	// Read lock for cache check
	parsersMutex.RLock()
	if parser, ok := parsers[lang]; ok {
		parsersMutex.RUnlock()
		return parser, nil
	}
	parsersMutex.RUnlock()

	// Parser not found - this shouldn't happen if initParsers was called
	// but we'll return an error rather than panicking
	return nil, fmt.Errorf("unsupported language: %s (supported: go, javascript, typescript, python)", language)
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
