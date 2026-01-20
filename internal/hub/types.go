// Package hub defines types for Hub API communication
// Complies with CODING_STANDARDS.md: Type definitions max 200 lines
package hub

import "time"

// ASTAnalysisRequest is sent to Hub for AST analysis
type ASTAnalysisRequest struct {
	Code     string   `json:"code"`
	Language string   `json:"language"`
	Filepath string   `json:"filepath,omitempty"`
	Analyses []string `json:"analyses"` // duplicate_functions, orphaned_code, unused_vars, etc.
}

// ASTAnalysisResponse is returned from Hub
type ASTAnalysisResponse struct {
	Findings []ASTFinding  `json:"findings"`
	Stats    AnalysisStats `json:"stats"`
}

// ASTFinding represents an issue found via AST analysis
type ASTFinding struct {
	Type        string `json:"type"` // duplicate_function, orphaned_code, etc.
	Message     string `json:"message"`
	File        string `json:"file"`
	Line        int    `json:"line"`
	Column      int    `json:"column"`
	Severity    string `json:"severity"` // error, warning, info
	Suggestion  string `json:"suggestion,omitempty"`
	CodeSnippet string `json:"code_snippet,omitempty"`
}

// AnalysisStats contains statistics about the analysis
type AnalysisStats struct {
	Duration      time.Duration `json:"duration"`
	LinesAnalyzed int           `json:"lines_analyzed"`
	FilesAnalyzed int           `json:"files_analyzed"`
	Method        string        `json:"method"` // ast or pattern
}

// VibeAnalysisRequest is sent to Hub for vibe coding detection
type VibeAnalysisRequest struct {
	CodebasePath string   `json:"codebase_path"`
	Files        []string `json:"files,omitempty"` // Specific files to analyze
	DeepAnalysis bool     `json:"deep_analysis"`   // Enable cross-file analysis
}

// VibeAnalysisResponse is returned from Hub
type VibeAnalysisResponse struct {
	Issues  []VibeIssue `json:"issues"`
	Summary VibeSummary `json:"summary"`
}

// VibeIssue represents a vibe coding issue
type VibeIssue struct {
	Type         string   `json:"type"` // duplicate_function, orphaned_code, unused_variable, etc.
	Description  string   `json:"description"`
	File         string   `json:"file"`
	Line         int      `json:"line"`
	Severity     string   `json:"severity"`
	Suggestion   string   `json:"suggestion"`
	RelatedFiles []string `json:"related_files,omitempty"` // For cross-file issues
}

// VibeSummary summarizes vibe issues
type VibeSummary struct {
	TotalIssues      int            `json:"total_issues"`
	IssuesByType     map[string]int `json:"issues_by_type"`
	IssuesBySeverity map[string]int `json:"issues_by_severity"`
	FilesWithIssues  int            `json:"files_with_issues"`
}

// StructureAnalysisRequest is sent to Hub for file structure analysis
type StructureAnalysisRequest struct {
	File     string `json:"file"`
	Language string `json:"language"`
	Content  string `json:"content"`
}

// StructureAnalysisResponse is returned from Hub
type StructureAnalysisResponse struct {
	File             string            `json:"file"`
	Lines            int               `json:"lines"`
	Status           string            `json:"status"` // ok, warning, critical
	Sections         []Section         `json:"sections"`
	SplitSuggestions []SplitSuggestion `json:"split_suggestions"`
	Complexity       int               `json:"complexity"`
}

// Section represents a logical section of a file
type Section struct {
	Name      string `json:"name"`
	Type      string `json:"type"` // function, class, interface, etc.
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
	Lines     int    `json:"lines"`
}

// SplitSuggestion suggests how to split an oversized file
type SplitSuggestion struct {
	TargetFile string   `json:"target_file"`
	Sections   []string `json:"sections"` // Section names to move
	Rationale  string   `json:"rationale"`
}

// HookPolicy defines policy for git hook execution
type HookPolicy struct {
	AuditEnabled         bool `json:"audit_enabled"`
	VibeCheckEnabled     bool `json:"vibe_check_enabled"`
	SecurityCheckEnabled bool `json:"security_check_enabled"`
	FileSizeCheckEnabled bool `json:"file_size_check_enabled"`
	AllowOverride        bool `json:"allow_override"`
	MaxOverridesPerDay   int  `json:"max_overrides_per_day"`
	MaxOverridesPerWeek  int  `json:"max_overrides_per_week"`
}

// TelemetryData is sent to Hub for metrics tracking
type TelemetryData struct {
	EventType string                 `json:"event_type"` // audit, fix, hook, etc.
	Timestamp time.Time              `json:"timestamp"`
	ProjectID string                 `json:"project_id,omitempty"`
	UserID    string                 `json:"user_id,omitempty"`
	Metadata  map[string]interface{} `json:"metadata"`
	Duration  time.Duration          `json:"duration,omitempty"`
	Success   bool                   `json:"success"`
	ErrorMsg  string                 `json:"error_msg,omitempty"`
}
