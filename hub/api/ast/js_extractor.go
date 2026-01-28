// Package ast provides JavaScript/TypeScript language-specific extraction
// Complies with CODING_STANDARDS.md: Language implementations max 400 lines
package ast

import (
	"context"
	"fmt"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

// JsExtractor implements LanguageExtractor for JavaScript and TypeScript
type JsExtractor struct {
	// Lang is "javascript" or "typescript" for parser selection
	Lang string
}

// ExtractFunctions extracts functions from JS/TS code
func (e *JsExtractor) ExtractFunctions(code, keyword string) ([]FunctionInfo, error) {
	parser, err := GetParser(e.Lang)
	if err != nil {
		return nil, fmt.Errorf("failed to get parser: %w", err)
	}
	ctx := context.Background()
	tree, err := parser.ParseCtx(ctx, nil, []byte(code))
	if err != nil {
		return nil, fmt.Errorf("failed to parse: %w", err)
	}
	defer tree.Close()
	rootNode := tree.RootNode()
	if rootNode == nil {
		return nil, fmt.Errorf("failed to get root node")
	}

	functions := []FunctionInfo{}
	keywordLower := strings.ToLower(keyword)

	TraverseAST(rootNode, func(node *sitter.Node) bool {
		funcInfo := extractFunctionFromNode(node, code, e.Lang)
		if funcInfo != nil {
			if keyword == "" || strings.Contains(strings.ToLower(funcInfo.Name), keywordLower) {
				functions = append(functions, *funcInfo)
			}
		}
		return true
	})
	return functions, nil
}

// ExtractImports extracts imports from JS/TS code
func (e *JsExtractor) ExtractImports(code string) ([]ImportInfo, error) {
	parser, err := GetParser(e.Lang)
	if err != nil {
		return nil, fmt.Errorf("failed to get parser: %w", err)
	}
	ctx := context.Background()
	tree, err := parser.ParseCtx(ctx, nil, []byte(code))
	if err != nil {
		return nil, fmt.Errorf("failed to parse: %w", err)
	}
	defer tree.Close()
	rootNode := tree.RootNode()
	if rootNode == nil {
		return nil, fmt.Errorf("failed to get root node")
	}

	imports := []ImportInfo{}
	TraverseAST(rootNode, func(node *sitter.Node) bool {
		if node.Type() == "import_statement" || node.Type() == "import_declaration" {
			for i := 0; i < int(node.ChildCount()); i++ {
				child := node.Child(i)
				if child != nil && (child.Type() == "string" || child.Type() == "import_clause") {
					if child.Type() == "string" {
						path := safeSlice(code, child.StartByte(), child.EndByte())
						if len(path) >= 2 {
							path = path[1 : len(path)-1]
						}
						imports = append(imports, ImportInfo{Path: path, IsPackage: false})
					}
					break
				}
			}
		}
		return true
	})
	return imports, nil
}

// ExtractSymbols extracts symbols from JS/TS AST
func (e *JsExtractor) ExtractSymbols(root *sitter.Node, code string) (map[string]*Symbol, error) {
	symbols := make(map[string]*Symbol)
	TraverseAST(root, func(node *sitter.Node) bool {
		if node.Type() == "identifier" || node.Type() == "property_identifier" {
			name := safeSlice(code, node.StartByte(), node.EndByte())
			if name != "" && name[0] != '_' {
				if _, exists := symbols[name]; !exists {
					symbols[name] = &Symbol{
						Name:       name,
						DeclNode:   node,
						UsageCount: 0,
						Position:   node.StartByte(),
						Kind:       "variable",
					}
				}
			}
		}
		return true
	})
	return symbols, nil
}
