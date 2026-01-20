// Package ast provides core AST analysis functionality
// Complies with CODING_STANDARDS.md: Core analysis max 300 lines
package ast

import (
	"context"
	"crypto/md5"
	"fmt"
	"time"
)

// getCacheKey generates a cache key for AST analysis
func getCacheKey(code string, language string, analyses []string) string {
	h := md5.New()
	h.Write([]byte(code))
	h.Write([]byte(language))
	for _, analysis := range analyses {
		h.Write([]byte(analysis))
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

// cleanExpiredCacheEntries removes expired cache entries
func cleanExpiredCacheEntries() {
	now := time.Now()
	for key, entry := range astCache {
		if now.After(entry.Expires) {
			delete(astCache, key)
		}
	}
}

// AnalyzeAST performs comprehensive AST analysis with panic recovery
func AnalyzeAST(code string, language string, analyses []string) (findings []ASTFinding, stats AnalysisStats, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("AST analysis panic: %v", r)
			findings = nil
			stats = AnalysisStats{}
		}
	}()

	return analyzeASTInternal(code, language, analyses)
}

// analyzeASTInternal performs the actual AST analysis logic
func analyzeASTInternal(code string, language string, analyses []string) ([]ASTFinding, AnalysisStats, error) {
	// Check cache first
	cacheKey := getCacheKey(code, language, analyses)
	cacheMutex.RLock()
	if entry, ok := astCache[cacheKey]; ok {
		if time.Now().Before(entry.Expires) {
			cacheMutex.RUnlock()
			return entry.Findings, entry.Stats, nil
		}
		// Cache expired, remove it
		delete(astCache, cacheKey)
	}
	cacheMutex.RUnlock()

	// Get parser for language
	parser, err := getParser(language)
	if err != nil {
		return nil, AnalysisStats{}, fmt.Errorf("parser error: %w", err)
	}

	// Parse code into AST
	parseStart := time.Now()
	ctx := context.Background()
	tree, err := parser.ParseCtx(ctx, nil, []byte(code))
	if err != nil {
		return nil, AnalysisStats{}, fmt.Errorf("parse error: %w", err)
	}
	parseTime := time.Since(parseStart).Milliseconds()

	if tree == nil {
		return nil, AnalysisStats{}, fmt.Errorf("failed to parse code")
	}
	defer tree.Close()

	rootNode := tree.RootNode()
	if rootNode == nil {
		return nil, AnalysisStats{}, fmt.Errorf("failed to get root node")
	}

	// Perform requested analyses
	analysisStart := time.Now()
	findings := []ASTFinding{}

	// Track which analyses to perform
	checkDuplicates := contains(analyses, "duplicates") || len(analyses) == 0
	checkUnused := contains(analyses, "unused") || len(analyses) == 0
	checkUnreachable := contains(analyses, "unreachable") || len(analyses) == 0
	checkEmptyCatch := contains(analyses, "empty_catch") || contains(analyses, "vibe") || len(analyses) == 0
	checkMissingAwait := contains(analyses, "missing_await") || contains(analyses, "vibe") || len(analyses) == 0
	checkBraceMismatch := contains(analyses, "brace_mismatch") || contains(analyses, "vibe") || len(analyses) == 0

	if checkDuplicates {
		duplicates := detectDuplicateFunctions(rootNode, code, language)
		findings = append(findings, duplicates...)
	}

	if checkUnused {
		unused := detectUnusedVariables(rootNode, code, language)
		findings = append(findings, unused...)
	}

	if checkUnreachable {
		unreachable := detectUnreachableCode(rootNode, code, language)
		findings = append(findings, unreachable...)
	}

	// Orphaned code detection (always enabled for vibe analysis)
	if contains(analyses, "orphaned") || contains(analyses, "duplicates") {
		orphaned := detectOrphanedCode(rootNode, code, language)
		findings = append(findings, orphaned...)
	}

	// Phase 7C: Additional AST detections
	if checkEmptyCatch {
		emptyCatch := detectEmptyCatchBlocks(rootNode, code, language)
		findings = append(findings, emptyCatch...)
	}

	if checkMissingAwait {
		missingAwait := detectMissingAwait(rootNode, code, language)
		findings = append(findings, missingAwait...)
	}

	// Brace/bracket mismatch detection (check parse errors)
	if checkBraceMismatch {
		braceMismatch := detectBraceMismatch(tree, code, language)
		findings = append(findings, braceMismatch...)
	}

	analysisTime := time.Since(analysisStart).Milliseconds()

	stats := AnalysisStats{
		ParseTime:    parseTime,
		AnalysisTime: analysisTime,
		NodesVisited: countNodes(rootNode),
	}

	// Cache the results
	cacheMutex.Lock()
	astCache[cacheKey] = &cacheEntry{
		Findings: findings,
		Stats:    stats,
		Expires:  time.Now().Add(cacheTTL),
	}

	// Clean up expired entries periodically (time-based, not size-based)
	if time.Since(lastCacheCleanup) > cacheCleanupInterval {
		cleanExpiredCacheEntries()
		lastCacheCleanup = time.Now()
	}
	// Also clean if cache is too large
	if len(astCache) > 1000 {
		cleanExpiredCacheEntries()
	}
	cacheMutex.Unlock()

	return findings, stats, nil
}

// AnalyzeASTWithValidation performs AST analysis with codebase validation
func AnalyzeASTWithValidation(code, language, filePath, projectRoot string, analyses []string) ([]ASTFinding, AnalysisStats, error) {
	findings, stats, err := AnalyzeAST(code, language, analyses)
	if err != nil {
		return findings, stats, err
	}

	// Detect edge cases in source code
	edgeCases := DetectEdgeCases(code, language)

	// Validate findings with timeout and concurrency
	err = ValidateFindingsConcurrent(findings, filePath, projectRoot, language, DefaultMaxConcurrency)
	if err != nil {
		// Log validation error but don't fail the entire analysis
		// Continue with edge case penalties
	}

	// Apply edge case penalties to confidence and recalculate AutoFixSafe
	for i := range findings {
		// Apply edge case penalty to existing confidence
		findings[i].Confidence -= edgeCases.ConfidencePenalty

		// Ensure confidence doesn't go below 0
		if findings[i].Confidence < 0 {
			findings[i].Confidence = 0
		}

		// Recalculate AutoFixSafe based on updated confidence
		findings[i].AutoFixSafe = DetermineAutoFixSafe(findings[i].Confidence, findings[i].Type)

		// Update reasoning if edge cases were detected
		if edgeCases.ConfidencePenalty > 0 {
			edgeCaseNote := ""
			if edgeCases.HasGeneratedCode {
				edgeCaseNote = "Generated code detected - never auto-fix"
			} else if edgeCases.HasReflection {
				edgeCaseNote = "Reflection usage detected - reduced confidence"
			} else if edgeCases.HasDynamicImport {
				edgeCaseNote = "Dynamic imports detected - reduced confidence"
			} else if edgeCases.HasPluginUsage {
				edgeCaseNote = "Plugin usage detected - reduced confidence"
			}
			if edgeCaseNote != "" {
				findings[i].Reasoning = edgeCaseNote + ". " + findings[i].Reasoning
			}
		}
	}

	return findings, stats, nil
}
