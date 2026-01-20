// Package ast provides utility function tests
// Complies with CODING_STANDARDS.md: Tests max 500 lines
package ast

import (
	"testing"
)

func TestDetectLanguage_FileExtension(t *testing.T) {
	t.Run("go_extension", func(t *testing.T) {
		// Given
		code := `package main`
		filePath := "main.go"

		// When
		language := DetectLanguage(code, filePath)

		// Then
		if language != "go" {
			t.Errorf("Expected language 'go', got '%s'", language)
		}
	})

	t.Run("javascript_extension", func(t *testing.T) {
		// Given
		code := `console.log('test');`
		filePath := "script.js"

		// When
		language := DetectLanguage(code, filePath)

		// Then
		if language != "javascript" {
			t.Errorf("Expected language 'javascript', got '%s'", language)
		}
	})

	t.Run("jsx_extension", func(t *testing.T) {
		// Given
		code := `const Component = () => <div>Test</div>;`
		filePath := "Component.jsx"

		// When
		language := DetectLanguage(code, filePath)

		// Then
		if language != "javascript" {
			t.Errorf("Expected language 'javascript', got '%s'", language)
		}
	})

	t.Run("typescript_extension", func(t *testing.T) {
		// Given
		code := `const x: number = 5;`
		filePath := "script.ts"

		// When
		language := DetectLanguage(code, filePath)

		// Then
		if language != "typescript" {
			t.Errorf("Expected language 'typescript', got '%s'", language)
		}
	})

	t.Run("tsx_extension", func(t *testing.T) {
		// Given
		code := `const Component: React.FC = () => <div>Test</div>;`
		filePath := "Component.tsx"

		// When
		language := DetectLanguage(code, filePath)

		// Then
		if language != "typescript" {
			t.Errorf("Expected language 'typescript', got '%s'", language)
		}
	})

	t.Run("python_extension", func(t *testing.T) {
		// Given
		code := `print("test")`
		filePath := "script.py"

		// When
		language := DetectLanguage(code, filePath)

		// Then
		if language != "python" {
			t.Errorf("Expected language 'python', got '%s'", language)
		}
	})
}

func TestDetectLanguage_CodePatterns(t *testing.T) {
	t.Run("go_package_keyword", func(t *testing.T) {
		// Given
		code := `package main

func main() {}`
		filePath := ""

		// When
		language := DetectLanguage(code, filePath)

		// Then
		if language != "go" {
			t.Errorf("Expected language 'go', got '%s'", language)
		}
	})

	t.Run("go_func_keyword", func(t *testing.T) {
		// Given
		code := `func test() {}`
		filePath := ""

		// When
		language := DetectLanguage(code, filePath)

		// Then
		if language != "go" {
			t.Errorf("Expected language 'go', got '%s'", language)
		}
	})

	t.Run("javascript_function", func(t *testing.T) {
		// Given
		code := `function test() {}`
		filePath := ""

		// When
		language := DetectLanguage(code, filePath)

		// Then
		if language != "javascript" {
			t.Errorf("Expected language 'javascript', got '%s'", language)
		}
	})

	t.Run("javascript_const", func(t *testing.T) {
		// Given
		code := `const x = 5;`
		filePath := ""

		// When
		language := DetectLanguage(code, filePath)

		// Then
		if language != "javascript" {
			t.Errorf("Expected language 'javascript', got '%s'", language)
		}
	})

	t.Run("javascript_arrow", func(t *testing.T) {
		// Given
		code := `const test = () => {};`
		filePath := ""

		// When
		language := DetectLanguage(code, filePath)

		// Then
		if language != "javascript" {
			t.Errorf("Expected language 'javascript', got '%s'", language)
		}
	})

	t.Run("typescript_interface", func(t *testing.T) {
		// Given
		code := `interface Test {
	name: string;
}`
		filePath := ""

		// When
		language := DetectLanguage(code, filePath)

		// Then
		if language != "typescript" {
			t.Errorf("Expected language 'typescript', got '%s'", language)
		}
	})

	t.Run("typescript_type", func(t *testing.T) {
		// Given
		code := `type Test = string;`
		filePath := ""

		// When
		language := DetectLanguage(code, filePath)

		// Then
		if language != "typescript" {
			t.Errorf("Expected language 'typescript', got '%s'", language)
		}
	})

	t.Run("python_def", func(t *testing.T) {
		// Given
		code := `def test():
	pass`
		filePath := ""

		// When
		language := DetectLanguage(code, filePath)

		// Then
		if language != "python" {
			t.Errorf("Expected language 'python', got '%s'", language)
		}
	})

	t.Run("python_import", func(t *testing.T) {
		// Given
		code := `import os
print("test")`
		filePath := ""

		// When
		language := DetectLanguage(code, filePath)

		// Then
		if language != "python" {
			t.Errorf("Expected language 'python', got '%s'", language)
		}
	})

	t.Run("python_shebang", func(t *testing.T) {
		// Given
		code := `#!/usr/bin/env python
print("test")`
		filePath := ""

		// When
		language := DetectLanguage(code, filePath)

		// Then
		if language != "python" {
			t.Errorf("Expected language 'python', got '%s'", language)
		}
	})
}

func TestDetectLanguage_Unknown(t *testing.T) {
	t.Run("empty_code", func(t *testing.T) {
		// Given
		code := ``
		filePath := ""

		// When
		language := DetectLanguage(code, filePath)

		// Then
		if language != "unknown" {
			t.Errorf("Expected language 'unknown', got '%s'", language)
		}
	})

	t.Run("unrecognized_pattern", func(t *testing.T) {
		// Given
		code := `some random text that doesn't match any pattern`
		filePath := ""

		// When
		language := DetectLanguage(code, filePath)

		// Then
		if language != "unknown" {
			t.Errorf("Expected language 'unknown', got '%s'", language)
		}
	})
}

func TestDetectLanguage_Priority(t *testing.T) {
	t.Run("file_extension_overrides_code", func(t *testing.T) {
		// Given
		// Code looks like Python but file extension is .go
		code := `def test():
	pass`
		filePath := "test.go"

		// When
		language := DetectLanguage(code, filePath)

		// Then
		// File extension should take priority
		if language != "go" {
			t.Errorf("Expected language 'go' (from file extension), got '%s'", language)
		}
	})
}
