// Package ast - Python-specific function extraction tests
// Complies with CODING_STANDARDS.md: Tests max 500 lines
package ast

import (
	"strings"
	"testing"
)

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

func TestExtractFunctions_ClassMethods(t *testing.T) {
	t.Run("python_class_method", func(t *testing.T) {
		// Given
		code := `class MyClass:
	def method(self):
		pass`
		keyword := "method"

		// When
		functions, err := ExtractFunctions(code, "python", keyword)

		// Then
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		if len(functions) == 0 {
			t.Fatal("Expected at least one function, got none")
		}
		// Should be formatted as ClassName.methodName or just methodName
		if !strings.Contains(functions[0].Name, "method") {
			t.Errorf("Expected function name to contain 'method', got '%s'", functions[0].Name)
		}
	})

	t.Run("python_static_method", func(t *testing.T) {
		// Given
		code := `class MyClass:
	@staticmethod
	def static_method():
		pass`
		keyword := "static_method"

		// When
		functions, err := ExtractFunctions(code, "python", keyword)

		// Then
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		if len(functions) == 0 {
			t.Fatal("Expected at least one function, got none")
		}
		if !strings.Contains(functions[0].Name, "static_method") {
			t.Errorf("Expected function name to contain 'static_method', got '%s'", functions[0].Name)
		}
	})

	t.Run("python_private_method", func(t *testing.T) {
		// Given
		code := `class MyClass:
	def _private_method(self):
		pass`
		keyword := "_private_method"

		// When
		functions, err := ExtractFunctions(code, "python", keyword)

		// Then
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		if len(functions) == 0 {
			t.Fatal("Expected at least one function, got none")
		}
		if functions[0].Visibility != "private" {
			t.Errorf("Expected visibility 'private', got '%s'", functions[0].Visibility)
		}
	})
}
