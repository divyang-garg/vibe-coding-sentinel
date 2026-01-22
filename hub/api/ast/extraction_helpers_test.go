// Package ast - Helper function extraction tests (error handling, visibility, parameters, return types, documentation)
// Complies with CODING_STANDARDS.md: Tests max 500 lines
package ast

import (
	"testing"
)

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

func TestExtractFunctions_Parameters(t *testing.T) {
	t.Run("go_parameters", func(t *testing.T) {
		// Given
		code := `package main

func calculate(x int, y float64) float64 {
	return float64(x) * y
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
		if len(functions[0].Parameters) != 2 {
			t.Errorf("Expected 2 parameters, got %d", len(functions[0].Parameters))
		}
		if functions[0].Parameters[0].Name != "x" {
			t.Errorf("Expected first parameter name 'x', got '%s'", functions[0].Parameters[0].Name)
		}
		if functions[0].Parameters[0].Type != "int" {
			t.Errorf("Expected first parameter type 'int', got '%s'", functions[0].Parameters[0].Type)
		}
	})

	t.Run("javascript_parameters", func(t *testing.T) {
		// Given
		code := `function calculate(x, y) {
	return x * y;
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
		if len(functions[0].Parameters) != 2 {
			t.Errorf("Expected 2 parameters, got %d", len(functions[0].Parameters))
		}
		if functions[0].Parameters[0].Name != "x" {
			t.Errorf("Expected first parameter name 'x', got '%s'", functions[0].Parameters[0].Name)
		}
	})

	t.Run("typescript_typed_parameters", func(t *testing.T) {
		// Given
		code := `function calculate(x: number, y: number): number {
	return x * y;
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
		if len(functions[0].Parameters) != 2 {
			t.Errorf("Expected 2 parameters, got %d", len(functions[0].Parameters))
		}
		if functions[0].Parameters[0].Type == "" {
			t.Error("Expected parameter type to be extracted for TypeScript")
		}
	})

	t.Run("python_parameters", func(t *testing.T) {
		// Given
		code := `def calculate(x, y):
	return x * y`
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
		if len(functions[0].Parameters) != 2 {
			t.Errorf("Expected 2 parameters, got %d", len(functions[0].Parameters))
		}
		if functions[0].Parameters[0].Name != "x" {
			t.Errorf("Expected first parameter name 'x', got '%s'", functions[0].Parameters[0].Name)
		}
	})

	t.Run("go_no_parameters", func(t *testing.T) {
		// Given
		code := `package main

func main() {
}`
		keyword := "main"

		// When
		functions, err := ExtractFunctions(code, "go", keyword)

		// Then
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		if len(functions) == 0 {
			t.Fatal("Expected at least one function, got none")
		}
		if len(functions[0].Parameters) != 0 {
			t.Errorf("Expected 0 parameters, got %d", len(functions[0].Parameters))
		}
	})
}

func TestExtractFunctions_ReturnTypes(t *testing.T) {
	t.Run("go_return_type", func(t *testing.T) {
		// Given
		code := `package main

func calculate(x int) float64 {
	return float64(x)
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
		if functions[0].ReturnType != "float64" {
			t.Errorf("Expected return type 'float64', got '%s'", functions[0].ReturnType)
		}
	})

	t.Run("typescript_return_type", func(t *testing.T) {
		// Given
		code := `function calculate(x: number): number {
	return x * 2;
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
		if functions[0].ReturnType == "" {
			t.Error("Expected return type to be extracted for TypeScript")
		}
	})

	t.Run("python_return_type", func(t *testing.T) {
		// Given
		code := `def calculate(x: int) -> int:
	return x * 2`
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
		if functions[0].ReturnType == "" {
			t.Error("Expected return type to be extracted for Python")
		}
	})

	t.Run("go_no_return_type", func(t *testing.T) {
		// Given
		code := `package main

func main() {
}`
		keyword := "main"

		// When
		functions, err := ExtractFunctions(code, "go", keyword)

		// Then
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		if len(functions) == 0 {
			t.Fatal("Expected at least one function, got none")
		}
		if functions[0].ReturnType != "" {
			t.Errorf("Expected empty return type, got '%s'", functions[0].ReturnType)
		}
	})
}

func TestExtractFunctions_Documentation(t *testing.T) {
	t.Run("go_doc_comment", func(t *testing.T) {
		// Given
		code := `package main

// calculate performs calculation
func calculate(x int) int {
	return x * 2
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
		if functions[0].Documentation == "" {
			t.Error("Expected documentation to be extracted for Go function")
		}
	})

	t.Run("python_docstring", func(t *testing.T) {
		// Given
		code := `def calculate(x):
	"""Calculate and return result"""
	return x * 2`
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
		if functions[0].Documentation == "" {
			t.Error("Expected documentation to be extracted for Python function")
		}
	})

	t.Run("javascript_jsdoc", func(t *testing.T) {
		// Given
		code := `/**
 * Calculate and return result
 * @param {number} x - Input value
 * @returns {number} Result
 */
function calculate(x) {
	return x * 2;
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
		if functions[0].Documentation == "" {
			t.Error("Expected documentation to be extracted for JavaScript function")
		}
	})
}
