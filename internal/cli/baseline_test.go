// Package cli provides tests for baseline command
package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestClearBaseline(t *testing.T) {
	// Setup: Create a temp directory structure
	tmpDir := t.TempDir()
	originalWD, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalWD)

	// Create sentinel directory
	os.MkdirAll(".sentinel", 0755)

	t.Run("clears existing baseline", func(t *testing.T) {
		// Create a baseline file
		baselinePath := ".sentinel/baseline.json"
		os.WriteFile(baselinePath, []byte(`{"version":"1.0","entries":[]}`), 0644)

		err := clearBaseline()
		if err != nil {
			t.Errorf("clearBaseline() error = %v", err)
		}

		// Verify file is removed
		if _, err := os.Stat(baselinePath); !os.IsNotExist(err) {
			t.Error("baseline file should be removed")
		}
	})

	t.Run("handles empty baseline gracefully", func(t *testing.T) {
		err := clearBaseline()
		if err != nil {
			t.Errorf("clearBaseline() on non-existent file should not error: %v", err)
		}
	})
}

func TestExportBaseline(t *testing.T) {
	tmpDir := t.TempDir()
	originalWD, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalWD)

	os.MkdirAll(".sentinel", 0755)

	t.Run("exports baseline to file", func(t *testing.T) {
		// Create baseline
		baselinePath := ".sentinel/baseline.json"
		os.WriteFile(baselinePath, []byte(`{"version":"1.0","entries":[{"pattern":"test","file":"test.js","line":1}]}`), 0644)

		exportPath := filepath.Join(tmpDir, "export.json")
		err := exportBaseline([]string{exportPath})
		if err != nil {
			t.Errorf("exportBaseline() error = %v", err)
		}

		// Verify export file exists
		if _, err := os.Stat(exportPath); os.IsNotExist(err) {
			t.Error("export file should exist")
		}
	})

	t.Run("requires export path argument", func(t *testing.T) {
		err := exportBaseline([]string{})
		if err == nil {
			t.Error("exportBaseline() should require path argument")
		}
	})
}

func TestImportBaseline(t *testing.T) {
	tmpDir := t.TempDir()
	originalWD, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalWD)

	os.MkdirAll(".sentinel", 0755)

	t.Run("imports baseline from file", func(t *testing.T) {
		// Create import file
		importPath := filepath.Join(tmpDir, "import.json")
		os.WriteFile(importPath, []byte(`{"version":"1.0","entries":[{"pattern":"imported","file":"imported.js","line":10}]}`), 0644)

		err := importBaseline([]string{importPath})
		if err != nil {
			t.Errorf("importBaseline() error = %v", err)
		}

		// Verify baseline was created
		if _, err := os.Stat(".sentinel/baseline.json"); os.IsNotExist(err) {
			t.Error("baseline file should exist after import")
		}
	})

	t.Run("requires import path argument", func(t *testing.T) {
		err := importBaseline([]string{})
		if err == nil {
			t.Error("importBaseline() should require path argument")
		}
	})

	t.Run("handles missing import file", func(t *testing.T) {
		err := importBaseline([]string{"/nonexistent/path.json"})
		if err == nil {
			t.Error("importBaseline() should error for missing file")
		}
	})
}

func TestRunBaseline(t *testing.T) {
	tmpDir := t.TempDir()
	originalWD, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalWD)

	os.MkdirAll(".sentinel", 0755)

	t.Run("routes add command", func(t *testing.T) {
		err := runBaseline([]string{"add", "test.js", "10", "test reason"})
		if err != nil {
			t.Errorf("runBaseline(add) error = %v", err)
		}
	})

	t.Run("routes clear command", func(t *testing.T) {
		err := runBaseline([]string{"clear"})
		if err != nil {
			t.Errorf("runBaseline(clear) error = %v", err)
		}
	})

	t.Run("routes help command", func(t *testing.T) {
		err := runBaseline([]string{"help"})
		if err != nil {
			t.Errorf("runBaseline(help) error = %v", err)
		}
	})

	t.Run("handles unknown command", func(t *testing.T) {
		err := runBaseline([]string{"unknown"})
		if err == nil {
			t.Error("runBaseline(unknown) should error")
		}
	})

	t.Run("shows baseline when no args", func(t *testing.T) {
		err := runBaseline([]string{})
		if err != nil {
			t.Errorf("runBaseline([]) error = %v", err)
		}
	})
}

func TestAddToBaseline(t *testing.T) {
	tmpDir := t.TempDir()
	originalWD, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalWD)

	os.MkdirAll(".sentinel", 0755)

	t.Run("adds entry with reason", func(t *testing.T) {
		err := addToBaseline([]string{"file.js", "42", "test reason"})
		if err != nil {
			t.Errorf("addToBaseline() error = %v", err)
		}

		// Verify entry was added
		baseline, _ := loadBaseline()
		if len(baseline.Entries) == 0 {
			t.Error("baseline should have entries")
		}
	})

	t.Run("adds entry without reason", func(t *testing.T) {
		// Clear first
		clearBaseline()

		err := addToBaseline([]string{"file.js", "50"})
		if err != nil {
			t.Errorf("addToBaseline() error = %v", err)
		}
	})

	t.Run("requires file and line arguments", func(t *testing.T) {
		err := addToBaseline([]string{"file.js"})
		if err == nil {
			t.Error("addToBaseline() should require 2 args")
		}
	})

	t.Run("validates line number", func(t *testing.T) {
		err := addToBaseline([]string{"file.js", "notanumber"})
		if err == nil {
			t.Error("addToBaseline() should validate line number")
		}
	})
}

func TestRemoveFromBaseline(t *testing.T) {
	tmpDir := t.TempDir()
	originalWD, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalWD)

	os.MkdirAll(".sentinel", 0755)

	t.Run("removes entry by index", func(t *testing.T) {
		// Add an entry first
		addToBaseline([]string{"file.js", "10", "test"})

		err := removeFromBaseline([]string{"1"})
		if err != nil {
			t.Errorf("removeFromBaseline() error = %v", err)
		}
	})

	t.Run("requires index argument", func(t *testing.T) {
		err := removeFromBaseline([]string{})
		if err == nil {
			t.Error("removeFromBaseline() should require index")
		}
	})

	t.Run("validates index number", func(t *testing.T) {
		err := removeFromBaseline([]string{"notanumber"})
		if err == nil {
			t.Error("removeFromBaseline() should validate index")
		}
	})

	t.Run("handles out of range index", func(t *testing.T) {
		// Clear and add one entry
		clearBaseline()
		addToBaseline([]string{"file.js", "10", "test"})

		err := removeFromBaseline([]string{"99"})
		if err == nil {
			t.Error("removeFromBaseline() should error for out of range")
		}
	})
}

func TestShowBaseline(t *testing.T) {
	tmpDir := t.TempDir()
	originalWD, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalWD)

	os.MkdirAll(".sentinel", 0755)

	t.Run("shows empty baseline message", func(t *testing.T) {
		clearBaseline()
		err := showBaseline()
		if err != nil {
			t.Errorf("showBaseline() error = %v", err)
		}
	})

	t.Run("shows baseline with entries", func(t *testing.T) {
		addToBaseline([]string{"file.js", "10", "test"})

		err := showBaseline()
		if err != nil {
			t.Errorf("showBaseline() error = %v", err)
		}
	})
}

func TestPrintBaselineHelp(t *testing.T) {
	err := printBaselineHelp()
	if err != nil {
		t.Errorf("printBaselineHelp() error = %v", err)
	}
}
