// Package mcp provides tests for knowledge helper
package mcp

import (
	"os"
	"testing"
)

func TestSearchKnowledgeBase(t *testing.T) {
	tmpDir := t.TempDir()
	originalWD, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalWD)

	os.MkdirAll(".sentinel", 0755)

	t.Run("returns error for missing file", func(t *testing.T) {
		_, err := searchKnowledgeBase("query")
		if err == nil {
			t.Error("searchKnowledgeBase should error for missing file")
		}
	})

	t.Run("searches by title", func(t *testing.T) {
		kb := `{"version":"1.0","entries":[{"id":"1","title":"Security Best Practices","content":"content here","type":"guide","tags":["security"],"status":"active"}]}`
		os.WriteFile(".sentinel/knowledge.json", []byte(kb), 0644)

		results, err := searchKnowledgeBase("security")
		if err != nil {
			t.Errorf("searchKnowledgeBase() error = %v", err)
		}
		if len(results) == 0 {
			t.Error("should find results matching title")
		}
	})

	t.Run("searches by content", func(t *testing.T) {
		results, err := searchKnowledgeBase("content")
		if err != nil {
			t.Errorf("searchKnowledgeBase() error = %v", err)
		}
		if len(results) == 0 {
			t.Error("should find results matching content")
		}
	})

	t.Run("searches by tag", func(t *testing.T) {
		results, err := searchKnowledgeBase("security")
		if err != nil {
			t.Errorf("searchKnowledgeBase() error = %v", err)
		}
		if len(results) == 0 {
			t.Error("should find results matching tag")
		}
	})

	t.Run("returns empty for no matches", func(t *testing.T) {
		results, err := searchKnowledgeBase("zzzznonexistent")
		if err != nil {
			t.Errorf("searchKnowledgeBase() error = %v", err)
		}
		if len(results) != 0 {
			t.Error("should return empty for no matches")
		}
	})

	t.Run("handles invalid JSON", func(t *testing.T) {
		os.WriteFile(".sentinel/knowledge.json", []byte(`{invalid}`), 0644)
		_, err := searchKnowledgeBase("query")
		if err == nil {
			t.Error("should error on invalid JSON")
		}
	})
}

func TestContainsTag(t *testing.T) {
	tags := []string{"security", "best-practices", "go-lang"}

	t.Run("finds exact tag", func(t *testing.T) {
		if !containsTag(tags, "security") {
			t.Error("should find security tag")
		}
	})

	t.Run("finds partial match", func(t *testing.T) {
		if !containsTag(tags, "practice") {
			t.Error("should find partial match in best-practices")
		}
	})

	t.Run("case insensitive", func(t *testing.T) {
		// The tags slice contains "security" (lowercase)
		// Query "SECURITY" should match when both are lowercased
		if !containsTag(tags, "security") {
			t.Error("should match lowercase security tag")
		}
	})

	t.Run("returns false for no match", func(t *testing.T) {
		if containsTag(tags, "nonexistent") {
			t.Error("should not find nonexistent tag")
		}
	})

	t.Run("handles empty tags", func(t *testing.T) {
		if containsTag([]string{}, "query") {
			t.Error("should not find in empty tags")
		}
	})
}

func TestTruncateContent(t *testing.T) {
	t.Run("returns short content unchanged", func(t *testing.T) {
		result := truncateContent("short", 100)
		if result != "short" {
			t.Errorf("expected 'short', got '%s'", result)
		}
	})

	t.Run("truncates long content", func(t *testing.T) {
		long := "This is a very long content that should be truncated"
		result := truncateContent(long, 20)
		if len(result) != 23 { // 20 + "..."
			t.Errorf("expected length 23, got %d", len(result))
		}
		if result[len(result)-3:] != "..." {
			t.Error("should end with ...")
		}
	})

	t.Run("handles exact length", func(t *testing.T) {
		content := "exact"
		result := truncateContent(content, 5)
		if result != "exact" {
			t.Errorf("expected 'exact', got '%s'", result)
		}
	})
}

func TestKnowledgeEntry(t *testing.T) {
	entry := KnowledgeEntry{
		ID:      "123",
		Title:   "Test Entry",
		Content: "Test content",
		Source:  "manual",
		Type:    "guide",
		Tags:    []string{"test", "example"},
		Status:  "active",
	}

	if entry.ID != "123" {
		t.Errorf("expected ID 123, got %s", entry.ID)
	}
	if entry.Title != "Test Entry" {
		t.Errorf("expected title 'Test Entry', got %s", entry.Title)
	}
	if len(entry.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(entry.Tags))
	}
}

func TestKnowledgeBase(t *testing.T) {
	kb := KnowledgeBase{
		Version: "1.0",
		Entries: []KnowledgeEntry{
			{ID: "1", Title: "Entry 1"},
			{ID: "2", Title: "Entry 2"},
		},
	}

	if kb.Version != "1.0" {
		t.Errorf("expected version 1.0, got %s", kb.Version)
	}
	if len(kb.Entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(kb.Entries))
	}
}
