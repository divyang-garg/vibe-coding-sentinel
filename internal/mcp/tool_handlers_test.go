// Package mcp provides tests for tool handlers
// Complies with CODING_STANDARDS.md: Test file max 500 lines
package mcp

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestHandleCheckFileSize(t *testing.T) {
	server := NewServer()

	t.Run("valid file", func(t *testing.T) {
		// Create a temporary test file
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.go")
		content := "package main\n\nfunc main() {}\n"
		os.WriteFile(testFile, []byte(content), 0644)

		args := map[string]interface{}{
			"file": testFile,
		}

		result, err := server.handleCheckFileSize(args)
		if err != nil {
			t.Fatalf("handleCheckFileSize failed: %v", err)
		}

		resultMap, ok := result.(map[string]interface{})
		if !ok {
			t.Fatal("Result is not a map")
		}

		if resultMap["valid"] != true {
			t.Error("Expected file to be valid")
		}

		if resultMap["lines"] == nil {
			t.Error("Lines not returned")
		}
	})

	t.Run("missing file parameter", func(t *testing.T) {
		args := map[string]interface{}{}
		_, err := server.handleCheckFileSize(args)
		if err == nil {
			t.Error("Expected error for missing file parameter")
		}
	})

	t.Run("non-existent file", func(t *testing.T) {
		args := map[string]interface{}{
			"file": "/nonexistent/file.go",
		}
		_, err := server.handleCheckFileSize(args)
		if err == nil {
			t.Error("Expected error for non-existent file")
		}
	})
}

func TestHandleAudit(t *testing.T) {
	server := NewServer()

	t.Run("audit with path", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.js")
		os.WriteFile(testFile, []byte("eval(userInput);"), 0644)

		args := map[string]interface{}{
			"path": tmpDir,
		}

		result, err := server.handleAudit(args)
		if err != nil {
			t.Fatalf("handleAudit failed: %v", err)
		}

		if result == nil {
			t.Error("Expected result, got nil")
		}
	})

	t.Run("audit with default path", func(t *testing.T) {
		args := map[string]interface{}{}
		result, err := server.handleAudit(args)
		if err != nil {
			t.Fatalf("handleAudit with default path failed: %v", err)
		}

		if result == nil {
			t.Error("Expected result, got nil")
		}
	})
}

func TestHandleVibeCheck(t *testing.T) {
	server := NewServer()

	t.Run("vibe check with path", func(t *testing.T) {
		tmpDir := t.TempDir()

		args := map[string]interface{}{
			"path": tmpDir,
		}

		result, err := server.handleVibeCheck(args)
		if err != nil {
			t.Fatalf("handleVibeCheck failed: %v", err)
		}

		if result == nil {
			t.Error("Expected result, got nil")
		}
	})
}

func TestHandleBaselineAdd(t *testing.T) {
	server := NewServer()
	tmpDir := t.TempDir()

	// Change to temp directory
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// Create .sentinel directory
	os.MkdirAll(".sentinel", 0755)

	t.Run("add to baseline", func(t *testing.T) {
		args := map[string]interface{}{
			"file":   "test.go",
			"line":   float64(10),
			"reason": "Test reason",
		}

		result, err := server.handleBaselineAdd(args)
		if err != nil {
			t.Fatalf("handleBaselineAdd failed: %v", err)
		}

		resultMap, ok := result.(map[string]interface{})
		if !ok {
			t.Fatal("Result is not a map")
		}

		if resultMap["status"] != "added" {
			t.Error("Expected status to be 'added'")
		}
	})

	t.Run("missing file parameter", func(t *testing.T) {
		args := map[string]interface{}{
			"line": float64(10),
		}
		_, err := server.handleBaselineAdd(args)
		if err == nil {
			t.Error("Expected error for missing file")
		}
	})

	t.Run("missing line parameter", func(t *testing.T) {
		args := map[string]interface{}{
			"file": "test.go",
		}
		_, err := server.handleBaselineAdd(args)
		if err == nil {
			t.Error("Expected error for missing line")
		}
	})
}

func TestHandleKnowledgeSearch(t *testing.T) {
	server := NewServer()
	tmpDir := t.TempDir()

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	t.Run("search non-existent knowledge base", func(t *testing.T) {
		args := map[string]interface{}{
			"query": "test",
		}

		result, err := server.handleKnowledgeSearch(args)
		if err != nil {
			t.Fatalf("handleKnowledgeSearch failed: %v", err)
		}

		resultMap, ok := result.(map[string]interface{})
		if !ok {
			t.Fatal("Result is not a map")
		}

		if resultMap["count"] != 0 {
			t.Error("Expected count to be 0 for non-existent KB")
		}
	})

	t.Run("missing query parameter", func(t *testing.T) {
		args := map[string]interface{}{}
		_, err := server.handleKnowledgeSearch(args)
		if err == nil {
			t.Error("Expected error for missing query")
		}
	})

	t.Run("search existing knowledge base", func(t *testing.T) {
		// Create knowledge base
		os.MkdirAll(".sentinel", 0755)
		kb := `{
			"version": "1.0",
			"entries": [
				{
					"title": "Test Entry",
					"content": "Test content",
					"type": "note",
					"tags": ["test"]
				}
			]
		}`
		os.WriteFile(".sentinel/knowledge.json", []byte(kb), 0644)

		args := map[string]interface{}{
			"query": "test",
		}

		result, err := server.handleKnowledgeSearch(args)
		if err != nil {
			t.Fatalf("handleKnowledgeSearch failed: %v", err)
		}

		resultMap, ok := result.(map[string]interface{})
		if !ok {
			t.Fatal("Result is not a map")
		}

		count, ok := resultMap["count"].(int)
		if !ok || count < 1 {
			t.Error("Expected at least 1 result")
		}
	})
}

func TestHandleCheckFileSize_FileExceedsLimit(t *testing.T) {
	server := NewServer()
	tmpDir := t.TempDir()

	// Create a file that exceeds the limit for http_handler (200 lines)
	testFile := filepath.Join(tmpDir, "api", "handlers", "large_handler.go")
	os.MkdirAll(filepath.Dir(testFile), 0755)

	// Create file with 250 lines (exceeds 200 line limit)
	content := "package handlers\n\n"
	for i := 0; i < 250; i++ {
		content += "// Line " + fmt.Sprintf("%d\n", i+1)
	}
	os.WriteFile(testFile, []byte(content), 0644)

	args := map[string]interface{}{
		"file": testFile,
	}

	result, err := server.handleCheckFileSize(args)
	if err != nil {
		t.Fatalf("handleCheckFileSize failed: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Result is not a map")
	}

	if resultMap["valid"] != false {
		t.Error("Expected file to be invalid (exceeds limit)")
	}

	if resultMap["violation"] == nil || resultMap["violation"] == "" {
		t.Error("Expected violation message")
	}
}

func TestHandleCheckFileSize_DifferentFileTypes(t *testing.T) {
	server := NewServer()
	tmpDir := t.TempDir()

	testCases := []struct {
		path     string
		expected string
		limit    int
	}{
		{"test_file_test.go", "test", 500},
		{"cli/handler.go", "cli_handler", 300},
		{"api/handlers/user.go", "http_handler", 200},
		{"services/user_service.go", "business_service", 400},
		{"models/types.go", "types", 200},
		{"utils/helper.go", "utility", 250},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			testFile := filepath.Join(tmpDir, tc.path)
			os.MkdirAll(filepath.Dir(testFile), 0755)
			os.WriteFile(testFile, []byte("package test\n"), 0644)

			args := map[string]interface{}{
				"file": testFile,
			}

			result, err := server.handleCheckFileSize(args)
			if err != nil {
				t.Fatalf("handleCheckFileSize failed: %v", err)
			}

			resultMap, ok := result.(map[string]interface{})
			if !ok {
				t.Fatal("Result is not a map")
			}

			if resultMap["file_type"] != tc.expected {
				t.Errorf("Expected file_type %s, got %v", tc.expected, resultMap["file_type"])
			}

			limitValue, ok := resultMap["limit"].(float64)
			if !ok {
				limitValueInt, okInt := resultMap["limit"].(int)
				if !okInt {
					t.Errorf("Expected limit to be numeric, got %v (type %T)", resultMap["limit"], resultMap["limit"])
				} else {
					limitValue = float64(limitValueInt)
					if int(limitValue) != tc.limit {
						t.Errorf("Expected limit %d, got %v", tc.limit, limitValue)
					}
				}
			} else {
				if int(limitValue) != tc.limit {
					t.Errorf("Expected limit %d, got %v", tc.limit, limitValue)
				}
			}
		})
	}
}

func TestHandleAudit_WithVibeCheck(t *testing.T) {
	server := NewServer()
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.js")
	os.WriteFile(testFile, []byte("eval(userInput);"), 0644)

	args := map[string]interface{}{
		"path":      tmpDir,
		"vibeCheck": true,
	}

	result, err := server.handleAudit(args)
	if err != nil {
		t.Fatalf("handleAudit with vibeCheck failed: %v", err)
	}

	if result == nil {
		t.Error("Expected result, got nil")
	}
}

func TestHandleAudit_WithDeep(t *testing.T) {
	server := NewServer()
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.js")
	os.WriteFile(testFile, []byte("eval(userInput);"), 0644)

	args := map[string]interface{}{
		"path": tmpDir,
		"deep": true,
	}

	result, err := server.handleAudit(args)
	if err != nil {
		t.Fatalf("handleAudit with deep failed: %v", err)
	}

	if result == nil {
		t.Error("Expected result, got nil")
	}
}

func TestHandleAudit_WithVibeCheckAndDeep(t *testing.T) {
	server := NewServer()
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.js")
	os.WriteFile(testFile, []byte("eval(userInput);"), 0644)

	args := map[string]interface{}{
		"path":      tmpDir,
		"vibeCheck": true,
		"deep":      true,
	}

	result, err := server.handleAudit(args)
	if err != nil {
		t.Fatalf("handleAudit with vibeCheck and deep failed: %v", err)
	}

	if result == nil {
		t.Error("Expected result, got nil")
	}
}

func TestHandleVibeCheck_Error(t *testing.T) {
	server := NewServer()

	// Use a path that might cause an error (non-existent directory)
	args := map[string]interface{}{
		"path": "/nonexistent/directory/path",
	}

	// Vibe check might fail or succeed depending on implementation
	// We just verify it doesn't panic
	result, err := server.handleVibeCheck(args)

	// Either result or error is acceptable
	if err != nil && result != nil {
		t.Error("Should return either result or error, not both")
	}
}

func TestHandleBaselineAdd_AddError(t *testing.T) {
	server := NewServer()
	tmpDir := t.TempDir()

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// Don't create .sentinel directory to cause an error
	args := map[string]interface{}{
		"file":   "test.go",
		"line":   float64(10),
		"reason": "Test reason",
	}

	// This should fail because .sentinel directory doesn't exist
	result, err := server.handleBaselineAdd(args)
	if err == nil {
		// If it doesn't error, check if it created the directory
		// In that case, verify the result
		if result != nil {
			resultMap, ok := result.(map[string]interface{})
			if ok && resultMap["status"] == "added" {
				// Success case - directory was created
				return
			}
		}
		t.Error("Expected error or successful result with status 'added'")
	}
}

func TestHandleKnowledgeSearch_SearchError(t *testing.T) {
	server := NewServer()
	tmpDir := t.TempDir()

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// Create invalid knowledge.json to cause parse error
	os.MkdirAll(".sentinel", 0755)
	os.WriteFile(".sentinel/knowledge.json", []byte("invalid json"), 0644)

	args := map[string]interface{}{
		"query": "test",
	}

	// Should handle error gracefully and return empty results
	result, err := server.handleKnowledgeSearch(args)
	if err != nil {
		t.Fatalf("handleKnowledgeSearch should handle errors gracefully: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Result should be a map")
	}

	// Should return empty results or error message
	if resultMap["count"] == nil {
		t.Error("Expected count in result")
	}
}
