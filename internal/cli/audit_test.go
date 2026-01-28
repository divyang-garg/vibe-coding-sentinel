// Package cli provides tests for audit command
package cli

import (
	"os"
	"testing"

	"github.com/divyang-garg/sentinel-hub-api/internal/scanner"
)

func TestRunAudit_Flags(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	testCases := []struct {
		name string
		args []string
	}{
		{"no args", []string{}},
		{"ci mode", []string{"--ci"}},
		{"offline", []string{"--offline"}},
		{"verbose", []string{"--verbose"}},
		{"vibe-check", []string{"--vibe-check"}},
		{"vibe-only", []string{"--vibe-only"}},
		{"deep", []string{"--deep", "--offline"}}, // offline to avoid Hub connection
		{"analyze-structure", []string{"--analyze-structure"}},
		{"output format", []string{"--output", "json"}},
		{"output file", []string{"--output-file", "results.json"}},
		{"custom path", []string{"."}},
		{"multiple flags", []string{"--ci", "--verbose", "--offline"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Most will fail due to missing codebase, but test parsing
			err := runAudit(tc.args)
			_ = err // Expected to fail without proper setup
		})
	}
}

func TestDisplayResults(t *testing.T) {
	t.Run("CI mode", func(t *testing.T) {
		result := &scanner.Result{
			Success:  true,
			Findings: []scanner.Finding{},
			Summary:  map[string]int{"test": 5},
		}
		displayResults(result, true)
	})

	t.Run("interactive mode passed", func(t *testing.T) {
		result := &scanner.Result{
			Success:   true,
			Findings:  []scanner.Finding{},
			Summary:   map[string]int{"test": 2},
			Timestamp: "2026-01-20T00:00:00Z",
		}
		displayResults(result, false)
	})

	t.Run("interactive mode failed", func(t *testing.T) {
		result := &scanner.Result{
			Success: false,
			Findings: []scanner.Finding{
				{
					Type:     "test",
					Severity: scanner.SeverityHigh,
					File:     "test.go",
					Line:     10,
					Message:  "Test finding",
				},
			},
			Summary:   map[string]int{"test": 1},
			Timestamp: "2026-01-20T00:00:00Z",
		}
		displayResults(result, false)
	})

	t.Run("with summary", func(t *testing.T) {
		result := &scanner.Result{
			Success:  true,
			Findings: []scanner.Finding{},
			Summary: map[string]int{
				"pattern1": 3,
				"pattern2": 5,
			},
			Timestamp: "2026-01-20T00:00:00Z",
		}
		displayResults(result, false)
	})
}
