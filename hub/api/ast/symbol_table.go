// Package ast provides multi-file symbol table for cross-file analysis
// Complies with CODING_STANDARDS.md: Utility modules max 250 lines
package ast

import (
	"fmt"
	"sync"

	sitter "github.com/smacker/go-tree-sitter"
)

// FileSymbol represents a symbol defined in a specific file
type FileSymbol struct {
	Name     string
	Kind     string // "function", "class", "variable", "type", "constant"
	FilePath string
	Line     int
	Column   int
	Exported bool // Is this symbol exported/public?
	DeclNode *sitter.Node
	Language string
}

// SymbolReference represents a reference to a symbol
type SymbolReference struct {
	Name     string
	FilePath string
	Line     int
	Column   int
	Kind     string // "import", "call", "usage"
}

// SymbolTable manages symbols across multiple files
type SymbolTable struct {
	symbols     map[string][]*FileSymbol // symbol name -> definitions
	references  map[string][]*SymbolReference
	mu          sync.RWMutex
	fileSymbols map[string]map[string]*FileSymbol // file path -> symbol name -> symbol
}

// NewSymbolTable creates a new symbol table
func NewSymbolTable() *SymbolTable {
	return &SymbolTable{
		symbols:     make(map[string][]*FileSymbol),
		references:  make(map[string][]*SymbolReference),
		fileSymbols: make(map[string]map[string]*FileSymbol),
	}
}

// AddSymbol adds a symbol definition to the table
func (st *SymbolTable) AddSymbol(symbol *FileSymbol) error {
	if symbol == nil {
		return fmt.Errorf("symbol cannot be nil")
	}
	if symbol.Name == "" {
		return fmt.Errorf("symbol name cannot be empty")
	}

	st.mu.Lock()
	defer st.mu.Unlock()

	// Initialize file map if needed
	if st.fileSymbols[symbol.FilePath] == nil {
		st.fileSymbols[symbol.FilePath] = make(map[string]*FileSymbol)
	}

	// Add to global symbols map
	st.symbols[symbol.Name] = append(st.symbols[symbol.Name], symbol)

	// Add to file-specific map
	st.fileSymbols[symbol.FilePath][symbol.Name] = symbol

	return nil
}

// AddReference adds a symbol reference to the table
func (st *SymbolTable) AddReference(ref *SymbolReference) {
	if ref == nil || ref.Name == "" {
		return
	}

	st.mu.Lock()
	defer st.mu.Unlock()

	st.references[ref.Name] = append(st.references[ref.Name], ref)
}

// GetSymbols returns all definitions of a symbol
func (st *SymbolTable) GetSymbols(name string) []*FileSymbol {
	st.mu.RLock()
	defer st.mu.RUnlock()

	return st.symbols[name]
}

// GetReferences returns all references to a symbol
func (st *SymbolTable) GetReferences(name string) []*SymbolReference {
	st.mu.RLock()
	defer st.mu.RUnlock()

	return st.references[name]
}

// GetFileSymbols returns all symbols defined in a file
func (st *SymbolTable) GetFileSymbols(filePath string) map[string]*FileSymbol {
	st.mu.RLock()
	defer st.mu.RUnlock()

	result := make(map[string]*FileSymbol)
	if fileSymbols, ok := st.fileSymbols[filePath]; ok {
		for name, symbol := range fileSymbols {
			result[name] = symbol
		}
	}
	return result
}

// FindUnusedExports finds exported symbols that are never referenced
func (st *SymbolTable) FindUnusedExports() []*FileSymbol {
	st.mu.RLock()
	defer st.mu.RUnlock()

	unused := []*FileSymbol{}
	for name, symbols := range st.symbols {
		// Only check exported symbols
		for _, symbol := range symbols {
			if !symbol.Exported {
				continue
			}

			// Check if there are any references (excluding the definition itself)
			refs := st.references[name]
			hasExternalRef := false
			for _, ref := range refs {
				if ref.FilePath != symbol.FilePath {
					hasExternalRef = true
					break
				}
			}

			if !hasExternalRef {
				unused = append(unused, symbol)
			}
		}
	}

	return unused
}

// FindUndefinedReferences finds references to symbols that don't exist
func (st *SymbolTable) FindUndefinedReferences() []*SymbolReference {
	st.mu.RLock()
	defer st.mu.RUnlock()

	undefined := []*SymbolReference{}
	for name, refs := range st.references {
		// Check if symbol is defined
		if len(st.symbols[name]) == 0 {
			undefined = append(undefined, refs...)
		}
	}

	return undefined
}

// ExtractSymbolsFromFile extracts all symbols from a parsed file
func ExtractSymbolsFromFile(rootNode *sitter.Node, code, filePath, language string) ([]*FileSymbol, error) {
	if rootNode == nil {
		return nil, fmt.Errorf("root node cannot be nil")
	}

	symbols := []*FileSymbol{}
	visitor := func(node *sitter.Node) bool {
		var symbol *FileSymbol

		switch language {
		case "go":
			symbol = extractGoSymbol(node, code, filePath, language)
		case "javascript", "typescript":
			symbol = extractJSSymbol(node, code, filePath, language)
		case "python":
			symbol = extractPythonSymbol(node, code, filePath, language)
		}

		if symbol != nil {
			symbols = append(symbols, symbol)
		}

		return true
	}

	traverseAST(rootNode, visitor)
	return symbols, nil
}

// extractGoSymbol extracts a symbol from a Go AST node
func extractGoSymbol(node *sitter.Node, code, filePath, language string) *FileSymbol {
	nodeType := node.Type()

	// Function declaration
	if nodeType == "function_declaration" || nodeType == "method_declaration" {
		name := extractFunctionName(node, code)
		if name == "" {
			return nil
		}

		line, col := getLineColumn(code, int(node.StartByte()))
		isExported := len(name) > 0 && name[0] >= 'A' && name[0] <= 'Z'

		return &FileSymbol{
			Name:     name,
			Kind:     "function",
			FilePath: filePath,
			Line:     line,
			Column:   col,
			Exported: isExported,
			DeclNode: node,
			Language: language,
		}
	}

	// Type declaration
	if nodeType == "type_declaration" {
		name := extractTypeName(node, code)
		if name == "" {
			return nil
		}

		line, col := getLineColumn(code, int(node.StartByte()))
		isExported := len(name) > 0 && name[0] >= 'A' && name[0] <= 'Z'

		return &FileSymbol{
			Name:     name,
			Kind:     "type",
			FilePath: filePath,
			Line:     line,
			Column:   col,
			Exported: isExported,
			DeclNode: node,
			Language: language,
		}
	}

	return nil
}

// extractJSSymbol extracts a symbol from a JavaScript/TypeScript AST node
func extractJSSymbol(node *sitter.Node, code, filePath, language string) *FileSymbol {
	nodeType := node.Type()

	// Function declaration
	if nodeType == "function_declaration" {
		name := extractFunctionName(node, code)
		if name == "" {
			return nil
		}

		line, col := getLineColumn(code, int(node.StartByte()))
		return &FileSymbol{
			Name:     name,
			Kind:     "function",
			FilePath: filePath,
			Line:     line,
			Column:   col,
			Exported: false, // JS doesn't have explicit exports in function_declaration
			DeclNode: node,
			Language: language,
		}
	}

	// Export statement
	if nodeType == "export_statement" {
		// Extract exported name from export
		name := extractExportName(node, code)
		if name == "" {
			return nil
		}

		line, col := getLineColumn(code, int(node.StartByte()))
		return &FileSymbol{
			Name:     name,
			Kind:     "export",
			FilePath: filePath,
			Line:     line,
			Column:   col,
			Exported: true,
			DeclNode: node,
			Language: language,
		}
	}

	return nil
}

// extractPythonSymbol extracts a symbol from a Python AST node
func extractPythonSymbol(node *sitter.Node, code, filePath, language string) *FileSymbol {
	nodeType := node.Type()

	// Function definition
	if nodeType == "function_definition" {
		name := extractFunctionName(node, code)
		if name == "" {
			return nil
		}

		line, col := getLineColumn(code, int(node.StartByte()))
		return &FileSymbol{
			Name:     name,
			Kind:     "function",
			FilePath: filePath,
			Line:     line,
			Column:   col,
			Exported: false, // Python doesn't have explicit exports
			DeclNode: node,
			Language: language,
		}
	}

	// Class definition
	if nodeType == "class_definition" {
		name := extractClassName(node, code)
		if name == "" {
			return nil
		}

		line, col := getLineColumn(code, int(node.StartByte()))
		return &FileSymbol{
			Name:     name,
			Kind:     "class",
			FilePath: filePath,
			Line:     line,
			Column:   col,
			Exported: false,
			DeclNode: node,
			Language: language,
		}
	}

	return nil
}

// Helper functions for extracting names
func extractFunctionName(node *sitter.Node, code string) string {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}
		if child.Type() == "identifier" || child.Type() == "field_identifier" {
			return safeSlice(code, child.StartByte(), child.EndByte())
		}
	}
	return ""
}

func extractTypeName(node *sitter.Node, code string) string {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}
		if child.Type() == "type_spec" {
			for j := 0; j < int(child.ChildCount()); j++ {
				grandchild := child.Child(j)
				if grandchild != nil && grandchild.Type() == "type_identifier" {
					return safeSlice(code, grandchild.StartByte(), grandchild.EndByte())
				}
			}
		}
	}
	return ""
}

func extractExportName(node *sitter.Node, code string) string {
	// Simplified - would need more complex logic for full export parsing
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}
		if child.Type() == "identifier" {
			return safeSlice(code, child.StartByte(), child.EndByte())
		}
	}
	return ""
}

func extractClassName(node *sitter.Node, code string) string {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}
		if child.Type() == "identifier" {
			return safeSlice(code, child.StartByte(), child.EndByte())
		}
	}
	return ""
}
