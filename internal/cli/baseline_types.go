// Package cli provides baseline type definitions
// Complies with CODING_STANDARDS.md: Type definitions max 200 lines
package cli

import "time"

// BaselineEntry represents an accepted finding
type BaselineEntry struct {
	Pattern string    `json:"pattern"`
	File    string    `json:"file"`
	Line    int       `json:"line"`
	Reason  string    `json:"reason"`
	AddedBy string    `json:"added_by"`
	AddedAt time.Time `json:"added_at"`
	Hash    string    `json:"hash"` // Hash of finding for exact matching
}

// Baseline represents the complete baseline file
type Baseline struct {
	Version string          `json:"version"`
	Entries []BaselineEntry `json:"entries"`
}
