// Package main business rule detection utilities
// Business rule detection functions using AST analysis
// Complies with CODING_STANDARDS.md: Utilities max 250 lines

package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"sentinel-hub-api/ast"

	sitter "github.com/smacker/go-tree-sitter"
)

// detectBusinessRuleImplementation searches codebase for business rule implementation using AST analysis.
// This is a functional implementation that scans the full codebase and uses AST parsing for accurate detection.
func detectBusinessRuleImplementation(rule KnowledgeItem, codebasePath string) ImplementationEvidence {
	evidence := ImplementationEvidence{
		Feature:     rule.Title,
		Files:       []string{},
		Functions:   []string{},
		Endpoints:   []string{},
		Tests:       []string{},
		Confidence:  0.0,
		LineNumbers: make(map[string][]int),
	}

	// Extract keywords from rule title and content using shared utility
	keywords := extractKeywords(rule.Title)
	if rule.Content != "" {
		contentKeywords := extractKeywords(rule.Content)
		keywords = append(keywords, contentKeywords...)
	}

	if len(keywords) == 0 {
		return evidence
	}

	// Build keyword map for faster lookup
	keywordMap := make(map[string]bool)
	for _, kw := range keywords {
		keywordMap[strings.ToLower(kw)] = true
	}

	// Scan codebase for source files (full codebase, not just hub/api)
	sourceFiles := scanSourceFiles(codebasePath)
	for _, file := range sourceFiles {
		content, err := os.ReadFile(file)
		if err != nil {
			// Skip files we can't read, continue processing other files
			// Error is silently ignored to allow processing to continue
			continue
		}

		// Try AST-based analysis first
		astMatches := detectBusinessRuleWithAST(string(content), file, keywordMap, keywords)
		if astMatches.Confidence > 0 {
			// AST found matches
			evidence.Files = appendIfNotExists(evidence.Files, file)
			evidence.Functions = append(evidence.Functions, astMatches.Functions...)
			for funcName, lines := range astMatches.LineNumbers {
				evidence.LineNumbers[funcName] = append(evidence.LineNumbers[funcName], lines...)
			}
			evidence.Confidence += astMatches.Confidence
		} else {
			// Fallback to keyword matching
			contentLower := strings.ToLower(string(content))
			matches := 0
			for _, keyword := range keywords {
				if strings.Contains(contentLower, strings.ToLower(keyword)) {
					matches++
				}
			}

			if matches >= len(keywords)/2 {
				evidence.Files = appendIfNotExists(evidence.Files, file)
				evidence.Confidence += 0.2 // Lower confidence for keyword-only matches
			}
		}
	}

	// Cap confidence at 1.0
	if evidence.Confidence > 1.0 {
		evidence.Confidence = 1.0
	}

	return evidence
}

// detectBusinessRuleWithAST uses AST to detect business rule implementation.
// Parameters:
//   - code: Source code content to analyze
//   - filePath: Path to the source file (used for language detection)
//   - keywordMap: Map of keywords for fast lookup (lowercase keys)
//   - keywords: List of keywords to match against
//
// Returns: ImplementationEvidence with detected functions, line numbers, and confidence score
func detectBusinessRuleWithAST(code string, filePath string, keywordMap map[string]bool, keywords []string) ImplementationEvidence {
	evidence := ImplementationEvidence{
		Files:       []string{},
		Functions:   []string{},
		LineNumbers: make(map[string][]int),
		Confidence:  0.0,
	}

	// Determine language from file path
	ext := strings.ToLower(filepath.Ext(filePath))
	var language string
	switch ext {
	case ".go":
		language = "go"
	case ".js", ".jsx":
		language = "javascript"
	case ".ts", ".tsx":
		language = "typescript"
	case ".py":
		language = "python"
	default:
		// Unsupported language, return empty evidence
		return evidence
	}

	// Parse code into AST
	parser, err := ast.GetParser(language)
	if err != nil {
		return evidence
	}

	ctx := context.Background()
	tree, err := parser.ParseCtx(ctx, nil, []byte(code))
	if err != nil || tree == nil {
		return evidence
	}
	defer tree.Close()

	rootNode := tree.RootNode()
	if rootNode == nil {
		return evidence
	}

	// Extract function and class definitions from AST
	ast.TraverseAST(rootNode, func(node *sitter.Node) bool {
		var funcName string
		var isFunction bool
		var line int

		if node.Type() == "function_declaration" || node.Type() == "method_declaration" {
			// Extract function name
			for i := 0; i < int(node.ChildCount()); i++ {
				child := node.Child(i)
				if child != nil {
					if child.Type() == "identifier" || child.Type() == "field_identifier" {
						funcName = code[child.StartByte():child.EndByte()]
						isFunction = true
						line, _ = getLineColumn(code, int(node.StartByte()))
						break
					}
				}
			}

			if isFunction && funcName != "" {
				funcNameLower := strings.ToLower(funcName)

				// Check if function name matches keywords
				matches := 0
				for keyword := range keywordMap {
					if strings.Contains(funcNameLower, keyword) || strings.Contains(keyword, funcNameLower) {
						matches++
					}
				}

				if matches > 0 {
					// Function name matches - high confidence
					evidence.Functions = appendIfNotExists(evidence.Functions, funcName)
					evidence.LineNumbers[funcName] = append(evidence.LineNumbers[funcName], line)
					evidence.Confidence += 0.5 // High weight for function name matches
				} else {
					// Check function signature and body for keyword matches
					funcCode := code[node.StartByte():node.EndByte()]
					funcCodeLower := strings.ToLower(funcCode)

					keywordMatches := 0
					for _, keyword := range keywords {
						if strings.Contains(funcCodeLower, strings.ToLower(keyword)) {
							keywordMatches++
						}
					}

					if keywordMatches >= len(keywords)/2 {
						// Keywords found in function - medium confidence
						evidence.Functions = appendIfNotExists(evidence.Functions, funcName)
						evidence.LineNumbers[funcName] = append(evidence.LineNumbers[funcName], line)
						evidence.Confidence += 0.3 // Medium weight for keyword matches in function
					}
				}
			}
		}

		return true
	})

	return evidence
}

// scanSourceFiles scans the codebase for source files (excluding test files and vendor directories).
// Parameters:
//   - codebasePath: Root directory to scan
//
// Returns: List of source file paths (Go, JavaScript, TypeScript, Python)
// Note: Excludes test files, vendor directories, and common build/cache directories
func scanSourceFiles(codebasePath string) []string {
	var files []string
	supportedExts := map[string]bool{
		".go": true, ".js": true, ".jsx": true, ".ts": true, ".tsx": true, ".py": true,
	}

	err := filepath.Walk(codebasePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we can't access
		}

		// Skip directories
		if info.IsDir() {
			// Skip common directories that shouldn't be scanned
			dirName := info.Name()
			if dirName == "vendor" || dirName == "node_modules" || dirName == ".git" ||
				dirName == ".idea" || dirName == ".vscode" || dirName == "__pycache__" {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if file has supported extension
		ext := strings.ToLower(filepath.Ext(path))
		if supportedExts[ext] {
			// Skip test files for now (can be enhanced later)
			if !strings.Contains(path, "_test.") && !strings.Contains(path, ".test.") &&
				!strings.Contains(path, ".spec.") && !strings.HasSuffix(path, "_test.go") {
				files = append(files, path)
			}
		}

		return nil
	})

	if err != nil {
		// Return empty slice on error to allow processing to continue
		// Error is silently ignored to prevent breaking the detection process
		return []string{}
	}

	if files == nil {
		return []string{}
	}
	return files
}
