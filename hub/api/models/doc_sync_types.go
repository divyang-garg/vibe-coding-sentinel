// Package doc_sync_types - Data structures for documentation synchronization
// Complies with CODING_STANDARDS.md: Models max 200 lines

package models

import (
	"time"
)

// =============================================================================
// DATA STRUCTURES
// =============================================================================

// StatusMarker represents a phase status marker extracted from documentation
type StatusMarker struct {
	Phase      string      `json:"phase"`
	Line       int         `json:"line"`
	Status     string      `json:"status"`      // "COMPLETE", "PENDING", "STUB", "PARTIAL"
	StatusIcon string      `json:"status_icon"` // "✅", "⏳", "🔴", "⚠️"
	Tasks      []PhaseTask `json:"tasks"`
	FilePath   string      `json:"file_path"`
}

// PhaseTask represents a task within a phase (renamed to avoid collision with Phase 14E Task)
type PhaseTask struct {
	Name     string `json:"name"`
	Days     string `json:"days"`
	Status   string `json:"status"` // "Done", "Pending", etc.
	Line     int    `json:"line"`
	Priority string `json:"priority,omitempty"`
}

// ImplementationEvidence represents evidence of code implementation
type ImplementationEvidence struct {
	Feature     string           `json:"feature"`
	Files       []string         `json:"files"`
	Functions   []string         `json:"functions"`
	Endpoints   []string         `json:"endpoints"`
	Tests       []string         `json:"tests"`
	Confidence  float64          `json:"confidence"`   // 0.0 to 1.0
	LineNumbers map[string][]int `json:"line_numbers"` // file -> line numbers
}

// Discrepancy represents a discrepancy between documentation and code
type Discrepancy struct {
	Type           string                 `json:"type"` // "status_mismatch", "missing_impl", "missing_doc", "partial_match"
	Phase          string                 `json:"phase"`
	Feature        string                 `json:"feature,omitempty"`
	DocStatus      string                 `json:"doc_status"`
	CodeStatus     string                 `json:"code_status"`
	Evidence       ImplementationEvidence `json:"evidence"`
	Recommendation string                 `json:"recommendation"`
	FilePath       string                 `json:"file_path"`
	LineNumber     int                    `json:"line_number"`
}

// DocSyncReport represents a complete doc-sync report
type DocSyncReport struct {
	ID            string        `json:"id"`
	ProjectID     string        `json:"project_id"`
	ReportType    string        `json:"report_type"` // "status_tracking", "business_rules", "all"
	InSync        []InSyncItem  `json:"in_sync"`
	Discrepancies []Discrepancy `json:"discrepancies"`
	Summary       SummaryStats  `json:"summary"`
	CreatedAt     time.Time     `json:"created_at"`
}

// InSyncItem represents an item that is in sync
type InSyncItem struct {
	Phase    string `json:"phase"`
	Status   string `json:"status"`
	Evidence string `json:"evidence"`
}

// SummaryStats represents summary statistics
type SummaryStats struct {
	TotalPhases       int     `json:"total_phases"`
	InSyncCount       int     `json:"in_sync_count"`
	DiscrepancyCount  int     `json:"discrepancy_count"`
	AverageConfidence float64 `json:"average_confidence"`
}

// DocSyncRequest represents a request for doc-sync analysis
type DocSyncRequest struct {
	ProjectID  string                 `json:"project_id"`
	ReportType string                 `json:"report_type"` // "status_tracking", "business_rules", "all"
	Options    map[string]interface{} `json:"options,omitempty"`
}

// DocSyncResponse represents the response from doc-sync analysis
type DocSyncResponse struct {
	Success       bool          `json:"success"`
	ReportID      string        `json:"report_id,omitempty"`
	InSync        []InSyncItem  `json:"in_sync"`
	Discrepancies []Discrepancy `json:"discrepancies"`
	Summary       SummaryStats  `json:"summary"`
	Message       string        `json:"message,omitempty"`
}
