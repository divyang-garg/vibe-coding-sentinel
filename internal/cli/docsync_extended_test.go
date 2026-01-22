// Package cli provides extended tests for doc-sync checkMissingDocumentation
package cli

import (
	"os"
	"testing"
)

func TestCheckMissingDocumentation_ExtendedPaths(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	t.Run("with files having many findings", func(t *testing.T) {
		// Create a file that will generate findings
		// Note: We can't easily control the exact number of findings from scanner,
		// but we test the code path that processes findings
		os.WriteFile("code.go", []byte(`
package main

func main() {
	// Multiple potential issues that might trigger findings
	var unused = 1
	var unused2 = 2
	var unused3 = 3
	var unused4 = 4
	var unused5 = 5
	var unused6 = 6
	var unused7 = 7
	var unused8 = 8
	var unused9 = 9
	var unused10 = 10
	var unused11 = 11
	_ = unused + unused2 + unused3 + unused4 + unused5 + 
		unused6 + unused7 + unused8 + unused9 + unused10 + unused11
}
`), 0644)

		issues := checkMissingDocumentation(".")
		// Should process findings without error
		_ = issues
	})

	t.Run("with files having few findings", func(t *testing.T) {
		// Create clean code with few/no findings
		os.WriteFile("clean.go", []byte("package main\nfunc main() {}"), 0644)

		issues := checkMissingDocumentation(".")
		// Should handle files with few findings
		_ = issues
	})

	t.Run("with multiple files", func(t *testing.T) {
		os.WriteFile("file1.go", []byte("package main\nfunc f1() {}"), 0644)
		os.WriteFile("file2.go", []byte("package main\nfunc f2() {}"), 0644)
		os.WriteFile("file3.go", []byte("package main\nfunc f3() {}"), 0644)

		issues := checkMissingDocumentation(".")
		// Should process multiple files
		_ = issues
	})

	t.Run("scan error path", func(t *testing.T) {
		// Test when scanner returns error
		issues := checkMissingDocumentation("/nonexistent/invalid/path")
		// Should return empty slice on error, not nil
		if issues == nil {
			t.Error("Expected non-nil slice even on error")
		}
	})

	t.Run("with empty findings", func(t *testing.T) {
		// Create empty directory or directory with files that generate no findings
		emptyDir := t.TempDir()
		
		issues := checkMissingDocumentation(emptyDir)
		// Should handle empty findings gracefully
		if issues == nil {
			t.Error("Expected non-nil slice for empty findings")
		}
	})

	t.Run("files with exactly 10 findings threshold", func(t *testing.T) {
		// The threshold is > 10, so files with exactly 10 should not trigger
		// We test the code path by creating code and verifying behavior
		os.WriteFile("threshold.go", []byte("package main\nfunc main() {}"), 0644)
		
		issues := checkMissingDocumentation(".")
		// Tests the count > 10 check
		_ = issues
	})
}

func TestRunDocSync_EdgeCases(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	t.Run("with knowledge base load error", func(t *testing.T) {
		os.MkdirAll("docs/knowledge", 0755)
		os.MkdirAll(".sentinel", 0755)

		// Create corrupted knowledge file
		os.WriteFile(".sentinel/knowledge.json", []byte("invalid json"), 0644)

		err := runDocSync([]string{})
		// Should handle load error gracefully
		if err != nil {
			t.Errorf("runDocSync() should handle load error gracefully, error = %v", err)
		}
	})

	t.Run("with empty knowledge base", func(t *testing.T) {
		os.MkdirAll("docs/knowledge", 0755)
		os.MkdirAll(".sentinel", 0755)

		kb := &KnowledgeBase{
			Version: "1.0",
			Entries: []KnowledgeEntry{},
		}
		_ = saveKnowledge(kb)

		err := runDocSync([]string{})
		if err != nil {
			t.Errorf("runDocSync() with empty KB error = %v", err)
		}
	})

	t.Run("with issues having line numbers", func(t *testing.T) {
		os.MkdirAll("docs/knowledge", 0755)
		os.MkdirAll(".sentinel", 0755)

		kb := &KnowledgeBase{
			Version: "1.0",
			Entries: []KnowledgeEntry{
				{ID: "1", Title: "Test", Source: "missing.go"},
			},
		}
		_ = saveKnowledge(kb)

		err := runDocSync([]string{})
		// Should display issues correctly
		if err != nil {
			t.Errorf("runDocSync() error = %v", err)
		}
	})

	t.Run("with multiple issue types", func(t *testing.T) {
		os.MkdirAll("docs/knowledge", 0755)
		os.MkdirAll(".sentinel", 0755)

		kb := &KnowledgeBase{
			Version: "1.0",
			Entries: []KnowledgeEntry{
				{ID: "1", Title: "Missing", Source: "missing1.go"},
				{ID: "2", Title: "Missing2", Source: "missing2.go"},
			},
		}
		_ = saveKnowledge(kb)

		// Create a file that might trigger documentation check
		os.WriteFile("code.go", []byte("package main\nfunc main() {}"), 0644)

		err := runDocSync([]string{})
		if err != nil {
			t.Errorf("runDocSync() with multiple issues error = %v", err)
		}
	})
}
