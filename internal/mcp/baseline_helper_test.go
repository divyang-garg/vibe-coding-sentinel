// Package mcp provides tests for baseline helper
package mcp

import (
	"os"
	"testing"
)

func TestLoadBaseline(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("returns empty baseline for missing file", func(t *testing.T) {
		baseline, err := loadBaseline(tmpDir + "/nonexistent.json")
		if baseline == nil {
			t.Error("loadBaseline should return empty baseline")
		}
		if err == nil {
			t.Error("loadBaseline should return error for missing file")
		}
		if baseline.Version != "1.0" {
			t.Errorf("expected version 1.0, got %s", baseline.Version)
		}
		if len(baseline.Entries) != 0 {
			t.Error("entries should be empty")
		}
	})

	t.Run("loads existing baseline", func(t *testing.T) {
		path := tmpDir + "/baseline.json"
		os.WriteFile(path, []byte(`{"version":"2.0","entries":[{"pattern":"test","file":"test.js","line":1}]}`), 0644)

		baseline, err := loadBaseline(path)
		if err != nil {
			t.Errorf("loadBaseline() error = %v", err)
		}
		if baseline.Version != "2.0" {
			t.Errorf("expected version 2.0, got %s", baseline.Version)
		}
		if len(baseline.Entries) != 1 {
			t.Errorf("expected 1 entry, got %d", len(baseline.Entries))
		}
	})

	t.Run("handles invalid JSON", func(t *testing.T) {
		path := tmpDir + "/invalid.json"
		os.WriteFile(path, []byte(`{invalid json}`), 0644)

		_, err := loadBaseline(path)
		if err == nil {
			t.Error("loadBaseline should error on invalid JSON")
		}
	})
}

func TestSaveBaseline(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("saves baseline to file", func(t *testing.T) {
		path := tmpDir + "/save_test.json"
		baseline := &Baseline{
			Version: "1.0",
			Entries: []BaselineEntry{
				{Pattern: "test", File: "test.js", Line: 10},
			},
		}

		err := saveBaseline(path, baseline)
		if err != nil {
			t.Errorf("saveBaseline() error = %v", err)
		}

		// Verify file exists
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Error("baseline file should exist")
		}

		// Verify content
		content, _ := os.ReadFile(path)
		if len(content) == 0 {
			t.Error("baseline file should have content")
		}
	})

	t.Run("creates parent directories", func(t *testing.T) {
		path := tmpDir + "/nested/dir/baseline.json"
		baseline := &Baseline{Version: "1.0", Entries: []BaselineEntry{}}

		err := saveBaseline(path, baseline)
		if err != nil {
			t.Errorf("saveBaseline() error = %v", err)
		}

		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Error("nested baseline file should exist")
		}
	})
}

func TestAddToBaselineFile(t *testing.T) {
	tmpDir := t.TempDir()
	originalWD, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalWD)

	t.Run("adds entry to baseline", func(t *testing.T) {
		err := addToBaselineFile("test.js", 42, "test reason")
		if err != nil {
			t.Errorf("addToBaselineFile() error = %v", err)
		}

		// Verify entry was added
		baseline, _ := loadBaseline(".sentinel/baseline.json")
		if len(baseline.Entries) == 0 {
			t.Error("baseline should have entries")
		}
	})

	t.Run("appends to existing baseline", func(t *testing.T) {
		err := addToBaselineFile("another.js", 100, "another reason")
		if err != nil {
			t.Errorf("addToBaselineFile() error = %v", err)
		}

		baseline, _ := loadBaseline(".sentinel/baseline.json")
		if len(baseline.Entries) < 2 {
			t.Error("baseline should have multiple entries")
		}
	})
}

func TestBaselineEntry(t *testing.T) {
	entry := BaselineEntry{
		Pattern: "test_pattern",
		File:    "file.js",
		Line:    42,
		Reason:  "test reason",
		AddedBy: "testuser",
		Hash:    "file.js:42",
	}

	if entry.Pattern != "test_pattern" {
		t.Errorf("expected pattern test_pattern, got %s", entry.Pattern)
	}
	if entry.Line != 42 {
		t.Errorf("expected line 42, got %d", entry.Line)
	}
	if entry.Hash != "file.js:42" {
		t.Errorf("expected hash file.js:42, got %s", entry.Hash)
	}
}
