// Package cli provides tests for doc-sync command
package cli

import (
	"os"
	"testing"
)

func TestRunDocSync(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	t.Run("no docs directory", func(t *testing.T) {
		err := runDocSync([]string{})
		if err != nil {
			t.Errorf("runDocSync() error = %v", err)
		}
	})

	t.Run("with docs directory", func(t *testing.T) {
		os.MkdirAll("docs/knowledge", 0755)
		err := runDocSync([]string{})
		if err != nil {
			t.Errorf("runDocSync() error = %v", err)
		}
	})

	t.Run("with custom path", func(t *testing.T) {
		err := runDocSync([]string{"."})
		if err != nil {
			t.Errorf("runDocSync() error = %v", err)
		}
	})
}

func TestCheckKnowledgeReferences(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	os.MkdirAll(".sentinel", 0755)

	// Create a test file
	testFile := "test.go"
	os.WriteFile(testFile, []byte("package main"), 0644)

	t.Run("valid reference", func(t *testing.T) {
		kb := &KnowledgeBase{
			Entries: []KnowledgeEntry{
				{
					ID:     "1",
					Title:  "Test",
					Source: testFile,
				},
			},
		}

		issues := checkKnowledgeReferences(kb, ".")
		if len(issues) > 0 {
			t.Errorf("Expected no issues for valid reference, got %d", len(issues))
		}
	})

	t.Run("missing reference", func(t *testing.T) {
		kb := &KnowledgeBase{
			Entries: []KnowledgeEntry{
				{
					ID:     "1",
					Title:  "Test",
					Source: "nonexistent.go",
				},
			},
		}

		issues := checkKnowledgeReferences(kb, ".")
		if len(issues) == 0 {
			t.Error("Expected issue for missing reference")
		}
	})

	t.Run("HTTP reference", func(t *testing.T) {
		kb := &KnowledgeBase{
			Entries: []KnowledgeEntry{
				{
					ID:     "1",
					Title:  "Test",
					Source: "http://example.com/doc",
				},
			},
		}

		issues := checkKnowledgeReferences(kb, ".")
		if len(issues) > 0 {
			t.Error("Expected no issues for HTTP reference")
		}
	})

	t.Run("empty source", func(t *testing.T) {
		kb := &KnowledgeBase{
			Entries: []KnowledgeEntry{
				{
					ID:    "1",
					Title: "Test",
					// No Source
				},
			},
		}

		issues := checkKnowledgeReferences(kb, ".")
		if len(issues) > 0 {
			t.Error("Expected no issues for empty source")
		}
	})
}

func TestCheckMissingDocumentation(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	t.Run("no findings", func(t *testing.T) {
		issues := checkMissingDocumentation(".")
		// Should not error, may return empty list
		_ = issues
	})

	t.Run("with code files", func(t *testing.T) {
		// Create a simple Go file
		os.WriteFile("test.go", []byte("package main\nfunc main() {}"), 0644)
		issues := checkMissingDocumentation(".")
		// Should complete without error
		_ = issues
	})
}
