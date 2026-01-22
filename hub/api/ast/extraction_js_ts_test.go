// Package ast - JavaScript and TypeScript function extraction tests
// Complies with CODING_STANDARDS.md: Tests max 500 lines
package ast

import (
	"testing"
)

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

	t.Run("extract_arrow_function", func(t *testing.T) {
		// Given
		code := `const calculate = (x, y) => x + y;
const handler = async () => {};`
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
		if functions[0].Name != "calculate" {
			t.Errorf("Expected function name 'calculate', got '%s'", functions[0].Name)
		}
	})

	t.Run("extract_arrow_function_async", func(t *testing.T) {
		// Given
		code := `const handler = async () => {
	return true;
};`
		keyword := "handler"

		// When
		functions, err := ExtractFunctions(code, "javascript", keyword)

		// Then
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		if len(functions) == 0 {
			t.Fatal("Expected at least one function, got none")
		}
		if functions[0].Name != "handler" {
			t.Errorf("Expected function name 'handler', got '%s'", functions[0].Name)
		}
	})

	t.Run("extract_function_expression", func(t *testing.T) {
		// Given
		code := `const func = function() {
	return true;
};`
		keyword := "func"

		// When
		functions, err := ExtractFunctions(code, "javascript", keyword)

		// Then
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		if len(functions) == 0 {
			t.Fatal("Expected at least one function, got none")
		}
		if functions[0].Name != "func" {
			t.Errorf("Expected function name 'func', got '%s'", functions[0].Name)
		}
	})

	t.Run("extract_named_function_expression", func(t *testing.T) {
		// Given
		code := `const named = function namedFunc() {
	return true;
};`
		keyword := "named"

		// When
		functions, err := ExtractFunctions(code, "javascript", keyword)

		// Then
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		if len(functions) == 0 {
			t.Fatal("Expected at least one function, got none")
		}
		// Should extract the variable name, not the function name
		if functions[0].Name != "named" {
			t.Errorf("Expected function name 'named', got '%s'", functions[0].Name)
		}
	})

	t.Run("extract_arrow_function_in_object", func(t *testing.T) {
		// Given
		code := `const obj = {
	method: () => {
		return true;
	}
};`
		keyword := "method"

		// When
		functions, err := ExtractFunctions(code, "javascript", keyword)

		// Then
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		if len(functions) == 0 {
			t.Fatal("Expected at least one function, got none")
		}
		if functions[0].Name != "method" {
			t.Errorf("Expected function name 'method', got '%s'", functions[0].Name)
		}
	})

	t.Run("extract_class_method", func(t *testing.T) {
		// Given
		code := `class MyClass {
	method() {
		return true;
	}
}`
		keyword := "method"

		// When
		functions, err := ExtractFunctions(code, "javascript", keyword)

		// Then
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		if len(functions) == 0 {
			t.Fatal("Expected at least one function, got none")
		}
		if functions[0].Name != "method" {
			t.Errorf("Expected function name 'method', got '%s'", functions[0].Name)
		}
	})

	t.Run("extract_static_method", func(t *testing.T) {
		// Given
		code := `class MyClass {
	static staticMethod() {
		return true;
	}
}`
		keyword := "staticMethod"

		// When
		functions, err := ExtractFunctions(code, "javascript", keyword)

		// Then
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		if len(functions) == 0 {
			t.Fatal("Expected at least one function, got none")
		}
		if functions[0].Name != "staticMethod" {
			t.Errorf("Expected function name 'staticMethod', got '%s'", functions[0].Name)
		}
	})

	t.Run("extract_private_method", func(t *testing.T) {
		// Given
		code := `class MyClass {
	#privateMethod() {
		return true;
	}
}`
		keyword := "privateMethod"

		// When
		functions, err := ExtractFunctions(code, "javascript", keyword)

		// Then
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		if len(functions) == 0 {
			t.Fatal("Expected at least one function, got none")
		}
		if functions[0].Name != "#privateMethod" {
			t.Errorf("Expected function name '#privateMethod', got '%s'", functions[0].Name)
		}
		if functions[0].Visibility != "private" {
			t.Errorf("Expected visibility 'private', got '%s'", functions[0].Visibility)
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

	t.Run("extract_arrow_function_typed", func(t *testing.T) {
		// Given
		code := `const calculate = (x: number, y: number): number => x + y;`
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
		if functions[0].Name != "calculate" {
			t.Errorf("Expected function name 'calculate', got '%s'", functions[0].Name)
		}
	})

	t.Run("extract_arrow_function_async_typed", func(t *testing.T) {
		// Given
		code := `const handler = async (): Promise<void> => {
	return;
};`
		keyword := "handler"

		// When
		functions, err := ExtractFunctions(code, "typescript", keyword)

		// Then
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		if len(functions) == 0 {
			t.Fatal("Expected at least one function, got none")
		}
		if functions[0].Name != "handler" {
			t.Errorf("Expected function name 'handler', got '%s'", functions[0].Name)
		}
	})

	t.Run("extract_class_method_typed", func(t *testing.T) {
		// Given
		code := `class MyClass {
	method(): number {
		return 42;
	}
}`
		keyword := "method"

		// When
		functions, err := ExtractFunctions(code, "typescript", keyword)

		// Then
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		if len(functions) == 0 {
			t.Fatal("Expected at least one function, got none")
		}
		if functions[0].Name != "method" {
			t.Errorf("Expected function name 'method', got '%s'", functions[0].Name)
		}
	})
}
