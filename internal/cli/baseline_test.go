// Package cli provides tests for baseline command
package cli

import (
	"os"
	"testing"
)

func TestRunBaseline_Commands(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	os.MkdirAll(".sentinel", 0755)

	t.Run("show baseline empty", func(t *testing.T) {
		err := runBaseline([]string{})
		if err != nil {
			t.Errorf("runBaseline() error = %v", err)
		}
	})

	t.Run("add entry", func(t *testing.T) {
		err := runBaseline([]string{"add", "test.js", "10", "Test reason"})
		if err != nil {
			t.Errorf("runBaseline(add) error = %v", err)
		}
	})

	t.Run("show with entries", func(t *testing.T) {
		err := runBaseline([]string{})
		if err != nil {
			t.Errorf("runBaseline() error = %v", err)
		}
	})

	t.Run("remove entry", func(t *testing.T) {
		// Add an entry first
		_ = runBaseline([]string{"add", "test2.js", "20", "Reason"})
		err := runBaseline([]string{"remove", "1"})
		if err != nil {
			t.Errorf("runBaseline(remove) error = %v", err)
		}
	})

	t.Run("clear baseline", func(t *testing.T) {
		_ = runBaseline([]string{"add", "test3.js", "30", "Reason"})
		err := runBaseline([]string{"clear"})
		if err != nil {
			t.Errorf("runBaseline(clear) error = %v", err)
		}
	})

	t.Run("export baseline", func(t *testing.T) {
		_ = runBaseline([]string{"add", "test4.js", "40", "Reason"})
		err := runBaseline([]string{"export", "baseline.json"})
		if err != nil {
			t.Errorf("runBaseline(export) error = %v", err)
		}
		if _, err := os.Stat("baseline.json"); os.IsNotExist(err) {
			t.Error("Expected baseline.json to be created")
		}
	})

	t.Run("import baseline", func(t *testing.T) {
		err := runBaseline([]string{"import", "baseline.json"})
		if err != nil {
			t.Errorf("runBaseline(import) error = %v", err)
		}
	})

	t.Run("help", func(t *testing.T) {
		err := runBaseline([]string{"help"})
		if err != nil {
			t.Errorf("runBaseline(help) error = %v", err)
		}
	})

	t.Run("unknown command", func(t *testing.T) {
		err := runBaseline([]string{"unknown"})
		if err == nil {
			t.Error("Expected error for unknown command")
		}
	})
}

func TestAddToBaseline(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	os.MkdirAll(".sentinel", 0755)

	t.Run("with reason", func(t *testing.T) {
		err := addToBaseline([]string{"test.js", "10", "Custom reason"})
		if err != nil {
			t.Errorf("addToBaseline() error = %v", err)
		}
	})

	t.Run("without reason", func(t *testing.T) {
		err := addToBaseline([]string{"test2.js", "20"})
		if err != nil {
			t.Errorf("addToBaseline() error = %v", err)
		}
	})

	t.Run("invalid args", func(t *testing.T) {
		err := addToBaseline([]string{"test.js"})
		if err == nil {
			t.Error("Expected error for insufficient args")
		}
	})

	t.Run("invalid line number", func(t *testing.T) {
		err := addToBaseline([]string{"test.js", "invalid"})
		if err == nil {
			t.Error("Expected error for invalid line number")
		}
	})
}

func TestRemoveFromBaseline(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	os.MkdirAll(".sentinel", 0755)

	// Setup: add some entries
	_ = addToBaseline([]string{"test.js", "10", "Reason 1"})
	_ = addToBaseline([]string{"test.js", "20", "Reason 2"})
	_ = addToBaseline([]string{"test.js", "30", "Reason 3"})

	t.Run("valid index", func(t *testing.T) {
		err := removeFromBaseline([]string{"2"})
		if err != nil {
			t.Errorf("removeFromBaseline() error = %v", err)
		}
	})

	t.Run("invalid index", func(t *testing.T) {
		err := removeFromBaseline([]string{"999"})
		if err == nil {
			t.Error("Expected error for out of range index")
		}
	})

	t.Run("invalid args", func(t *testing.T) {
		err := removeFromBaseline([]string{})
		if err == nil {
			t.Error("Expected error for missing args")
		}
	})

	t.Run("invalid index format", func(t *testing.T) {
		err := removeFromBaseline([]string{"invalid"})
		if err == nil {
			t.Error("Expected error for invalid index format")
		}
	})
}

func TestClearBaseline(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	os.MkdirAll(".sentinel", 0755)

	t.Run("clear existing", func(t *testing.T) {
		_ = addToBaseline([]string{"test.js", "10", "Reason"})
		err := clearBaseline()
		if err != nil {
			t.Errorf("clearBaseline() error = %v", err)
		}
	})

	t.Run("clear empty", func(t *testing.T) {
		err := clearBaseline()
		if err != nil {
			t.Errorf("clearBaseline() on empty error = %v", err)
		}
	})
}

func TestShowBaseline(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	os.MkdirAll(".sentinel", 0755)

	t.Run("empty baseline", func(t *testing.T) {
		err := showBaseline()
		if err != nil {
			t.Errorf("showBaseline() error = %v", err)
		}
	})

	t.Run("with entries", func(t *testing.T) {
		_ = addToBaseline([]string{"test.js", "10", "Test reason"})
		err := showBaseline()
		if err != nil {
			t.Errorf("showBaseline() error = %v", err)
		}
	})
}
