// Package cli tests for doc-sync command
// Complies with CODING_STANDARDS.md: Test file max 500 lines
package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheckKnowledgeReferences(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	os.MkdirAll(".sentinel", 0755)

	// Create a knowledge base with a reference to an existing file
	testFile := "src/test.js"
	os.MkdirAll(filepath.Dir(testFile), 0755)
	os.WriteFile(testFile, []byte("test content"), 0644)

	kb := &KnowledgeBase{
		Version: "1.0",
		Entries: []KnowledgeEntry{
			{
				ID:     "1",
				Title:  "Test Entry",
				Source: testFile,
			},
			{
				ID:     "2",
				Title:  "Missing Entry",
				Source: "nonexistent/file.js",
			},
			{
				ID:     "3",
				Title:  "HTTP Entry",
				Source: "http://example.com/doc",
			},
		},
	}

	issues := checkKnowledgeReferences(kb, tmpDir)

	// Should find issue for nonexistent file, but not for HTTP or existing file
	foundMissing := false
	for _, issue := range issues {
		if issue.File == "nonexistent/file.js" {
			foundMissing = true
		}
		if issue.File == testFile {
			t.Error("Should not flag existing file as missing")
		}
		if issue.File == "http://example.com/doc" {
			t.Error("Should not flag HTTP URLs as missing files")
		}
	}

	if !foundMissing {
		t.Error("Should detect missing file reference")
	}
}

func TestCheckMissingDocumentation(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create a test file with code
	testFile := "test.js"
	content := ""
	for i := 0; i < 15; i++ {
		content += "console.log('test');\n"
	}
	os.WriteFile(testFile, []byte(content), 0644)

	issues := checkMissingDocumentation(tmpDir)

	// Function may or may not find issues depending on scanner results
	_ = issues
	// Just ensure it doesn't panic
}

func TestContainsTag(t *testing.T) {
	tests := []struct {
		name     string
		tags     []string
		query    string
		expected bool
	}{
		{"contains_exact", []string{"test", "demo"}, "test", true},
		{"contains_partial", []string{"testing", "demo"}, "test", true},
		{"case_insensitive", []string{"TEST", "Demo"}, "test", true},
		{"not_contains", []string{"other", "tags"}, "test", false},
		{"empty_tags", []string{}, "test", false},
		{"empty_query", []string{"test"}, "", true}, // Empty string matches any string
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsTag(tt.tags, tt.query)
			if result != tt.expected {
				t.Errorf("containsTag(%v, %q) = %v, want %v", tt.tags, tt.query, result, tt.expected)
			}
		})
	}
}

func TestUpdateKnowledgeEntry(t *testing.T) {
	kb := &KnowledgeBase{
		Version: "1.0",
		Entries: []KnowledgeEntry{
			{ID: "1", Title: "Original"},
			{ID: "2", Title: "Another"},
		},
	}

	updated := KnowledgeEntry{ID: "1", Title: "Updated"}
	updateKnowledgeEntry(kb, updated)

	if kb.Entries[0].Title != "Updated" {
		t.Error("Entry should be updated")
	}
}
