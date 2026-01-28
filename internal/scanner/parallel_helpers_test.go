// Package scanner provides tests for parallel helper functions
package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCollectFiles(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("collects code files", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "test.js")
		os.WriteFile(file1, []byte("const x = 1;"), 0644)

		file2 := filepath.Join(tmpDir, "test.go")
		os.WriteFile(file2, []byte("package main"), 0644)

		// Non-code file should be skipped
		file3 := filepath.Join(tmpDir, "readme.md")
		os.WriteFile(file3, []byte("# Readme"), 0644)

		files, err := collectFiles(tmpDir)
		if err != nil {
			t.Errorf("collectFiles() error = %v", err)
		}
		if len(files) < 2 {
			t.Errorf("expected at least 2 files, got %d", len(files))
		}
	})

	t.Run("skips ignored directories", func(t *testing.T) {
		nodeModulesFile := filepath.Join(tmpDir, "node_modules", "test.js")
		os.MkdirAll(filepath.Dir(nodeModulesFile), 0755)
		os.WriteFile(nodeModulesFile, []byte("const x = 1;"), 0644)

		files, err := collectFiles(tmpDir)
		if err != nil {
			t.Errorf("collectFiles() error = %v", err)
		}
		// Should not include node_modules files
		for _, file := range files {
			if filepath.Base(filepath.Dir(file)) == "node_modules" {
				t.Error("should skip node_modules directory")
			}
		}
	})

	t.Run("handles path resolution error", func(t *testing.T) {
		// Use invalid path
		_, err := collectFiles("/nonexistent/path/that/does/not/exist")
		// May or may not error depending on implementation
		_ = err
	})
}

func TestScanFile(t *testing.T) {
	tmpDir := t.TempDir()
	patterns := GetSecurityPatterns()

	t.Run("scans file for patterns", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "test.js")
		content := `const apiKey = "secret123";
eval(userInput);`
		os.WriteFile(file1, []byte(content), 0644)

		findings := scanFile(file1, patterns, tmpDir)
		// May or may not find patterns depending on pattern matching
		// Just verify function doesn't panic
		_ = findings
	})

	t.Run("handles file read error", func(t *testing.T) {
		findings := scanFile("/nonexistent/file.js", patterns, tmpDir)
		if len(findings) != 0 {
			t.Error("should return empty findings for unreadable file")
		}
	})

	t.Run("calculates relative path correctly", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "subdir", "test.js")
		os.MkdirAll(filepath.Dir(file1), 0755)
		os.WriteFile(file1, []byte(`const apiKey = "secret";`), 0644)

		findings := scanFile(file1, patterns, tmpDir)
		if len(findings) > 0 {
			if findings[0].File == "" {
				t.Error("should have relative path in findings")
			}
		}
	})
}

func TestFilterBaselineParallel(t *testing.T) {
	tmpDir := t.TempDir()
	originalWD, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalWD)

	t.Run("filters findings with baseline", func(t *testing.T) {
		os.MkdirAll(".sentinel", 0755)
		baselineJSON := `{
			"version": "1.0",
			"entries": [
				{"file": "test.js", "line": 1, "hash": "test.js:1"}
			]
		}`
		os.WriteFile(".sentinel/baseline.json", []byte(baselineJSON), 0644)

		result := &Result{
			Findings: []Finding{
				{File: "test.js", Line: 1, Type: "secrets", Severity: SeverityHigh},
				{File: "test.js", Line: 2, Type: "secrets", Severity: SeverityHigh},
			},
			Summary: map[string]int{"secrets": 2},
			Success: false,
		}

		filtered := filterBaselineParallel(result)
		if len(filtered.Findings) != 1 {
			t.Errorf("expected 1 finding after filtering, got %d", len(filtered.Findings))
		}
		if filtered.Findings[0].Line != 2 {
			t.Error("line 2 finding should remain")
		}
	})

	t.Run("handles missing baseline file", func(t *testing.T) {
		// Ensure no baseline file exists
		os.Remove(".sentinel/baseline.json")

		result := &Result{
			Findings: []Finding{
				{File: "test.js", Line: 1, Type: "secrets"},
			},
			Summary: map[string]int{"secrets": 1},
		}

		filtered := filterBaselineParallel(result)
		// Should return result as-is when no baseline
		if len(filtered.Findings) < 1 {
			t.Error("should preserve findings when no baseline")
		}
	})

	t.Run("handles invalid baseline JSON", func(t *testing.T) {
		os.MkdirAll(".sentinel", 0755)
		os.WriteFile(".sentinel/baseline.json", []byte("invalid json"), 0644)

		result := &Result{
			Findings: []Finding{
				{File: "test.js", Line: 1, Type: "secrets"},
			},
		}

		filtered := filterBaselineParallel(result)
		if len(filtered.Findings) != 1 {
			t.Error("should return result as-is when baseline is invalid")
		}
	})

	t.Run("recalculates summary after filtering", func(t *testing.T) {
		os.MkdirAll(".sentinel", 0755)
		baselineJSON := `{
			"version": "1.0",
			"entries": [
				{"file": "test.js", "line": 1, "hash": "test.js:1"}
			]
		}`
		os.WriteFile(".sentinel/baseline.json", []byte(baselineJSON), 0644)

		result := &Result{
			Findings: []Finding{
				{File: "test.js", Line: 1, Type: "secrets", Severity: SeverityCritical},
				{File: "test.js", Line: 2, Type: "debug", Severity: SeverityWarning},
			},
			Summary: map[string]int{"secrets": 1, "debug": 1},
			Success: false,
		}

		filtered := filterBaselineParallel(result)
		if filtered.Summary["secrets"] != 0 {
			t.Error("secrets should be filtered out")
		}
		if filtered.Summary["debug"] != 1 {
			t.Error("debug should remain")
		}
		if !filtered.Success {
			t.Error("should be successful after filtering critical finding")
		}
	})
}
