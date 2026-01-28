// Package ast provides language support interfaces for dynamic language integration
// Complies with CODING_STANDARDS.md: Type definitions max 200 lines
package ast

import (
	sitter "github.com/smacker/go-tree-sitter"
)

// LanguageDetector defines language-specific detection capabilities
type LanguageDetector interface {
	// Security middleware detection
	DetectSecurityMiddleware(root *sitter.Node, code string) []ASTFinding

	// Code quality detections
	DetectUnused(root *sitter.Node, code string) []ASTFinding
	DetectDuplicates(root *sitter.Node, code string) []ASTFinding
	DetectUnreachable(root *sitter.Node, code string) []ASTFinding
	DetectAsync(root *sitter.Node, code string) []ASTFinding

	// Security vulnerability detections
	DetectSQLInjection(root *sitter.Node, code string) []SecurityVulnerability
	DetectXSS(root *sitter.Node, code string) []SecurityVulnerability
	DetectCommandInjection(root *sitter.Node, code string) []SecurityVulnerability
	DetectCrypto(root *sitter.Node, code string) []SecurityVulnerability
}

// LanguageExtractor defines language-specific extraction capabilities
type LanguageExtractor interface {
	ExtractFunctions(code, keyword string) ([]FunctionInfo, error)
	ExtractImports(code string) ([]ImportInfo, error)
	ExtractSymbols(root *sitter.Node, code string) (map[string]*Symbol, error)
}

// LanguageNodeTypes defines language-specific AST node types
type LanguageNodeTypes struct {
	FunctionDeclaration []string
	MethodDeclaration   []string
	VariableDeclaration []string
	ClassDeclaration    []string
	ImportStatement     []string
}

// LanguageSupport provides complete language support
// This is the main interface that language implementations must satisfy
type LanguageSupport interface {
	GetLanguage() string
	GetDetector() LanguageDetector
	GetExtractor() LanguageExtractor
	GetNodeTypes() LanguageNodeTypes
}
