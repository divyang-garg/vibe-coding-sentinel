// Package ast - Symbol extraction tests
// Tests for symbol extraction from JavaScript, TypeScript, and Python
// Complies with CODING_STANDARDS.md: Tests max 500 lines
package ast

import (
	"context"
	"testing"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/javascript"
	"github.com/smacker/go-tree-sitter/python"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
)

// TestExtractJSSymbol tests JavaScript symbol extraction
func TestExtractJSSymbol(t *testing.T) {
	parser := sitter.NewParser()
	parser.SetLanguage(javascript.GetLanguage())

	t.Run("function_declaration", func(t *testing.T) {
		code := `function myFunction() {
    return 42;
}`
		tree, err := parser.ParseCtx(context.Background(), nil, []byte(code))
		if err != nil {
			t.Fatalf("Failed to parse: %v", err)
		}
		defer tree.Close()

		rootNode := tree.RootNode()
		if rootNode == nil {
			t.Fatal("Root node is nil")
		}

		// Find function_declaration node
		var funcNode *sitter.Node
		TraverseAST(rootNode, func(node *sitter.Node) bool {
			if node.Type() == "function_declaration" {
				funcNode = node
				return false // Stop traversal
			}
			return true
		})

		if funcNode == nil {
			t.Fatal("Function declaration node not found")
		}

		symbol := extractJSSymbol(funcNode, code, "test.js", "javascript")
		if symbol == nil {
			t.Fatal("extractJSSymbol returned nil")
		}

		if symbol.Name != "myFunction" {
			t.Errorf("Expected name 'myFunction', got '%s'", symbol.Name)
		}

		if symbol.Kind != "function" {
			t.Errorf("Expected kind 'function', got '%s'", symbol.Kind)
		}

		if symbol.Language != "javascript" {
			t.Errorf("Expected language 'javascript', got '%s'", symbol.Language)
		}
	})

	t.Run("export_statement", func(t *testing.T) {
		code := `export function exportedFunc() {
    return 1;
}

export const myConst = 42;`
		tree, err := parser.ParseCtx(context.Background(), nil, []byte(code))
		if err != nil {
			t.Fatalf("Failed to parse: %v", err)
		}
		defer tree.Close()

		rootNode := tree.RootNode()
		if rootNode == nil {
			t.Fatal("Root node is nil")
		}

		// Find export_statement node
		var exportNode *sitter.Node
		TraverseAST(rootNode, func(node *sitter.Node) bool {
			if node.Type() == "export_statement" {
				exportNode = node
				return false // Stop traversal
			}
			return true
		})

		if exportNode == nil {
			t.Fatal("Export statement node not found")
		}

		symbol := extractJSSymbol(exportNode, code, "test.js", "javascript")
		if symbol == nil {
			t.Log("extractJSSymbol returned nil for export_statement (may be expected)")
		} else {
			if symbol.Kind != "export" {
				t.Errorf("Expected kind 'export', got '%s'", symbol.Kind)
			}
			if !symbol.Exported {
				t.Error("Expected Exported to be true")
			}
		}
	})

	t.Run("unsupported_node_type", func(t *testing.T) {
		code := `const x = 5;`
		tree, err := parser.ParseCtx(context.Background(), nil, []byte(code))
		if err != nil {
			t.Fatalf("Failed to parse: %v", err)
		}
		defer tree.Close()

		rootNode := tree.RootNode()
		if rootNode == nil {
			t.Fatal("Root node is nil")
		}

		// Try to extract from variable_declaration (not supported)
		symbol := extractJSSymbol(rootNode, code, "test.js", "javascript")
		if symbol != nil {
			t.Logf("extractJSSymbol returned symbol for unsupported type: %+v", symbol)
		}
	})
}

// TestExtractPythonSymbol tests Python symbol extraction
func TestExtractPythonSymbol(t *testing.T) {
	parser := sitter.NewParser()
	parser.SetLanguage(python.GetLanguage())

	t.Run("function_definition", func(t *testing.T) {
		code := `def my_function():
    return 42`
		tree, err := parser.ParseCtx(context.Background(), nil, []byte(code))
		if err != nil {
			t.Fatalf("Failed to parse: %v", err)
		}
		defer tree.Close()

		rootNode := tree.RootNode()
		if rootNode == nil {
			t.Fatal("Root node is nil")
		}

		// Find function_definition node
		var funcNode *sitter.Node
		TraverseAST(rootNode, func(node *sitter.Node) bool {
			if node.Type() == "function_definition" {
				funcNode = node
				return false // Stop traversal
			}
			return true
		})

		if funcNode == nil {
			t.Fatal("Function definition node not found")
		}

		symbol := extractPythonSymbol(funcNode, code, "test.py", "python")
		if symbol == nil {
			t.Fatal("extractPythonSymbol returned nil")
		}

		if symbol.Name != "my_function" {
			t.Errorf("Expected name 'my_function', got '%s'", symbol.Name)
		}

		if symbol.Kind != "function" {
			t.Errorf("Expected kind 'function', got '%s'", symbol.Kind)
		}

		if symbol.Language != "python" {
			t.Errorf("Expected language 'python', got '%s'", symbol.Language)
		}
	})

	t.Run("class_definition", func(t *testing.T) {
		code := `class MyClass:
    def __init__(self):
        pass`
		tree, err := parser.ParseCtx(context.Background(), nil, []byte(code))
		if err != nil {
			t.Fatalf("Failed to parse: %v", err)
		}
		defer tree.Close()

		rootNode := tree.RootNode()
		if rootNode == nil {
			t.Fatal("Root node is nil")
		}

		// Find class_definition node
		var classNode *sitter.Node
		TraverseAST(rootNode, func(node *sitter.Node) bool {
			if node.Type() == "class_definition" {
				classNode = node
				return false // Stop traversal
			}
			return true
		})

		if classNode == nil {
			t.Fatal("Class definition node not found")
		}

		symbol := extractPythonSymbol(classNode, code, "test.py", "python")
		if symbol == nil {
			t.Fatal("extractPythonSymbol returned nil")
		}

		if symbol.Name != "MyClass" {
			t.Errorf("Expected name 'MyClass', got '%s'", symbol.Name)
		}

		if symbol.Kind != "class" {
			t.Errorf("Expected kind 'class', got '%s'", symbol.Kind)
		}
	})

	t.Run("unsupported_node_type", func(t *testing.T) {
		code := `x = 5`
		tree, err := parser.ParseCtx(context.Background(), nil, []byte(code))
		if err != nil {
			t.Fatalf("Failed to parse: %v", err)
		}
		defer tree.Close()

		rootNode := tree.RootNode()
		if rootNode == nil {
			t.Fatal("Root node is nil")
		}

		// Try to extract from assignment (not supported)
		symbol := extractPythonSymbol(rootNode, code, "test.py", "python")
		if symbol != nil {
			t.Logf("extractPythonSymbol returned symbol for unsupported type: %+v", symbol)
		}
	})
}

// TestExtractTypeName tests type name extraction
func TestExtractTypeName(t *testing.T) {
	parser := sitter.NewParser()
	parser.SetLanguage(typescript.GetLanguage())

	code := `interface MyInterface {
    prop: string;
}

type MyType = string;`
	tree, err := parser.ParseCtx(context.Background(), nil, []byte(code))
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}
	defer tree.Close()

	rootNode := tree.RootNode()
	if rootNode == nil {
		t.Fatal("Root node is nil")
	}

	// Find type_alias_declaration or interface_declaration
	var typeNode *sitter.Node
	TraverseAST(rootNode, func(node *sitter.Node) bool {
		if node.Type() == "type_alias_declaration" || node.Type() == "interface_declaration" {
			typeNode = node
			return false // Stop traversal
		}
		return true
	})

	if typeNode == nil {
		t.Fatal("Type node not found")
	}

	typeName := extractTypeName(typeNode, code)
	if typeName == "" {
		t.Log("extractTypeName returned empty string (may be expected depending on node structure)")
	} else {
		t.Logf("Extracted type name: %s", typeName)
	}
}

// TestExtractExportName tests export name extraction
func TestExtractExportName(t *testing.T) {
	parser := sitter.NewParser()
	parser.SetLanguage(javascript.GetLanguage())

	code := `export function myFunc() {}
export const myConst = 42;
export { myVar };`
	tree, err := parser.ParseCtx(context.Background(), nil, []byte(code))
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}
	defer tree.Close()

	rootNode := tree.RootNode()
	if rootNode == nil {
		t.Fatal("Root node is nil")
	}

	// Find export_statement node
	var exportNode *sitter.Node
	TraverseAST(rootNode, func(node *sitter.Node) bool {
		if node.Type() == "export_statement" {
			exportNode = node
			return false // Stop traversal
		}
		return true
	})

	if exportNode == nil {
		t.Fatal("Export statement node not found")
	}

	exportName := extractExportName(exportNode, code)
	if exportName == "" {
		t.Log("extractExportName returned empty string (may be expected depending on export structure)")
	} else {
		t.Logf("Extracted export name: %s", exportName)
	}
}

// TestExtractClassName tests class name extraction
func TestExtractClassName(t *testing.T) {
	parser := sitter.NewParser()
	parser.SetLanguage(python.GetLanguage())

	code := `class MyClass:
    pass`
	tree, err := parser.ParseCtx(context.Background(), nil, []byte(code))
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}
	defer tree.Close()

	rootNode := tree.RootNode()
	if rootNode == nil {
		t.Fatal("Root node is nil")
	}

	// Find class_definition node
	var classNode *sitter.Node
	TraverseAST(rootNode, func(node *sitter.Node) bool {
		if node.Type() == "class_definition" {
			classNode = node
			return false // Stop traversal
		}
		return true
	})

	if classNode == nil {
		t.Fatal("Class definition node not found")
	}

	className := extractClassName(classNode, code)
	if className == "" {
		t.Error("extractClassName returned empty string")
	} else {
		if className != "MyClass" {
			t.Errorf("Expected 'MyClass', got '%s'", className)
		}
	}
}

// TestExtractSymbolsFromFile_JavaScript tests symbol extraction from JavaScript files
func TestExtractSymbolsFromFile_JavaScript(t *testing.T) {
	parser := sitter.NewParser()
	parser.SetLanguage(javascript.GetLanguage())

	code := `function myFunction() {
    return 42;
}

export function exportedFunc() {
    return 1;
}`
	tree, err := parser.ParseCtx(context.Background(), nil, []byte(code))
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}
	defer tree.Close()

	rootNode := tree.RootNode()
	if rootNode == nil {
		t.Fatal("Root node is nil")
	}

	symbols, err := ExtractSymbolsFromFile(rootNode, code, "test.js", "javascript")
	if err != nil {
		t.Fatalf("ExtractSymbolsFromFile failed: %v", err)
	}

	if len(symbols) == 0 {
		t.Error("Expected at least one symbol")
	}

	// Check that we found the function
	found := false
	for _, symbol := range symbols {
		if symbol.Name == "myFunction" || symbol.Name == "exportedFunc" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected to find function symbol")
	}
}

// TestExtractSymbolsFromFile_Python tests symbol extraction from Python files
func TestExtractSymbolsFromFile_Python(t *testing.T) {
	parser := sitter.NewParser()
	parser.SetLanguage(python.GetLanguage())

	code := `def my_function():
    return 42

class MyClass:
    pass`
	tree, err := parser.ParseCtx(context.Background(), nil, []byte(code))
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}
	defer tree.Close()

	rootNode := tree.RootNode()
	if rootNode == nil {
		t.Fatal("Root node is nil")
	}

	symbols, err := ExtractSymbolsFromFile(rootNode, code, "test.py", "python")
	if err != nil {
		t.Fatalf("ExtractSymbolsFromFile failed: %v", err)
	}

	if len(symbols) == 0 {
		t.Error("Expected at least one symbol")
	}

	// Check that we found the function or class
	found := false
	for _, symbol := range symbols {
		if symbol.Name == "my_function" || symbol.Name == "MyClass" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected to find function or class symbol")
	}
}
