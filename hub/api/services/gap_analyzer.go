// Gap Analysis Engine - Main Analysis Functions
// Identifies discrepancies between documented business rules and code implementation
// Complies with CODING_STANDARDS.md: Business Services max 400 lines

package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"sentinel-hub-api/pkg/database"
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
				Description:    fmt.Sprintf("Function '%s' implements business logic but is not documented as a business rule", pattern.FunctionName),
				Recommendation: fmt.Sprintf("Document function '%s' as a business rule in knowledge base", pattern.FunctionName),
				Severity:       severity,
			})
		}
	}

	return gaps, nil
}

// matchesDocumentedRule checks if a pattern matches any documented rule
func matchesDocumentedRule(pattern BusinessLogicPattern, documentedRules []KnowledgeItem) bool {
	patternKeyword := strings.ToLower(pattern.Keyword)
	patternFunction := strings.ToLower(pattern.FunctionName)

	for _, rule := range documentedRules {
		ruleTitle := strings.ToLower(rule.Title)
		ruleContent := strings.ToLower(rule.Content)

		// Check keyword match
		if patternKeyword != "" && (strings.Contains(ruleTitle, patternKeyword) || strings.Contains(ruleContent, patternKeyword)) {
			return true
		}

		// Check function name match
		if strings.Contains(ruleTitle, patternFunction) || strings.Contains(ruleContent, patternFunction) {
			return true
		}

		// Check semantic similarity
		similarity := semanticSimilarity(patternFunction, ruleTitle)
		if similarity > 0.5 {
			return true
		}
	}

	return false
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

// determineSeverityFromPattern determines severity based on pattern characteristics
func determineSeverityFromPattern(pattern BusinessLogicPattern) string {
	funcLower := strings.ToLower(pattern.FunctionName)
	criticalKeywords := []string{"payment", "transaction", "security", "auth", "validate", "check"}

	for _, keyword := range criticalKeywords {
		if strings.Contains(funcLower, keyword) {
			return "high"
		}
	}

	return "medium"
}

// Helper functions
func determineSeverity(rule KnowledgeItem) string {
	// Determine severity based on rule content or title
	// Check for critical keywords in title or content
	content := strings.ToLower(rule.Title + " " + rule.Content)
	criticalKeywords := []string{"security", "payment", "transaction", "critical", "important"}

	for _, keyword := range criticalKeywords {
		if strings.Contains(content, keyword) {
			return "high"
		}
	}

	return "medium"
}

func checkTestCoverage(ctx context.Context, knowledgeItemID string) (bool, error) {
	query := `
		SELECT COUNT(*) 
		FROM test_requirements 
		WHERE knowledge_item_id = $1
	`
	var count int
	err := database.QueryRowWithTimeout(ctx, db, query, knowledgeItemID).Scan(&count)
	if err != nil {
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
	err = database.QueryRowWithTimeout(queryCtx, db, query, report.ProjectID, string(gapsJSON), string(summaryJSON)).Scan(&reportID)
	if err != nil {
		return "", fmt.Errorf("failed to store gap report: %w", err)
	}

	return reportID, nil
}
