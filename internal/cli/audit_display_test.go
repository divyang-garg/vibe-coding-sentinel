// Package cli provides additional tests for audit display functions
package cli

import (
	"testing"

	"github.com/divyang-garg/sentinel-hub-api/internal/scanner"
)

func TestDisplayResults_DetailedScenarios(t *testing.T) {
	t.Run("with findings and summary", func(t *testing.T) {
		result := &scanner.Result{
			Success: false,
			Findings: []scanner.Finding{
				{
					Type:     "pattern1",
					Severity: scanner.SeverityCritical,
					File:     "test1.go",
					Line:     10,
					Message:  "Critical issue",
				},
				{
					Type:     "pattern2",
					Severity: scanner.SeverityHigh,
					File:     "test2.go",
					Line:     20,
					Message:  "High severity issue",
				},
			},
			Summary: map[string]int{
				"pattern1": 1,
				"pattern2": 1,
			},
			Timestamp: "2026-01-20T00:00:00Z",
		}
		displayResults(result, false)
	})

	t.Run("empty findings", func(t *testing.T) {
		result := &scanner.Result{
			Success:   true,
			Findings:  []scanner.Finding{},
			Summary:   map[string]int{},
			Timestamp: "2026-01-20T00:00:00Z",
		}
		displayResults(result, false)
	})

	t.Run("multiple severity levels", func(t *testing.T) {
		result := &scanner.Result{
			Success: false,
			Findings: []scanner.Finding{
				{Type: "test", Severity: scanner.SeverityCritical, File: "test.go", Line: 1},
				{Type: "test", Severity: scanner.SeverityHigh, File: "test.go", Line: 2},
				{Type: "test", Severity: scanner.SeverityMedium, File: "test.go", Line: 3},
				{Type: "test", Severity: scanner.SeverityLow, File: "test.go", Line: 4},
				{Type: "test", Severity: scanner.SeverityWarning, File: "test.go", Line: 5},
			},
			Summary:   map[string]int{"test": 5},
			Timestamp: "2026-01-20T00:00:00Z",
		}
		displayResults(result, false)
	})
}
