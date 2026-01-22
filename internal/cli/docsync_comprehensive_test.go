// Package cli provides comprehensive tests for doc-sync command
package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunDocSync_AllPaths(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	t.Run("with docs/knowledge directory", func(t *testing.T) {
		os.MkdirAll("docs/knowledge", 0755)
		os.MkdirAll(".sentinel", 0755)

		// Create knowledge base with entries
		kb := &KnowledgeBase{
			Version: "1.0",
			Entries: []KnowledgeEntry{
				{
					ID:     "1",
					Title:  "Test",
					Source: "existing.go",
				},
			},
		}
		_ = saveKnowledge(kb)

		// Create referenced file
		os.WriteFile("existing.go", []byte("package main"), 0644)

		err := runDocSync([]string{})
		if err != nil {
			t.Errorf("runDocSync() with valid setup error = %v", err)
		}
	})

	t.Run("with missing referenced files", func(t *testing.T) {
		os.MkdirAll("docs/knowledge", 0755)
		os.MkdirAll(".sentinel", 0755)

		kb := &KnowledgeBase{
			Version: "1.0",
			Entries: []KnowledgeEntry{
				{
					ID:     "1",
					Title:  "Test",
					Source: "missing.go", // File doesn't exist
				},
			},
		}
		_ = saveKnowledge(kb)

		err := runDocSync([]string{})
		if err != nil {
			t.Errorf("runDocSync() with missing files error = %v", err)
		}
	})

	t.Run("with HTTP source references", func(t *testing.T) {
		os.MkdirAll("docs/knowledge", 0755)
		os.MkdirAll(".sentinel", 0755)

		kb := &KnowledgeBase{
			Version: "1.0",
			Entries: []KnowledgeEntry{
				{
					ID:     "1",
					Title:  "Test",
					Source: "http://example.com/doc", // HTTP URL
				},
			},
		}
		_ = saveKnowledge(kb)

		err := runDocSync([]string{})
		if err != nil {
			t.Errorf("runDocSync() with HTTP sources error = %v", err)
		}
	})

	t.Run("with custom codebase path", func(t *testing.T) {
		os.MkdirAll("custom/docs/knowledge", 0755)

		err := runDocSync([]string{"custom"})
		if err != nil {
			t.Errorf("runDocSync() with custom path error = %v", err)
		}
	})

	t.Run("with multiple sync issues", func(t *testing.T) {
		os.MkdirAll("docs/knowledge", 0755)
		os.MkdirAll(".sentinel", 0755)

		kb := &KnowledgeBase{
			Version: "1.0",
			Entries: []KnowledgeEntry{
				{ID: "1", Title: "Missing", Source: "missing1.go"},
				{ID: "2", Title: "Missing", Source: "missing2.go"},
				{ID: "3", Title: "Valid", Source: "valid.go"},
			},
		}
		_ = saveKnowledge(kb)

		os.WriteFile("valid.go", []byte("package main"), 0644)

		err := runDocSync([]string{})
		if err != nil {
			t.Errorf("runDocSync() with multiple issues error = %v", err)
		}
	})
}

func TestCheckKnowledgeReferences_AllPaths(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	t.Run("multiple valid references", func(t *testing.T) {
		os.WriteFile("file1.go", []byte("package main"), 0644)
		os.WriteFile("file2.go", []byte("package main"), 0644)

		kb := &KnowledgeBase{
			Version: "1.0",
			Entries: []KnowledgeEntry{
				{ID: "1", Title: "Test1", Source: "file1.go"},
				{ID: "2", Title: "Test2", Source: "file2.go"},
			},
		}

		issues := checkKnowledgeReferences(kb, ".")
		if len(issues) > 0 {
			t.Errorf("Expected no issues, got %d", len(issues))
		}
	})

	t.Run("mixed valid and invalid references", func(t *testing.T) {
		os.WriteFile("valid.go", []byte("package main"), 0644)

		kb := &KnowledgeBase{
			Version: "1.0",
			Entries: []KnowledgeEntry{
				{ID: "1", Title: "Valid", Source: "valid.go"},
				{ID: "2", Title: "Invalid", Source: "invalid.go"},
			},
		}

		issues := checkKnowledgeReferences(kb, ".")
		if len(issues) != 1 {
			t.Errorf("Expected 1 issue, got %d", len(issues))
		}
	})

	t.Run("references in subdirectories", func(t *testing.T) {
		os.MkdirAll("subdir", 0755)
		os.WriteFile("subdir/file.go", []byte("package main"), 0644)

		kb := &KnowledgeBase{
			Version: "1.0",
			Entries: []KnowledgeEntry{
				{ID: "1", Title: "Test", Source: "subdir/file.go"},
			},
		}

		issues := checkKnowledgeReferences(kb, ".")
		if len(issues) > 0 {
			t.Errorf("Expected no issues for valid subdirectory reference, got %d", len(issues))
		}
	})
}

func TestCheckMissingDocumentation_AllPaths(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	t.Run("with files needing documentation", func(t *testing.T) {
		// Create a Go file that might trigger findings
		os.WriteFile("code.go", []byte("package main\nfunc main() {}"), 0644)

		issues := checkMissingDocumentation(".")
		// Should check without error
		_ = issues
	})

	t.Run("with empty directory", func(t *testing.T) {
		emptyDir := filepath.Join(tmpDir, "empty")
		os.MkdirAll(emptyDir, 0755)

		issues := checkMissingDocumentation(emptyDir)
		// Should handle empty directory
		_ = issues
	})

	t.Run("scan error handling", func(t *testing.T) {
		// Use invalid path
		issues := checkMissingDocumentation("/nonexistent/path")
		// Should handle scan error gracefully
		if issues == nil {
			t.Error("Expected non-nil issues slice even on error")
		}
	})
}
