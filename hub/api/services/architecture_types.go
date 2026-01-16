// Package architecture_types - Data structures for architecture analysis
// Complies with CODING_STANDARDS.md: Models max 200 lines

package services

import ()

// FileContent represents a file to be analyzed
type FileContent struct {
	Path     string `json:"path"`
	Content  string `json:"content"`
	Language string `json:"language"`
}

// ArchitectureAnalysisRequest represents a request for architecture analysis
type ArchitectureAnalysisRequest struct {
	Files []FileContent `json:"files"`
}

// ArchitectureAnalysisResponse represents the response from architecture analysis
type ArchitectureAnalysisResponse struct {
	OversizedFiles   []FileAnalysisResult `json:"oversizedFiles"`
	ModuleGraph      ModuleGraph          `json:"moduleGraph"`
	DependencyIssues []DependencyIssue    `json:"dependencyIssues"`
	Recommendations  []string             `json:"recommendations"`
}

// FileAnalysisResult represents analysis result for a single file
type FileAnalysisResult struct {
	File            string           `json:"file"`
	Lines           int              `json:"lines"`
	Status          string           `json:"status"` // ok, warning, critical, oversized
	Sections        []FileSection    `json:"sections,omitempty"`
	SplitSuggestion *SplitSuggestion `json:"splitSuggestion,omitempty"`
}

// FileSection represents a logical section within a file
type FileSection struct {
	StartLine   int    `json:"startLine"`
	EndLine     int    `json:"endLine"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Lines       int    `json:"lines"`
}

// SplitSuggestion represents a suggestion for splitting a file
type SplitSuggestion struct {
	Reason                string         `json:"reason"`
	ProposedFiles         []ProposedFile `json:"proposedFiles"`
	MigrationInstructions []string       `json:"migrationInstructions"` // Text instructions only, not executable
	EstimatedEffort       string         `json:"estimatedEffort"`
}

// ProposedFile represents a proposed file in a split suggestion
type ProposedFile struct {
	Path     string   `json:"path"`
	Lines    int      `json:"lines"`
	Contents []string `json:"contents"` // Function/class names to move
}

// ModuleGraph represents the module dependency graph
type ModuleGraph struct {
	Nodes []ModuleNode `json:"nodes"`
	Edges []ModuleEdge `json:"edges"`
}

// ModuleNode represents a node in the module graph
type ModuleNode struct {
	Path  string `json:"path"`
	Lines int    `json:"lines"`
	Type  string `json:"type"` // component, service, utility, etc.
}

// ModuleEdge represents an edge in the module graph
type ModuleEdge struct {
	From string `json:"from"`
	To   string `json:"to"`
	Type string `json:"type"` // import, extends, implements
}

// DependencyIssue represents a dependency issue found in the codebase
type DependencyIssue struct {
	Type        string   `json:"type"` // circular, tight_coupling, god_module
	Severity    string   `json:"severity"`
	Files       []string `json:"files"`
	Description string   `json:"description"`
	Suggestion  string   `json:"suggestion"`
}
