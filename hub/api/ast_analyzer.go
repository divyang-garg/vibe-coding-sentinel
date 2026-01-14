// AST Analyzer - Phase 6 Implementation
// Tree-sitter based code analysis for vibe coding detection

package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"
	"github.com/smacker/go-tree-sitter/javascript"
	"github.com/smacker/go-tree-sitter/python"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
)

// AnalysisStats tracks performance metrics for AST analysis
type AnalysisStats struct {
	ParseTime    int64
	AnalysisTime int64
	NodesVisited int
}

// LanguageParser maps language names to Tree-sitter parsers
type LanguageParser struct {
	Language string
	Parser   *sitter.Parser
}

var parsers = make(map[string]*sitter.Parser)

// AST cache for performance optimization
type cacheEntry struct {
	Findings []ASTFinding
	Stats    AnalysisStats
	Expires  time.Time
}

var astCache = make(map[string]*cacheEntry)
var cacheMutex sync.RWMutex
var cacheTTL = 5 * time.Minute // Cache AST results for 5 minutes
var lastCacheCleanup time.Time
var cacheCleanupInterval = 5 * time.Minute

// initParsers initializes Tree-sitter parsers for supported languages
func initParsers() {
	// Go parser
	goParser := sitter.NewParser()
	goParser.SetLanguage(golang.GetLanguage())
	parsers["go"] = goParser
	parsers["golang"] = goParser

	// JavaScript parser
	jsParser := sitter.NewParser()
	jsParser.SetLanguage(javascript.GetLanguage())
	parsers["javascript"] = jsParser
	parsers["js"] = jsParser
	parsers["jsx"] = jsParser

	// TypeScript parser
	tsParser := sitter.NewParser()
	tsParser.SetLanguage(typescript.GetLanguage())
	parsers["typescript"] = tsParser
	parsers["ts"] = tsParser
	parsers["tsx"] = tsParser

	// Python parser
	pyParser := sitter.NewParser()
	pyParser.SetLanguage(python.GetLanguage())
	parsers["python"] = pyParser
	parsers["py"] = pyParser
}

// getParser returns the appropriate parser for a language
func getParser(language string) (*sitter.Parser, error) {
	// Normalize language name
	lang := normalizeLanguage(language)

	// Initialize parsers if not already done
	if len(parsers) == 0 {
		initParsers()
	}

	if parser, ok := parsers[lang]; ok {
		return parser, nil
	}

	return nil, fmt.Errorf("unsupported language: %s (supported: go, javascript, typescript, python)", language)
}

// normalizeLanguage normalizes language names to standard keys
func normalizeLanguage(lang string) string {
	lang = strings.ToLower(lang)
	switch lang {
	case "js", "javascript", "jsx":
		return "javascript"
	case "ts", "typescript", "tsx":
		return "typescript"
	case "py", "python":
		return "python"
	case "go", "golang":
		return "go"
	default:
		return lang
	}
}

// getCacheKey generates a cache key from code, language, and analyses
func getCacheKey(code string, language string, analyses []string) string {
	hash := sha256.Sum256([]byte(code + language + strings.Join(analyses, ",")))
	return hex.EncodeToString(hash[:])
}

// analyzeAST performs AST analysis on code (with caching)
func analyzeAST(code string, language string, analyses []string) ([]ASTFinding, AnalysisStats, error) {
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

// cleanExpiredCacheEntries removes expired cache entries
func cleanExpiredCacheEntries() {
	now := time.Now()
	for key, entry := range astCache {
		if now.After(entry.Expires) {
			delete(astCache, key)
		}
	}
}

// detectDuplicateFunctions detects duplicate function definitions
func detectDuplicateFunctions(root *sitter.Node, code string, language string) []ASTFinding {
	findings := []ASTFinding{}
	functionMap := make(map[string][]*sitter.Node)

	// Traverse AST to find all function definitions
	traverseAST(root, func(node *sitter.Node) bool {
		var funcName string
		var isFunction bool

		switch language {
		case "go":
			if node.Type() == "function_declaration" || node.Type() == "method_declaration" {
				// For method_declaration, format is: receiver method_name
				// For function_declaration, format is: func_name
				for i := 0; i < int(node.ChildCount()); i++ {
					child := node.Child(i)
					if child != nil {
						if child.Type() == "identifier" {
							funcName = code[child.StartByte():child.EndByte()]
							isFunction = true
							break
						} else if child.Type() == "parameter_list" {
							// This is a method receiver - get the method name after it
							continue
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
				// Get function name - could be identifier or property_identifier
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
				// Arrow functions assigned to variables - check parent
				parent := node.Parent()
				if parent != nil && parent.Type() == "variable_declarator" {
					// Get the variable name (which is the function name)
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
				// Get function name
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
			functionMap[funcName] = append(functionMap[funcName], node)
		}

		return true // Continue traversal
	})

	// Check for duplicates
	for funcName, nodes := range functionMap {
		if len(nodes) > 1 {
			// Found duplicate - report all occurrences
			for _, node := range nodes {
				startLine, startCol := getLineColumn(code, int(node.StartByte()))
				endLine, endCol := getLineColumn(code, int(node.EndByte()))

				findings = append(findings, ASTFinding{
					Type:       "duplicate_function",
					Severity:   "error",
					Line:       startLine,
					Column:     startCol,
					EndLine:    endLine,
					EndColumn:  endCol,
					Message:    fmt.Sprintf("Duplicate function definition: '%s' is defined %d times", funcName, len(nodes)),
					Code:       code[node.StartByte():node.EndByte()],
					Suggestion: fmt.Sprintf("Remove duplicate definition of '%s' or rename one of them", funcName),
				})
			}
		}
	}

	return findings
}

// detectUnusedVariables detects unused variable declarations
func detectUnusedVariables(root *sitter.Node, code string, language string) []ASTFinding {
	findings := []ASTFinding{}
	variableDeclarations := make(map[string]*sitter.Node)
	variableUsages := make(map[string]bool)

	// First pass: collect all variable declarations
	traverseAST(root, func(node *sitter.Node) bool {
		var varName string
		var isDeclaration bool

		switch language {
		case "go":
			if node.Type() == "short_var_declaration" || node.Type() == "var_declaration" {
				// Get variable name
				for i := 0; i < int(node.ChildCount()); i++ {
					child := node.Child(i)
					if child != nil && child.Type() == "identifier" {
						varName = code[child.StartByte():child.EndByte()]
						isDeclaration = true
						break
					}
				}
			}
		case "javascript", "typescript":
			if node.Type() == "variable_declaration" {
				// Get variable name
				for i := 0; i < int(node.ChildCount()); i++ {
					child := node.Child(i)
					if child != nil && child.Type() == "variable_declarator" {
						// Get identifier from declarator
						for j := 0; j < int(child.ChildCount()); j++ {
							grandchild := child.Child(j)
							if grandchild != nil && grandchild.Type() == "identifier" {
								varName = code[grandchild.StartByte():grandchild.EndByte()]
								isDeclaration = true
								break
							}
						}
						break
					}
				}
			}
		case "python":
			if node.Type() == "assignment" {
				// Get variable name
				for i := 0; i < int(node.ChildCount()); i++ {
					child := node.Child(i)
					if child != nil && child.Type() == "identifier" {
						varName = code[child.StartByte():child.EndByte()]
						isDeclaration = true
						break
					}
				}
			}
		}

		if isDeclaration && varName != "" {
			variableDeclarations[varName] = node
		}

		return true
	})

	// Second pass: collect all variable usages (excluding declarations)
	traverseAST(root, func(node *sitter.Node) bool {
		if node.Type() == "identifier" {
			varName := code[node.StartByte():node.EndByte()]

			// Check if this identifier is a usage (not a declaration)
			if declNode, isDeclared := variableDeclarations[varName]; isDeclared {
				// Only count as usage if it's not the declaration itself
				// Also check parent to avoid counting declaration nodes
				parent := node.Parent()
				isInDeclaration := false

				// Check if this identifier is part of a declaration
				if parent != nil {
					parentType := parent.Type()
					switch language {
					case "go":
						if parentType == "short_var_declaration" || parentType == "var_declaration" {
							isInDeclaration = true
						}
					case "javascript", "typescript":
						if parentType == "variable_declarator" {
							isInDeclaration = true
						}
					case "python":
						if parentType == "assignment" {
							isInDeclaration = true
						}
					}
				}

				if !isInDeclaration && node.StartByte() != declNode.StartByte() {
					variableUsages[varName] = true
				}
			}
		}
		return true
	})

	// Find unused variables
	for varName, declNode := range variableDeclarations {
		if !variableUsages[varName] {
			startLine, startCol := getLineColumn(code, int(declNode.StartByte()))
			endLine, endCol := getLineColumn(code, int(declNode.EndByte()))

			findings = append(findings, ASTFinding{
				Type:       "unused_variable",
				Severity:   "warning",
				Line:       startLine,
				Column:     startCol,
				EndLine:    endLine,
				EndColumn:  endCol,
				Message:    fmt.Sprintf("Unused variable: '%s' is declared but never used", varName),
				Code:       code[declNode.StartByte():declNode.EndByte()],
				Suggestion: fmt.Sprintf("Remove unused variable '%s' or use it in your code", varName),
			})
		}
	}

	return findings
}

// detectUnreachableCode detects code after return/throw statements
func detectUnreachableCode(root *sitter.Node, code string, language string) []ASTFinding {
	findings := []ASTFinding{}

	// Track return/throw/raise statements and their parent blocks (Phase 7C enhancement)
	type returnInfo struct {
		node     *sitter.Node
		parent   *sitter.Node
		stmtType string // "return", "throw", "raise"
	}
	returns := []returnInfo{}

	// First pass: collect all return/throw/raise statements
	traverseAST(root, func(node *sitter.Node) bool {
		var stmtType string
		isTerminal := false

		switch language {
		case "go":
			if node.Type() == "return_statement" || node.Type() == "return" {
				stmtType = "return"
				isTerminal = true
			}
		case "javascript", "typescript":
			if node.Type() == "return_statement" || node.Type() == "return" {
				stmtType = "return"
				isTerminal = true
			} else if node.Type() == "throw_statement" || node.Type() == "throw" {
				stmtType = "throw"
				isTerminal = true
			}
		case "python":
			if node.Type() == "return_statement" || node.Type() == "return" {
				stmtType = "return"
				isTerminal = true
			} else if node.Type() == "raise_statement" || node.Type() == "raise" {
				stmtType = "raise"
				isTerminal = true
			}
		}

		if isTerminal {
			parent := node.Parent()
			if parent != nil {
				returns = append(returns, returnInfo{node: node, parent: parent, stmtType: stmtType})
			}
		}
		return true
	})

	// Second pass: check for statements after returns
	for _, ret := range returns {
		returnIndex := -1
		// Find return statement index in parent's children
		for i := 0; i < int(ret.parent.ChildCount()); i++ {
			if ret.parent.Child(i) == ret.node {
				returnIndex = i
				break
			}
		}

		// Check for statements after return
		if returnIndex >= 0 {
			for i := returnIndex + 1; i < int(ret.parent.ChildCount()); i++ {
				nextChild := ret.parent.Child(i)
				if nextChild != nil && isStatementNode(nextChild, language) {
					// Skip comments and whitespace
					childCode := code[nextChild.StartByte():nextChild.EndByte()]
					if strings.TrimSpace(childCode) == "" || strings.HasPrefix(strings.TrimSpace(childCode), "//") {
						continue
					}

					startLine, startCol := getLineColumn(code, int(nextChild.StartByte()))
					endLine, endCol := getLineColumn(code, int(nextChild.EndByte()))

					stmtTypeMsg := "return statement"
					if ret.stmtType == "throw" {
						stmtTypeMsg = "throw statement"
					} else if ret.stmtType == "raise" {
						stmtTypeMsg = "raise statement"
					}

					findings = append(findings, ASTFinding{
						Type:       "unreachable_code",
						Severity:   "warning",
						Line:       startLine,
						Column:     startCol,
						EndLine:    endLine,
						EndColumn:  endCol,
						Message:    fmt.Sprintf("Unreachable code: statements after %s", stmtTypeMsg),
						Code:       childCode,
						Suggestion: fmt.Sprintf("Remove unreachable code or move %s", stmtTypeMsg),
					})
				}
			}
		}
	}

	return findings
}

// detectOrphanedCode detects code outside valid scope (e.g., statements outside functions/classes)
func detectOrphanedCode(root *sitter.Node, code string, language string) []ASTFinding {
	findings := []ASTFinding{}

	// Track valid scopes (functions, classes, modules)
	validScopes := make(map[*sitter.Node]bool)

	// First pass: identify valid scopes
	traverseAST(root, func(node *sitter.Node) bool {
		switch language {
		case "go":
			if node.Type() == "function_declaration" || node.Type() == "method_declaration" {
				validScopes[node] = true
			}
		case "javascript", "typescript":
			if node.Type() == "function_declaration" || node.Type() == "function" ||
				node.Type() == "class_declaration" || node.Type() == "arrow_function" {
				validScopes[node] = true
			}
		case "python":
			if node.Type() == "function_definition" || node.Type() == "class_definition" {
				validScopes[node] = true
			}
		}
		return true
	})

	// Helper to check if a node is within a valid scope
	isInValidScope := func(node *sitter.Node) bool {
		current := node
		for current != nil {
			if validScopes[current] {
				return true
			}
			current = current.Parent()
		}
		return false
	}

	// Second pass: find orphaned statements (statements not in valid scope)
	traverseAST(root, func(node *sitter.Node) bool {
		// Check for top-level statements that aren't in valid scopes
		if isStatementNode(node, language) && !isInValidScope(node) {
			// Check if parent is the root (top-level statement)
			parent := node.Parent()
			if parent != nil && parent == root {
				// This is an orphaned top-level statement
				startLine, startCol := getLineColumn(code, int(node.StartByte()))
				endLine, endCol := getLineColumn(code, int(node.EndByte()))

				// Skip import/export statements (they're valid at top level)
				nodeCode := code[node.StartByte():node.EndByte()]
				if strings.Contains(strings.ToLower(nodeCode), "import") ||
					strings.Contains(strings.ToLower(nodeCode), "export") ||
					strings.Contains(strings.ToLower(nodeCode), "package") {
					return true
				}

				findings = append(findings, ASTFinding{
					Type:       "orphaned_code",
					Severity:   "error",
					Line:       startLine,
					Column:     startCol,
					EndLine:    endLine,
					EndColumn:  endCol,
					Message:    "Orphaned code: statement outside function/class scope",
					Code:       nodeCode,
					Suggestion: "Move this code into a function or remove it if unnecessary",
				})
			}
		}
		return true
	})

	return findings
}

// Helper functions

func traverseAST(node *sitter.Node, visitor func(*sitter.Node) bool) {
	if node == nil {
		return
	}

	if !visitor(node) {
		return
	}

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child != nil {
			traverseAST(child, visitor)
		}
	}
}

func countNodes(node *sitter.Node) int {
	if node == nil {
		return 0
	}

	count := 1
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child != nil {
			count += countNodes(child)
		}
	}

	return count
}

func getLineColumn(code string, byteOffset int) (line, column int) {
	line = 1
	column = 0

	for i := 0; i < byteOffset && i < len(code); i++ {
		if code[i] == '\n' {
			line++
			column = 0
		} else {
			column++
		}
	}

	return line, column
}

func isStatementNode(node *sitter.Node, language string) bool {
	var stmtTypes []string

	switch language {
	case "go":
		stmtTypes = []string{
			"expression_statement",
			"if_statement",
			"for_statement",
			"short_var_declaration",
			"var_declaration",
			"assignment_statement",
			"call_expression",
			"return_statement",
		}
	case "javascript", "typescript":
		stmtTypes = []string{
			"expression_statement",
			"if_statement",
			"for_statement",
			"while_statement",
			"for_in_statement",
			"for_of_statement",
			"variable_declaration",
			"assignment_expression",
			"call_expression",
			"return_statement",
		}
	case "python":
		stmtTypes = []string{
			"expression_statement",
			"if_statement",
			"for_statement",
			"while_statement",
			"assignment",
			"call",
			"return_statement",
		}
	default:
		stmtTypes = []string{
			"expression_statement",
			"if_statement",
			"for_statement",
			"while_statement",
			"assignment",
			"call_expression",
		}
	}

	nodeType := node.Type()
	for _, stmtType := range stmtTypes {
		if nodeType == stmtType {
			return true
		}
	}

	return false
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// detectEmptyCatchBlocks detects empty catch/except blocks (Phase 7C)
func detectEmptyCatchBlocks(root *sitter.Node, code string, language string) []ASTFinding {
	findings := []ASTFinding{}

	traverseAST(root, func(node *sitter.Node) bool {
		var catchBlock *sitter.Node
		var catchStart, catchEnd uint32

		switch language {
		case "javascript", "typescript":
			if node.Type() == "catch_clause" {
				catchBlock = node
				catchStart = node.StartByte()
				catchEnd = node.EndByte()
			}
		case "python":
			if node.Type() == "except_clause" {
				catchBlock = node
				catchStart = node.StartByte()
				catchEnd = node.EndByte()
			}
		case "go":
			// Go doesn't have catch blocks, but has error handling patterns
			// Check for empty error handling in if err != nil blocks
			if node.Type() == "if_statement" {
				// Check if this is an error check with empty body
				condition := ""
				for i := 0; i < int(node.ChildCount()); i++ {
					child := node.Child(i)
					if child != nil && child.Type() == "condition" {
						condition = code[child.StartByte():child.EndByte()]
						if strings.Contains(condition, "err") && strings.Contains(condition, "!=") {
							// Check if body is empty
							for j := 0; j < int(node.ChildCount()); j++ {
								body := node.Child(j)
								if body != nil && body.Type() == "block" {
									bodyCode := code[body.StartByte():body.EndByte()]
									if strings.TrimSpace(bodyCode) == "{}" || strings.TrimSpace(bodyCode) == "" {
										catchBlock = node
										catchStart = node.StartByte()
										catchEnd = node.EndByte()
										break
									}
								}
							}
						}
						break
					}
				}
			}
		}

		if catchBlock != nil {
			// Check if catch block is empty (only whitespace/comments)
			blockCode := code[catchStart:catchEnd]
			trimmed := strings.TrimSpace(blockCode)

			// Remove catch/except keyword and braces to check if body is empty
			bodyStart := strings.Index(trimmed, "{")
			bodyEnd := strings.LastIndex(trimmed, "}")
			if bodyStart >= 0 && bodyEnd > bodyStart {
				bodyContent := trimmed[bodyStart+1 : bodyEnd]
				bodyContent = strings.TrimSpace(bodyContent)

				// Check if body is empty or only contains comments
				isEmpty := bodyContent == "" ||
					strings.HasPrefix(bodyContent, "//") ||
					strings.HasPrefix(bodyContent, "/*") ||
					strings.HasPrefix(bodyContent, "#")

				if isEmpty {
					startLine, startCol := getLineColumn(code, int(catchStart))
					endLine, endCol := getLineColumn(code, int(catchEnd))

					langSpecific := "catch"
					if language == "python" {
						langSpecific = "except"
					} else if language == "go" {
						langSpecific = "error handling"
					}

					findings = append(findings, ASTFinding{
						Type:       "empty_catch",
						Severity:   "warning",
						Line:       startLine,
						Column:     startCol,
						EndLine:    endLine,
						EndColumn:  endCol,
						Message:    fmt.Sprintf("Empty %s block: errors are silently ignored", langSpecific),
						Code:       blockCode,
						Suggestion: fmt.Sprintf("Add error handling logic in %s block or log the error", langSpecific),
					})
				}
			}
		}

		return true
	})

	return findings
}

// detectMissingAwait detects async function calls without await (Phase 7C)
func detectMissingAwait(root *sitter.Node, code string, language string) []ASTFinding {
	findings := []ASTFinding{}

	// Only for JavaScript/TypeScript and Python (async/await languages)
	if language != "javascript" && language != "typescript" && language != "python" {
		return findings
	}

	// Track async function definitions
	asyncFunctions := make(map[string]bool)

	// First pass: identify async functions
	traverseAST(root, func(node *sitter.Node) bool {
		var funcName string
		isAsync := false

		switch language {
		case "javascript", "typescript":
			if node.Type() == "function_declaration" || node.Type() == "arrow_function" {
				// Check for async keyword
				for i := 0; i < int(node.ChildCount()); i++ {
					child := node.Child(i)
					if child != nil {
						if child.Type() == "async" || strings.Contains(code[child.StartByte():child.EndByte()], "async") {
							isAsync = true
						}
						if child.Type() == "identifier" || child.Type() == "property_identifier" {
							funcName = code[child.StartByte():child.EndByte()]
						}
					}
				}
			}
		case "python":
			if node.Type() == "function_definition" {
				// Check for async keyword
				for i := 0; i < int(node.ChildCount()); i++ {
					child := node.Child(i)
					if child != nil {
						if child.Type() == "async" || strings.Contains(code[child.StartByte():child.EndByte()], "async") {
							isAsync = true
						}
						if child.Type() == "identifier" {
							funcName = code[child.StartByte():child.EndByte()]
						}
					}
				}
			}
		}

		if isAsync && funcName != "" {
			asyncFunctions[funcName] = true
		}

		return true
	})

	// Second pass: find calls to async functions without await
	traverseAST(root, func(node *sitter.Node) bool {
		if node.Type() == "call_expression" {
			// Get function name being called
			var callName string
			for i := 0; i < int(node.ChildCount()); i++ {
				child := node.Child(i)
				if child != nil {
					if child.Type() == "identifier" || child.Type() == "member_expression" {
						callName = code[child.StartByte():child.EndByte()]
						// Extract just the function name (handle obj.method)
						if strings.Contains(callName, ".") {
							parts := strings.Split(callName, ".")
							callName = parts[len(parts)-1]
						}
						break
					}
				}
			}

			// Check if this is a call to an async function
			if asyncFunctions[callName] {
				// Check if parent has await
				parent := node.Parent()
				hasAwait := false
				if parent != nil {
					parentCode := code[parent.StartByte():parent.EndByte()]
					// Check if await is in the parent context
					// Look backwards from the call to find await
					callStart := int(node.StartByte())
					contextStart := callStart - 50 // Check 50 bytes before
					if contextStart < 0 {
						contextStart = 0
					}
					contextCode := code[contextStart:callStart]
					hasAwait = strings.Contains(contextCode, "await") || strings.Contains(parentCode, "await")
				}

				if !hasAwait {
					startLine, startCol := getLineColumn(code, int(node.StartByte()))
					endLine, endCol := getLineColumn(code, int(node.EndByte()))

					findings = append(findings, ASTFinding{
						Type:       "missing_await",
						Severity:   "warning",
						Line:       startLine,
						Column:     startCol,
						EndLine:    endLine,
						EndColumn:  endCol,
						Message:    fmt.Sprintf("Async function '%s' called without await", callName),
						Code:       code[node.StartByte():node.EndByte()],
						Suggestion: "Add 'await' keyword before the async function call",
					})
				}
			}
		}

		return true
	})

	return findings
}

// detectBraceMismatch detects brace/bracket mismatches from parser errors (Phase 7C)
func detectBraceMismatch(tree *sitter.Tree, code string, language string) []ASTFinding {
	findings := []ASTFinding{}

	// Check for parse errors (Tree-sitter marks errors in the tree)
	rootNode := tree.RootNode()
	if rootNode == nil {
		return findings
	}

	// Check if tree has errors (Tree-sitter includes ERROR nodes)
	hasError := false
	traverseAST(rootNode, func(node *sitter.Node) bool {
		if node.Type() == "ERROR" || strings.Contains(node.Type(), "ERROR") {
			hasError = true
			// Try to identify if it's a brace mismatch
			errorCode := code[node.StartByte():node.EndByte()]

			// Count braces/brackets in the error region
			openBraces := strings.Count(errorCode, "{") + strings.Count(errorCode, "[") + strings.Count(errorCode, "(")
			closeBraces := strings.Count(errorCode, "}") + strings.Count(errorCode, "]") + strings.Count(errorCode, ")")

			if openBraces != closeBraces {
				startLine, startCol := getLineColumn(code, int(node.StartByte()))
				endLine, endCol := getLineColumn(code, int(node.EndByte()))

				diff := openBraces - closeBraces
				var message string
				if diff > 0 {
					message = fmt.Sprintf("Missing %d closing brace(s)/bracket(s)", diff)
				} else {
					message = fmt.Sprintf("Extra %d closing brace(s)/bracket(s)", -diff)
				}

				findings = append(findings, ASTFinding{
					Type:       "brace_mismatch",
					Severity:   "error",
					Line:       startLine,
					Column:     startCol,
					EndLine:    endLine,
					EndColumn:  endCol,
					Message:    message,
					Code:       errorCode,
					Suggestion: "Check for mismatched braces, brackets, or parentheses",
				})
			}
			return false // Don't traverse children of error nodes
		}
		return true
	})

	// If no explicit ERROR nodes, check for unbalanced braces in the code
	if !hasError {
		// Simple brace counting (can be enhanced)
		openCount := 0
		closeCount := 0
		lines := strings.Split(code, "\n")

		for lineNum, line := range lines {
			openCount += strings.Count(line, "{") + strings.Count(line, "[") + strings.Count(line, "(")
			closeCount += strings.Count(line, "}") + strings.Count(line, "]") + strings.Count(line, ")")

			// If we have a significant mismatch, report it
			diff := openCount - closeCount
			if diff > 3 || diff < -3 {
				findings = append(findings, ASTFinding{
					Type:       "brace_mismatch",
					Severity:   "error",
					Line:       lineNum + 1,
					Column:     0,
					EndLine:    lineNum + 1,
					EndColumn:  len(line),
					Message:    fmt.Sprintf("Possible brace mismatch detected (difference: %d)", diff),
					Code:       line,
					Suggestion: "Check for missing or extra braces, brackets, or parentheses",
				})
				break // Report first significant mismatch
			}
		}
	}

	return findings
}

// =============================================================================
// Phase 6F: Cross-File Analysis Implementation
// =============================================================================

// SymbolInfo represents a symbol (function, class, variable) definition
type SymbolInfo struct {
	Name       string
	Type       string // "function", "class", "variable", "export"
	File       string
	Line       int
	Column     int
	Signature  string // Function signature (parameters, return type)
	Visibility string // "public", "private", "exported"
	Language   string
}

// SymbolTable tracks all symbols across files
type SymbolTable struct {
	Symbols map[string][]SymbolInfo // symbol name -> list of definitions
	Imports map[string][]string     // file -> list of imported symbols
	Exports map[string][]string     // file -> list of exported symbols
}

// buildSymbolTable builds a symbol table from multiple files (Phase 6F)
func buildSymbolTable(files []struct {
	Path     string
	Code     string
	Language string
}) *SymbolTable {
	table := &SymbolTable{
		Symbols: make(map[string][]SymbolInfo),
		Imports: make(map[string][]string),
		Exports: make(map[string][]string),
	}

	for _, file := range files {
		parser, err := getParser(file.Language)
		if err != nil {
			continue
		}

		ctx := context.Background()
		tree, err := parser.ParseCtx(ctx, nil, []byte(file.Code))
		if err != nil || tree == nil {
			continue
		}

		rootNode := tree.RootNode()
		if rootNode == nil {
			tree.Close() // Close before continue to prevent resource leak
			continue
		}

		// Extract symbols from this file
		extractSymbols(rootNode, file.Code, file.Path, file.Language, table)
		// Extract imports/exports
		extractImportsExports(rootNode, file.Code, file.Path, file.Language, table)
		tree.Close() // Close after use to prevent resource leak
	}

	return table
}

// extractSymbols extracts function/class/variable definitions from a file
func extractSymbols(root *sitter.Node, code string, filePath string, language string, table *SymbolTable) {
	traverseAST(root, func(node *sitter.Node) bool {
		var symbol SymbolInfo
		var found bool

		switch language {
		case "go":
			if node.Type() == "function_declaration" || node.Type() == "method_declaration" {
				// Get function name
				for i := 0; i < int(node.ChildCount()); i++ {
					child := node.Child(i)
					if child != nil {
						if child.Type() == "identifier" || child.Type() == "field_identifier" {
							symbol.Name = code[child.StartByte():child.EndByte()]
							symbol.Type = "function"
							found = true
							break
						}
					}
				}
				// Extract signature
				if found {
					symbol.Signature = extractFunctionSignature(node, code, language)
				}
			}
		case "javascript", "typescript":
			if node.Type() == "function_declaration" || node.Type() == "function" {
				// Get function name
				for i := 0; i < int(node.ChildCount()); i++ {
					child := node.Child(i)
					if child != nil {
						if child.Type() == "identifier" || child.Type() == "property_identifier" {
							symbol.Name = code[child.StartByte():child.EndByte()]
							symbol.Type = "function"
							found = true
							break
						}
					}
				}
				if found {
					symbol.Signature = extractFunctionSignature(node, code, language)
				}
			} else if node.Type() == "class_declaration" {
				// Get class name
				for i := 0; i < int(node.ChildCount()); i++ {
					child := node.Child(i)
					if child != nil && child.Type() == "type_identifier" {
						symbol.Name = code[child.StartByte():child.EndByte()]
						symbol.Type = "class"
						found = true
						break
					}
				}
			}
		case "python":
			if node.Type() == "function_definition" {
				// Get function name
				for i := 0; i < int(node.ChildCount()); i++ {
					child := node.Child(i)
					if child != nil && child.Type() == "identifier" {
						symbol.Name = code[child.StartByte():child.EndByte()]
						symbol.Type = "function"
						found = true
						break
					}
				}
				if found {
					symbol.Signature = extractFunctionSignature(node, code, language)
				}
			} else if node.Type() == "class_definition" {
				// Get class name
				for i := 0; i < int(node.ChildCount()); i++ {
					child := node.Child(i)
					if child != nil && child.Type() == "identifier" {
						symbol.Name = code[child.StartByte():child.EndByte()]
						symbol.Type = "class"
						found = true
						break
					}
				}
			}
		}

		if found && symbol.Name != "" {
			startLine, startCol := getLineColumn(code, int(node.StartByte()))
			symbol.File = filePath
			symbol.Line = startLine
			symbol.Column = startCol
			symbol.Language = language
			symbol.Visibility = "public" // Default, can be enhanced

			table.Symbols[symbol.Name] = append(table.Symbols[symbol.Name], symbol)
		}

		return true
	})
}

// extractFunctionSignature extracts function signature (parameters, return type)
func extractFunctionSignature(node *sitter.Node, code string, language string) string {
	var signature strings.Builder

	// Find parameter list
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child != nil {
			if child.Type() == "parameter_list" || child.Type() == "formal_parameters" {
				params := code[child.StartByte():child.EndByte()]
				signature.WriteString(params)
				break
			}
		}
	}

	// For typed languages, try to find return type
	if language == "go" || language == "typescript" {
		// Look for return type annotation
		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(i)
			if child != nil {
				if child.Type() == "type_identifier" || child.Type() == "type_annotation" {
					returnType := code[child.StartByte():child.EndByte()]
					if !strings.Contains(signature.String(), returnType) {
						signature.WriteString(" -> ")
						signature.WriteString(returnType)
					}
					break
				}
			}
		}
	}

	result := signature.String()
	if result == "" {
		return "()"
	}
	return result
}

// extractImportsExports extracts import and export statements
func extractImportsExports(root *sitter.Node, code string, filePath string, language string, table *SymbolTable) {
	var imports []string
	var exports []string

	traverseAST(root, func(node *sitter.Node) bool {
		switch language {
		case "go":
			if node.Type() == "import_declaration" || node.Type() == "import_spec_list" {
				// Extract imported package names
				for i := 0; i < int(node.ChildCount()); i++ {
					child := node.Child(i)
					if child != nil {
						if child.Type() == "package_identifier" || child.Type() == "import_spec" {
							importName := code[child.StartByte():child.EndByte()]
							imports = append(imports, importName)
						}
					}
				}
			}
		case "javascript", "typescript":
			if node.Type() == "import_statement" {
				// Extract imported symbols
				for i := 0; i < int(node.ChildCount()); i++ {
					child := node.Child(i)
					if child != nil {
						if child.Type() == "import_specifier" || child.Type() == "namespace_import" {
							importName := code[child.StartByte():child.EndByte()]
							imports = append(imports, importName)
						}
					}
				}
			} else if node.Type() == "export_statement" || node.Type() == "export_declaration" {
				// Extract exported symbols
				for i := 0; i < int(node.ChildCount()); i++ {
					child := node.Child(i)
					if child != nil {
						if child.Type() == "export_specifier" || child.Type() == "identifier" {
							exportName := code[child.StartByte():child.EndByte()]
							exports = append(exports, exportName)
						}
					}
				}
			}
		case "python":
			if node.Type() == "import_statement" || node.Type() == "import_from_statement" {
				// Extract imported module/names
				for i := 0; i < int(node.ChildCount()); i++ {
					child := node.Child(i)
					if child != nil {
						if child.Type() == "dotted_name" || child.Type() == "identifier" {
							importName := code[child.StartByte():child.EndByte()]
							imports = append(imports, importName)
						}
					}
				}
			}
		}

		return true
	})

	if len(imports) > 0 {
		table.Imports[filePath] = imports
	}
	if len(exports) > 0 {
		table.Exports[filePath] = exports
	}
}

// resolveCrossFileReferences resolves imports to actual definitions
func resolveCrossFileReferences(table *SymbolTable) map[string]string {
	resolutions := make(map[string]string) // import -> definition file

	for filePath, imports := range table.Imports {
		for _, imp := range imports {
			// Try to find this symbol in the symbol table
			if symbols, exists := table.Symbols[imp]; exists {
				// Find the first definition that's not in the same file
				for _, sym := range symbols {
					if sym.File != filePath {
						resolutions[filePath+":"+imp] = sym.File
						break
					}
				}
			}
		}
	}

	return resolutions
}

// detectSignatureMismatches detects function signature mismatches across files (Phase 6F)
func detectSignatureMismatches(table *SymbolTable) []ASTFinding {
	findings := []ASTFinding{}

	// Check for duplicate function names with different signatures
	for funcName, definitions := range table.Symbols {
		if len(definitions) < 2 {
			continue // Need at least 2 definitions to compare
		}

		// Group by signature
		signatureGroups := make(map[string][]SymbolInfo)
		for _, def := range definitions {
			if def.Type == "function" {
				sig := def.Signature
				if sig == "" {
					sig = "()" // Default for functions without explicit signature
				}
				signatureGroups[sig] = append(signatureGroups[sig], def)
			}
		}

		// If we have multiple signature groups, we have mismatches
		if len(signatureGroups) > 1 {
			// Report all definitions with mismatched signatures
			for sig, defs := range signatureGroups {
				if len(defs) > 0 {
					// Find other signatures for comparison
					var otherSigs []string
					for otherSig := range signatureGroups {
						if otherSig != sig {
							otherSigs = append(otherSigs, otherSig)
						}
					}

					for _, def := range defs {
						// Include file path in message for cross-file findings
						message := fmt.Sprintf("%s: Function '%s' has signature mismatch: '%s' vs '%s'", def.File, funcName, sig, strings.Join(otherSigs, ", "))
						findings = append(findings, ASTFinding{
							Type:       "signature_mismatch",
							Severity:   "error",
							Line:       def.Line,
							Column:     def.Column,
							EndLine:    def.Line,
							EndColumn:  def.Column + len(def.Name),
							Message:    message,
							Code:       def.Name,
							Suggestion: fmt.Sprintf("Ensure all definitions of '%s' have the same signature, or rename conflicting functions", funcName),
						})
					}
				}
			}
		}
	}

	return findings
}

// detectImportExportMismatches detects import/export mismatches (Phase 6F)
func detectImportExportMismatches(table *SymbolTable) []ASTFinding {
	findings := []ASTFinding{}

	// Check for imports that don't have corresponding exports
	for filePath, imports := range table.Imports {
		for _, imp := range imports {
			// Check if this symbol is exported from any file
			foundExport := false
			for _, exports := range table.Exports {
				for _, exp := range exports {
					if exp == imp {
						foundExport = true
						break
					}
				}
				if foundExport {
					break
				}
			}

			// Also check if it's defined in symbol table (might be internal)
			if !foundExport {
				if _, exists := table.Symbols[imp]; !exists {
					// Imported but not exported and not defined
					findings = append(findings, ASTFinding{
						Type:       "import_mismatch",
						Severity:   "warning",
						Line:       1, // Import statements are usually at top
						Column:     0,
						EndLine:    1,
						EndColumn:  0,
						Message:    fmt.Sprintf("Import '%s' in %s has no corresponding export or definition", imp, filePath),
						Code:       imp,
						Suggestion: fmt.Sprintf("Check if '%s' is exported from the source file or remove the import", imp),
					})
				}
			}
		}
	}

	return findings
}

// analyzeCrossFile performs cross-file analysis on multiple files (Phase 6F)
func analyzeCrossFile(files []struct {
	Path     string
	Code     string
	Language string
}) ([]ASTFinding, AnalysisStats, error) {
	analysisStart := time.Now()

	// Build symbol table from all files
	table := buildSymbolTable(files)

	// Resolve cross-file references
	resolutions := resolveCrossFileReferences(table)
	_ = resolutions // May be used for future enhancements

	// Detect signature mismatches
	signatureFindings := detectSignatureMismatches(table)

	// Detect import/export mismatches
	importFindings := detectImportExportMismatches(table)

	// Combine all findings
	allFindings := append(signatureFindings, importFindings...)

	analysisTime := time.Since(analysisStart).Milliseconds()

	stats := AnalysisStats{
		ParseTime:    0, // Cross-file doesn't track individual parse times
		AnalysisTime: analysisTime,
		NodesVisited: len(table.Symbols), // Approximate
	}

	return allFindings, stats, nil
}
