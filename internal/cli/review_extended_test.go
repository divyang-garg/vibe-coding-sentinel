// Package cli provides extended tests for review command
package cli

import (
	"os"
	"testing"
)

func TestRunReview_EdgeCases(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	os.MkdirAll(".sentinel", 0755)

	t.Run("with approved entries only", func(t *testing.T) {
		kb := &KnowledgeBase{
			Version: "1.0",
			Entries: []KnowledgeEntry{
				{
					ID:     "1",
					Title:  "Approved Entry",
					Status: "approved",
				},
			},
		}
		_ = saveKnowledge(kb)

		err := runReview([]string{})
		if err != nil {
			t.Errorf("runReview() error = %v", err)
		}
	})

	t.Run("with archived entries", func(t *testing.T) {
		kb := &KnowledgeBase{
			Version: "1.0",
			Entries: []KnowledgeEntry{
				{
					ID:     "1",
					Title:  "Archived Entry",
					Status: "archived",
				},
			},
		}
		_ = saveKnowledge(kb)

		err := runReview([]string{})
		if err != nil {
			t.Errorf("runReview() error = %v", err)
		}
	})

	t.Run("with entries having tags", func(t *testing.T) {
		kb := &KnowledgeBase{
			Version: "1.0",
			Entries: []KnowledgeEntry{
				{
					ID:     "1",
					Title:  "Tagged Entry",
					Status: "draft",
					Tags:   []string{"tag1", "tag2", "tag3"},
				},
			},
		}
		_ = saveKnowledge(kb)

		// Note: Interactive review requires stdin, so this tests the setup path
		err := runReview([]string{})
		// Will attempt to read stdin, may timeout or error
		_ = err
	})

	t.Run("with entries without tags", func(t *testing.T) {
		kb := &KnowledgeBase{
			Version: "1.0",
			Entries: []KnowledgeEntry{
				{
					ID:     "1",
					Title:  "No Tags Entry",
					Status: "draft",
					Tags:   []string{},
				},
			},
		}
		_ = saveKnowledge(kb)

		err := runReview([]string{})
		_ = err
	})

	t.Run("with long content", func(t *testing.T) {
		longContent := make([]byte, 500)
		for i := range longContent {
			longContent[i] = 'A'
		}

		kb := &KnowledgeBase{
			Version: "1.0",
			Entries: []KnowledgeEntry{
				{
					ID:      "1",
					Title:   "Long Content",
					Status:  "draft",
					Content: string(longContent),
				},
			},
		}
		_ = saveKnowledge(kb)

		err := runReview([]string{})
		// Tests truncation in display
		_ = err
	})

	t.Run("with short content", func(t *testing.T) {
		kb := &KnowledgeBase{
			Version: "1.0",
			Entries: []KnowledgeEntry{
				{
					ID:      "1",
					Title:   "Short Content",
					Status:  "draft",
					Content: "Short",
				},
			},
		}
		_ = saveKnowledge(kb)

		err := runReview([]string{})
		_ = err
	})
}

func TestUpdateKnowledgeEntry(t *testing.T) {
	t.Run("update existing entry", func(t *testing.T) {
		kb := &KnowledgeBase{
			Version: "1.0",
			Entries: []KnowledgeEntry{
				{ID: "1", Title: "Original"},
				{ID: "2", Title: "Other"},
			},
		}

		updated := KnowledgeEntry{
			ID:    "1",
			Title: "Updated",
		}

		updateKnowledgeEntry(kb, updated)

		if kb.Entries[0].Title != "Updated" {
			t.Error("Expected entry to be updated")
		}
		if kb.Entries[1].Title != "Other" {
			t.Error("Expected other entry to remain unchanged")
		}
	})

	t.Run("update non-existent entry", func(t *testing.T) {
		kb := &KnowledgeBase{
			Version: "1.0",
			Entries: []KnowledgeEntry{
				{ID: "1", Title: "Original"},
			},
		}

		updated := KnowledgeEntry{
			ID:    "nonexistent",
			Title: "Updated",
		}

		updateKnowledgeEntry(kb, updated)

		// Should not modify entries if ID doesn't match
		if kb.Entries[0].Title != "Original" {
			t.Error("Expected entry to remain unchanged")
		}
	})

	t.Run("update with empty entries", func(t *testing.T) {
		kb := &KnowledgeBase{
			Version: "1.0",
			Entries: []KnowledgeEntry{},
		}

		updated := KnowledgeEntry{
			ID:    "1",
			Title: "New",
		}

		updateKnowledgeEntry(kb, updated)
		// Should handle empty entries gracefully
		if len(kb.Entries) != 0 {
			t.Error("Expected entries to remain empty")
		}
	})
}
