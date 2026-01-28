// Package ast provides Go language-specific extraction implementation
// Complies with CODING_STANDARDS.md: Language implementations max 400 lines
package ast

import (
	"context"
	"fmt"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

// GoExtractor implements LanguageExtractor for Go
type GoExtractor struct{}

// ExtractFunctions extracts functions from Go code
func (e *GoExtractor) ExtractFunctions(code, keyword string) ([]FunctionInfo, error) {
	parser, err := GetParser("go")
	if err != nil {
		return nil, fmt.Errorf("failed to get Go parser: %w", err)
	}

	ctx := context.Background()
	tree, err := parser.ParseCtx(ctx, nil, []byte(code))
	if err != nil {
		return nil, fmt.Errorf("failed to parse Go code: %w", err)
	}
	defer tree.Close()

	rootNode := tree.RootNode()
	if rootNode == nil {
		return nil, fmt.Errorf("failed to get root node")
	}

	functions := []FunctionInfo{}
	TraverseAST(rootNode, func(node *sitter.Node) bool {
		if node.Type() == "function_declaration" || node.Type() == "method_declaration" {
			funcName := ""
			for i := 0; i < int(node.ChildCount()); i++ {
				child := node.Child(i)
				if child != nil {
					if child.Type() == "identifier" {
						funcName = safeSlice(code, child.StartByte(), child.EndByte())
						break
					} else if child.Type() == "field_identifier" {
						funcName = safeSlice(code, child.StartByte(), child.EndByte())
						break
					}
				}
			}

			if funcName != "" {
				if keyword == "" || containsIgnoreCase(funcName, keyword) {
					functions = append(functions, FunctionInfo{
						Name:   funcName,
						Line:   int(node.StartPoint().Row) + 1,
						Column: int(node.StartPoint().Column) + 1,
					})
				}
			}
		}
		return true
	})

	return functions, nil
}

// ExtractImports extracts imports from Go code
func (e *GoExtractor) ExtractImports(code string) ([]ImportInfo, error) {
	parser, err := GetParser("go")
	if err != nil {
		return nil, fmt.Errorf("failed to get Go parser: %w", err)
	}

	ctx := context.Background()
	tree, err := parser.ParseCtx(ctx, nil, []byte(code))
	if err != nil {
		return nil, fmt.Errorf("failed to parse Go code: %w", err)
	}
	defer tree.Close()

	rootNode := tree.RootNode()
	if rootNode == nil {
		return nil, fmt.Errorf("failed to get root node")
	}

	imports := []ImportInfo{}
	TraverseAST(rootNode, func(node *sitter.Node) bool {
		if node.Type() == "import_declaration" {
			// Extract import path
			for i := 0; i < int(node.ChildCount()); i++ {
				child := node.Child(i)
				if child != nil && child.Type() == "interpreted_string_literal" {
					importPath := safeSlice(code, child.StartByte(), child.EndByte())
					// Remove quotes
					if len(importPath) >= 2 {
						importPath = importPath[1 : len(importPath)-1]
					}
					imports = append(imports, ImportInfo{
						Path:      importPath,
						IsPackage: true,
					})
				}
			}
		}
		return true
	})

	return imports, nil
}

// ExtractSymbols extracts symbols from Go code
func (e *GoExtractor) ExtractSymbols(root *sitter.Node, code string) (map[string]*Symbol, error) {
	// Use existing symbol extraction logic
	// This is a simplified version - can be enhanced
	symbols := make(map[string]*Symbol)

	TraverseAST(root, func(node *sitter.Node) bool {
		if node.Type() == "identifier" {
			name := safeSlice(code, node.StartByte(), node.EndByte())
			if name != "" {
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

// Helper function
func containsIgnoreCase(s, substr string) bool {
	sLower := strings.ToLower(s)
	substrLower := strings.ToLower(substr)
	return strings.Contains(sLower, substrLower)
}
