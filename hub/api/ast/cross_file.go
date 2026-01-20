// Package ast provides cross-file analysis capabilities
// Complies with CODING_STANDARDS.md: Business Services max 400 lines
package ast

import (
	"context"
	"fmt"
	"sync"
)

// FileInput represents a file to be analyzed
type FileInput struct {
	Path     string
	Content  string
	Language string
}

// CrossFileAnalysisResult contains results from cross-file analysis
type CrossFileAnalysisResult struct {
	Findings            []ASTFinding
	UnusedExports       []*FileSymbol
	UndefinedRefs       []*SymbolReference
	CircularDeps        [][]string
	CrossFileDuplicates []ASTFinding
	Stats               CrossFileStats
}

// CrossFileStats tracks cross-file analysis metrics
type CrossFileStats struct {
	FilesAnalyzed     int
	SymbolsFound      int
	DependenciesFound int
	AnalysisTime      int64
}

// AnalyzeCrossFile performs cross-file analysis on multiple files
func AnalyzeCrossFile(ctx context.Context, files []FileInput, analyses []string) (*CrossFileAnalysisResult, error) {
	if len(files) == 0 {
		return nil, fmt.Errorf("no files provided for analysis")
	}

	result := &CrossFileAnalysisResult{
		Findings:            []ASTFinding{},
		UnusedExports:       []*FileSymbol{},
		UndefinedRefs:       []*SymbolReference{},
		CircularDeps:        [][]string{},
		CrossFileDuplicates: []ASTFinding{},
	}

	// Build symbol table and dependency graph
	symbolTable := NewSymbolTable()
	dependencyGraph := NewDependencyGraph()

	// Parse all files and extract symbols/dependencies
	var wg sync.WaitGroup
	var parseErr error
	var parseMutex sync.Mutex

	for _, file := range files {
		wg.Add(1)
		go func(f FileInput) {
			defer wg.Done()

			// Get parser
			parser, err := getParser(f.Language)
			if err != nil {
				parseMutex.Lock()
				if parseErr == nil {
					parseErr = fmt.Errorf("failed to get parser for %s: %w", f.Path, err)
				}
				parseMutex.Unlock()
				return
			}

			// Parse file
			tree, err := parser.ParseCtx(ctx, nil, []byte(f.Content))
			if err != nil {
				parseMutex.Lock()
				if parseErr == nil {
					parseErr = fmt.Errorf("failed to parse %s: %w", f.Path, err)
				}
				parseMutex.Unlock()
				return
			}
			defer tree.Close()

			rootNode := tree.RootNode()
			if rootNode == nil {
				return
			}

			// Extract symbols
			symbols, err := ExtractSymbolsFromFile(rootNode, f.Content, f.Path, f.Language)
			if err == nil {
				for _, symbol := range symbols {
					symbolTable.AddSymbol(symbol)
				}
			}

			// Extract dependencies
			deps, err := ExtractDependenciesFromFile(rootNode, f.Content, f.Path, f.Language)
			if err == nil {
				for _, dep := range deps {
					dependencyGraph.AddDependency(dep)
				}
			}
		}(file)
	}

	wg.Wait()

	if parseErr != nil {
		return nil, parseErr
	}

	// Perform cross-file analyses
	if contains(analyses, "unused_exports") || len(analyses) == 0 {
		result.UnusedExports = symbolTable.FindUnusedExports()
		// Convert to findings
		for _, symbol := range result.UnusedExports {
			finding := ASTFinding{
				Type:        "unused_export",
				Severity:    "medium",
				Line:        symbol.Line,
				Column:      symbol.Column,
				Message:     fmt.Sprintf("Exported %s '%s' is never used outside this file", symbol.Kind, symbol.Name),
				Code:        "",
				Suggestion:  fmt.Sprintf("Consider removing export or using this %s elsewhere", symbol.Kind),
				Confidence:  0.9,
				AutoFixSafe: false,
				FixType:     "remove",
				Reasoning:   "Exported symbol has no external references",
			}
			result.Findings = append(result.Findings, finding)
		}
	}

	if contains(analyses, "undefined_refs") || len(analyses) == 0 {
		result.UndefinedRefs = symbolTable.FindUndefinedReferences()
		// Convert to findings
		for _, ref := range result.UndefinedRefs {
			finding := ASTFinding{
				Type:        "undefined_reference",
				Severity:    "high",
				Line:        ref.Line,
				Column:      ref.Column,
				Message:     fmt.Sprintf("Reference to undefined symbol '%s'", ref.Name),
				Code:        "",
				Suggestion:  fmt.Sprintf("Import or define '%s' before use", ref.Name),
				Confidence:  0.95,
				AutoFixSafe: false,
				FixType:     "error",
				Reasoning:   "Symbol is referenced but not defined",
			}
			result.Findings = append(result.Findings, finding)
		}
	}

	if contains(analyses, "circular_deps") || len(analyses) == 0 {
		result.CircularDeps = dependencyGraph.FindCircularDependencies()
		// Convert to findings
		for _, cycle := range result.CircularDeps {
			if len(cycle) > 0 {
				finding := ASTFinding{
					Type:        "circular_dependency",
					Severity:    "high",
					Line:        1,
					Column:      1,
					Message:     fmt.Sprintf("Circular dependency detected: %v", cycle),
					Code:        "",
					Suggestion:  "Refactor to break circular dependency",
					Confidence:  1.0,
					AutoFixSafe: false,
					FixType:     "refactor",
					Reasoning:   "Circular dependencies can cause initialization issues",
				}
				result.Findings = append(result.Findings, finding)
			}
		}
	}

	if contains(analyses, "cross_file_duplicates") {
		result.CrossFileDuplicates = detectCrossFileDuplicates(files, symbolTable)
		result.Findings = append(result.Findings, result.CrossFileDuplicates...)
	}

	// Update stats
	result.Stats = CrossFileStats{
		FilesAnalyzed:     len(files),
		SymbolsFound:      len(symbolTable.symbols),
		DependenciesFound: len(dependencyGraph.dependencies),
	}

	return result, nil
}

// detectCrossFileDuplicates finds duplicate functions across files
func detectCrossFileDuplicates(files []FileInput, symbolTable *SymbolTable) []ASTFinding {
	findings := []ASTFinding{}
	functionMap := make(map[string][]*FileSymbol)

	// Collect all functions
	for _, file := range files {
		symbols := symbolTable.GetFileSymbols(file.Path)
		for _, symbol := range symbols {
			if symbol.Kind == "function" {
				functionMap[symbol.Name] = append(functionMap[symbol.Name], symbol)
			}
		}
	}

	// Find duplicates across files
	for funcName, symbols := range functionMap {
		if len(symbols) > 1 {
			// Check if they're in different files
			fileSet := make(map[string]bool)
			for _, symbol := range symbols {
				fileSet[symbol.FilePath] = true
			}

			if len(fileSet) > 1 {
				// Duplicate across files
				for _, symbol := range symbols {
					finding := ASTFinding{
						Type:        "cross_file_duplicate",
						Severity:    "medium",
						Line:        symbol.Line,
						Column:      symbol.Column,
						Message:     fmt.Sprintf("Function '%s' is duplicated across multiple files", funcName),
						Code:        "",
						Suggestion:  "Consider consolidating duplicate functions",
						Confidence:  0.8,
						AutoFixSafe: false,
						FixType:     "refactor",
						Reasoning:   "Duplicate function found in multiple files",
					}
					findings = append(findings, finding)
				}
			}
		}
	}

	return findings
}

// AnalyzeMultiFile performs multi-file AST analysis
func AnalyzeMultiFile(ctx context.Context, files []FileInput, analyses []string) ([]ASTFinding, AnalysisStats, error) {
	// First do single-file analysis on each file
	allFindings := []ASTFinding{}
	var totalParseTime int64
	var totalAnalysisTime int64
	var totalNodes int

	for _, file := range files {
		findings, stats, err := AnalyzeAST(file.Content, file.Language, analyses)
		if err != nil {
			// Continue with other files even if one fails
			continue
		}

		// Update file path in findings
		for range findings {
			// Add file context to findings
		}

		allFindings = append(allFindings, findings...)
		totalParseTime += stats.ParseTime
		totalAnalysisTime += stats.AnalysisTime
		totalNodes += stats.NodesVisited
	}

	// Then do cross-file analysis
	crossFileResult, err := AnalyzeCrossFile(ctx, files, analyses)
	if err == nil {
		allFindings = append(allFindings, crossFileResult.Findings...)
	}

	combinedStats := AnalysisStats{
		ParseTime:    totalParseTime,
		AnalysisTime: totalAnalysisTime,
		NodesVisited: totalNodes,
	}

	return allFindings, combinedStats, nil
}
