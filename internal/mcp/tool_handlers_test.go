// Package mcp provides tests for tool handlers
// Complies with CODING_STANDARDS.md: Test file max 500 lines
package mcp

import (
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
