// Package ast provides Python language-specific extraction
// Complies with CODING_STANDARDS.md: Language implementations max 400 lines
package ast

import (
	"context"
	"fmt"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

// PythonExtractor implements LanguageExtractor for Python
type PythonExtractor struct{}

// ExtractFunctions extracts functions from Python code
func (e *PythonExtractor) ExtractFunctions(code, keyword string) ([]FunctionInfo, error) {
	parser, err := GetParser("python")
	if err != nil {
		return nil, fmt.Errorf("failed to get Python parser: %w", err)
	}
	ctx := context.Background()
	tree, err := parser.ParseCtx(ctx, nil, []byte(code))
	if err != nil {
		return nil, fmt.Errorf("failed to parse Python code: %w", err)
	}
	defer tree.Close()
	rootNode := tree.RootNode()
	if rootNode == nil {
		return nil, fmt.Errorf("failed to get root node")
	}

	functions := []FunctionInfo{}
	keywordLower := strings.ToLower(keyword)

	TraverseAST(rootNode, func(node *sitter.Node) bool {
		funcInfo := extractFunctionFromNode(node, code, "python")
		if funcInfo != nil {
			if keyword == "" || strings.Contains(strings.ToLower(funcInfo.Name), keywordLower) {
				functions = append(functions, *funcInfo)
			}
		}
		return true
	})
	return functions, nil
}

// ExtractImports extracts imports from Python code
func (e *PythonExtractor) ExtractImports(code string) ([]ImportInfo, error) {
	parser, err := GetParser("python")
	if err != nil {
		return nil, fmt.Errorf("failed to get Python parser: %w", err)
	}
	ctx := context.Background()
	tree, err := parser.ParseCtx(ctx, nil, []byte(code))
	if err != nil {
		return nil, fmt.Errorf("failed to parse Python code: %w", err)
	}
	defer tree.Close()
	rootNode := tree.RootNode()
	if rootNode == nil {
		return nil, fmt.Errorf("failed to get root node")
	}

	imports := []ImportInfo{}
	TraverseAST(rootNode, func(node *sitter.Node) bool {
		if node.Type() == "import_statement" || node.Type() == "import_from_statement" {
			for i := 0; i < int(node.ChildCount()); i++ {
				child := node.Child(i)
				if child != nil && (child.Type() == "dotted_name" || child.Type() == "relative_import") {
					path := safeSlice(code, child.StartByte(), child.EndByte())
					imports = append(imports, ImportInfo{Path: path, IsPackage: true})
					break
				}
			}
		}
		return true
	})
	return imports, nil
}

// ExtractSymbols extracts symbols from Python AST
func (e *PythonExtractor) ExtractSymbols(root *sitter.Node, code string) (map[string]*Symbol, error) {
	symbols := make(map[string]*Symbol)
	TraverseAST(root, func(node *sitter.Node) bool {
		if node.Type() == "identifier" {
			name := safeSlice(code, node.StartByte(), node.EndByte())
			if name != "" && name != "self" && name[0] != '_' {
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
