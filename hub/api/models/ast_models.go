// Package models - AST-specific data models
// Complies with CODING_STANDARDS.md: Data Models max 200 lines
package models

// ASTAnalysisRequest represents a request for single-file AST analysis
type ASTAnalysisRequest struct {
	Code      string   `json:"code" validate:"required"`
	Language  string   `json:"language" validate:"required"`
	Analyses  []string `json:"analyses,omitempty"` // e.g., ["duplicates", "unused", "unreachable"]
	FilePath  string   `json:"file_path,omitempty"`
}

// MultiFileASTRequest represents a request for multi-file AST analysis
type MultiFileASTRequest struct {
	Files       []FileInput `json:"files" validate:"required"`
	Analyses    []string    `json:"analyses,omitempty"`
	ProjectRoot string      `json:"project_root,omitempty"`
}

// FileInput represents a file to be analyzed
type FileInput struct {
	Path     string `json:"path" validate:"required"`
	Content  string `json:"content" validate:"required"`
	Language string `json:"language" validate:"required"`
}

// SecurityASTRequest represents a request for security-focused AST analysis
type SecurityASTRequest struct {
	Code     string   `json:"code" validate:"required"`
	Language string   `json:"language" validate:"required"`
	Severity string   `json:"severity,omitempty"` // "critical", "high", "medium", "low", "all"
}

// CrossFileASTRequest represents a request for cross-file dependency analysis
type CrossFileASTRequest struct {
	Files       []FileInput `json:"files" validate:"required"`
	ProjectRoot string      `json:"project_root,omitempty"`
}

// ASTAnalysisResponse represents the response from AST analysis
type ASTAnalysisResponse struct {
	Findings []ASTFinding      `json:"findings"`
	Stats    ASTAnalysisStats  `json:"stats"`
	Language string            `json:"language"`
	FilePath string            `json:"file_path,omitempty"`
}

// MultiFileASTResponse represents the response from multi-file AST analysis
type MultiFileASTResponse struct {
	Findings []ASTFinding      `json:"findings"`
	Stats    ASTAnalysisStats  `json:"stats"`
	Files    []string          `json:"files"`
}

// SecurityASTResponse represents the response from security AST analysis
type SecurityASTResponse struct {
	Findings    []ASTFinding     `json:"findings"`
	Stats       ASTAnalysisStats  `json:"stats"`
	Vulnerabilities []Vulnerability `json:"vulnerabilities"`
	RiskScore   float64          `json:"risk_score"`
}

// CrossFileASTResponse represents the response from cross-file analysis
type CrossFileASTResponse struct {
	Findings          []ASTFinding     `json:"findings"`
	UnusedExports     []ExportSymbol   `json:"unused_exports,omitempty"`
	UndefinedRefs     []SymbolRef      `json:"undefined_refs,omitempty"`
	CircularDeps      [][]string       `json:"circular_deps,omitempty"`
	CrossFileDuplicates []ASTFinding   `json:"cross_file_duplicates,omitempty"`
	Stats             CrossFileStats   `json:"stats"`
}

// ASTFinding represents a single finding from AST analysis
type ASTFinding struct {
	Type        string  `json:"type"`
	Severity    string  `json:"severity"`
	Line        int     `json:"line"`
	Column      int     `json:"column"`
	EndLine     int     `json:"end_line,omitempty"`
	EndColumn   int     `json:"end_column,omitempty"`
	Message     string  `json:"message"`
	Code        string  `json:"code,omitempty"`
	Suggestion  string  `json:"suggestion,omitempty"`
	Confidence  float64 `json:"confidence,omitempty"`
	AutoFixSafe bool    `json:"auto_fix_safe,omitempty"`
	FixType     string  `json:"fix_type,omitempty"`
	Reasoning   string  `json:"reasoning,omitempty"`
	FilePath    string  `json:"file_path,omitempty"`
}

// ASTAnalysisStats tracks performance metrics for AST analysis
type ASTAnalysisStats struct {
	ParseTime    int64 `json:"parse_time_ms"`
	AnalysisTime int64 `json:"analysis_time_ms"`
	NodesVisited int   `json:"nodes_visited"`
}

// Vulnerability represents a security vulnerability found
type Vulnerability struct {
	Type        string  `json:"type"` // "sql_injection", "xss", "command_injection", etc.
	Severity    string  `json:"severity"`
	Line        int     `json:"line"`
	Column      int     `json:"column"`
	Message     string  `json:"message"`
	Code        string  `json:"code,omitempty"`
	Description string  `json:"description"`
	Remediation string  `json:"remediation"`
	Confidence  float64 `json:"confidence"`
	FilePath    string  `json:"file_path,omitempty"`
}

// ExportSymbol represents an exported symbol
type ExportSymbol struct {
	Name     string `json:"name"`
	Kind     string `json:"kind"` // "function", "class", "variable", "type"
	FilePath string `json:"file_path"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
}

// SymbolRef represents a symbol reference
type SymbolRef struct {
	Name     string `json:"name"`
	FilePath string `json:"file_path"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
	Kind     string `json:"kind"` // "import", "call", "usage"
}

// CrossFileStats tracks cross-file analysis metrics
type CrossFileStats struct {
	FilesAnalyzed     int `json:"files_analyzed"`
	SymbolsFound      int `json:"symbols_found"`
	DependenciesFound int `json:"dependencies_found"`
	AnalysisTime      int64 `json:"analysis_time_ms"`
}

// SupportedAnalysesResponse lists supported languages and analyses
type SupportedAnalysesResponse struct {
	Languages []LanguageSupport `json:"languages"`
	Analyses  []string          `json:"analyses"`
}

// LanguageSupport describes support for a language
type LanguageSupport struct {
	Name      string   `json:"name"`
	Aliases   []string `json:"aliases,omitempty"`
	Supported bool     `json:"supported"`
}
