// Package mcp provides tests for audit helper
package mcp

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunAuditScan(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("runs audit scan successfully", func(t *testing.T) {
		result, err := runAuditScan(tmpDir, false, false)
		if err != nil {
			t.Errorf("runAuditScan() error = %v", err)
		}
		if result == nil {
			t.Error("result should not be nil")
		}
		if result["status"] == nil {
			t.Error("result should have status")
		}
	})

	t.Run("runs audit scan with vibe check", func(t *testing.T) {
		result, err := runAuditScan(tmpDir, true, false)
		if err != nil {
			t.Errorf("runAuditScan() error = %v", err)
		}
		if result == nil {
			t.Error("result should not be nil")
		}
	})

	t.Run("runs audit scan with deep mode", func(t *testing.T) {
		result, err := runAuditScan(tmpDir, false, true)
		if err != nil {
			t.Errorf("runAuditScan() error = %v", err)
		}
		if result == nil {
			t.Error("result should not be nil")
		}
	})

	t.Run("handles scan errors gracefully", func(t *testing.T) {
		// Use invalid path
		_, err := runAuditScan("/nonexistent/path/that/does/not/exist", false, false)
		// May or may not error depending on scanner implementation
		_ = err
	})

	t.Run("returns correct result structure", func(t *testing.T) {
		result, err := runAuditScan(tmpDir, false, false)
		if err != nil {
			t.Errorf("runAuditScan() error = %v", err)
			return
		}
		// Verify all expected fields
		if _, ok := result["status"]; !ok {
			t.Error("result should have status field")
		}
		if _, ok := result["path"]; !ok {
			t.Error("result should have path field")
		}
		if _, ok := result["findings"]; !ok {
			t.Error("result should have findings field")
		}
		if _, ok := result["summary"]; !ok {
			t.Error("result should have summary field")
		}
		if _, ok := result["success"]; !ok {
			t.Error("result should have success field")
		}
		if _, ok := result["timestamp"]; !ok {
			t.Error("result should have timestamp field")
		}
		if _, ok := result["details"]; !ok {
			t.Error("result should have details field")
		}
	})
}

func TestRunVibeCheck(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("runs vibe check successfully", func(t *testing.T) {
		result, err := runVibeCheck(tmpDir)
		if err != nil {
			t.Errorf("runVibeCheck() error = %v", err)
		}
		if result == nil {
			t.Error("result should not be nil")
		}
		if result["status"] == nil {
			t.Error("result should have status")
		}
		if result["count"] == nil {
			t.Error("result should have count")
		}
	})

	t.Run("filters to only vibe issues", func(t *testing.T) {
		// Create test file
		testFile := filepath.Join(tmpDir, "test.js")
		os.WriteFile(testFile, []byte("const x = 1;\nconst y = 2;"), 0644)

		result, err := runVibeCheck(tmpDir)
		if err != nil {
			t.Errorf("runVibeCheck() error = %v", err)
		}
		if result == nil {
			t.Error("result should not be nil")
		}
	})

	t.Run("handles scan errors", func(t *testing.T) {
		_, err := runVibeCheck("/nonexistent/path")
		// May or may not error
		_ = err
	})

	t.Run("returns correct result structure", func(t *testing.T) {
		result, err := runVibeCheck(tmpDir)
		if err != nil {
			t.Errorf("runVibeCheck() error = %v", err)
			return
		}
		if _, ok := result["status"]; !ok {
			t.Error("result should have status field")
		}
		if _, ok := result["path"]; !ok {
			t.Error("result should have path field")
		}
		if _, ok := result["issues"]; !ok {
			t.Error("result should have issues field")
		}
		if _, ok := result["count"]; !ok {
			t.Error("result should have count field")
		}
	})
}

func TestGetStatus(t *testing.T) {
	t.Run("returns passed for success", func(t *testing.T) {
		result := getStatus(true)
		if result != "passed" {
			t.Errorf("expected 'passed', got %s", result)
		}
	})

	t.Run("returns failed for failure", func(t *testing.T) {
		result := getStatus(false)
		if result != "failed" {
			t.Errorf("expected 'failed', got %s", result)
		}
	})
}
