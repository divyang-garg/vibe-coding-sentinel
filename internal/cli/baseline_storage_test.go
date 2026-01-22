// Package cli provides tests for baseline storage functions
package cli

import (
	"os"
	"testing"
)

func TestLoadBaseline(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	os.MkdirAll(".sentinel", 0755)

	t.Run("new file", func(t *testing.T) {
		baseline, err := loadBaseline()
		if err == nil {
			t.Error("Expected error when loading non-existent file")
		}
		if baseline == nil {
			t.Error("Expected non-nil baseline even on error")
		}
		if baseline.Version != "1.0" {
			t.Errorf("Expected version 1.0, got %s", baseline.Version)
		}
	})

	t.Run("corrupted file", func(t *testing.T) {
		os.WriteFile(".sentinel/baseline.json", []byte("invalid json"), 0644)
		baseline, err := loadBaseline()
		if err == nil {
			t.Error("Expected error when loading corrupted file")
		}
		if baseline != nil {
			t.Error("Expected nil baseline on parse error")
		}
	})

	t.Run("valid baseline", func(t *testing.T) {
		validBaseline := `{"version":"1.0","entries":[{"pattern":"test","file":"test.js","line":10}]}`
		os.WriteFile(".sentinel/baseline.json", []byte(validBaseline), 0644)
		baseline, err := loadBaseline()
		if err != nil {
			t.Fatalf("loadBaseline() error = %v", err)
		}
		if baseline == nil {
			t.Fatal("Expected non-nil baseline")
		}
		if len(baseline.Entries) != 1 {
			t.Errorf("Expected 1 entry, got %d", len(baseline.Entries))
		}
	})
}

func TestSaveBaseline(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	t.Run("creates directory", func(t *testing.T) {
		baseline := &Baseline{
			Version: "1.0",
			Entries: []BaselineEntry{},
		}
		err := saveBaseline(baseline)
		if err != nil {
			t.Errorf("saveBaseline() error = %v", err)
		}
		if _, err := os.Stat(".sentinel/baseline.json"); os.IsNotExist(err) {
			t.Error("Expected baseline.json to be created")
		}
	})

	t.Run("saves with entries", func(t *testing.T) {
		baseline := &Baseline{
			Version: "1.0",
			Entries: []BaselineEntry{
				{
					Pattern: "test",
					File:    "test.js",
					Line:    10,
				},
			},
		}
		err := saveBaseline(baseline)
		if err != nil {
			t.Errorf("saveBaseline() error = %v", err)
		}
		// Verify by loading
		loaded, err := loadBaseline()
		if err != nil {
			t.Fatalf("loadBaseline() error = %v", err)
		}
		if len(loaded.Entries) != 1 {
			t.Errorf("Expected 1 entry after save, got %d", len(loaded.Entries))
		}
	})
}

func TestGetBaselinePath(t *testing.T) {
	path := getBaselinePath()
	if path != ".sentinel/baseline.json" {
		t.Errorf("Expected .sentinel/baseline.json, got %s", path)
	}
}

func TestExportBaseline(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	os.MkdirAll(".sentinel", 0755)

	// Setup: add a baseline entry
	baseline := &Baseline{
		Version: "1.0",
		Entries: []BaselineEntry{
			{Pattern: "test", File: "test.js", Line: 10},
		},
	}
	_ = saveBaseline(baseline)

	t.Run("valid export", func(t *testing.T) {
		err := exportBaseline([]string{"export.json"})
		if err != nil {
			t.Errorf("exportBaseline() error = %v", err)
		}
		if _, err := os.Stat("export.json"); os.IsNotExist(err) {
			t.Error("Expected export.json to be created")
		}
	})

	t.Run("missing args", func(t *testing.T) {
		err := exportBaseline([]string{})
		if err == nil {
			t.Error("Expected error for missing args")
		}
	})
}

func TestImportBaseline(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	os.MkdirAll(".sentinel", 0755)

	t.Run("valid import", func(t *testing.T) {
		importJSON := `{"version":"1.0","entries":[{"pattern":"imported","file":"import.js","line":20}]}`
		os.WriteFile("import.json", []byte(importJSON), 0644)

		err := importBaseline([]string{"import.json"})
		if err != nil {
			t.Errorf("importBaseline() error = %v", err)
		}

		// Verify import
		baseline, _ := loadBaseline()
		if len(baseline.Entries) != 1 {
			t.Errorf("Expected 1 imported entry, got %d", len(baseline.Entries))
		}
	})

	t.Run("missing file", func(t *testing.T) {
		err := importBaseline([]string{"nonexistent.json"})
		if err == nil {
			t.Error("Expected error for missing file")
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		os.WriteFile("invalid.json", []byte("invalid json"), 0644)
		err := importBaseline([]string{"invalid.json"})
		if err == nil {
			t.Error("Expected error for invalid JSON")
		}
	})

	t.Run("missing args", func(t *testing.T) {
		err := importBaseline([]string{})
		if err == nil {
			t.Error("Expected error for missing args")
		}
	})
}
