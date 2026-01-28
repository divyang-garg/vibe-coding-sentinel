// Gap Analysis Engine - Pattern Extraction Functions
// Extracts business logic patterns from codebase
// Complies with CODING_STANDARDS.md: Business Services max 400 lines

package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sentinel-hub-api/ast"
)

// BusinessLogicPattern represents a business logic pattern found in code
type BusinessLogicPattern struct {
	FilePath     string `json:"file_path"`
	FunctionName string `json:"function_name"`
	Keyword      string `json:"keyword"`
	LineNumber   int    `json:"line_number"`
	Signature    string `json:"signature,omitempty"`
}

// extractBusinessLogicPatterns extracts business logic patterns from codebase
// Uses AST analysis for accurate function detection and line numbers
// Deprecated: Use extractBusinessLogicPatternsEnhanced for context support
func extractBusinessLogicPatterns(codebasePath string) ([]BusinessLogicPattern, error) {
	// Use AST-based extraction for better accuracy
	return extractBusinessLogicPatternsAST(codebasePath)
}

// extractBusinessLogicPatternsEnhanced extracts business logic patterns using comprehensive AST analysis
// with context support for cancellation and timeout handling
func extractBusinessLogicPatternsEnhanced(ctx context.Context, codebasePath string) ([]BusinessLogicPattern, error) {
	// Check context cancellation before starting
	if ctx.Err() != nil {
		return []BusinessLogicPattern{}, ctx.Err()
	}

	// Validate codebase path exists
	if _, err := os.Stat(codebasePath); os.IsNotExist(err) {
		return []BusinessLogicPattern{}, fmt.Errorf("codebase path does not exist: %s", codebasePath)
	}

	var patterns []BusinessLogicPattern

	// Map file extensions to AST language strings
	extToLang := map[string]string{
		".go": "go",
		".js": "javascript",
		".ts": "typescript",
		".py": "python",
	}

	// Walk codebase and analyze each file
	err := filepath.Walk(codebasePath, func(path string, info os.FileInfo, err error) error {
		// Check context cancellation in loop
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if err != nil {
			LogWarn(ctx, "Error walking path %s: %v", path, err)
			return nil // Continue processing other files
		}

		// Skip non-code files
		if info.IsDir() || !isCodeFile(path) {
			return nil
		}

		// Determine language from extension
		ext := strings.ToLower(filepath.Ext(path))
		language, ok := extToLang[ext]
		if !ok {
			return nil // Skip unsupported languages
		}

		// Read file content
		content, err := os.ReadFile(path)
		if err != nil {
			LogWarn(ctx, "Failed to read file %s: %v", path, err)
			return nil // Continue processing other files
		}

		// Use AST.AnalyzeAST for comprehensive analysis
		analyses := []string{"duplicates", "orphaned", "unused"}
		findings, stats, err := ast.AnalyzeAST(string(content), language, analyses)
		if err != nil {
			LogWarn(ctx, "AST analysis failed for %s: %v, falling back to simple extraction", path, err)
			// Fallback to simple extraction
			filePatterns := extractPatternsFromCodeFallback(ctx, path, string(content), language)
			patterns = append(patterns, filePatterns...)
			return nil
		}

		// Extract functions using AST
		functions, err := ast.ExtractFunctions(string(content), language, "")
		if err != nil {
			LogWarn(ctx, "Function extraction failed for %s: %v, falling back to simple extraction", path, err)
			// Fallback to simple extraction
			filePatterns := extractPatternsFromCodeFallback(ctx, path, string(content), language)
			patterns = append(patterns, filePatterns...)
			return nil
		}

		// Convert to BusinessLogicPattern with enhanced detection
		for _, fn := range functions {
			pattern := convertToBusinessPattern(ctx, path, fn, findings, stats)
			if pattern != nil && isBusinessLogicPattern(*pattern) {
				patterns = append(patterns, *pattern)
			}
		}

		return nil
	})

	if err != nil {
		LogError(ctx, "Failed to walk codebase %s: %v", codebasePath, err)
		return patterns, fmt.Errorf("failed to extract business logic patterns: %w", err)
	}

	return patterns, nil
}

// convertToBusinessPattern converts AST FunctionInfo to BusinessLogicPattern with enhanced classification
func convertToBusinessPattern(ctx context.Context, filePath string, fn ast.FunctionInfo, findings []ast.ASTFinding, stats ast.AnalysisStats) *BusinessLogicPattern {
	// Check if function contains business keywords
	funcNameLower := strings.ToLower(fn.Name)
	keyword := extractKeywordFromFunctionName(funcNameLower)

	// If no keyword in name, check function code
	if keyword == "" {
		keyword = extractKeywordFromFunctionName(strings.ToLower(fn.Code))
	}

	// Classify pattern type
	patternType := classifyBusinessPattern(ctx, fn, findings, stats)

	// Only include if it has business keywords or is classified as business logic
	if keyword != "" || containsBusinessKeywordsInName(funcNameLower) || patternType != "general" {
		return &BusinessLogicPattern{
			FilePath:     filePath,
			FunctionName: fn.Name,
			Keyword:      keyword,
			LineNumber:   fn.Line,
			Signature:    fn.Code, // Full function code
		}
	}

	return nil
}

// classifyBusinessPattern classifies a pattern based on AST analysis and function characteristics
func classifyBusinessPattern(ctx context.Context, fn ast.FunctionInfo, findings []ast.ASTFinding, stats ast.AnalysisStats) string {
	// Check context cancellation
	if ctx.Err() != nil {
		return "general"
	}

	funcNameLower := strings.ToLower(fn.Name)

	// Check for CRUD operations
	crudKeywords := []string{"create", "update", "delete", "get", "fetch", "save", "remove"}
	for _, keyword := range crudKeywords {
		if strings.Contains(funcNameLower, keyword) {
			return "crud_operation"
		}
	}

	// Check for validation patterns
	validationKeywords := []string{"validate", "check", "verify", "ensure"}
	for _, keyword := range validationKeywords {
		if strings.Contains(funcNameLower, keyword) {
			return "validation"
		}
	}

	// Check for business workflow patterns
	workflowKeywords := []string{"process", "execute", "handle", "workflow"}
	for _, keyword := range workflowKeywords {
		if strings.Contains(funcNameLower, keyword) {
			return "workflow"
		}
	}

	// Check AST findings for business logic indicators
	for _, finding := range findings {
		// If finding indicates business logic (e.g., complex control flow, data transformations)
		if finding.Type == "business_logic" || finding.Severity == "high" {
			return "business_logic"
		}
	}

	// Use stats to influence classification - high node count suggests complex business logic
	if stats.NodesVisited > 50 {
		LogDebug(ctx, "Function %s classified as business_logic based on high node count (%d)", fn.Name, stats.NodesVisited)
		return "business_logic"
	}

	return "general"
}

// extractPatternsFromCodeFallback extracts patterns using fallback method when AST fails
func extractPatternsFromCodeFallback(ctx context.Context, filePath, code, language string) []BusinessLogicPattern {
	// Check context cancellation
	if ctx.Err() != nil {
		return []BusinessLogicPattern{}
	}

	// Use language-specific extraction if available, otherwise use simple pattern matching
	// For now, use simple extraction but log language for debugging
	LogDebug(ctx, "Using fallback pattern extraction for %s file (language: %s)", filePath, language)
	return extractBusinessLogicPatternsSimple(filePath, code)
}

// matchesPatternToRule checks if a pattern matches a documented rule using AST evidence
// Uses detectBusinessRuleImplementation results for accurate matching
func matchesPatternToRule(ctx context.Context, pattern BusinessLogicPattern, rule KnowledgeItem, evidence ImplementationEvidence) bool {
	// High confidence match (> 0.7) indicates strong match
	if evidence.Confidence > 0.7 {
		LogDebug(ctx, "Pattern %s matches rule %s with high confidence %.2f", pattern.FunctionName, rule.Title, evidence.Confidence)
		return true
	}

	// Check function name match
	patternFuncLower := strings.ToLower(pattern.FunctionName)
	ruleTitleLower := strings.ToLower(rule.Title)
	ruleContentLower := strings.ToLower(rule.Content)

	// Check title match
	if strings.Contains(ruleTitleLower, patternFuncLower) || strings.Contains(patternFuncLower, ruleTitleLower) {
		LogDebug(ctx, "Pattern %s matches rule %s by function name in title", pattern.FunctionName, rule.Title)
		return true
	}

	// Check content match
	if strings.Contains(ruleContentLower, patternFuncLower) || strings.Contains(patternFuncLower, ruleContentLower) {
		LogDebug(ctx, "Pattern %s matches rule %s by function name in content", pattern.FunctionName, rule.Title)
		return true
	}

	// Check semantic similarity with higher threshold for better accuracy
	// Use simple string matching as fallback (semantic similarity would require NLP library)
	// For now, check if key words match
	patternWords := strings.Fields(patternFuncLower)
	ruleWords := strings.Fields(ruleTitleLower)
	commonWords := 0
	for _, pw := range patternWords {
		for _, rw := range ruleWords {
			if pw == rw && len(pw) > 3 { // Only count words longer than 3 chars
				commonWords++
				break
			}
		}
	}
	if len(patternWords) > 0 && len(ruleWords) > 0 {
		similarity := float64(commonWords) / float64(len(patternWords)+len(ruleWords)-commonWords)
		if similarity > 0.3 { // Lower threshold for word matching
			LogDebug(ctx, "Pattern %s matches rule %s by word similarity %.2f", pattern.FunctionName, rule.Title, similarity)
			return true
		}
	}

	// Check if pattern is in evidence functions (from detectBusinessRuleImplementation)
	for _, funcName := range evidence.Functions {
		if strings.EqualFold(funcName, pattern.FunctionName) {
			LogDebug(ctx, "Pattern %s matches rule %s via implementation evidence", pattern.FunctionName, rule.Title)
			return true
		}
	}

	// Check if pattern file is in evidence files
	for _, file := range evidence.Files {
		if strings.Contains(file, pattern.FilePath) || strings.Contains(pattern.FilePath, file) {
			LogDebug(ctx, "Pattern %s matches rule %s via file evidence", pattern.FunctionName, rule.Title)
			return true
		}
	}

	return false
}

// extractBusinessLogicPatternsAST extracts business logic patterns using AST analysis
func extractBusinessLogicPatternsAST(codebasePath string) ([]BusinessLogicPattern, error) {
	var patterns []BusinessLogicPattern

	// Map file extensions to AST language strings
	extToLang := map[string]string{
		".go": "go",
		".js": "javascript",
		".ts": "typescript",
		".py": "python",
	}

	// Walk codebase and collect code files
	err := filepath.Walk(codebasePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip non-code files
		if info.IsDir() || !isCodeFile(path) {
			return nil
		}

		// Determine language from extension
		ext := strings.ToLower(filepath.Ext(path))
		language, ok := extToLang[ext]
		if !ok {
			return nil // Skip unsupported languages
		}

		// Read file content
		content, err := os.ReadFile(path)
		if err != nil {
			return nil // Skip files that can't be read
		}

		// Use AST analyzer to extract function definitions
		// We parse the code directly using AST to get accurate function definitions
		// If AST analysis fails, fall back to simple pattern matching
		filePatterns := extractPatternsFromCode(path, string(content), language)
		patterns = append(patterns, filePatterns...)

		return nil
	})

	return patterns, err
}

// extractPatternsFromCode extracts function definitions directly from code using AST
// Falls back to pattern matching if AST extraction fails or no patterns found
func extractPatternsFromCode(filePath, code string, language string) []BusinessLogicPattern {
	var patterns []BusinessLogicPattern

	// Try AST extraction first
	functions, err := ast.ExtractFunctions(code, language, "")
	if err == nil && len(functions) > 0 {
		// Convert FunctionInfo to BusinessLogicPattern
		for _, fn := range functions {
			// Check if function contains business keywords
			funcNameLower := strings.ToLower(fn.Name)
			keyword := extractKeywordFromFunctionName(funcNameLower)

			// If no keyword in name, check function code
			if keyword == "" {
				keyword = extractKeywordFromFunctionName(strings.ToLower(fn.Code))
			}

			// Only include if it has business keywords
			if keyword != "" || containsBusinessKeywordsInName(funcNameLower) {
				patterns = append(patterns, BusinessLogicPattern{
					FilePath:     filePath,
					FunctionName: fn.Name,
					Keyword:      keyword,
					LineNumber:   fn.Line,
					Signature:    fn.Code, // Full function code
				})
			}
		}

		if len(patterns) > 0 {
			return patterns
		}
	}

	// Fallback to pattern matching if AST fails or no patterns found
	return extractBusinessLogicPatternsSimple(filePath, code)
}

// extractKeywordFromFunctionName extracts keyword from function name
func extractKeywordFromFunctionName(name string) string {
	keywords := []string{"order", "payment", "user", "account", "transaction", "rule", "validate", "check"}
	nameLower := strings.ToLower(name)
	for _, keyword := range keywords {
		if strings.Contains(nameLower, keyword) {
			return keyword
		}
	}
	return ""
}

// containsBusinessKeywordsInName checks if function name contains business keywords
func containsBusinessKeywordsInName(name string) bool {
	keywords := []string{"order", "payment", "user", "account", "transaction", "rule", "validate", "check", "process", "create", "update", "delete"}
	nameLower := strings.ToLower(name)
	for _, keyword := range keywords {
		if strings.Contains(nameLower, keyword) {
			return true
		}
	}
	return false
}

// extractBusinessLogicPatternsSimple is a fallback for when AST analysis fails
func extractBusinessLogicPatternsSimple(path, content string) []BusinessLogicPattern {
	// Fallback to simple pattern matching if AST fails
	var patterns []BusinessLogicPattern
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if strings.Contains(line, "func ") && containsBusinessKeywords(line) {
			funcName := extractFunctionNameGap(line)
			if funcName != "" {
				patterns = append(patterns, BusinessLogicPattern{
					FilePath:     path,
					FunctionName: funcName,
					Keyword:      extractKeyword(line),
					LineNumber:   i + 1,
					Signature:    "",
				})
			}
		}
	}
	return patterns
}

// isBusinessLogicPattern checks if a pattern represents business logic
func isBusinessLogicPattern(pattern BusinessLogicPattern) bool {
	// Check function name for business keywords
	funcLower := strings.ToLower(pattern.FunctionName)
	businessKeywords := []string{"order", "payment", "user", "account", "transaction", "rule", "validate", "check", "process", "create", "update", "delete"}

	for _, keyword := range businessKeywords {
		if strings.Contains(funcLower, keyword) {
			return true
		}
	}

	return false
}

// extractKeywordFromFunction extracts a keyword from function name or content
func extractKeywordFromFunction(funcName, content string) string {
	keywords := []string{"order", "payment", "user", "account", "transaction", "rule", "validate", "check"}
	funcLower := strings.ToLower(funcName)

	// First check function name
	for _, keyword := range keywords {
		if strings.Contains(funcLower, keyword) {
			return keyword
		}
	}

	// If not found in function name, check content
	if content != "" {
		contentLower := strings.ToLower(content)
		for _, keyword := range keywords {
			if strings.Contains(contentLower, keyword) {
				return keyword
			}
		}
	}

	return ""
}

// semanticSimilarity calculates a simple similarity score between two strings
func semanticSimilarity(s1, s2 string) float64 {
	s1Lower := strings.ToLower(s1)
	s2Lower := strings.ToLower(s2)

	// Exact match
	if s1Lower == s2Lower {
		return 1.0
	}

	// Substring match
	if strings.Contains(s1Lower, s2Lower) || strings.Contains(s2Lower, s1Lower) {
		return 0.7
	}

	// Common words
	s1Words := strings.Fields(s1Lower)
	s2Words := strings.Fields(s2Lower)

	commonCount := 0
	for _, w1 := range s1Words {
		for _, w2 := range s2Words {
			if w1 == w2 && len(w1) > 2 {
				commonCount++
			}
		}
	}

	if len(s1Words) == 0 || len(s2Words) == 0 {
		return 0.0
	}

	return float64(commonCount) / float64(len(s1Words)+len(s2Words)-commonCount)
}
