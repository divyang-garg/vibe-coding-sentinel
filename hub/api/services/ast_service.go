// Package services provides AST analysis business logic
// Complies with CODING_STANDARDS.md: Business Services max 400 lines
package services

import (
	"context"
	"fmt"
	"time"

	"sentinel-hub-api/ast"
	"sentinel-hub-api/models"
)

// ASTService defines the interface for AST analysis operations
type ASTService interface {
	AnalyzeAST(ctx context.Context, req models.ASTAnalysisRequest) (*models.ASTAnalysisResponse, error)
	AnalyzeMultiFile(ctx context.Context, req models.MultiFileASTRequest) (*models.MultiFileASTResponse, error)
	AnalyzeSecurity(ctx context.Context, req models.SecurityASTRequest) (*models.SecurityASTResponse, error)
	AnalyzeCrossFile(ctx context.Context, req models.CrossFileASTRequest) (*models.CrossFileASTResponse, error)
	GetSupportedAnalyses(ctx context.Context) (*models.SupportedAnalysesResponse, error)
}

// ASTServiceImpl implements ASTService
type ASTServiceImpl struct {
	// Dependencies can be added here if needed
}

// NewASTService creates a new AST service instance
func NewASTService() ASTService {
	return &ASTServiceImpl{}
}

// AnalyzeAST performs single-file AST analysis
func (s *ASTServiceImpl) AnalyzeAST(ctx context.Context, req models.ASTAnalysisRequest) (*models.ASTAnalysisResponse, error) {
	if req.Code == "" {
		return nil, fmt.Errorf("code is required")
	}
	if req.Language == "" {
		return nil, fmt.Errorf("language is required")
	}

	// Perform AST analysis
	findings, stats, err := ast.AnalyzeAST(req.Code, req.Language, req.Analyses)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze AST: %w", err)
	}

	// Convert AST findings to model findings
	modelFindings := make([]models.ASTFinding, len(findings))
	for i, f := range findings {
		modelFindings[i] = models.ASTFinding{
			Type:        f.Type,
			Severity:    f.Severity,
			Line:        f.Line,
			Column:      f.Column,
			EndLine:     f.EndLine,
			EndColumn:   f.EndColumn,
			Message:     f.Message,
			Code:        f.Code,
			Suggestion:  f.Suggestion,
			Confidence:  f.Confidence,
			AutoFixSafe: f.AutoFixSafe,
			FixType:     f.FixType,
			Reasoning:   f.Reasoning,
			FilePath:    req.FilePath,
		}
	}

	response := &models.ASTAnalysisResponse{
		Findings: modelFindings,
		Stats: models.ASTAnalysisStats{
			ParseTime:    stats.ParseTime,
			AnalysisTime: stats.AnalysisTime,
			NodesVisited: stats.NodesVisited,
		},
		Language: req.Language,
		FilePath: req.FilePath,
	}

	return response, nil
}

// AnalyzeMultiFile performs multi-file AST analysis
func (s *ASTServiceImpl) AnalyzeMultiFile(ctx context.Context, req models.MultiFileASTRequest) (*models.MultiFileASTResponse, error) {
	if len(req.Files) == 0 {
		return nil, fmt.Errorf("at least one file is required")
	}

	// Convert model files to AST file inputs
	astFiles := make([]ast.FileInput, len(req.Files))
	for i, f := range req.Files {
		if f.Content == "" {
			return nil, fmt.Errorf("file content is required for file: %s", f.Path)
		}
		if f.Language == "" {
			return nil, fmt.Errorf("language is required for file: %s", f.Path)
		}
		astFiles[i] = ast.FileInput{
			Path:     f.Path,
			Content:  f.Content,
			Language: f.Language,
		}
	}

	// Perform multi-file analysis
	startTime := time.Now()
	findings, stats, err := ast.AnalyzeMultiFile(ctx, astFiles, req.Analyses)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze multi-file AST: %w", err)
	}
	analysisTime := time.Since(startTime).Milliseconds()

	// Convert AST findings to model findings
	modelFindings := make([]models.ASTFinding, len(findings))
	filePaths := make([]string, len(req.Files))
	for i, f := range findings {
		modelFindings[i] = models.ASTFinding{
			Type:        f.Type,
			Severity:    f.Severity,
			Line:        f.Line,
			Column:      f.Column,
			EndLine:     f.EndLine,
			EndColumn:   f.EndColumn,
			Message:     f.Message,
			Code:        f.Code,
			Suggestion:  f.Suggestion,
			Confidence:  f.Confidence,
			AutoFixSafe: f.AutoFixSafe,
			FixType:     f.FixType,
			Reasoning:   f.Reasoning,
		}
	}
	for i, f := range req.Files {
		filePaths[i] = f.Path
	}

	response := &models.MultiFileASTResponse{
		Findings: modelFindings,
		Stats: models.ASTAnalysisStats{
			ParseTime:    stats.ParseTime,
			AnalysisTime: stats.AnalysisTime,
			NodesVisited: stats.NodesVisited,
		},
		Files: filePaths,
	}

	// Update analysis time if needed
	if response.Stats.AnalysisTime == 0 {
		response.Stats.AnalysisTime = analysisTime
	}

	return response, nil
}

// AnalyzeSecurity performs security-focused AST analysis
func (s *ASTServiceImpl) AnalyzeSecurity(ctx context.Context, req models.SecurityASTRequest) (*models.SecurityASTResponse, error) {
	if req.Code == "" {
		return nil, fmt.Errorf("code is required")
	}
	if req.Language == "" {
		return nil, fmt.Errorf("language is required")
	}

	// Perform security analysis
	vulnerabilities, findings, stats, err := ast.AnalyzeSecurity(ctx, req.Code, req.Language, req.Severity)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze security: %w", err)
	}

	// Convert AST findings to model findings
	modelFindings := make([]models.ASTFinding, len(findings))
	for i, f := range findings {
		modelFindings[i] = models.ASTFinding{
			Type:        f.Type,
			Severity:    f.Severity,
			Line:        f.Line,
			Column:      f.Column,
			EndLine:     f.EndLine,
			EndColumn:   f.EndColumn,
			Message:     f.Message,
			Code:        f.Code,
			Suggestion:  f.Suggestion,
			Confidence:  f.Confidence,
			AutoFixSafe: f.AutoFixSafe,
			FixType:     f.FixType,
			Reasoning:   f.Reasoning,
		}
	}

	// Convert vulnerabilities
	modelVulns := make([]models.Vulnerability, len(vulnerabilities))
	for i, v := range vulnerabilities {
		modelVulns[i] = models.Vulnerability{
			Type:        v.Type,
			Severity:    v.Severity,
			Line:        v.Line,
			Column:      v.Column,
			Message:     v.Message,
			Code:        v.Code,
			Description: v.Description,
			Remediation: v.Remediation,
			Confidence:  v.Confidence,
		}
	}

	// Calculate risk score
	riskScore := calculateRiskScore(modelVulns)

	response := &models.SecurityASTResponse{
		Findings: modelFindings,
		Stats: models.ASTAnalysisStats{
			ParseTime:    stats.ParseTime,
			AnalysisTime: stats.AnalysisTime,
			NodesVisited: stats.NodesVisited,
		},
		Vulnerabilities: modelVulns,
		RiskScore:       riskScore,
	}

	return response, nil
}

// AnalyzeCrossFile performs cross-file dependency analysis
func (s *ASTServiceImpl) AnalyzeCrossFile(ctx context.Context, req models.CrossFileASTRequest) (*models.CrossFileASTResponse, error) {
	if len(req.Files) == 0 {
		return nil, fmt.Errorf("at least one file is required")
	}

	// Convert model files to AST file inputs
	astFiles := make([]ast.FileInput, len(req.Files))
	for i, f := range req.Files {
		if f.Content == "" {
			return nil, fmt.Errorf("file content is required for file: %s", f.Path)
		}
		if f.Language == "" {
			return nil, fmt.Errorf("language is required for file: %s", f.Path)
		}
		astFiles[i] = ast.FileInput{
			Path:     f.Path,
			Content:  f.Content,
			Language: f.Language,
		}
	}

	// Perform cross-file analysis
	crossFileResult, err := ast.AnalyzeCrossFile(ctx, astFiles, []string{})
	if err != nil {
		return nil, fmt.Errorf("failed to analyze cross-file: %w", err)
	}

	// Convert findings
	modelFindings := make([]models.ASTFinding, len(crossFileResult.Findings))
	for i, f := range crossFileResult.Findings {
		modelFindings[i] = models.ASTFinding{
			Type:        f.Type,
			Severity:    f.Severity,
			Line:        f.Line,
			Column:      f.Column,
			EndLine:     f.EndLine,
			EndColumn:   f.EndColumn,
			Message:     f.Message,
			Code:        f.Code,
			Suggestion:  f.Suggestion,
			Confidence:  f.Confidence,
			AutoFixSafe: f.AutoFixSafe,
			FixType:     f.FixType,
			Reasoning:   f.Reasoning,
		}
	}

	// Convert unused exports
	unusedExports := make([]models.ExportSymbol, len(crossFileResult.UnusedExports))
	for i, exp := range crossFileResult.UnusedExports {
		unusedExports[i] = models.ExportSymbol{
			Name:     exp.Name,
			Kind:     exp.Kind,
			FilePath: exp.FilePath,
			Line:     exp.Line,
			Column:   exp.Column,
		}
	}

	// Convert undefined refs
	undefinedRefs := make([]models.SymbolRef, len(crossFileResult.UndefinedRefs))
	for i, ref := range crossFileResult.UndefinedRefs {
		undefinedRefs[i] = models.SymbolRef{
			Name:     ref.Name,
			FilePath: ref.FilePath,
			Line:     ref.Line,
			Column:   ref.Column,
			Kind:     ref.Kind,
		}
	}

	// Convert cross-file duplicates
	crossFileDups := make([]models.ASTFinding, len(crossFileResult.CrossFileDuplicates))
	for i, f := range crossFileResult.CrossFileDuplicates {
		crossFileDups[i] = models.ASTFinding{
			Type:        f.Type,
			Severity:    f.Severity,
			Line:        f.Line,
			Column:      f.Column,
			EndLine:     f.EndLine,
			EndColumn:   f.EndColumn,
			Message:     f.Message,
			Code:        f.Code,
			Suggestion:  f.Suggestion,
			Confidence:  f.Confidence,
			AutoFixSafe: f.AutoFixSafe,
			FixType:     f.FixType,
			Reasoning:   f.Reasoning,
		}
	}

	response := &models.CrossFileASTResponse{
		Findings:            modelFindings,
		UnusedExports:       unusedExports,
		UndefinedRefs:       undefinedRefs,
		CircularDeps:        crossFileResult.CircularDeps,
		CrossFileDuplicates: crossFileDups,
		Stats: models.CrossFileStats{
			FilesAnalyzed:     crossFileResult.Stats.FilesAnalyzed,
			SymbolsFound:      crossFileResult.Stats.SymbolsFound,
			DependenciesFound: crossFileResult.Stats.DependenciesFound,
			AnalysisTime:      crossFileResult.Stats.AnalysisTime,
		},
	}

	return response, nil
}

// GetSupportedAnalyses returns supported languages and analyses
func (s *ASTServiceImpl) GetSupportedAnalyses(ctx context.Context) (*models.SupportedAnalysesResponse, error) {
	languages := []models.LanguageSupport{
		{Name: "go", Aliases: []string{"golang"}, Supported: true},
		{Name: "javascript", Aliases: []string{"js", "jsx"}, Supported: true},
		{Name: "typescript", Aliases: []string{"ts", "tsx"}, Supported: true},
		{Name: "python", Aliases: []string{"py"}, Supported: true},
	}

	analyses := []string{
		"duplicates",
		"unused",
		"unreachable",
		"orphaned",
		"empty_catch",
		"missing_await",
		"brace_mismatch",
		"unused_exports",
		"undefined_refs",
		"circular_deps",
		"cross_file_duplicates",
	}

	return &models.SupportedAnalysesResponse{
		Languages: languages,
		Analyses:  analyses,
	}, nil
}

// calculateRiskScore calculates a risk score based on vulnerabilities
func calculateRiskScore(vulns []models.Vulnerability) float64 {
	if len(vulns) == 0 {
		return 0.0
	}

	score := 0.0
	for _, vuln := range vulns {
		severityWeight := 0.0
		switch vuln.Severity {
		case "critical":
			severityWeight = 10.0
		case "high":
			severityWeight = 7.0
		case "medium":
			severityWeight = 4.0
		case "low":
			severityWeight = 1.0
		}
		score += severityWeight * vuln.Confidence
	}

	// Normalize to 0-100 scale
	maxScore := float64(len(vulns)) * 10.0
	if maxScore > 0 {
		score = (score / maxScore) * 100.0
	}

	return score
}
