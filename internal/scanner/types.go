// Package scanner provides security scanning types
// Complies with CODING_STANDARDS.md: Type definitions max 200 lines
package scanner

// Severity represents the severity level of a finding
type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityMedium   Severity = "medium"
	SeverityLow      Severity = "low"
	SeverityWarning  Severity = "warning"
	SeverityInfo     Severity = "info"
)

// Finding represents a single security finding
type Finding struct {
	Type     string   `json:"type"`
	Severity Severity `json:"severity"`
	File     string   `json:"file"`
	Line     int      `json:"line"`
	Column   int      `json:"column,omitempty"`
	Message  string   `json:"message"`
	Pattern  string   `json:"pattern,omitempty"`
	Code     string   `json:"code,omitempty"`
}

// Result contains all scan findings and summary
type Result struct {
	Success   bool           `json:"success"`
	Findings  []Finding      `json:"findings"`
	Summary   map[string]int `json:"summary"`
	Timestamp string         `json:"timestamp"`
}

// ScanOptions configures the scan
type ScanOptions struct {
	CodebasePath     string
	CIMode           bool
	Verbose          bool
	VibeCheck        bool
	VibeOnly         bool
	Deep             bool
	AnalyzeStructure bool
	Offline          bool
}
