// Package services provides comprehensive multi-layer code analysis
// Complies with CODING_STANDARDS.md: Business Services max 400 lines
package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"sentinel-hub-api/ast"
)

// simpleLogger provides a minimal logger implementation
type simpleLogger struct{}

func (l *simpleLogger) Warn(ctx context.Context, msg string, fields map[string]interface{}) {
	// Minimal logging - can be enhanced with actual logger
}

func (l *simpleLogger) Error(ctx context.Context, msg string, err error, fields ...map[string]interface{}) {
	// Minimal logging - can be enhanced with actual logger
}

func (l *simpleLogger) Info(ctx context.Context, msg string, fields ...map[string]interface{}) {
	// Minimal logging - can be enhanced with actual logger
}

func (l *simpleLogger) Debug(ctx context.Context, msg string, fields ...map[string]interface{}) {
	// Minimal logging - can be enhanced with actual logger
}

// ComprehensiveAnalysisService handles multi-layer code analysis
type ComprehensiveAnalysisService struct {
	astService      ASTService
	knowledgeService KnowledgeService
	logger         Logger
}

// NewComprehensiveAnalysisService creates a new comprehensive analysis service
func NewComprehensiveAnalysisService(astService ASTService, knowledgeService KnowledgeService, logger Logger) *ComprehensiveAnalysisService {
	return &ComprehensiveAnalysisService{
		astService:      astService,
		knowledgeService: knowledgeService,
		logger:          logger,
	}
}

// ExecuteAnalysis performs comprehensive multi-layer code analysis
func (s *ComprehensiveAnalysisService) ExecuteAnalysis(ctx context.Context, req ComprehensiveAnalysisRequest) (*ComprehensiveAnalysisResult, error) {
	if req.ProjectID == "" {
		return nil, fmt.Errorf("project_id is required")
	}

	// Check context cancellation
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// Determine codebase path
	codebasePath := req.CodebasePath
	if codebasePath == "" {
		codebasePath = "."
	}

	// Detect layers to analyze
	layersToAnalyze, err := s.detectLayers(ctx, codebasePath, req.Mode, req.Files)
	if err != nil {
		s.logger.Warn(ctx, "Failed to detect layers, using default layers", map[string]interface{}{
			"error": err.Error(),
		})
		layersToAnalyze = []string{"ui", "api", "database", "logic", "integration", "tests"}
	}

	// Execute layer analysis in parallel
	layerResults := s.executeLayerAnalysis(ctx, codebasePath, layersToAnalyze, req.Depth, req.Files)

	// Aggregate results
	result := s.aggregateResults(layerResults, req)

	// Add business context if requested
	if req.IncludeBusinessContext {
		businessCtx, err := s.getBusinessContext(ctx, req.ProjectID, req.Feature)
		if err != nil {
			s.logger.Warn(ctx, "Failed to get business context", map[string]interface{}{
				"error": err.Error(),
			})
		} else {
			result.BusinessContext = businessCtx
		}
	}

	return result, nil
}

// detectLayers detects which layers exist in the codebase
func (s *ComprehensiveAnalysisService) detectLayers(ctx context.Context, codebasePath, mode string, specifiedFiles []string) ([]string, error) {
	// Check context cancellation before starting
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	if mode == "manual" && len(specifiedFiles) > 0 {
		// In manual mode, analyze based on file paths
		return s.detectLayersFromFiles(specifiedFiles), nil
	}

	// Auto mode: detect layers from codebase structure
	var layers []string
	if _, err := os.Stat(codebasePath); err != nil {
		return layers, fmt.Errorf("codebase path does not exist: %w", err)
	}

	// Check context cancellation during file system operations
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// Check for UI layer (frontend files)
	if s.hasLayerFiles(codebasePath, []string{".jsx", ".tsx", ".vue", ".html", ".css"}) {
		layers = append(layers, "ui")
	}

	// Check for API layer (API handlers, routes)
	if s.hasLayerFiles(codebasePath, []string{"handler", "route", "endpoint", "controller"}) {
		layers = append(layers, "api")
	}

	// Check for database layer (migrations, models, repositories)
	if s.hasLayerFiles(codebasePath, []string{"migration", "model", "repository", "schema"}) {
		layers = append(layers, "database")
	}

	// Check for logic layer (services, business logic)
	if s.hasLayerFiles(codebasePath, []string{"service", "business", "logic"}) {
		layers = append(layers, "logic")
	}

	// Check for integration layer (external services, APIs)
	if s.hasLayerFiles(codebasePath, []string{"integration", "client", "external"}) {
		layers = append(layers, "integration")
	}

	// Check for test layer (test files)
	if s.hasLayerFiles(codebasePath, []string{"_test.go", "_test.js", "_test.ts", ".test.", ".spec."}) {
		layers = append(layers, "tests")
	}

	// Default to all layers if none detected
	if len(layers) == 0 {
		layers = []string{"ui", "api", "database", "logic", "integration", "tests"}
	}

	return layers, nil
}

// detectLayersFromFiles detects layers from file paths
func (s *ComprehensiveAnalysisService) detectLayersFromFiles(files []string) []string {
	layerMap := make(map[string]bool)
	for _, file := range files {
		lowerFile := strings.ToLower(file)
		if strings.Contains(lowerFile, "ui") || strings.Contains(lowerFile, "component") ||
			strings.Contains(lowerFile, ".jsx") || strings.Contains(lowerFile, ".tsx") {
			layerMap["ui"] = true
		}
		if strings.Contains(lowerFile, "api") || strings.Contains(lowerFile, "handler") ||
			strings.Contains(lowerFile, "route") || strings.Contains(lowerFile, "endpoint") {
			layerMap["api"] = true
		}
		if strings.Contains(lowerFile, "database") || strings.Contains(lowerFile, "migration") ||
			strings.Contains(lowerFile, "model") || strings.Contains(lowerFile, "repository") {
			layerMap["database"] = true
		}
		if strings.Contains(lowerFile, "service") || strings.Contains(lowerFile, "business") ||
			strings.Contains(lowerFile, "logic") {
			layerMap["logic"] = true
		}
		if strings.Contains(lowerFile, "integration") || strings.Contains(lowerFile, "client") ||
			strings.Contains(lowerFile, "external") {
			layerMap["integration"] = true
		}
		if strings.Contains(lowerFile, "test") || strings.Contains(lowerFile, "spec") {
			layerMap["tests"] = true
		}
	}

	var layers []string
	for layer := range layerMap {
		layers = append(layers, layer)
	}

	if len(layers) == 0 {
		layers = []string{"ui", "api", "database", "logic", "integration", "tests"}
	}

	return layers
}

// hasLayerFiles checks if codebase has files matching layer patterns
func (s *ComprehensiveAnalysisService) hasLayerFiles(codebasePath string, patterns []string) bool {
	found := false
	filepath.Walk(codebasePath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		lowerPath := strings.ToLower(path)
		for _, pattern := range patterns {
			if strings.Contains(lowerPath, pattern) {
				found = true
				return filepath.SkipDir
			}
		}
		return nil
	})
	return found
}

// executeLayerAnalysis executes analysis for all layers in parallel
func (s *ComprehensiveAnalysisService) executeLayerAnalysis(ctx context.Context, codebasePath string, layers []string, depth string, files []string) []LayerAnalysis {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var results []LayerAnalysis

	for _, layer := range layers {
		wg.Add(1)
		go func(layerName string) {
			defer wg.Done()

			// Check context cancellation
			if ctx.Err() != nil {
				return
			}

			analysis, err := s.analyzeLayer(ctx, codebasePath, layerName, depth, files)
			if err != nil {
				s.logger.Warn(ctx, fmt.Sprintf("Failed to analyze layer: %s", layerName), map[string]interface{}{
					"error": err.Error(),
				})
				// Return empty analysis on error
				analysis = LayerAnalysis{
					Layer:      layerName,
					Files:      []string{},
					Findings:   []LayerFinding{},
					Issues:     []Issue{},
					AnalyzedAt: time.Now().Format(time.RFC3339),
				}
			}

			mu.Lock()
			results = append(results, analysis)
			mu.Unlock()
		}(layer)
	}

	wg.Wait()
	return results
}

// analyzeLayer analyzes a specific layer
func (s *ComprehensiveAnalysisService) analyzeLayer(ctx context.Context, codebasePath, layer, depth string, specifiedFiles []string) (LayerAnalysis, error) {
	// Find files for this layer
	files := s.findLayerFiles(codebasePath, layer, specifiedFiles)

	// Perform analysis based on depth
	var findings []LayerFinding
	var issues []Issue
	var dependencies []Dependency
	var qualityScore float64

	if depth == "deep" {
		// Deep analysis: full AST analysis
		findings, issues, dependencies, qualityScore = s.performDeepAnalysis(ctx, files, layer)
	} else {
		// Shallow analysis: quick syntax and structure check
		findings, issues, dependencies, qualityScore = s.performShallowAnalysis(ctx, files, layer)
	}

	return LayerAnalysis{
		Layer:        layer,
		Files:        files,
		Findings:     findings,
		QualityScore: qualityScore,
		Issues:       issues,
		Dependencies: dependencies,
		AnalyzedAt:   time.Now().Format(time.RFC3339),
	}, nil
}

// findLayerFiles finds files belonging to a specific layer
func (s *ComprehensiveAnalysisService) findLayerFiles(codebasePath, layer string, specifiedFiles []string) []string {
	var files []string

	// If specific files provided, filter by layer
	if len(specifiedFiles) > 0 {
		for _, file := range specifiedFiles {
			if s.fileBelongsToLayer(file, layer) {
				files = append(files, file)
			}
		}
		return files
	}

	// Otherwise, scan codebase for layer files
	patterns := s.getLayerPatterns(layer)
	filepath.Walk(codebasePath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		// Skip hidden and build directories
		if strings.Contains(path, ".git") || strings.Contains(path, "node_modules") ||
			strings.Contains(path, "vendor") || strings.Contains(path, ".idea") {
			return nil
		}

		lowerPath := strings.ToLower(path)
		for _, pattern := range patterns {
			if strings.Contains(lowerPath, pattern) {
				relPath, _ := filepath.Rel(codebasePath, path)
				files = append(files, relPath)
				break
			}
		}
		return nil
	})

	return files
}

// fileBelongsToLayer checks if a file belongs to a specific layer
func (s *ComprehensiveAnalysisService) fileBelongsToLayer(file, layer string) bool {
	lowerFile := strings.ToLower(file)
	patterns := s.getLayerPatterns(layer)
	for _, pattern := range patterns {
		if strings.Contains(lowerFile, pattern) {
			return true
		}
	}
	return false
}

// getLayerPatterns returns file patterns for a layer
func (s *ComprehensiveAnalysisService) getLayerPatterns(layer string) []string {
	patterns := map[string][]string{
		"ui":         {".jsx", ".tsx", ".vue", "component", "ui", "frontend"},
		"api":        {"handler", "route", "endpoint", "controller", "api"},
		"database":   {"migration", "model", "repository", "schema", "database"},
		"logic":      {"service", "business", "logic"},
		"integration": {"integration", "client", "external"},
		"tests":      {"_test.go", "_test.js", "_test.ts", ".test.", ".spec."},
	}
	if p, ok := patterns[layer]; ok {
		return p
	}
	return []string{}
}

// performDeepAnalysis performs comprehensive AST-based analysis
func (s *ComprehensiveAnalysisService) performDeepAnalysis(ctx context.Context, files []string, _ string) ([]LayerFinding, []Issue, []Dependency, float64) {
	var findings []LayerFinding
	var issues []Issue
	var dependencies []Dependency
	totalScore := 0.0
	fileCount := 0

	for _, file := range files {
		if ctx.Err() != nil {
			break
		}

		// Check if file exists
		if _, err := os.Stat(file); err != nil {
			continue
		}

		// Read file content
		content, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		// Skip empty files
		if len(content) == 0 {
			continue
		}

		// Determine language from file extension
		language := s.detectLanguage(file)

		// Skip unsupported languages
		if language == "" {
			continue
		}

		// Perform AST analysis with error recovery
		astFindings, _, err := ast.AnalyzeAST(string(content), language, []string{
			"unused", "unreachable", "duplicates", "orphaned",
		})
		if err != nil {
			// Log error but continue with other files
			s.logger.Debug(ctx, fmt.Sprintf("AST analysis failed for %s: %v", file, err))
			continue
		}

		// Convert AST findings to layer findings
		for _, f := range astFindings {
			findings = append(findings, LayerFinding{
				Type:       f.Type,
				Severity:   f.Severity,
				Line:       f.Line,
				Message:    f.Message,
				Suggestion: f.Suggestion,
				FilePath:   file,
			})

			issues = append(issues, Issue{
				Type:       f.Type,
				Severity:   f.Severity,
				Line:       f.Line,
				Message:    f.Message,
				Suggestion: f.Suggestion,
				FilePath:   file,
			})
		}

		// Extract dependencies
		deps := s.extractDependencies(string(content), language)
		dependencies = append(dependencies, deps...)

		// Calculate quality score for this file
		score := s.calculateFileQualityScore(astFindings)
		totalScore += score
		fileCount++
	}

	avgScore := 0.0
	if fileCount > 0 {
		avgScore = totalScore / float64(fileCount)
	}

	return findings, issues, dependencies, avgScore
}

// performShallowAnalysis performs quick syntax and structure analysis
func (s *ComprehensiveAnalysisService) performShallowAnalysis(ctx context.Context, files []string, _ string) ([]LayerFinding, []Issue, []Dependency, float64) {
	var findings []LayerFinding
	var issues []Issue
	var dependencies []Dependency
	totalScore := 100.0
	fileCount := 0

	for _, file := range files {
		if ctx.Err() != nil {
			break
		}

		// Check if file exists
		if _, err := os.Stat(file); err != nil {
			continue
		}

		// Read file content
		content, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		// Skip empty files
		if len(content) == 0 {
			continue
		}

		language := s.detectLanguage(file)

		// Skip unsupported languages
		if language == "" {
			continue
		}

		// Quick syntax validation with error recovery
		parser, err := ast.GetParser(language)
		if err == nil {
			// Use recover to catch any panics from tree-sitter
			func() {
				defer func() {
					if r := recover(); r != nil {
						s.logger.Debug(ctx, fmt.Sprintf("Parser panic recovered for %s: %v", file, r))
					}
				}()
				tree, parseErr := parser.ParseCtx(ctx, nil, content)
				if parseErr != nil {
					issues = append(issues, Issue{
						Type:     "syntax_error",
						Severity: "high",
						Line:     0,
						Message:  fmt.Sprintf("Syntax error in %s: %v", file, parseErr),
						FilePath: file,
					})
					totalScore -= 10.0
				} else if tree != nil {
					tree.Close()
				}
			}()
		}

		// Extract basic dependencies
		deps := s.extractDependencies(string(content), language)
		dependencies = append(dependencies, deps...)

		fileCount++
	}

	avgScore := totalScore
	if fileCount > 0 && totalScore < 100.0 {
		avgScore = totalScore / float64(fileCount)
	}

	return findings, issues, dependencies, avgScore
}

// detectLanguage detects programming language from file extension
func (s *ComprehensiveAnalysisService) detectLanguage(file string) string {
	ext := strings.ToLower(filepath.Ext(file))
	langMap := map[string]string{
		".go":   "go",
		".js":   "javascript",
		".ts":   "typescript",
		".jsx":  "javascript",
		".tsx":  "typescript",
		".py":   "python",
		".java": "java",
		".rb":   "ruby",
		".php":  "php",
	}
	if lang, ok := langMap[ext]; ok {
		return lang
	}
	return "go" // Default
}

// extractDependencies extracts dependencies from code
func (s *ComprehensiveAnalysisService) extractDependencies(code, language string) []Dependency {
	var deps []Dependency

	switch language {
	case "go":
		// Extract Go imports
		lines := strings.Split(code, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "import") || strings.HasPrefix(line, "\"") || strings.HasPrefix(line, "`") {
				if strings.Contains(line, "\"") {
					parts := strings.Split(line, "\"")
					if len(parts) >= 2 {
						dep := Dependency{
							Name: parts[1],
							Type: s.classifyDependency(parts[1]),
						}
						deps = append(deps, dep)
					}
				}
			}
		}
	case "javascript", "typescript":
		// Extract JS/TS imports
		lines := strings.Split(code, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "import") || strings.HasPrefix(line, "require") {
				if strings.Contains(line, "from") || strings.Contains(line, "require") {
					parts := strings.Fields(line)
					for i, part := range parts {
						if part == "from" && i+1 < len(parts) {
							depName := strings.Trim(parts[i+1], "\"'")
							dep := Dependency{
								Name: depName,
								Type: s.classifyDependency(depName),
							}
							deps = append(deps, dep)
						}
					}
				}
			}
		}
	}

	return deps
}

// classifyDependency classifies dependency type
func (s *ComprehensiveAnalysisService) classifyDependency(name string) string {
	if strings.HasPrefix(name, ".") || strings.HasPrefix(name, "./") || strings.HasPrefix(name, "../") {
		return "internal"
	}
	if strings.Contains(name, "github.com") || strings.Contains(name, "npm") ||
		strings.Contains(name, "pypi") || strings.Contains(name, "maven") {
		return "external"
	}
	return "standard_library"
}

// calculateFileQualityScore calculates quality score for a file
func (s *ComprehensiveAnalysisService) calculateFileQualityScore(findings []ast.ASTFinding) float64 {
	score := 100.0
	for _, finding := range findings {
		switch finding.Severity {
		case "critical":
			score -= 10.0
		case "high":
			score -= 5.0
		case "medium":
			score -= 2.0
		case "low":
			score -= 1.0
		}
	}
	if score < 0 {
		score = 0
	}
	return score
}

// aggregateResults aggregates layer analysis results
func (s *ComprehensiveAnalysisService) aggregateResults(layers []LayerAnalysis, req ComprehensiveAnalysisRequest) *ComprehensiveAnalysisResult {
	totalScore := 0.0
	layerCount := 0

	for _, layer := range layers {
		totalScore += layer.QualityScore
		if layer.QualityScore > 0 {
			layerCount++
		}
	}

	overallScore := 0.0
	if layerCount > 0 {
		overallScore = totalScore / float64(layerCount)
	}

	return &ComprehensiveAnalysisResult{
		ProjectID:    req.ProjectID,
		Feature:      req.Feature,
		Mode:         req.Mode,
		Depth:        req.Depth,
		Layers:       layers,
		OverallScore: overallScore,
		AnalyzedAt:   time.Now().Format(time.RFC3339),
	}
}

// getBusinessContext retrieves business context for the analysis
func (s *ComprehensiveAnalysisService) getBusinessContext(ctx context.Context, projectID, feature string) (*BusinessContextResponse, error) {
	if s.knowledgeService == nil {
		return nil, fmt.Errorf("knowledge service not available")
	}

	req := BusinessContextRequest{
		ProjectID: projectID,
		Feature:   feature,
	}

	return s.knowledgeService.GetBusinessContext(ctx, req)
}
