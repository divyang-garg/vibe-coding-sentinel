// Phase 12: Gap Analysis Engine
// Identifies discrepancies between documented business rules and code implementation

package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	sitter "github.com/smacker/go-tree-sitter"
)

// GapType represents the type of gap identified
// GapType, Gap, and GapAnalysisReport are defined in types.go

// analyzeGaps performs comprehensive gap analysis between business rules and code
func analyzeGaps(ctx context.Context, projectID string, codebasePath string, options map[string]interface{}) (*GapAnalysisReport, error) {
	// Check cache first
	if cached, ok := getCachedGapAnalysis(projectID, codebasePath); ok {
		LogInfo(ctx, "Returning cached gap analysis for project %s", projectID)
		return cached, nil
	}

	// Validate project ID
	if err := ValidateUUID(projectID); err != nil {
		return nil, fmt.Errorf("invalid project ID: %w", err)
	}

	// Validate project exists
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM projects WHERE id = $1)"
	err := db.QueryRowContext(ctx, query, projectID).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("failed to verify project: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("project not found: %s", projectID)
	}

	// Validate codebase path
	if err := ValidateDirectory(codebasePath); err != nil {
		return nil, fmt.Errorf("invalid codebase path: %w", err)
	}

	// Validate options
	if options == nil {
		options = make(map[string]interface{})
	}

	// Validate option types
	if includeTests, ok := options["includeTests"]; ok {
		if _, ok := includeTests.(bool); !ok {
			return nil, fmt.Errorf("includeTests must be a boolean")
		}
	}
	if reverseCheck, ok := options["reverseCheck"]; ok {
		if _, ok := reverseCheck.(bool); !ok {
			return nil, fmt.Errorf("reverseCheck must be a boolean")
		}
	}

	var gaps []Gap

	// Extract business rules from knowledge base
	rules, err := extractBusinessRules(ctx, projectID, nil, "", nil)
	if err != nil {
		LogError(ctx, "Failed to extract business rules for project %s: %v", projectID, err)
		return nil, fmt.Errorf("failed to extract business rules for project %s: %w", projectID, err)
	}
	if rules == nil {
		rules = []KnowledgeItem{} // Ensure non-nil slice
	}

	includeTests := true
	if opt, ok := options["includeTests"].(bool); ok {
		includeTests = opt
	}

	// Analyze each rule for implementation gaps
	for _, rule := range rules {
		// Check code implementation
		evidence := detectBusinessRuleImplementation(rule, codebasePath)

		// Convert ImplementationEvidence to map[string]interface{}
		evidenceMap := map[string]interface{}{
			"feature":      evidence.Feature,
			"files":        evidence.Files,
			"functions":    evidence.Functions,
			"endpoints":    evidence.Endpoints,
			"tests":        evidence.Tests,
			"confidence":   evidence.Confidence,
			"line_numbers": evidence.LineNumbers,
		}

		if evidence.Confidence < 0.3 {
			// Rule documented but not implemented
			gaps = append(gaps, Gap{
				Type:            GapMissingImpl,
				KnowledgeItemID: rule.ID,
				RuleTitle:       rule.Title,
				Description:     fmt.Sprintf("Business rule '%s' is documented but not implemented in code", rule.Title),
				Evidence:        evidenceMap,
				Recommendation:  fmt.Sprintf("Implement business rule '%s' in code", rule.Title),
				Severity:        determineSeverity(rule),
			})
		} else if evidence.Confidence < 0.7 {
			// Partially implemented
			gaps = append(gaps, Gap{
				Type:            GapPartial,
				KnowledgeItemID: rule.ID,
				RuleTitle:       rule.Title,
				Description:     fmt.Sprintf("Business rule '%s' is partially implemented (confidence: %.2f%%)", rule.Title, evidence.Confidence*100),
				Evidence:        evidenceMap,
				Recommendation:  fmt.Sprintf("Complete implementation of business rule '%s'", rule.Title),
				Severity:        determineSeverity(rule),
			})
		}

		// Check test coverage if requested
		if includeTests {
			hasTests, err := checkTestCoverage(ctx, rule.ID)
			if err != nil {
				LogWarn(ctx, "Failed to check test coverage for rule %s (project: %s): %v", rule.ID, projectID, err)
				// Continue but mark as unknown
				gaps = append(gaps, Gap{
					Type:            GapTestsMissing,
					KnowledgeItemID: rule.ID,
					RuleTitle:       rule.Title,
					Description:     fmt.Sprintf("Unable to verify test coverage for '%s' (error: %v)", rule.Title, err),
					Recommendation:  "Manually verify test coverage",
					Severity:        "low",
				})
				continue
			}

			if !hasTests {
				gaps = append(gaps, Gap{
					Type:            GapTestsMissing,
					KnowledgeItemID: rule.ID,
					RuleTitle:       rule.Title,
					Description:     fmt.Sprintf("Business rule '%s' has no test coverage", rule.Title),
					Recommendation:  fmt.Sprintf("Add tests for business rule '%s'", rule.Title),
					Severity:        determineSeverity(rule),
				})
			}
		}
	}

	// Reverse check: find code patterns not documented as rules
	if reverseCheck, ok := options["reverseCheck"].(bool); ok && reverseCheck {
		undocumentedGaps, err := analyzeUndocumentedCode(ctx, projectID, codebasePath, rules)
		if err != nil {
			LogError(ctx, "Failed to analyze undocumented code (project: %s, path: %s): %v", projectID, codebasePath, err)
			// Continue with documented gaps only
		} else {
			gaps = append(gaps, undocumentedGaps...)
		}
	}

	// Generate summary
	summary := generateGapSummary(gaps)

	report := &GapAnalysisReport{
		ProjectID: projectID,
		Gaps:      gaps,
		Summary:   summary,
		CreatedAt: getCurrentTimestamp(),
	}

	// Store in cache for future requests
	setCachedGapAnalysis(projectID, codebasePath, report)
	LogInfo(ctx, "Cached gap analysis for project %s (path: %s)", projectID, codebasePath)

	return report, nil
}

// analyzeUndocumentedCode finds code patterns not documented as business rules
func analyzeUndocumentedCode(ctx context.Context, projectID string, codebasePath string, documentedRules []KnowledgeItem) ([]Gap, error) {
	var gaps []Gap

	// Use AST analyzer to extract business logic patterns from code
	// This is a simplified version - full implementation would use Phase 6 AST analysis
	patterns, err := extractBusinessLogicPatterns(codebasePath)
	if err != nil {
		return nil, fmt.Errorf("failed to extract business logic patterns: %w", err)
	}

	// Compare patterns against documented rules using improved matching
	for _, pattern := range patterns {
		if !matchesDocumentedRule(pattern, documentedRules) {
			severity := determineSeverityFromPattern(pattern)
			gaps = append(gaps, Gap{
				Type:           GapMissingDoc,
				RuleTitle:      pattern.FunctionName,
				FilePath:       pattern.FilePath,
				LineNumber:     pattern.LineNumber,
				Description:    fmt.Sprintf("Function '%s' at %s:%d is not documented as a business rule", pattern.FunctionName, pattern.FilePath, pattern.LineNumber),
				Recommendation: fmt.Sprintf("Document business rule for '%s' or remove if not needed", pattern.FunctionName),
				Severity:       severity,
			})
		}
	}

	return gaps, nil
}

// matchesDocumentedRule checks if a pattern matches any documented rule
func matchesDocumentedRule(pattern BusinessLogicPattern, documentedRules []KnowledgeItem) bool {
	funcLower := strings.ToLower(pattern.FunctionName)

	for _, rule := range documentedRules {
		ruleTitleLower := strings.ToLower(rule.Title)
		ruleContentLower := strings.ToLower(rule.Content)

		// Check if function name appears in rule title or content
		if strings.Contains(ruleTitleLower, funcLower) || strings.Contains(ruleContentLower, funcLower) {
			return true
		}

		// Check if keyword matches
		if pattern.Keyword != "" {
			if strings.Contains(ruleTitleLower, strings.ToLower(pattern.Keyword)) ||
				strings.Contains(ruleContentLower, strings.ToLower(pattern.Keyword)) {
				return true
			}
		}

		// Semantic similarity check (simple fuzzy matching)
		if semanticSimilarity(funcLower, ruleTitleLower) > 0.7 {
			return true
		}
	}

	return false
}

// semanticSimilarity calculates a simple similarity score between two strings
func semanticSimilarity(s1, s2 string) float64 {
	// Simple word-based similarity
	words1 := strings.Fields(s1)
	words2 := strings.Fields(s2)

	if len(words1) == 0 || len(words2) == 0 {
		return 0.0
	}

	matches := 0
	for _, w1 := range words1 {
		for _, w2 := range words2 {
			if w1 == w2 {
				matches++
				break
			}
		}
	}

	return float64(matches) / float64(len(words1))
}

// determineSeverityFromPattern determines severity based on pattern characteristics
func determineSeverityFromPattern(pattern BusinessLogicPattern) string {
	funcLower := strings.ToLower(pattern.FunctionName)

	// Critical keywords
	criticalKeywords := []string{"payment", "transaction", "security", "auth", "password"}
	for _, keyword := range criticalKeywords {
		if strings.Contains(funcLower, keyword) {
			return "critical"
		}
	}

	// High priority keywords
	highKeywords := []string{"order", "user", "account", "validate"}
	for _, keyword := range highKeywords {
		if strings.Contains(funcLower, keyword) {
			return "high"
		}
	}

	return "medium"
}

// BusinessLogicPattern represents a business logic pattern found in code
type BusinessLogicPattern struct {
	FilePath     string
	FunctionName string
	Keyword      string
	LineNumber   int    // Line number from AST (accurate)
	Signature    string // Function signature
}

// extractBusinessLogicPatterns extracts business logic patterns from codebase
// Uses AST analysis for accurate function detection and line numbers
func extractBusinessLogicPatterns(codebasePath string) ([]BusinessLogicPattern, error) {
	// Use AST-based extraction for better accuracy
	return extractBusinessLogicPatternsAST(codebasePath)
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
func extractPatternsFromCode(filePath, code string, language string) []BusinessLogicPattern {
	var patterns []BusinessLogicPattern

	// Get parser for language
	parser, err := getParser(language)
	if err != nil {
		return patterns
	}

	// Parse code into AST
	ctx := context.Background()
	tree, err := parser.ParseCtx(ctx, nil, []byte(code))
	if err != nil {
		return patterns
	}

	if tree == nil {
		return patterns
	}
	defer tree.Close()

	rootNode := tree.RootNode()
	if rootNode == nil {
		return patterns
	}

	// Traverse AST to find function definitions
	traverseAST(rootNode, func(node *sitter.Node) bool {
		var funcName string
		var isFunction bool
		var startLine int

		switch language {
		case "go":
			if node.Type() == "function_declaration" || node.Type() == "method_declaration" {
				startLine, _ = getLineColumn(code, int(node.StartByte()))
				for i := 0; i < int(node.ChildCount()); i++ {
					child := node.Child(i)
					if child != nil {
						if child.Type() == "identifier" {
							funcName = code[child.StartByte():child.EndByte()]
							isFunction = true
							break
						} else if child.Type() == "field_identifier" {
							// Method name in method_declaration
							funcName = code[child.StartByte():child.EndByte()]
							isFunction = true
							break
						}
					}
				}
			}
		case "javascript", "typescript":
			if node.Type() == "function_declaration" || node.Type() == "function" {
				startLine, _ = getLineColumn(code, int(node.StartByte()))
				for i := 0; i < int(node.ChildCount()); i++ {
					child := node.Child(i)
					if child != nil {
						if child.Type() == "identifier" || child.Type() == "property_identifier" {
							funcName = code[child.StartByte():child.EndByte()]
							isFunction = true
							break
						}
					}
				}
			} else if node.Type() == "arrow_function" {
				// Arrow functions assigned to variables
				parent := node.Parent()
				if parent != nil && parent.Type() == "variable_declarator" {
					startLine, _ = getLineColumn(code, int(node.StartByte()))
					for i := 0; i < int(parent.ChildCount()); i++ {
						child := parent.Child(i)
						if child != nil && child.Type() == "identifier" {
							funcName = code[child.StartByte():child.EndByte()]
							isFunction = true
							break
						}
					}
				}
			}
		case "python":
			if node.Type() == "function_definition" {
				startLine, _ = getLineColumn(code, int(node.StartByte()))
				for i := 0; i < int(node.ChildCount()); i++ {
					child := node.Child(i)
					if child != nil && child.Type() == "identifier" {
						funcName = code[child.StartByte():child.EndByte()]
						isFunction = true
						break
					}
				}
			}
		}

		if isFunction && funcName != "" {
			signature := extractFunctionSignature(node, code, language)
			keyword := extractKeywordFromFunction(funcName, code[node.StartByte():node.EndByte()])

			patterns = append(patterns, BusinessLogicPattern{
				FilePath:     filePath,
				FunctionName: funcName,
				Keyword:      keyword,
				LineNumber:   startLine,
				Signature:    signature,
			})
		}

		return true // Continue traversal
	})

	return patterns
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

	// Check signature/content for business keywords
	if strings.Contains(strings.ToLower(pattern.Signature), "order") ||
		strings.Contains(strings.ToLower(pattern.Signature), "payment") ||
		strings.Contains(strings.ToLower(pattern.Signature), "user") {
		return true
	}

	return false
}

// extractKeywordFromFunction extracts a keyword from function name or content
func extractKeywordFromFunction(funcName, content string) string {
	keywords := []string{"order", "payment", "user", "account", "transaction", "rule", "validate"}
	funcLower := strings.ToLower(funcName)
	contentLower := strings.ToLower(content)

	for _, keyword := range keywords {
		if strings.Contains(funcLower, keyword) || strings.Contains(contentLower, keyword) {
			return keyword
		}
	}
	return ""
}

// Helper functions
func determineSeverity(rule KnowledgeItem) string {
	// Determine severity based on rule priority or category
	if strings.Contains(strings.ToLower(rule.Content), "critical") {
		return "critical"
	}
	if strings.Contains(strings.ToLower(rule.Content), "high") {
		return "high"
	}
	return "medium"
}

func checkTestCoverage(ctx context.Context, knowledgeItemID string) (bool, error) {
	// Check if tests exist for this knowledge item using Phase 10 test coverage tracker
	query := `SELECT COUNT(*) FROM test_coverage WHERE knowledge_item_id = $1`
	var count int

	// Use the same pattern as in doc_sync.go
	ctx, cancel := context.WithTimeout(ctx, getQueryTimeout())
	defer cancel()

	err := queryRowWithTimeout(ctx, query, knowledgeItemID).Scan(&count)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("failed to check test coverage: %w", err)
	}
	return count > 0, nil
}

func generateGapSummary(gaps []Gap) map[string]interface{} {
	summary := make(map[string]interface{})
	summary["total"] = len(gaps)

	byType := make(map[GapType]int)
	bySeverity := make(map[string]int)

	for _, gap := range gaps {
		byType[gap.Type]++
		bySeverity[gap.Severity]++
	}

	summary["by_type"] = map[string]int{
		"missing_impl":  byType[GapMissingImpl],
		"missing_doc":   byType[GapMissingDoc],
		"partial_match": byType[GapPartial],
		"tests_missing": byType[GapTestsMissing],
	}

	summary["by_severity"] = bySeverity

	return summary
}

func isCodeFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".go" || ext == ".js" || ext == ".ts" || ext == ".py" || ext == ".java"
}

func containsBusinessKeywords(line string) bool {
	keywords := []string{"order", "payment", "user", "account", "transaction", "rule", "validate", "check"}
	lineLower := strings.ToLower(line)
	for _, keyword := range keywords {
		if strings.Contains(lineLower, keyword) {
			return true
		}
	}
	return false
}

func extractFunctionNameGap(line string) string {
	// Simple extraction - full implementation would use AST
	parts := strings.Fields(line)
	for i, part := range parts {
		if part == "func" && i+1 < len(parts) {
			funcName := strings.TrimSpace(parts[i+1])
			// Remove parameters
			if idx := strings.Index(funcName, "("); idx > 0 {
				return funcName[:idx]
			}
			return funcName
		}
	}
	return ""
}

func extractKeyword(line string) string {
	keywords := []string{"order", "payment", "user", "account", "transaction"}
	lineLower := strings.ToLower(line)
	for _, keyword := range keywords {
		if strings.Contains(lineLower, keyword) {
			return keyword
		}
	}
	return ""
}

func getCurrentTimestamp() string {
	return time.Now().Format(time.RFC3339)
}

// storeGapReport stores a gap analysis report in the database
// Returns the report ID or error
func storeGapReport(ctx context.Context, report *GapAnalysisReport) (string, error) {
	// Marshal gaps to JSON
	gapsJSON, err := json.Marshal(report.Gaps)
	if err != nil {
		return "", fmt.Errorf("failed to marshal gaps: %w", err)
	}

	// Marshal summary to JSON
	summaryJSON, err := json.Marshal(report.Summary)
	if err != nil {
		return "", fmt.Errorf("failed to marshal summary: %w", err)
	}

	// Insert into database using timeout helper
	query := `
		INSERT INTO gap_reports (project_id, gaps, summary, created_at)
		VALUES ($1, $2, $3, NOW())
		RETURNING id
	`

	queryCtx, cancel := context.WithTimeout(ctx, getQueryTimeout())
	defer cancel()

	var reportID string
	err = queryRowWithTimeout(queryCtx, query, report.ProjectID, string(gapsJSON), string(summaryJSON)).Scan(&reportID)
	if err != nil {
		return "", fmt.Errorf("failed to store gap report: %w", err)
	}

	return reportID, nil
}
