// Package ast - Symbol table operations tests
// Tests for scope stack and symbol table operations
// Complies with CODING_STANDARDS.md: Tests max 500 lines
package ast

import (
	"testing"
)

// TestScopeStack tests scope stack functionality
func TestScopeStack(t *testing.T) {
	t.Run("NewScopeStack", func(t *testing.T) {
		stack := NewScopeStack()
		if stack == nil {
			t.Fatal("NewScopeStack returned nil")
		}

		if stack.Current == nil {
			t.Fatal("Current scope is nil")
		}

		if stack.Current.Symbols == nil {
			t.Fatal("Symbols map is nil")
		}
	})

	t.Run("Push", func(t *testing.T) {
		stack := NewScopeStack()
		initialScope := stack.Current

		stack.Push("function", 0, 100)
		if stack.Current == initialScope {
			t.Error("Push did not create new scope")
		}

		if stack.Current.Name != "function" {
			t.Errorf("Expected scope name 'function', got '%s'", stack.Current.Name)
		}

		if stack.Current.Parent != initialScope {
			t.Error("New scope parent is not initial scope")
		}
	})

	t.Run("Pop", func(t *testing.T) {
		stack := NewScopeStack()
		rootScope := stack.Current

		stack.Push("function", 0, 100)
		functionScope := stack.Current

		stack.Push("block", 10, 50)
		_ = stack.Current // blockScope

		// Pop should return to function scope
		stack.Pop()
		if stack.Current != functionScope {
			t.Error("Pop did not return to function scope")
		}

		// Pop should return to root scope
		stack.Pop()
		if stack.Current != rootScope {
			t.Error("Pop did not return to root scope")
		}

		// Pop from root should not change (no parent)
		stack.Pop()
		if stack.Current != rootScope {
			t.Error("Pop from root should not change scope")
		}
	})

	t.Run("PushPopSequence", func(t *testing.T) {
		stack := NewScopeStack()
		rootScope := stack.Current

		// Push multiple scopes
		stack.Push("level1", 0, 100)
		stack.Push("level2", 10, 50)
		stack.Push("level3", 20, 30)

		if stack.Current.Name != "level3" {
			t.Errorf("Expected 'level3', got '%s'", stack.Current.Name)
		}

		// Pop back to root
		stack.Pop()
		stack.Pop()
		stack.Pop()

		if stack.Current != rootScope {
			t.Error("After popping all, should be at root scope")
		}
	})
}

// TestSymbolTable_AddReference tests adding references to symbol table
func TestSymbolTable_AddReference(t *testing.T) {
	st := NewSymbolTable()

	t.Run("valid_reference", func(t *testing.T) {
		ref := &SymbolReference{
			Name:     "myFunc",
			FilePath: "file1.js",
			Line:     10,
			Column:   5,
			Kind:     "call",
		}

		st.AddReference(ref)

		refs := st.GetReferences("myFunc")
		if len(refs) == 0 {
			t.Error("Expected to find reference")
		}

		if refs[0].Name != "myFunc" {
			t.Errorf("Expected name 'myFunc', got '%s'", refs[0].Name)
		}
	})

	t.Run("nil_reference", func(t *testing.T) {
		st.AddReference(nil)
		// Should not panic
	})

	t.Run("empty_name", func(t *testing.T) {
		ref := &SymbolReference{
			Name: "",
		}
		st.AddReference(ref)
		// Should not panic
	})
}

// TestSymbolTable_GetFileSymbols tests getting symbols by file
func TestSymbolTable_GetFileSymbols(t *testing.T) {
	st := NewSymbolTable()

	symbol1 := &FileSymbol{
		Name:     "func1",
		Kind:     "function",
		FilePath: "file1.js",
		Line:     1,
		Column:   1,
		Language: "javascript",
	}

	symbol2 := &FileSymbol{
		Name:     "func2",
		Kind:     "function",
		FilePath: "file1.js",
		Line:     10,
		Column:   1,
		Language: "javascript",
	}

	symbol3 := &FileSymbol{
		Name:     "func3",
		Kind:     "function",
		FilePath: "file2.js",
		Line:     1,
		Column:   1,
		Language: "javascript",
	}

	if err := st.AddSymbol(symbol1); err != nil {
		t.Fatalf("AddSymbol failed: %v", err)
	}
	if err := st.AddSymbol(symbol2); err != nil {
		t.Fatalf("AddSymbol failed: %v", err)
	}
	if err := st.AddSymbol(symbol3); err != nil {
		t.Fatalf("AddSymbol failed: %v", err)
	}

	fileSymbols := st.GetFileSymbols("file1.js")
	if len(fileSymbols) != 2 {
		t.Errorf("Expected 2 symbols in file1.js, got %d", len(fileSymbols))
	}

	fileSymbols2 := st.GetFileSymbols("file2.js")
	if len(fileSymbols2) != 1 {
		t.Errorf("Expected 1 symbol in file2.js, got %d", len(fileSymbols2))
	}

	fileSymbols3 := st.GetFileSymbols("nonexistent.js")
	if len(fileSymbols3) != 0 {
		t.Errorf("Expected 0 symbols in nonexistent file, got %d", len(fileSymbols3))
	}
}
