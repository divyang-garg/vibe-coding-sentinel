// Package cli provides extended tests for history command
package cli

import (
	"os"
	"testing"
	"time"
)

func TestRunHistory_AllPaths(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	os.MkdirAll(".sentinel", 0755)

	t.Run("with JSON output", func(t *testing.T) {
		// Create some history
		history := &AuditHistory{
			Entries: []AuditHistoryEntry{
				{
					Timestamp:     time.Now(),
					Success:       true,
					TotalFindings: 5,
					ByType:        map[string]int{"test": 5},
					BySeverity:    map[string]int{"high": 2},
					Duration:      time.Second,
				},
			},
		}
		_ = saveHistory(history)

		err := runHistory([]string{"--json"})
		if err != nil {
			t.Errorf("runHistory() with JSON error = %v", err)
		}
	})

	t.Run("with last flag", func(t *testing.T) {
		history := &AuditHistory{
			Entries: []AuditHistoryEntry{
				{Timestamp: time.Now(), Success: true, TotalFindings: 1},
				{Timestamp: time.Now(), Success: true, TotalFindings: 2},
				{Timestamp: time.Now(), Success: true, TotalFindings: 3},
			},
		}
		_ = saveHistory(history)

		err := runHistory([]string{"--last", "2"})
		if err != nil {
			t.Errorf("runHistory() with --last error = %v", err)
		}
	})

	t.Run("with short last flag", func(t *testing.T) {
		history := &AuditHistory{
			Entries: []AuditHistoryEntry{
				{Timestamp: time.Now(), Success: true, TotalFindings: 1},
			},
		}
		_ = saveHistory(history)

		err := runHistory([]string{"-n", "1"})
		if err != nil {
			t.Errorf("runHistory() with -n flag error = %v", err)
		}
	})

	t.Run("with trend calculation", func(t *testing.T) {
		history := &AuditHistory{
			Entries: []AuditHistoryEntry{
				{
					Timestamp:     time.Now().Add(-time.Hour),
					Success:       true,
					TotalFindings: 10,
				},
				{
					Timestamp:     time.Now(),
					Success:       true,
					TotalFindings: 15,
				},
			},
		}
		_ = saveHistory(history)

		err := runHistory([]string{})
		if err != nil {
			t.Errorf("runHistory() with trend error = %v", err)
		}
	})

	t.Run("with decreasing trend", func(t *testing.T) {
		history := &AuditHistory{
			Entries: []AuditHistoryEntry{
				{
					Timestamp:     time.Now().Add(-time.Hour),
					Success:       true,
					TotalFindings: 20,
				},
				{
					Timestamp:     time.Now(),
					Success:       true,
					TotalFindings: 10,
				},
			},
		}
		_ = saveHistory(history)

		err := runHistory([]string{})
		if err != nil {
			t.Errorf("runHistory() with decreasing trend error = %v", err)
		}
	})

	t.Run("with no change trend", func(t *testing.T) {
		history := &AuditHistory{
			Entries: []AuditHistoryEntry{
				{
					Timestamp:     time.Now().Add(-time.Hour),
					Success:       true,
					TotalFindings: 10,
				},
				{
					Timestamp:     time.Now(),
					Success:       true,
					TotalFindings: 10,
				},
			},
		}
		_ = saveHistory(history)

		err := runHistory([]string{})
		if err != nil {
			t.Errorf("runHistory() with no change trend error = %v", err)
		}
	})

	t.Run("with byType display", func(t *testing.T) {
		history := &AuditHistory{
			Entries: []AuditHistoryEntry{
				{
					Timestamp:     time.Now(),
					Success:       true,
					TotalFindings: 5,
					ByType: map[string]int{
						"pattern1": 2,
						"pattern2": 3,
					},
				},
			},
		}
		_ = saveHistory(history)

		err := runHistory([]string{})
		if err != nil {
			t.Errorf("runHistory() with byType error = %v", err)
		}
	})

	t.Run("with failed audit", func(t *testing.T) {
		history := &AuditHistory{
			Entries: []AuditHistoryEntry{
				{
					Timestamp:     time.Now(),
					Success:       false,
					TotalFindings: 10,
				},
			},
		}
		_ = saveHistory(history)

		err := runHistory([]string{})
		if err != nil {
			t.Errorf("runHistory() with failed audit error = %v", err)
		}
	})

	t.Run("last flag without value", func(t *testing.T) {
		err := runHistory([]string{"--last"})
		// Should handle missing value gracefully
		_ = err
	})
}

func TestAddToHistory(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	os.MkdirAll(".sentinel", 0755)

	t.Run("add successful audit", func(t *testing.T) {
		err := AddToHistory(true, 5, map[string]int{"test": 5}, map[string]int{"high": 2}, time.Second)
		if err != nil {
			t.Errorf("AddToHistory() error = %v", err)
		}
	})

	t.Run("add failed audit", func(t *testing.T) {
		err := AddToHistory(false, 10, map[string]int{"test": 10}, map[string]int{"critical": 1}, time.Second*2)
		if err != nil {
			t.Errorf("AddToHistory() error = %v", err)
		}
	})

	t.Run("truncate to 100 entries", func(t *testing.T) {
		// Add more than 100 entries
		for i := 0; i < 105; i++ {
			_ = AddToHistory(true, i, map[string]int{"test": i}, map[string]int{"low": i}, time.Second)
		}

		history, _ := loadHistory()
		if len(history.Entries) > 100 {
			t.Errorf("Expected max 100 entries, got %d", len(history.Entries))
		}
	})
}

func TestLoadHistory(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	os.MkdirAll(".sentinel", 0755)

	t.Run("new file", func(t *testing.T) {
		history, err := loadHistory()
		if err == nil {
			t.Error("Expected error when loading non-existent file")
		}
		if history == nil {
			t.Error("Expected non-nil history even on error")
		}
	})

	t.Run("corrupted file", func(t *testing.T) {
		os.WriteFile(".sentinel/audit-history.json", []byte("invalid json"), 0644)
		history, err := loadHistory()
		if err == nil {
			t.Error("Expected error when loading corrupted file")
		}
		if history != nil {
			t.Error("Expected nil history on parse error")
		}
	})
}

func TestSaveHistory(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	t.Run("creates directory", func(t *testing.T) {
		history := &AuditHistory{
			Entries: []AuditHistoryEntry{},
		}
		err := saveHistory(history)
		if err != nil {
			t.Errorf("saveHistory() error = %v", err)
		}
	})

	t.Run("saves with entries", func(t *testing.T) {
		history := &AuditHistory{
			Entries: []AuditHistoryEntry{
				{
					Timestamp:     time.Now(),
					Success:       true,
					TotalFindings: 5,
				},
			},
		}
		err := saveHistory(history)
		if err != nil {
			t.Errorf("saveHistory() error = %v", err)
		}
	})
}
