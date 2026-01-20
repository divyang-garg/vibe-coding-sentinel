// Package ast - Cross-file analysis tests
// Complies with CODING_STANDARDS.md: Tests max 500 lines
package ast

import (
	"context"
	"testing"
)

func TestAnalyzeCrossFile(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		files := []FileInput{
			{
				Path:     "file1.go",
				Content:  "package main\nfunc ExportedFunc() {}\n",
				Language: "go",
			},
			{
				Path:     "file2.go",
				Content:  "package main\nfunc test() { ExportedFunc() }\n",
				Language: "go",
			},
		}

		result, err := AnalyzeCrossFile(ctx, files, []string{})
		if err != nil {
			t.Fatalf("AnalyzeCrossFile failed: %v", err)
		}
		if result == nil {
			t.Fatal("AnalyzeCrossFile returned nil result")
		}
	})

	t.Run("empty_files", func(t *testing.T) {
		_, err := AnalyzeCrossFile(ctx, []FileInput{}, []string{})
		if err == nil {
			t.Error("Expected error for empty files")
		}
	})
}

func TestSymbolTable(t *testing.T) {
	t.Run("add_and_get_symbol", func(t *testing.T) {
		st := NewSymbolTable()
		symbol := &FileSymbol{
			Name:     "testFunc",
			Kind:     "function",
			FilePath: "test.go",
			Line:     1,
			Column:   1,
			Exported: true,
		}

		err := st.AddSymbol(symbol)
		if err != nil {
			t.Fatalf("AddSymbol failed: %v", err)
		}

		symbols := st.GetSymbols("testFunc")
		if len(symbols) == 0 {
			t.Error("Expected to find symbol")
		}
	})

	t.Run("find_unused_exports", func(t *testing.T) {
		st := NewSymbolTable()
		symbol := &FileSymbol{
			Name:     "unusedFunc",
			Kind:     "function",
			FilePath: "test.go",
			Line:     1,
			Column:   1,
			Exported: true,
		}
		st.AddSymbol(symbol)

		unused := st.FindUnusedExports()
		if len(unused) == 0 {
			t.Error("Expected to find unused export")
		}
	})
}

func TestDependencyGraph(t *testing.T) {
	t.Run("add_and_get_dependency", func(t *testing.T) {
		dg := NewDependencyGraph()
		dep := &Dependency{
			FromFile: "file1.go",
			ToFile:   "file2.go",
			Type:     "import",
			Line:     1,
			Column:   1,
		}

		err := dg.AddDependency(dep)
		if err != nil {
			t.Fatalf("AddDependency failed: %v", err)
		}

		deps := dg.GetDependencies("file1.go")
		if len(deps) == 0 {
			t.Error("Expected to find dependency")
		}
	})

	t.Run("circular_dependency", func(t *testing.T) {
		dg := NewDependencyGraph()
		dep1 := &Dependency{
			FromFile: "file1.go",
			ToFile:   "file2.go",
			Type:     "import",
		}
		dep2 := &Dependency{
			FromFile: "file2.go",
			ToFile:   "file1.go",
			Type:     "import",
		}

		dg.AddDependency(dep1)
		dg.AddDependency(dep2)

		cycles := dg.FindCircularDependencies()
		if len(cycles) == 0 {
			t.Error("Expected to find circular dependency")
		}
	})
}
