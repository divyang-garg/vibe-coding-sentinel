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
