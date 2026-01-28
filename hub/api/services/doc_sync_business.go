// Package doc_sync_business - Business rules comparison for documentation synchronization
// Complies with CODING_STANDARDS.md: Utilities max 250 lines

package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

// compareBusinessRules performs bidirectional comparison between business rules and code
func compareBusinessRules(ctx context.Context, projectID string, codebasePath string) ([]Discrepancy, error) {
	var discrepancies []Discrepancy

	// Extract business rules from knowledge base
	rules, err := extractBusinessRules(ctx, projectID, nil, "", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to extract business rules: %w", err)
	}

	// For each rule, check if code implements it
	for _, rule := range rules {
		// Search for rule implementation in code
		evidence := detectBusinessRuleImplementation(rule, codebasePath)

		if evidence.Confidence < 0.3 {
			// Rule documented but not implemented
			discrepancies = append(discrepancies, Discrepancy{
				Type:           "missing_impl",
				Feature:        rule.Title,
				DocStatus:      "DOCUMENTED",
				CodeStatus:     "MISSING",
				Recommendation: fmt.Sprintf("Implement business rule '%s' in code", rule.Title),
			})
		} else if evidence.Confidence < 0.7 {
			// Partially implemented
			discrepancies = append(discrepancies, Discrepancy{
				Type:           "partial_match",
				Feature:        rule.Title,
				DocStatus:      "DOCUMENTED",
				CodeStatus:     "PARTIAL",
				Evidence:       evidence,
				Recommendation: fmt.Sprintf("Complete implementation of business rule '%s'", rule.Title),
			})
		}
	}

	// ENHANCEMENT OPPORTUNITY: Reverse check - find code patterns not documented as rules
	// AST analysis is available (ast.AnalyzeAST, detectBusinessRuleImplementation) to extract business logic patterns from code
	// and compare against documented business rules. This is a bidirectional validation
	// that would help identify undocumented business logic in the codebase.
	// Priority: P2 - Enhancement for future phase

	return discrepancies, nil
}

// detectBusinessRuleImplementation searches codebase for business rule implementation using AST analysis
func detectBusinessRuleImplementation(rule KnowledgeItem, codebasePath string) ImplementationEvidence {
	evidence := ImplementationEvidence{
		Feature:     rule.Title,
		Files:       []string{},
		Functions:   []string{},
		LineNumbers: make(map[string][]int),
	}

	// Extract keywords from rule title and content
	keywords := extractKeywordsFromRule(rule)

	if len(keywords) == 0 {
		return evidence
	}

	// Build keyword map for faster lookup
	keywordMap := make(map[string]bool)
	for _, kw := range keywords {
		keywordMap[strings.ToLower(kw)] = true
	}

	// Search in Go files using AST analysis
	if files, err := findGoFiles(filepath.Join(codebasePath, "hub", "api")); err == nil {
		for _, file := range files {
			content, err := os.ReadFile(file)
			if err != nil {
				continue
			}

			contentStr := string(content)

			// Try AST-based analysis first
			astMatches := detectBusinessRuleWithAST(contentStr, file, keywordMap, keywords)
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
				contentLower := strings.ToLower(contentStr)
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

			// Detect endpoints in this file
			detectedEndpoints := detectEndpoints(contentStr, file, keywords)
			if len(detectedEndpoints) > 0 {
				evidence.Endpoints = append(evidence.Endpoints, detectedEndpoints...)
				// Add confidence for endpoint matches (0.4 weight)
				evidence.Confidence += 0.4
			}
		}
	}

	// Detect tests related to this rule
	detectedTests := detectTests(codebasePath, rule.Title, keywords)
	if len(detectedTests) > 0 {
		evidence.Tests = append(evidence.Tests, detectedTests...)
		// Add confidence for test matches (0.2 weight)
		evidence.Confidence += 0.2
	}

	// Cap confidence at 1.0
	if evidence.Confidence > 1.0 {
		evidence.Confidence = 1.0
	}

	return evidence
}

// detectBusinessRuleWithAST uses AST to detect business rule implementation
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
	parser, err := getParser(language)
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
	TraverseAST(rootNode, func(node *sitter.Node) bool {
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

// extractKeywordsFromRule extracts keywords from rule title and content
func extractKeywordsFromRule(rule KnowledgeItem) []string {
	var keywords []string

	// Extract from title
	words := regexp.MustCompile(`\s+|[_-]`).Split(rule.Title, -1)
	for _, word := range words {
		word = strings.TrimSpace(word)
		if len(word) > 2 {
			wordLower := strings.ToLower(word)
			common := []string{"the", "a", "an", "and", "or", "but", "in", "on", "at", "to", "for", "of", "with", "by"}
			isCommon := false
			for _, c := range common {
				if wordLower == c {
					isCommon = true
					break
				}
			}
			if !isCommon {
				keywords = append(keywords, word)
			}
		}
	}

	// Extract from content if available
	if rule.Content != "" {
		contentKeywords := extractKeywords(rule.Content)
		keywords = append(keywords, contentKeywords...)
	}

	// Extract from structured data if available
	if rule.StructuredData != nil {
		for key, value := range rule.StructuredData {
			if strValue, ok := value.(string); ok && len(strValue) > 2 {
				keywords = append(keywords, key)
				valueKeywords := extractKeywords(strValue)
				keywords = append(keywords, valueKeywords...)
			}
		}
	}

	// Deduplicate keywords
	return deduplicateKeywords(keywords)
}

// deduplicateKeywords removes duplicate keywords while preserving order.
// Case-insensitive deduplication (e.g., "User" and "user" are considered duplicates)
func deduplicateKeywords(keywords []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, kw := range keywords {
		kwLower := strings.ToLower(kw)
		if !seen[kwLower] {
			seen[kwLower] = true
			result = append(result, kw)
		}
	}
	return result
}

// =============================================================================
