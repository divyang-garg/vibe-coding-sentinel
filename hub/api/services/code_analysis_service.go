// Package services provides code analysis business logic.
// Complies with CODING_STANDARDS.md: Business Services max 400 lines
package services

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"sentinel-hub-api/models"
)

// CodeAnalysisServiceImpl implements CodeAnalysisService
type CodeAnalysisServiceImpl struct {
	// Analysis patterns and rules
	patterns map[string]*regexp.Regexp
}

// NewCodeAnalysisService creates a new code analysis service instance
func NewCodeAnalysisService() CodeAnalysisService {
	service := &CodeAnalysisServiceImpl{
		patterns: make(map[string]*regexp.Regexp),
	}
	service.initializePatterns()
	return service
}

// AnalyzeCode performs comprehensive code analysis
func (s *CodeAnalysisServiceImpl) AnalyzeCode(ctx context.Context, req models.CodeAnalysisRequest) (interface{}, error) {
	if req.Code == "" {
		return nil, fmt.Errorf("code is required")
	}
	if req.Language == "" {
		return nil, fmt.Errorf("language is required")
	}

	analysis := map[string]interface{}{
		"language":      req.Language,
		"code_length":   len(req.Code),
		"lines_count":   strings.Count(req.Code, "\n") + 1,
		"complexity":    s.analyzeComplexity(req.Code, req.Language),
		"quality_score": s.calculateQualityScore(req.Code, req.Language),
		"issues":        s.identifyIssues(req.Code, req.Language),
		"suggestions":   s.generateSuggestions(req.Code, req.Language),
		"analyzed_at":   time.Now().Format(time.RFC3339),
	}

	return analysis, nil
}

// LintCode performs code linting
func (s *CodeAnalysisServiceImpl) LintCode(ctx context.Context, req models.CodeLintRequest) (interface{}, error) {
	if req.Code == "" {
		return nil, fmt.Errorf("code is required")
	}

	issues := s.identifyIssues(req.Code, req.Language)
	if len(req.Rules) > 0 {
		issues = s.filterIssuesByRules(issues, req.Rules)
	}

	return map[string]interface{}{
		"language":           req.Language,
		"issues":             issues,
		"issue_count":        len(issues),
		"severity_breakdown": s.calculateSeverityBreakdown(issues),
		"linted_at":          time.Now().Format(time.RFC3339),
	}, nil
}

// RefactorCode suggests code refactoring
func (s *CodeAnalysisServiceImpl) RefactorCode(ctx context.Context, req models.CodeRefactorRequest) (interface{}, error) {
	if req.Code == "" {
		return nil, fmt.Errorf("code is required")
	}

	suggestions := s.generateRefactoringSuggestions(req.Code, req.Language, req.Action)

	return map[string]interface{}{
		"language":          req.Language,
		"action":            req.Action,
		"suggestions":       suggestions,
		"confidence_score":  0.85,
		"estimated_savings": s.estimateRefactoringSavings(suggestions),
		"generated_at":      time.Now().Format(time.RFC3339),
	}, nil
}

// GenerateDocumentation generates code documentation
func (s *CodeAnalysisServiceImpl) GenerateDocumentation(ctx context.Context, req models.DocumentationRequest) (interface{}, error) {
	if req.Code == "" {
		return nil, fmt.Errorf("code is required")
	}

	docs := s.extractDocumentation(req.Code, req.Language)

	return map[string]interface{}{
		"language":      req.Language,
		"format":        req.Format,
		"documentation": docs,
		"coverage":      s.calculateDocumentationCoverage(docs, req.Code),
		"quality_score": s.assessDocumentationQuality(docs),
		"generated_at":  time.Now().Format(time.RFC3339),
	}, nil
}

// ValidateCode validates code syntax and structure
func (s *CodeAnalysisServiceImpl) ValidateCode(ctx context.Context, req models.CodeValidationRequest) (interface{}, error) {
	if req.Code == "" {
		return nil, fmt.Errorf("code is required")
	}

	validation := map[string]interface{}{
		"language":     req.Language,
		"is_valid":     s.validateSyntax(req.Code, req.Language),
		"errors":       s.findSyntaxErrors(req.Code, req.Language),
		"warnings":     s.findPotentialIssues(req.Code, req.Language),
		"compliance":   s.checkStandardsCompliance(req.Code, req.Language),
		"validated_at": time.Now().Format(time.RFC3339),
	}

	return validation, nil
}

// initializePatterns sets up regex patterns for analysis
func (s *CodeAnalysisServiceImpl) initializePatterns() {
	patterns := map[string]string{
		"function":     `func\s+\w+\s*\(`,
		"class":        `type\s+\w+\s+struct`,
		"interface":    `type\s+\w+\s+interface`,
		"import":       `import\s+`,
		"comment":      `//.*|/\*.*?\*/`,
		"error":        `fmt\.Errorf|errors\.New`,
		"test":         `func\s+Test\w+`,
		"global_var":   `var\s+[A-Z]\w*`,
		"magic_number": `\b\d{2,}\b`,
		"long_line":    `.{100,}`,
	}

	for name, pattern := range patterns {
		s.patterns[name] = regexp.MustCompile(pattern)
	}
}

// analyzeComplexity calculates code complexity metrics
func (s *CodeAnalysisServiceImpl) analyzeComplexity(code, language string) map[string]interface{} {
	lines := strings.Split(code, "\n")

	// Use language-specific patterns for function detection
	var functionPattern *regexp.Regexp
	switch language {
	case "go":
		functionPattern = regexp.MustCompile(`func\s+\w+\s*\(`)
	case "javascript", "typescript":
		functionPattern = regexp.MustCompile(`(function\s+\w+|const\s+\w+\s*=\s*(async\s+)?\(|function\s*\(`)
	case "python":
		functionPattern = regexp.MustCompile(`def\s+\w+\s*\(`)
	default:
		// Fallback to Go pattern
		functionPattern = s.patterns["function"]
	}

	functionCount := len(functionPattern.FindAllString(code, -1))

	cyclomatic := 1 // Base complexity
	// Add complexity for control structures (language-agnostic patterns)
	cyclomatic += strings.Count(code, "if ")
	cyclomatic += strings.Count(code, "for ")
	cyclomatic += strings.Count(code, "switch ")
	cyclomatic += strings.Count(code, "case ")

	// Language-specific complexity additions
	switch language {
	case "python":
		cyclomatic += strings.Count(code, "elif ")
		cyclomatic += strings.Count(code, "while ")
		cyclomatic += strings.Count(code, "except ")
	case "javascript", "typescript":
		cyclomatic += strings.Count(code, "else if")
		cyclomatic += strings.Count(code, "catch ")
		cyclomatic += strings.Count(code, "while ")
	}

	// Calculate language-specific thresholds
	maxRecommendedComplexity := 10
	switch language {
	case "go":
		maxRecommendedComplexity = 10
	case "python":
		maxRecommendedComplexity = 8 // Python functions tend to be simpler
	case "javascript", "typescript":
		maxRecommendedComplexity = 12 // JS/TS can handle slightly more complexity
	}

	return map[string]interface{}{
		"cyclomatic":               cyclomatic,
		"functions":                functionCount,
		"lines":                    len(lines),
		"avg_function_length":      len(lines) / max(functionCount, 1),
		"language":                 language,
		"max_recommended":          maxRecommendedComplexity,
		"complexity_exceeds_limit": cyclomatic > maxRecommendedComplexity,
	}
}

// calculateQualityScore computes an overall quality score
func (s *CodeAnalysisServiceImpl) calculateQualityScore(code, language string) float64 {
	score := 100.0

	// Deduct points for issues
	issues := s.identifyIssues(code, language)
	score -= float64(len(issues)) * 2.0

	// Deduct for complexity
	complexity := s.analyzeComplexity(code, language)
	if cyclomatic, ok := complexity["cyclomatic"].(int); ok && cyclomatic > 10 {
		score -= float64(cyclomatic-10) * 1.5
	}

	// Ensure score stays within bounds
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return score
}

// identifyIssues finds code quality issues
func (s *CodeAnalysisServiceImpl) identifyIssues(code, language string) []map[string]interface{} {
	var issues []map[string]interface{}

	lines := strings.Split(code, "\n")
	for i, line := range lines {
		lineNum := i + 1

		// Check for long lines
		if len(line) > 100 {
			issues = append(issues, map[string]interface{}{
				"type":     "long_line",
				"severity": "minor",
				"message":  "Line exceeds 100 characters",
				"line":     lineNum,
			})
		}

		// Check for magic numbers
		if language == "go" && s.patterns["magic_number"].MatchString(line) {
			issues = append(issues, map[string]interface{}{
				"type":     "magic_number",
				"severity": "minor",
				"message":  "Magic number detected",
				"line":     lineNum,
			})
		}
	}

	return issues
}

// generateSuggestions provides improvement suggestions
func (s *CodeAnalysisServiceImpl) generateSuggestions(code, language string) []string {
	var suggestions []string

	complexity := s.analyzeComplexity(code, language)
	if cyclomatic, ok := complexity["cyclomatic"].(int); ok && cyclomatic > 10 {
		suggestions = append(suggestions, "Consider breaking down complex functions into smaller ones")
	}

	if len(s.identifyIssues(code, language)) > 5 {
		suggestions = append(suggestions, "Address code quality issues to improve maintainability")
	}

	functionCount := len(s.patterns["function"].FindAllString(code, -1))
	if functionCount > 20 {
		suggestions = append(suggestions, "Consider splitting into multiple files for better organization")
	}

	return suggestions
}

// AnalyzeSecurity performs security-focused analysis
func (s *CodeAnalysisServiceImpl) AnalyzeSecurity(ctx context.Context, req models.SecurityASTRequest) (interface{}, error) {
	// Use AST service for security analysis
	astService := NewASTService()
	result, err := astService.AnalyzeSecurity(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze security: %w", err)
	}
	// Convert *models.SecurityASTResponse to interface{} for interface compliance
	return result, nil
}

// AnalyzeVibe performs comprehensive vibe coding detection analysis with quality metrics.
//
// This method analyzes code for "vibe" issues - code quality problems that affect
// maintainability, readability, and technical debt. It provides a comprehensive
// analysis including:
//   - Vibe issues: Code quality problems detected via AST analysis
//   - Duplicate functions: Similar or duplicate code patterns
//   - Orphaned code: Unused functions, variables, and dead code
//   - Quality metrics: Comprehensive quality scores across multiple dimensions
//   - Maintainability index: Halstead-based maintainability metric (0-100)
//   - Technical debt: Estimated effort and cost to resolve issues
//   - Refactoring priorities: Ranked list of refactoring recommendations
//
// The analysis uses AST-based parsing for accurate code understanding and provides
// actionable insights for code improvement.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - req: CodeAnalysisRequest containing code and language to analyze
//
// Returns:
//   - Map containing all analysis results including quality metrics, technical debt,
//     maintainability index, and refactoring priorities
//   - error if code or language is missing, or if analysis fails
//
// Example:
//
//	req := models.CodeAnalysisRequest{
//	    Code:     "package main\nfunc main() {}",
//	    Language: "go",
//	}
//	result, err := service.AnalyzeVibe(ctx, req)
//	if err != nil {
//	    // Handle error
//	}
//	// Access quality metrics: result["quality_metrics"]
//	// Access technical debt: result["technical_debt"]
func (s *CodeAnalysisServiceImpl) AnalyzeVibe(ctx context.Context, req models.CodeAnalysisRequest) (interface{}, error) {
	if req.Code == "" {
		return nil, fmt.Errorf("code is required")
	}
	if req.Language == "" {
		return nil, fmt.Errorf("language is required")
	}

	// Get basic vibe issues
	vibeIssues := s.identifyVibeIssues(req.Code, req.Language)
	duplicates := s.findDuplicateFunctions(req.Code, req.Language)
	orphaned := s.findOrphanedCode(req.Code, req.Language)

	// Calculate quality metrics
	qualityMetrics := s.calculateQualityMetrics(ctx, req.Code, req.Language, vibeIssues, duplicates, orphaned)

	// Calculate maintainability index
	maintainabilityIndex := s.calculateMaintainabilityIndex(ctx, req.Code, req.Language)

	// Estimate technical debt
	technicalDebt := s.estimateTechnicalDebt(ctx, req.Code, req.Language, vibeIssues, duplicates, orphaned)

	// Calculate refactoring priorities
	refactoringPriorities := s.calculateRefactoringPriority(ctx, req.Code, req.Language, vibeIssues, duplicates, orphaned)

	analysis := map[string]interface{}{
		"language":              req.Language,
		"vibe_issues":           vibeIssues,
		"duplicate_functions":   duplicates,
		"orphaned_code":         orphaned,
		"quality_metrics":       qualityMetrics,
		"maintainability_index": maintainabilityIndex,
		"technical_debt":        technicalDebt,
		"refactoring_priority":  refactoringPriorities,
		"analyzed_at":           time.Now().Format(time.RFC3339),
	}
	return analysis, nil
}

// AnalyzeComprehensive performs comprehensive feature analysis
func (s *CodeAnalysisServiceImpl) AnalyzeComprehensive(ctx context.Context, req ComprehensiveAnalysisRequest) (interface{}, error) {
	if req.ProjectID == "" {
		return nil, fmt.Errorf("project_id is required")
	}

	// Initialize comprehensive analysis service with dependencies
	var knowledgeService KnowledgeService
	if db != nil {
		knowledgeService = NewKnowledgeService(db)
	}

	// Create logger (use a simple logger if not available)
	// Note: simpleLogger is defined in code_analysis_comprehensive.go
	logger := &simpleLogger{}

	comprehensiveService := NewComprehensiveAnalysisService(
		NewASTService(),
		knowledgeService,
		logger,
	)

	// Execute comprehensive analysis
	result, err := comprehensiveService.ExecuteAnalysis(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute comprehensive analysis: %w", err)
	}

	return result, nil
}

// AnalyzeIntent performs intent clarification analysis
func (s *CodeAnalysisServiceImpl) AnalyzeIntent(ctx context.Context, req IntentAnalysisRequest) (interface{}, error) {
	if req.Prompt == "" {
		return nil, fmt.Errorf("prompt is required")
	}

	// Use intent analyzer
	var contextData *ContextData
	if req.IncludeContext {
		gitStatusRaw := extractGitStatus(req.CodebasePath)
		gitStatus := make(map[string]string)
		for k, v := range gitStatusRaw {
			if str, ok := v.(string); ok {
				gitStatus[k] = str
			} else {
				gitStatus[k] = fmt.Sprintf("%v", v)
			}
		}
		contextData = &ContextData{
			RecentFiles:      extractRecentFiles(req.CodebasePath),
			GitStatus:        gitStatus,
			ProjectStructure: extractProjectStructure(req.CodebasePath),
		}
	}

	result, err := AnalyzeIntent(ctx, req.Prompt, contextData, req.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze intent: %w", err)
	}

	return result, nil
}

// AnalyzeDocSync performs documentation synchronization analysis
func (s *CodeAnalysisServiceImpl) AnalyzeDocSync(ctx context.Context, req DocSyncRequest) (interface{}, error) {
	if req.ProjectID == "" {
		return nil, fmt.Errorf("project_id is required")
	}

	// Use doc sync analyzer - use project ID as codebase path fallback
	result, err := analyzeDocSync(ctx, req, "")
	if err != nil {
		return nil, fmt.Errorf("failed to analyze doc sync: %w", err)
	}

	return result, nil
}

// AnalyzeBusinessRules performs business rules compliance analysis
func (s *CodeAnalysisServiceImpl) AnalyzeBusinessRules(ctx context.Context, req BusinessRulesAnalysisRequest) (interface{}, error) {
	if req.ProjectID == "" {
		return nil, fmt.Errorf("project_id is required")
	}
	if req.CodebasePath == "" {
		return nil, fmt.Errorf("codebase_path is required")
	}

	// Extract business rules
	rules, err := extractBusinessRules(ctx, req.ProjectID, req.RuleIDs, "", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to extract business rules: %w", err)
	}

	// Analyze compliance for each rule
	var findings []BusinessContextFinding
	for _, rule := range rules {
		evidence := detectBusinessRuleImplementation(rule, req.CodebasePath)
		if evidence.Confidence < 0.5 {
			location := ""
			if len(evidence.Files) > 0 {
				location = evidence.Files[0]
			}
			findings = append(findings, BusinessContextFinding{
				Type:      "business_rule_violation",
				RuleID:    rule.ID,
				RuleTitle: rule.Title,
				Location:  location,
				Issue:     fmt.Sprintf("Business rule '%s' may not be properly implemented (confidence: %.2f)", rule.Title, evidence.Confidence),
				Severity:  "high",
			})
		}
	}

	return map[string]interface{}{
		"project_id":      req.ProjectID,
		"rules_checked":   len(rules),
		"findings":        findings,
		"compliance_rate": calculateComplianceRate(rules, findings),
		"analyzed_at":     time.Now().Format(time.RFC3339),
	}, nil
}
