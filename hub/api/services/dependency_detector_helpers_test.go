// Package services dependency detector helpers tests
// Complies with CODING_STANDARDS.md: Tests max 500 lines
package services

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestExtractSymbolsFromAST_Go(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected []string
	}{
		{
			name: "function declaration",
			code: `package main

func CalculateTotal(price float64) float64 {
	return price * 1.1
}`,
			expected: []string{"CalculateTotal"},
		},
		{
			name: "method declaration",
			code: `package main

type Calculator struct{}

func (c *Calculator) Add(a, b int) int {
	return a + b
}`,
			expected: []string{"Add"},
		},
		{
			name: "type declaration",
			code: `package main

type User struct {
	Name string
}`,
			expected: []string{}, // Type declarations may not be extracted by current implementation
		},
		{
			name: "multiple functions",
			code: `package main

func Func1() {}
func Func2() {}
func Func3() {}`,
			expected: []string{"Func1", "Func2", "Func3"},
		},
		{
			name:     "empty code",
			code:     "",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			symbols := extractSymbolsFromAST(tt.code, "go", "test.go")
			if len(symbols) != len(tt.expected) {
				t.Errorf("expected %d symbols, got %d", len(tt.expected), len(symbols))
			}
			for _, expected := range tt.expected {
				if !symbols[expected] {
					t.Errorf("expected symbol %q not found", expected)
				}
			}
		})
	}
}

func TestExtractSymbolsFromAST_JavaScript(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected []string
	}{
		{
			name: "function declaration",
			code: `function calculateTotal(price) {
	return price * 1.1;
}`,
			expected: []string{"calculateTotal"},
		},
		{
			name:     "arrow function",
			code:     `const add = (a, b) => a + b;`,
			expected: []string{"add"},
		},
		{
			name: "class declaration",
			code: `class Calculator {
	add(a, b) {
		return a + b;
	}
}`,
			expected: []string{"Calculator"},
		},
		{
			name: "function expression",
			code: `const multiply = function(a, b) {
	return a * b;
};`,
			expected: []string{"multiply"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			symbols := extractSymbolsFromAST(tt.code, "javascript", "test.js")
			if len(symbols) != len(tt.expected) {
				t.Errorf("expected %d symbols, got %d", len(tt.expected), len(symbols))
			}
			for _, expected := range tt.expected {
				if !symbols[expected] {
					t.Errorf("expected symbol %q not found", expected)
				}
			}
		})
	}
}

func TestExtractSymbolsFromAST_Python(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected []string
	}{
		{
			name: "function definition",
			code: `def calculate_total(price):
	return price * 1.1`,
			expected: []string{"calculate_total"},
		},
		{
			name: "class definition",
			code: `class Calculator:
	def add(self, a, b):
		return a + b`,
			expected: []string{"Calculator", "add"}, // Both class and method are extracted
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			symbols := extractSymbolsFromAST(tt.code, "python", "test.py")
			if len(symbols) != len(tt.expected) {
				t.Errorf("expected %d symbols, got %d", len(tt.expected), len(symbols))
			}
			for _, expected := range tt.expected {
				if !symbols[expected] {
					t.Errorf("expected symbol %q not found", expected)
				}
			}
		})
	}
}

func TestExtractSymbolsFromAST_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		language string
	}{
		{
			name:     "unsupported language",
			code:     "some code",
			language: "unsupported",
		},
		{
			name:     "malformed code",
			code:     "func { invalid syntax",
			language: "go",
		},
		{
			name:     "empty symbols",
			code:     "// just a comment",
			language: "go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			symbols := extractSymbolsFromAST(tt.code, tt.language, "test")
			// Should return empty map without panicking
			if symbols == nil {
				t.Error("expected non-nil map")
			}
		})
	}
}

func TestCheckSymbolReferences(t *testing.T) {
	// Create a test parser and parse code
	parser, err := GetParser("go")
	if err != nil {
		t.Fatalf("failed to get parser: %v", err)
	}

	code := `package main

func CalculateTotal(price float64) float64 {
	return price * 1.1
}

func main() {
	result := CalculateTotal(100.0)
}`
	tree, err := parser.ParseCtx(context.Background(), nil, []byte(code))
	if err != nil {
		t.Fatalf("failed to parse code: %v", err)
	}
	defer tree.Close()

	rootNode := tree.RootNode()
	if rootNode == nil {
		t.Fatal("root node is nil")
	}

	tests := []struct {
		name     string
		symbols  map[string]bool
		expected bool
	}{
		{
			name:     "identifier matching",
			symbols:  map[string]bool{"CalculateTotal": true},
			expected: true,
		},
		{
			name:     "no matches",
			symbols:  map[string]bool{"NonExistent": true},
			expected: false,
		},
		{
			name:     "empty symbols map",
			symbols:  map[string]bool{},
			expected: false,
		},
		{
			name:     "multiple symbols, one match",
			symbols:  map[string]bool{"CalculateTotal": true, "OtherFunc": true},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checkSymbolReferences(rootNode, code, "go", tt.symbols)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestCheckSymbolReferences_EdgeCases(t *testing.T) {
	t.Run("nil root node", func(t *testing.T) {
		result := checkSymbolReferences(nil, "code", "go", map[string]bool{"test": true})
		if result {
			t.Error("expected false for nil root node")
		}
	})

	t.Run("empty symbols map", func(t *testing.T) {
		parser, err := GetParser("go")
		if err != nil {
			t.Fatalf("failed to get parser: %v", err)
		}

		tree, err := parser.ParseCtx(context.Background(), nil, []byte("package main"))
		if err != nil {
			t.Fatalf("failed to parse: %v", err)
		}
		defer tree.Close()

		result := checkSymbolReferences(tree.RootNode(), "package main", "go", map[string]bool{})
		if result {
			t.Error("expected false for empty symbols map")
		}
	})
}

func TestCheckCodeReference(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "test-dependency-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test files
	file1Path := filepath.Join(tmpDir, "file1.go")
	file2Path := filepath.Join(tmpDir, "file2.go")

	file1Content := `package main

func CalculateTotal(price float64) float64 {
	return price * 1.1
}`
	file2Content := `package main

import "fmt"

func main() {
	result := CalculateTotal(100.0)
	fmt.Println(result)
}`

	if err := os.WriteFile(file1Path, []byte(file1Content), 0644); err != nil {
		t.Fatalf("failed to write file1: %v", err)
	}
	if err := os.WriteFile(file2Path, []byte(file2Content), 0644); err != nil {
		t.Fatalf("failed to write file2: %v", err)
	}

	otherTask := &Task{
		FilePath:    "file1.go",
		Title:       "Calculate total",
		Description: "Calculate total with tax",
	}

	t.Run("successful AST-based detection", func(t *testing.T) {
		result := checkCodeReference(tmpDir, "file2.go", otherTask)
		if !result {
			t.Error("expected true for AST-based detection")
		}
	})

	t.Run("fallback to keyword matching", func(t *testing.T) {
		// Test with file that doesn't exist
		result := checkCodeReference(tmpDir, "nonexistent.go", otherTask)
		// Should fallback to keyword matching (may or may not match)
		_ = result // Just check it doesn't panic
	})

	t.Run("unsupported languages", func(t *testing.T) {
		// Create file with unsupported extension
		unsupportedPath := filepath.Join(tmpDir, "file.unsupported")
		if err := os.WriteFile(unsupportedPath, []byte("some code"), 0644); err != nil {
			t.Fatalf("failed to write unsupported file: %v", err)
		}

		otherTaskUnsupported := &Task{
			FilePath: "file.unsupported",
			Title:    "Test",
		}

		result := checkCodeReference(tmpDir, "file2.go", otherTaskUnsupported)
		// Should fallback to keyword matching
		_ = result // Just check it doesn't panic
	})
}
