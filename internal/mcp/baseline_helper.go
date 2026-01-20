// Package mcp provides baseline helper to avoid circular dependencies
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package mcp

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// BaselineEntry represents an accepted finding
type BaselineEntry struct {
	Pattern string    `json:"pattern"`
	File    string    `json:"file"`
	Line    int       `json:"line"`
	Reason  string    `json:"reason"`
	AddedBy string    `json:"added_by"`
	AddedAt time.Time `json:"added_at"`
	Hash    string    `json:"hash"`
}

// Baseline represents the complete baseline file
type Baseline struct {
	Version string          `json:"version"`
	Entries []BaselineEntry `json:"entries"`
}

// addToBaselineFile adds an entry to the baseline
func addToBaselineFile(file string, line int, reason string) error {
	baselinePath := ".sentinel/baseline.json"

	// Load existing baseline
	baseline, err := loadBaseline(baselinePath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to load baseline: %w", err)
	}

	// Get current user
	user := os.Getenv("USER")
	if user == "" {
		user = "mcp-server"
	}

	// Create new entry
	entry := BaselineEntry{
		Pattern: "manual_baseline",
		File:    file,
		Line:    line,
		Reason:  reason,
		AddedBy: user,
		AddedAt: time.Now(),
		Hash:    fmt.Sprintf("%s:%d", file, line),
	}

	// Add to baseline
	baseline.Entries = append(baseline.Entries, entry)

	// Save baseline
	if err := saveBaseline(baselinePath, baseline); err != nil {
		return fmt.Errorf("failed to save baseline: %w", err)
	}

	return nil
}

// loadBaseline loads the baseline from disk
func loadBaseline(path string) (*Baseline, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return &Baseline{Version: "1.0", Entries: []BaselineEntry{}}, err
	}

	var baseline Baseline
	if err := json.Unmarshal(data, &baseline); err != nil {
		return nil, err
	}

	return &baseline, nil
}

// saveBaseline saves the baseline to disk
func saveBaseline(path string, baseline *Baseline) error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(baseline, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
