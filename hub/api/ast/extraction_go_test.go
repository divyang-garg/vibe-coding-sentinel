// Package ast - Go-specific function extraction tests
// Complies with CODING_STANDARDS.md: Tests max 500 lines
package ast

import (
	"testing"
)

func TestExtractFunctions_Go(t *testing.T) {
	t.Run("extract_simple_function", func(t *testing.T) {
		// Given
		code := `package main

func calculateTotal(price float64, quantity int) float64 {
	return price * float64(quantity)
}

func main() {
	calculateTotal(10.5, 2)
}`
		keyword := "calculate"

		// When
		functions, err := ExtractFunctions(code, "go", keyword)

		// Then
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		if len(functions) == 0 {
			t.Fatal("Expected at least one function, got none")
		}
		if functions[0].Name != "calculateTotal" {
			t.Errorf("Expected function name 'calculateTotal', got '%s'", functions[0].Name)
		}
		if functions[0].Language != "go" {
			t.Errorf("Expected language 'go', got '%s'", functions[0].Language)
		}
		if functions[0].Line != 3 {
			t.Errorf("Expected line 3, got %d", functions[0].Line)
		}
	})

	t.Run("extract_method", func(t *testing.T) {
		// Given
		code := `package main

type Calculator struct{}

func (c *Calculator) Add(a, b int) int {
	return a + b
}`
		keyword := "Add"

		// When
		functions, err := ExtractFunctions(code, "go", keyword)

		// Then
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		if len(functions) == 0 {
			t.Fatal("Expected at least one function, got none")
		}
		if functions[0].Name != "Add" {
			t.Errorf("Expected function name 'Add', got '%s'", functions[0].Name)
		}
	})

	t.Run("extract_all_functions_no_keyword", func(t *testing.T) {
		// Given
		code := `package main

func func1() {}
func func2() {}
func func3() {}`
		keyword := ""

		// When
		functions, err := ExtractFunctions(code, "go", keyword)

		// Then
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		if len(functions) != 3 {
			t.Errorf("Expected 3 functions, got %d", len(functions))
		}
	})

	t.Run("case_insensitive_keyword", func(t *testing.T) {
		// Given
		code := `package main

func CalculateTotal() {}
func calculatePrice() {}`
		keyword := "CALCULATE"

		// When
		functions, err := ExtractFunctions(code, "go", keyword)

		// Then
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		if len(functions) != 2 {
			t.Errorf("Expected 2 functions, got %d", len(functions))
		}
	})
}
