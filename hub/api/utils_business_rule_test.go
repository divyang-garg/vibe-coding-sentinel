// Package main tests for business rule detection
// Complies with CODING_STANDARDS.md: Tests max 500 lines

package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectBusinessRuleImplementation_Basic(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Given
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "user_service.go")
		code := `
package main

func CreateUser(name string) error {
	return nil
}

func UpdateUser(id string, name string) error {
	return nil
}
`
		if err := os.WriteFile(testFile, []byte(code), 0644); err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}

		rule := KnowledgeItem{
			Title:   "User Management",
			Content: "Users can be created and updated",
		}

		// When
		evidence := detectBusinessRuleImplementation(rule, tmpDir)

		// Then
		if evidence.Confidence == 0.0 {
			t.Error("Expected non-zero confidence")
		}

		if len(evidence.Functions) == 0 {
			t.Error("Expected to find at least one function")
		}
	})
}

func TestDetectBusinessRuleImplementation_NoMatch(t *testing.T) {
	t.Run("unrelated_code", func(t *testing.T) {
		// Given
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "unrelated.go")
		code := `
package main

func DoSomething() {}
`
		if err := os.WriteFile(testFile, []byte(code), 0644); err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}

		rule := KnowledgeItem{
			Title:   "Payment Processing",
			Content: "Handle payment transactions",
		}

		// When
		evidence := detectBusinessRuleImplementation(rule, tmpDir)

		// Then
		// Should have low or zero confidence
		if evidence.Confidence > 0.5 {
			t.Errorf("Expected low confidence for unrelated code, got %.2f", evidence.Confidence)
		}
	})
}

func TestDetectBusinessRuleWithAST_Go(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Given
		code := `
package main

func ProcessOrder(orderID string) error {
	return nil
}

func CancelOrder(orderID string) error {
	return nil
}
`
		keywordMap := map[string]bool{"order": true, "process": true}
		keywords := []string{"order", "process"}

		// When
		evidence := detectBusinessRuleWithAST(code, "order_service.go", keywordMap, keywords)

		// Then
		if evidence.Confidence == 0.0 {
			t.Error("Expected non-zero confidence for matching function")
		}

		if len(evidence.Functions) == 0 {
			t.Error("Expected to find ProcessOrder function")
		}
	})
}

func TestDetectBusinessRuleWithAST_UnsupportedLanguage(t *testing.T) {
	t.Run("unsupported_language", func(t *testing.T) {
		// Given
		code := `some code`
		keywordMap := map[string]bool{"test": true}
		keywords := []string{"test"}

		// When
		evidence := detectBusinessRuleWithAST(code, "file.unknown", keywordMap, keywords)

		// Then
		// Should return empty evidence for unsupported language
		if evidence.Confidence != 0.0 {
			t.Error("Expected zero confidence for unsupported language")
		}
	})
}

func TestScanSourceFiles(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Given
		tmpDir := t.TempDir()
		testFiles := []struct {
			path string
			code string
		}{
			{"main.go", "package main"},
			{"user.go", "package main"},
			{"user_test.go", "package main"},        // Should be excluded
			{"vendor/package.go", "package vendor"}, // Should be excluded
			{"script.js", "console.log('test');"},
			{"test.py", "def test(): pass"},
		}

		for _, f := range testFiles {
			fullPath := filepath.Join(tmpDir, f.path)
			if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
				t.Fatalf("Failed to create directory: %v", err)
			}
			if err := os.WriteFile(fullPath, []byte(f.code), 0644); err != nil {
				t.Fatalf("Failed to write file: %v", err)
			}
		}

		// When
		files := scanSourceFiles(tmpDir)

		// Then
		// Should find main.go, user.go, script.js, test.py
		// Should exclude user_test.go and vendor/package.go
		if len(files) < 4 {
			t.Errorf("Expected at least 4 source files, got %d", len(files))
		}

		// Check that test files are excluded
		for _, file := range files {
			if filepath.Base(file) == "user_test.go" {
				t.Error("Test files should be excluded")
			}
			if filepath.Base(filepath.Dir(file)) == "vendor" {
				t.Error("Vendor directory should be excluded")
			}
		}
	})
}

func TestScanSourceFiles_EmptyDirectory(t *testing.T) {
	t.Run("empty_directory", func(t *testing.T) {
		// Given
		tmpDir := t.TempDir()

		// When
		files := scanSourceFiles(tmpDir)

		// Then
		if len(files) != 0 {
			t.Errorf("Expected empty slice for empty directory, got %d files", len(files))
		}
	})
}

func TestScanSourceFiles_InvalidPath(t *testing.T) {
	t.Run("invalid_path", func(t *testing.T) {
		// Given
		invalidPath := "/nonexistent/path/that/does/not/exist"

		// When
		files := scanSourceFiles(invalidPath)

		// Then
		// Should return empty slice, not panic
		if files == nil {
			t.Error("Expected empty slice, not nil")
		}
	})
}
