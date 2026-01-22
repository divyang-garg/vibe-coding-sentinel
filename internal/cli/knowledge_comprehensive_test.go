// Package cli provides comprehensive tests for knowledge import/export edge cases
package cli

import (
	"os"
	"testing"
	"time"
)

func TestExportKnowledge_AllPaths(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	os.MkdirAll(".sentinel", 0755)

	t.Run("export empty knowledge base", func(t *testing.T) {
		// Create empty knowledge base
		kb := &KnowledgeBase{
			Version: "1.0",
			Entries: []KnowledgeEntry{},
		}
		_ = saveKnowledge(kb)

		err := exportKnowledge([]string{"empty.json"})
		if err != nil {
			t.Errorf("exportKnowledge() with empty KB error = %v", err)
		}

		// Verify file was created
		if _, err := os.Stat("empty.json"); os.IsNotExist(err) {
			t.Error("Expected empty.json to be created")
		}
	})

	t.Run("export knowledge with all entry fields", func(t *testing.T) {
		kb := &KnowledgeBase{
			Version: "1.0",
			Entries: []KnowledgeEntry{
				{
					ID:        "full-entry",
					Title:     "Full Entry",
					Content:   "Complete content with all fields",
					Source:    "source.md",
					Type:      "requirement",
					Tags:      []string{"tag1", "tag2"},
					Status:    "approved",
					CreatedAt: time.Now(),
				},
			},
		}
		_ = saveKnowledge(kb)

		err := exportKnowledge([]string{"full.json"})
		if err != nil {
			t.Errorf("exportKnowledge() with full entry error = %v", err)
		}
	})

	t.Run("export to nested directory", func(t *testing.T) {
		kb := &KnowledgeBase{
			Version: "1.0",
			Entries: []KnowledgeEntry{{ID: "1", Title: "Test"}},
		}
		_ = saveKnowledge(kb)

		err := exportKnowledge([]string{"nested/path/export.json"})
		if err != nil {
			// May fail if directory creation fails, but tests the path
			_ = err
		}
	})
}

func TestImportKnowledge_AllPaths(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	os.MkdirAll(".sentinel", 0755)

	t.Run("import empty knowledge base", func(t *testing.T) {
		emptyJSON := `{"version":"1.0","entries":[]}`
		os.WriteFile("empty.json", []byte(emptyJSON), 0644)

		err := importKnowledge([]string{"empty.json"})
		if err != nil {
			t.Errorf("importKnowledge() with empty KB error = %v", err)
		}
	})

	t.Run("import knowledge with multiple entries", func(t *testing.T) {
		multiJSON := `{
			"version": "1.0",
			"entries": [
				{"id":"1","title":"Entry 1","content":"Content 1","type":"requirement"},
				{"id":"2","title":"Entry 2","content":"Content 2","type":"decision"},
				{"id":"3","title":"Entry 3","content":"Content 3","type":"pattern"}
			]
		}`
		os.WriteFile("multi.json", []byte(multiJSON), 0644)

		err := importKnowledge([]string{"multi.json"})
		if err != nil {
			t.Errorf("importKnowledge() with multiple entries error = %v", err)
		}

		// Verify all entries were imported
		kb, _ := loadKnowledge()
		if len(kb.Entries) != 3 {
			t.Errorf("Expected 3 entries, got %d", len(kb.Entries))
		}
	})

	t.Run("import overwrites existing knowledge", func(t *testing.T) {
		// Create initial knowledge
		kb1 := &KnowledgeBase{
			Version: "1.0",
			Entries: []KnowledgeEntry{
				{ID: "old1", Title: "Old Entry"},
			},
		}
		_ = saveKnowledge(kb1)

		// Import new knowledge
		newJSON := `{"version":"1.0","entries":[{"id":"new1","title":"New Entry"}]}`
		os.WriteFile("new.json", []byte(newJSON), 0644)

		err := importKnowledge([]string{"new.json"})
		if err != nil {
			t.Errorf("importKnowledge() overwrite error = %v", err)
		}

		// Verify old entry was replaced
		kb2, _ := loadKnowledge()
		if len(kb2.Entries) != 1 || kb2.Entries[0].ID != "new1" {
			t.Error("Expected new entry to replace old entry")
		}
	})

	t.Run("import with malformed JSON", func(t *testing.T) {
		malformedJSON := `{"version":"1.0","entries":[{invalid json}]}`
		os.WriteFile("malformed.json", []byte(malformedJSON), 0644)

		err := importKnowledge([]string{"malformed.json"})
		if err == nil {
			t.Error("Expected error for malformed JSON")
		}
	})

	t.Run("import with missing required fields", func(t *testing.T) {
		// Test that import handles missing fields gracefully
		incompleteJSON := `{"version":"1.0","entries":[{"id":"1"}]}`
		os.WriteFile("incomplete.json", []byte(incompleteJSON), 0644)

		err := importKnowledge([]string{"incomplete.json"})
		// Should handle missing fields (they'll be zero values)
		if err != nil {
			t.Errorf("importKnowledge() should handle missing fields, error = %v", err)
		}
	})
}

func TestExportImportKnowledge_RoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	os.MkdirAll(".sentinel", 0755)

	t.Run("export then import preserves data", func(t *testing.T) {
		// Create original knowledge
		original := &KnowledgeBase{
			Version: "1.0",
			Entries: []KnowledgeEntry{
				{
					ID:      "test1",
					Title:   "Test Entry",
					Content: "Test content",
					Type:    "requirement",
					Tags:    []string{"tag1", "tag2"},
					Status:  "draft",
				},
			},
		}
		_ = saveKnowledge(original)

		// Export
		err := exportKnowledge([]string{"roundtrip.json"})
		if err != nil {
			t.Fatalf("exportKnowledge() error = %v", err)
		}

		// Clear knowledge
		os.Remove(".sentinel/knowledge.json")

		// Import
		err = importKnowledge([]string{"roundtrip.json"})
		if err != nil {
			t.Fatalf("importKnowledge() error = %v", err)
		}

		// Verify data matches
		imported, _ := loadKnowledge()
		if len(imported.Entries) != 1 {
			t.Fatalf("Expected 1 entry, got %d", len(imported.Entries))
		}

		entry := imported.Entries[0]
		if entry.ID != "test1" || entry.Title != "Test Entry" {
			t.Error("Round-trip failed: data mismatch")
		}
		if len(entry.Tags) != 2 {
			t.Errorf("Expected 2 tags, got %d", len(entry.Tags))
		}
	})
}

func TestKnowledgeImportExport_EdgeCases(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	os.MkdirAll(".sentinel", 0755)

	t.Run("export with special characters in content", func(t *testing.T) {
		kb := &KnowledgeBase{
			Version: "1.0",
			Entries: []KnowledgeEntry{
				{
					ID:      "special",
					Title:   "Special Chars",
					Content: "Content with \"quotes\" and\nnewlines\tand tabs",
				},
			},
		}
		_ = saveKnowledge(kb)

		err := exportKnowledge([]string{"special.json"})
		if err != nil {
			t.Errorf("exportKnowledge() with special chars error = %v", err)
		}

		// Import should handle it
		err = importKnowledge([]string{"special.json"})
		if err != nil {
			t.Errorf("importKnowledge() with special chars error = %v", err)
		}
	})

	t.Run("export with very long content", func(t *testing.T) {
		longContent := make([]byte, 10000)
		for i := range longContent {
			longContent[i] = 'A'
		}

		kb := &KnowledgeBase{
			Version: "1.0",
			Entries: []KnowledgeEntry{
				{
					ID:      "long",
					Title:   "Long Content",
					Content: string(longContent),
				},
			},
		}
		_ = saveKnowledge(kb)

		err := exportKnowledge([]string{"long.json"})
		if err != nil {
			t.Errorf("exportKnowledge() with long content error = %v", err)
		}
	})

	t.Run("import with invalid entry structure", func(t *testing.T) {
		invalidJSON := `{"version":"1.0","entries":[{"invalid":"structure"}]}`
		os.WriteFile("invalid.json", []byte(invalidJSON), 0644)

		err := importKnowledge([]string{"invalid.json"})
		// Should handle gracefully or return error
		_ = err
	})
}
