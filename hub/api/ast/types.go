// Package ast provides AST analysis types
// Complies with CODING_STANDARDS.md: Type definitions max 200 lines
package ast

import (
	"sync"
	"time"

	sitter "github.com/smacker/go-tree-sitter"
)

// ASTFinding represents a single finding from AST analysis
type ASTFinding struct {
	Type       string `json:"type"` // duplicate_function, unused_variable, etc.
	Severity   string `json:"severity"`
	Line       int    `json:"line"`
	Column     int    `json:"column"`
	EndLine    int    `json:"endLine"`
	EndColumn  int    `json:"endColumn"`
	Message    string `json:"message"`
	Code       string `json:"code"`
	Suggestion string `json:"suggestion"`
	// AutoFix metadata for vibe coding
	Confidence  float64 `json:"confidence"`  // 0.0-1.0 probability
	AutoFixSafe bool    `json:"autoFixSafe"` // Safe for auto-refactor
	FixType     string  `json:"fixType"`     // "delete", "refactor", "comment"
	Reasoning   string  `json:"reasoning"`   // Explanation for decision
	Validated   bool    `json:"validated"`   // Was codebase validation run
}

// AnalysisStats tracks performance metrics for AST analysis
type AnalysisStats struct {
	ParseTime    int64
	AnalysisTime int64
	NodesVisited int
}

// LanguageParser maps language names to Tree-sitter parsers
type LanguageParser struct {
	Language string
	Parser   *sitter.Parser
}

// AST cache for performance optimization
type cacheEntry struct {
	Findings []ASTFinding
	Stats    AnalysisStats
	Expires  time.Time
}

var astCache = make(map[string]*cacheEntry)
var cacheMutex sync.RWMutex
var cacheTTL = 5 * time.Minute // Cache AST results for 5 minutes
var lastCacheCleanup time.Time
var cacheCleanupInterval = 5 * time.Minute

// DetectionConfig controls which patterns are excluded from detection
type DetectionConfig struct {
	ExcludedFunctions []string // Function names to never flag
	ExcludedPrefixes  []string // Prefixes to exclude (Test*, Example*)
	TrustExported     bool     // Skip exported (uppercase) functions
}

// DefaultConfig returns production-safe defaults
func DefaultConfig() DetectionConfig {
	return DetectionConfig{
		ExcludedFunctions: []string{"main", "init"},
		ExcludedPrefixes:  []string{"Test", "Example", "Benchmark"},
		TrustExported:     true,
	}
}

// Scope represents a lexical scope for variable tracking
type Scope struct {
	Parent   *Scope
	Name     string
	Symbols  map[string]*Symbol
	StartPos uint32
	EndPos   uint32
}

// Symbol represents a declared identifier
type Symbol struct {
	Name       string
	DeclNode   *sitter.Node
	UsageCount int
	Position   uint32
	Kind       string // "variable", "function", "parameter"
}

// ScopeStack manages nested scopes during AST traversal
type ScopeStack struct {
	Current *Scope
}

// NewScopeStack creates a new scope stack with a root scope
func NewScopeStack() *ScopeStack {
	return &ScopeStack{
		Current: &Scope{Symbols: make(map[string]*Symbol)},
	}
}

// Push creates a new scope and makes it current
func (s *ScopeStack) Push(name string, start, end uint32) {
	newScope := &Scope{
		Parent:   s.Current,
		Name:     name,
		Symbols:  make(map[string]*Symbol),
		StartPos: start,
		EndPos:   end,
	}
	s.Current = newScope
}

// Pop returns to the parent scope
func (s *ScopeStack) Pop() {
	if s.Current.Parent != nil {
		s.Current = s.Current.Parent
	}
}

// SecurityVulnerability represents a security vulnerability found in code
type SecurityVulnerability struct {
	Type        string  `json:"type"`     // "sql_injection", "xss", "command_injection", etc.
	Severity    string  `json:"severity"` // "critical", "high", "medium", "low"
	Line        int     `json:"line"`
	Column      int     `json:"column"`
	Message     string  `json:"message"`
	Code        string  `json:"code,omitempty"`
	Description string  `json:"description"`
	Remediation string  `json:"remediation"`
	Confidence  float64 `json:"confidence"` // 0.0-1.0
	FilePath    string  `json:"file_path,omitempty"`
}

// FunctionInfo represents extracted function information
type FunctionInfo struct {
	Name         string            `json:"name"`
	Language     string            `json:"language"`
	Line         int               `json:"line"`
	Column       int               `json:"column"`
	EndLine      int               `json:"endLine"`
	EndColumn    int               `json:"endColumn"`
	Parameters   []ParameterInfo   `json:"parameters,omitempty"`
	ReturnType   string            `json:"returnType,omitempty"`
	Visibility   string            `json:"visibility"` // "public", "private", "exported"
	Documentation string           `json:"documentation,omitempty"`
	Code         string            `json:"code"` // Full function code
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// ParameterInfo represents a function parameter
type ParameterInfo struct {
	Name string `json:"name"`
	Type string `json:"type,omitempty"`
}
