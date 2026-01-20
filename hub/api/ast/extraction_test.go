// Package ast provides function extraction tests
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

func TestExtractFunctions_JavaScript(t *testing.T) {
	t.Run("extract_function_declaration", func(t *testing.T) {
		// Given
		code := `function calculateTotal(price, quantity) {
	return price * quantity;
}

function processOrder(orderId) {
	return orderId;
}`
		keyword := "calculate"

		// When
		functions, err := ExtractFunctions(code, "javascript", keyword)

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
		if functions[0].Language != "javascript" {
			t.Errorf("Expected language 'javascript', got '%s'", functions[0].Language)
		}
	})
}

func TestExtractFunctions_TypeScript(t *testing.T) {
	t.Run("extract_function_declaration", func(t *testing.T) {
		// Given
		code := `function calculateTotal(price: number, quantity: number): number {
	return price * quantity;
}`
		keyword := "calculate"

		// When
		functions, err := ExtractFunctions(code, "typescript", keyword)

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
		if functions[0].Language != "typescript" {
			t.Errorf("Expected language 'typescript', got '%s'", functions[0].Language)
		}
	})
}

func TestExtractFunctions_Python(t *testing.T) {
	t.Run("extract_function_definition", func(t *testing.T) {
		// Given
		code := `def calculate_total(price, quantity):
	return price * quantity

def process_order(order_id):
	return order_id`
		keyword := "calculate"

		// When
		functions, err := ExtractFunctions(code, "python", keyword)

		// Then
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		if len(functions) == 0 {
			t.Fatal("Expected at least one function, got none")
		}
		if functions[0].Name != "calculate_total" {
			t.Errorf("Expected function name 'calculate_total', got '%s'", functions[0].Name)
		}
		if functions[0].Language != "python" {
			t.Errorf("Expected language 'python', got '%s'", functions[0].Language)
		}
	})
}

func TestExtractFunctions_ErrorHandling(t *testing.T) {
	t.Run("unsupported_language", func(t *testing.T) {
		// Given
		code := `function test() {}`
		keyword := "test"

		// When
		functions, err := ExtractFunctions(code, "unsupported", keyword)

		// Then
		if err == nil {
			t.Fatal("Expected error for unsupported language, got nil")
		}
		if functions != nil {
			t.Errorf("Expected nil functions, got %v", functions)
		}
	})

	t.Run("invalid_code", func(t *testing.T) {
		// Given
		code := `invalid code syntax {{{{`
		keyword := "test"

		// When
		functions, err := ExtractFunctions(code, "go", keyword)

		// Then
		// Should return empty slice, not error, for backward compatibility
		if err != nil {
			t.Fatalf("Expected no error for invalid code, got: %v", err)
		}
		if len(functions) != 0 {
			t.Errorf("Expected empty functions slice, got %d functions", len(functions))
		}
	})

	t.Run("empty_code", func(t *testing.T) {
		// Given
		code := ``
		keyword := "test"

		// When
		functions, err := ExtractFunctions(code, "go", keyword)

		// Then
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		if len(functions) != 0 {
			t.Errorf("Expected empty functions slice, got %d functions", len(functions))
		}
	})
}

func TestExtractFunctionByName(t *testing.T) {
	t.Run("exact_match", func(t *testing.T) {
		// Given
		code := `package main

func calculateTotal() {}
func processOrder() {}`
		funcName := "calculateTotal"

		// When
		funcInfo, err := ExtractFunctionByName(code, "go", funcName)

		// Then
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		if funcInfo == nil {
			t.Fatal("Expected function info, got nil")
		}
		if funcInfo.Name != funcName {
			t.Errorf("Expected function name '%s', got '%s'", funcName, funcInfo.Name)
		}
	})

	t.Run("not_found", func(t *testing.T) {
		// Given
		code := `package main

func calculateTotal() {}`
		funcName := "nonExistent"

		// When
		funcInfo, err := ExtractFunctionByName(code, "go", funcName)

		// Then
		if err == nil {
			t.Fatal("Expected error for non-existent function, got nil")
		}
		if funcInfo != nil {
			t.Errorf("Expected nil function info, got %v", funcInfo)
		}
	})
}

func TestExtractFunctions_Visibility(t *testing.T) {
	t.Run("go_exported", func(t *testing.T) {
		// Given
		code := `package main

func ExportedFunction() {}
func privateFunction() {}`
		keyword := ""

		// When
		functions, err := ExtractFunctions(code, "go", keyword)

		// Then
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		if len(functions) != 2 {
			t.Fatalf("Expected 2 functions, got %d", len(functions))
		}

		// Find exported function
		var exported *FunctionInfo
		for i := range functions {
			if functions[i].Name == "ExportedFunction" {
				exported = &functions[i]
				break
			}
		}
		if exported == nil {
			t.Fatal("Expected to find ExportedFunction")
		}
		if exported.Visibility != "exported" {
			t.Errorf("Expected visibility 'exported', got '%s'", exported.Visibility)
		}

		// Find private function
		var private *FunctionInfo
		for i := range functions {
			if functions[i].Name == "privateFunction" {
				private = &functions[i]
				break
			}
		}
		if private == nil {
			t.Fatal("Expected to find privateFunction")
		}
		if private.Visibility != "private" {
			t.Errorf("Expected visibility 'private', got '%s'", private.Visibility)
		}
	})

	t.Run("python_private", func(t *testing.T) {
		// Given
		code := `def public_function():
	pass

def _private_function():
	pass`
		keyword := ""

		// When
		functions, err := ExtractFunctions(code, "python", keyword)

		// Then
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		if len(functions) != 2 {
			t.Fatalf("Expected 2 functions, got %d", len(functions))
		}

		// Find private function
		var private *FunctionInfo
		for i := range functions {
			if functions[i].Name == "_private_function" {
				private = &functions[i]
				break
			}
		}
		if private == nil {
			t.Fatal("Expected to find _private_function")
		}
		if private.Visibility != "private" {
			t.Errorf("Expected visibility 'private', got '%s'", private.Visibility)
		}
	})
}
