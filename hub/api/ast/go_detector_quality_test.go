// Package ast provides tests for Go detector code quality methods
// Complies with CODING_STANDARDS.md: Tests max 500 lines
package ast

import "testing"

// Test GoDetector_DetectUnused
func TestGoDetector_DetectUnused(t *testing.T) {
	detector := &GoDetector{}

	tests := []struct {
		name     string
		code     string
		expected int // Minimum expected findings
	}{
		{
			name: "Unused variable",
			code: `
package main

func test() {
	var unused int = 42
	var used int = 10
	println(used)
}
`,
			expected: 1,
		},
		{
			name: "Unused short variable declaration",
			code: `
package main

func test() {
	unused := 42
	used := 10
	println(used)
}
`,
			expected: 1,
		},
		{
			name: "All variables used",
			code: `
package main

func test() {
	var a int = 1
	var b int = 2
	println(a + b)
}
`,
			expected: 0,
		},
		{
			name: "Multiple unused variables",
			code: `
package main

func test() {
	var unused1 int = 1
	var unused2 int = 2
	var used int = 3
	println(used)
}
`,
			expected: 2,
		},
		{
			name: "No variables",
			code: `
package main

func test() {
	println("no variables")
}
`,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree, rootNode := parseGoCode(t, tt.code)
			defer tree.Close()
			findings := detector.DetectUnused(rootNode, tt.code)

			if len(findings) < tt.expected {
				t.Errorf("Expected at least %d findings, got %d", tt.expected, len(findings))
			}

			// Verify findings have required fields
			for _, finding := range findings {
				if finding.Type != "unused_variable" {
					t.Errorf("Expected finding type 'unused_variable', got '%s'", finding.Type)
				}
				if finding.Message == "" {
					t.Error("Finding message should not be empty")
				}
			}
		})
	}
}

// Test GoDetector_DetectDuplicates
func TestGoDetector_DetectDuplicates(t *testing.T) {
	detector := &GoDetector{}

	tests := []struct {
		name     string
		code     string
		expected int // Minimum expected findings
	}{
		{
			name: "Duplicate function",
			code: `
package main

func test() {
	return 1
}

func test() {
	return 2
}
`,
			expected: 2,
		},
		{
			name: "No duplicates",
			code: `
package main

func func1() {
	return 1
}

func func2() {
	return 2
}
`,
			expected: 0,
		},
		{
			name: "Multiple duplicates",
			code: `
package main

func duplicate() {
	return 1
}

func duplicate() {
	return 2
}

func duplicate() {
	return 3
}
`,
			expected: 3,
		},
		{
			name: "No functions",
			code: `
package main

var x = 1
`,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree, rootNode := parseGoCode(t, tt.code)
			defer tree.Close()
			findings := detector.DetectDuplicates(rootNode, tt.code)

			if len(findings) < tt.expected {
				t.Errorf("Expected at least %d findings, got %d", tt.expected, len(findings))
			}

			// Verify findings have required fields
			for _, finding := range findings {
				if finding.Type != "duplicate_function" {
					t.Errorf("Expected finding type 'duplicate_function', got '%s'", finding.Type)
				}
			}
		})
	}
}

// Test GoDetector_DetectUnreachable
func TestGoDetector_DetectUnreachable(t *testing.T) {
	detector := &GoDetector{}

	code := `
package main

func test() {
	return
	println("unreachable")
}
`

	tree, rootNode := parseGoCode(t, code)
	defer tree.Close()
	findings := detector.DetectUnreachable(rootNode, code)

	// Go detector delegates to detectUnreachableCodeGo; code after return is unreachable
	if len(findings) < 1 {
		t.Errorf("Expected at least 1 unreachable_code finding, got %d", len(findings))
	}
	var found bool
	for _, f := range findings {
		if f.Type == "unreachable_code" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected finding type 'unreachable_code'")
	}
}

// Test GoDetector_DetectAsync
func TestGoDetector_DetectAsync(t *testing.T) {
	detector := &GoDetector{}

	code := `
package main

func test() {
	// Async not applicable to Go
}
`

	tree, rootNode := parseGoCode(t, code)
	defer tree.Close()
	findings := detector.DetectAsync(rootNode, code)

	// Should return empty (not applicable to Go)
	if len(findings) != 0 {
		t.Errorf("Expected 0 findings (not applicable), got %d", len(findings))
	}
}

// Test GoDetector_RegistryIntegration
func TestGoDetector_RegistryIntegration(t *testing.T) {
	// Verify Go detector is available through registry
	detector := GetLanguageDetector("go")
	if detector == nil {
		t.Fatal("Go detector should be available through registry")
	}

	// Verify it's a GoDetector
	goDetector, ok := detector.(*GoDetector)
	if !ok {
		t.Fatal("Detector should be *GoDetector")
	}

	// Test that it works
	code := `
package main

import "net/http"

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if strings.HasPrefix(auth, "Bearer ") {
			// JWT
		}
		next.ServeHTTP(w, r)
	})
}
`

	tree, rootNode := parseGoCode(t, code)
	defer tree.Close()
	findings := goDetector.DetectSecurityMiddleware(rootNode, code)

	if len(findings) == 0 {
		t.Error("Expected at least one finding for JWT middleware")
	}
}
