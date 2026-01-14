// Phase 14E: Task Verification Engine
// Multi-factor verification for task completion

package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	sitter "github.com/smacker/go-tree-sitter"
)

// VerificationFactors represents the weights for different verification factors
type VerificationFactors struct {
	CodeExistence float64 // 0.4
	CodeUsage     float64 // 0.3
	TestCoverage  float64 // 0.2
	Integration   float64 // 0.1
}

// DefaultVerificationFactors returns default weights
func DefaultVerificationFactors() VerificationFactors {
	return VerificationFactors{
		CodeExistence: 0.4,
		CodeUsage:     0.3,
		TestCoverage:  0.2,
		Integration:   0.1,
	}
}

// VerifyTask verifies a task using multi-factor verification
func VerifyTask(ctx context.Context, taskID string, codebasePath string, force bool) (*VerifyTaskResponse, error) {
	// Get task
	task, err := GetTask(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	// Check cache first (unless force)
	if !force {
		if cachedVerification, found := GetCachedVerification(taskID); found {
			return cachedVerification, nil
		}
	}

	// Run multi-factor verification
	factors := DefaultVerificationFactors()
	verifications := []TaskVerification{}

	// 1. Code Existence Verification
	codeExistenceVerification, err := verifyCodeExistence(ctx, task, codebasePath)
	if err != nil {
		LogError(ctx, "Code existence verification failed: %v", err)
		codeExistenceVerification = TaskVerification{
			TaskID:           taskID,
			VerificationType: "code_existence",
			Status:           "failed",
			Confidence:       0.0,
		}
	}
	verifications = append(verifications, codeExistenceVerification)

	// 2. Code Usage Verification
	codeUsageVerification, err := verifyCodeUsage(ctx, task, codebasePath)
	if err != nil {
		LogError(ctx, "Code usage verification failed: %v", err)
		codeUsageVerification = TaskVerification{
			TaskID:           taskID,
			VerificationType: "code_usage",
			Status:           "failed",
			Confidence:       0.0,
		}
	}
	verifications = append(verifications, codeUsageVerification)

	// 3. Test Coverage Verification
	testCoverageVerification, err := verifyTestCoverage(ctx, task, codebasePath)
	if err != nil {
		LogError(ctx, "Test coverage verification failed: %v", err)
		testCoverageVerification = TaskVerification{
			TaskID:           taskID,
			VerificationType: "test_coverage",
			Status:           "failed",
			Confidence:       0.0,
		}
	}
	verifications = append(verifications, testCoverageVerification)

	// 4. Integration Verification
	integrationVerification, err := verifyIntegration(ctx, task, codebasePath)
	if err != nil {
		LogError(ctx, "Integration verification failed: %v", err)
		integrationVerification = TaskVerification{
			TaskID:           taskID,
			VerificationType: "integration",
			Status:           "failed",
			Confidence:       0.0,
		}
	}
	verifications = append(verifications, integrationVerification)

	// Calculate overall confidence
	overallConfidence := calculateOverallConfidence(verifications, factors)

	// Store verifications
	for _, verification := range verifications {
		if err := storeVerification(ctx, verification); err != nil {
			LogError(ctx, "Failed to store verification: %v", err)
		}
	}

	// Update task verification status
	now := time.Now()
	query := `
		UPDATE tasks 
		SET verification_confidence = $1, verified_at = $2, updated_at = $3
		WHERE id = $4
	`
	_, err = execWithTimeout(ctx, query, overallConfidence, now, now, taskID)
	if err != nil {
		LogError(ctx, "Failed to update task verification: %v", err)
	}

	// Determine status based on confidence
	status := determineTaskStatus(overallConfidence, verifications)

	// Build evidence map
	evidence := buildEvidenceMap(verifications)

	response := &VerifyTaskResponse{
		TaskID:            taskID,
		OverallConfidence: overallConfidence,
		Verifications:     verifications,
		Status:            status,
		Evidence:          evidence,
	}

	// Cache the verification result
	SetCachedVerification(taskID, response)

	return response, nil
}

// verifyCodeExistence verifies if code mentioned in task exists
func verifyCodeExistence(ctx context.Context, task *Task, codebasePath string) (TaskVerification, error) {
	verification := TaskVerification{
		TaskID:           task.ID,
		VerificationType: "code_existence",
		Status:           "pending",
		Confidence:       0.0,
		Evidence:         make(map[string]interface{}),
	}

	// Extract keywords from task title and description
	keywords := extractKeywords(task.Title + " " + task.Description)
	if len(keywords) == 0 {
		verification.Status = "failed"
		verification.Confidence = 0.0
		return verification, nil
	}

	// Search for keywords in codebase
	foundFiles := []string{}
	foundFunctions := []string{}
	totalMatches := 0

	err := filepath.Walk(codebasePath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		// Skip test files for code existence check
		if strings.Contains(path, "_test.") || strings.Contains(path, ".test.") {
			return nil
		}

		// Check file extension
		ext := filepath.Ext(path)
		if ext != ".go" && ext != ".js" && ext != ".ts" && ext != ".py" && ext != ".java" {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		contentStr := strings.ToLower(string(content))
		matches := 0
		for _, keyword := range keywords {
			if strings.Contains(contentStr, strings.ToLower(keyword)) {
				matches++
			}
		}

		if matches > 0 {
			relativePath, _ := filepath.Rel(codebasePath, path)
			foundFiles = append(foundFiles, relativePath)
			totalMatches += matches
		}

		return nil
	})

	if err != nil {
		return verification, fmt.Errorf("failed to scan codebase: %w", err)
	}

	// Calculate confidence based on matches
	confidence := 0.0
	if len(foundFiles) > 0 {
		// More files = higher confidence
		fileConfidence := float64(len(foundFiles)) / 10.0
		if fileConfidence > 1.0 {
			fileConfidence = 1.0
		}

		// More matches = higher confidence
		matchConfidence := float64(totalMatches) / float64(len(keywords)*5)
		if matchConfidence > 1.0 {
			matchConfidence = 1.0
		}

		confidence = (fileConfidence + matchConfidence) / 2.0
	}

	verification.Confidence = confidence
	if confidence > 0.5 {
		verification.Status = "verified"
	} else {
		verification.Status = "failed"
	}

	verification.Evidence["files"] = foundFiles
	verification.Evidence["functions"] = foundFunctions
	verification.Evidence["matches"] = totalMatches

	return verification, nil
}

// verifyCodeUsage verifies if code is actually used (called/referenced)
func verifyCodeUsage(ctx context.Context, task *Task, codebasePath string) (TaskVerification, error) {
	verification := TaskVerification{
		TaskID:           task.ID,
		VerificationType: "code_usage",
		Status:           "pending",
		Confidence:       0.0,
		Evidence:         make(map[string]interface{}),
	}

	// Get code existence first
	codeExistenceVerification, err := verifyCodeExistence(ctx, task, codebasePath)
	if err != nil {
		return verification, err
	}

	files, _ := codeExistenceVerification.Evidence["files"].([]string)
	if len(files) == 0 {
		verification.Confidence = 0.0
		verification.Status = "failed"
		verification.Evidence["files"] = files
		verification.Evidence["call_sites"] = []string{}
		return verification, nil
	}

	// Extract keywords from task
	keywords := extractKeywords(task.Title + " " + task.Description)
	if len(keywords) == 0 {
		verification.Confidence = 0.0
		verification.Status = "failed"
		verification.Evidence["files"] = files
		verification.Evidence["call_sites"] = []string{}
		return verification, nil
	}

	// Find call sites using AST-based analysis
	callSites, err := extractFunctionCallsAST(codebasePath, keywords, files)
	if err != nil {
		LogError(ctx, "AST analysis failed, using fallback: %v", err)
		// Fallback to simple heuristic
		callSites = []string{}
	}

	// Calculate confidence based on call sites and file count
	confidence := 0.0
	if len(callSites) > 0 {
		// High confidence if call sites found
		confidence = 0.9
		verification.Status = "verified"
	} else if len(files) > 1 {
		// Code appears in multiple files, likely used
		confidence = 0.7
		verification.Status = "verified"
	} else if len(files) == 1 {
		// Code in one file, moderate confidence
		confidence = 0.5
		verification.Status = "verified"
	} else {
		confidence = 0.0
		verification.Status = "failed"
	}

	verification.Confidence = confidence
	verification.Evidence["files"] = files
	verification.Evidence["call_sites"] = callSites

	return verification, nil
}

// extractFunctionCallsAST extracts function calls and references matching keywords using AST analysis
func extractFunctionCallsAST(codebasePath string, keywords []string, sourceFiles []string) ([]string, error) {
	var callSites []string
	var astFailed bool

	// Build a map of keywords for faster lookup
	keywordMap := make(map[string]bool)
	for _, kw := range keywords {
		keywordMap[strings.ToLower(kw)] = true
	}

	// Try AST-based analysis first
	for _, file := range sourceFiles {
		fullPath := filepath.Join(codebasePath, file)

		// Skip test files for call site detection
		if strings.Contains(fullPath, "_test.") || strings.Contains(fullPath, ".test.") {
			continue
		}

		content, err := os.ReadFile(fullPath)
		if err != nil {
			continue
		}

		// Determine language from file extension
		lang := detectLanguageFromFile(file)
		if lang == "" {
			continue
		}

		// Get parser for language
		parser, err := getParser(lang)
		if err != nil {
			astFailed = true
			continue
		}

		// Parse code into AST
		ctx := context.Background()
		tree, err := parser.ParseCtx(ctx, nil, content)
		if err != nil || tree == nil {
			astFailed = true
			continue
		}
		defer tree.Close()

		rootNode := tree.RootNode()
		if rootNode == nil {
			astFailed = true
			continue
		}

		// Extract function calls from AST
		fileCallSites := extractCallSitesFromAST(rootNode, string(content), lang, keywordMap, file)
		callSites = append(callSites, fileCallSites...)
	}

	// Fallback to regex if AST failed for all files
	if len(callSites) == 0 && astFailed {
		return extractFunctionCallsRegex(codebasePath, keywords, sourceFiles)
	}

	return callSites, nil
}

// detectLanguageFromFile detects language from file extension
func detectLanguageFromFile(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".go":
		return "go"
	case ".js", ".jsx":
		return "javascript"
	case ".ts", ".tsx":
		return "typescript"
	case ".py":
		return "python"
	case ".java":
		return "java"
	default:
		return ""
	}
}

// extractCallSitesFromAST extracts function call sites from AST matching keywords
func extractCallSitesFromAST(root *sitter.Node, code string, language string, keywordMap map[string]bool, filePath string) []string {
	var callSites []string

	traverseAST(root, func(node *sitter.Node) bool {
		var funcName string
		var isCall bool

		switch language {
		case "go":
			if node.Type() == "call_expression" {
				// Get function name from call_expression
				for i := 0; i < int(node.ChildCount()); i++ {
					child := node.Child(i)
					if child != nil && (child.Type() == "identifier" || child.Type() == "field_expression" || child.Type() == "selector_expression") {
						// Extract identifier from expression
						funcName = extractIdentifierFromNode(child, code)
						isCall = true
						break
					}
				}
			}
		case "javascript", "typescript":
			if node.Type() == "call_expression" {
				// Get function name from call_expression
				for i := 0; i < int(node.ChildCount()); i++ {
					child := node.Child(i)
					if child != nil && (child.Type() == "identifier" || child.Type() == "member_expression" || child.Type() == "property_identifier") {
						funcName = extractIdentifierFromNode(child, code)
						isCall = true
						break
					}
				}
			}
		case "python":
			if node.Type() == "call" {
				// Get function name from call
				for i := 0; i < int(node.ChildCount()); i++ {
					child := node.Child(i)
					if child != nil && (child.Type() == "identifier" || child.Type() == "attribute") {
						funcName = extractIdentifierFromNode(child, code)
						isCall = true
						break
					}
				}
			}
		}

		if isCall && funcName != "" {
			funcNameLower := strings.ToLower(funcName)
			// Check if function name matches any keyword
			for keyword := range keywordMap {
				if strings.Contains(funcNameLower, keyword) || strings.Contains(keyword, funcNameLower) {
					line, _ := getLineColumn(code, int(node.StartByte()))
					callSite := fmt.Sprintf("%s:%d", filePath, line+1)
					callSites = append(callSites, callSite)
					break
				}
			}
		}

		return true
	})

	return callSites
}

// extractIdentifierFromNode extracts identifier name from AST node
func extractIdentifierFromNode(node *sitter.Node, code string) string {
	if node == nil {
		return ""
	}

	// If node is identifier, return it directly
	if node.Type() == "identifier" || node.Type() == "property_identifier" || node.Type() == "field_identifier" {
		return code[node.StartByte():node.EndByte()]
	}

	// If node is member/selector expression, get the last identifier
	if node.Type() == "member_expression" || node.Type() == "selector_expression" || node.Type() == "field_expression" {
		// Get the property/field part (last child)
		childCount := int(node.ChildCount())
		if childCount > 0 {
			lastChild := node.Child(childCount - 1)
			if lastChild != nil {
				return extractIdentifierFromNode(lastChild, code)
			}
		}
	}

	// Try to find identifier in children
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child != nil {
			if child.Type() == "identifier" || child.Type() == "property_identifier" || child.Type() == "field_identifier" {
				return code[child.StartByte():child.EndByte()]
			}
		}
	}

	return ""
}

// extractFunctionCallsRegex is fallback regex-based implementation
func extractFunctionCallsRegex(codebasePath string, keywords []string, sourceFiles []string) ([]string, error) {
	var callSites []string

	// Build a map of keywords for faster lookup
	keywordMap := make(map[string]bool)
	for _, kw := range keywords {
		keywordMap[strings.ToLower(kw)] = true
	}

	// Patterns for function calls in different languages
	callPatterns := map[string]*regexp.Regexp{
		".go":   regexp.MustCompile(`\b([A-Za-z_][A-Za-z0-9_]*)\s*\(`),
		".js":   regexp.MustCompile(`\b([A-Za-z_][A-Za-z0-9_]*)\s*\(`),
		".ts":   regexp.MustCompile(`\b([A-Za-z_][A-Za-z0-9_]*)\s*\(`),
		".py":   regexp.MustCompile(`\b([A-Za-z_][A-Za-z0-9_]*)\s*\(`),
		".java": regexp.MustCompile(`\b([A-Za-z_][A-Za-z0-9_]*)\s*\(`),
	}

	// Process only files in sourceFiles list
	for _, file := range sourceFiles {
		fullPath := filepath.Join(codebasePath, file)

		// Skip test files
		if strings.Contains(fullPath, "_test.") || strings.Contains(fullPath, ".test.") {
			continue
		}

		ext := filepath.Ext(fullPath)
		pattern, ok := callPatterns[ext]
		if !ok {
			continue
		}

		content, err := os.ReadFile(fullPath)
		if err != nil {
			continue
		}

		contentStr := string(content)
		lines := strings.Split(contentStr, "\n")

		// Find function calls matching keywords
		for lineNum, line := range lines {
			matches := pattern.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				if len(match) > 1 {
					funcName := strings.ToLower(match[1])
					// Check if function name matches any keyword
					for keyword := range keywordMap {
						if strings.Contains(funcName, keyword) || strings.Contains(keyword, funcName) {
							callSite := fmt.Sprintf("%s:%d", file, lineNum+1)
							callSites = append(callSites, callSite)
							break
						}
					}
				}
			}
		}
	}

	return callSites, nil
}

// verifyTestCoverage verifies if tests exist for the task
func verifyTestCoverage(ctx context.Context, task *Task, codebasePath string) (TaskVerification, error) {
	verification := TaskVerification{
		TaskID:           task.ID,
		VerificationType: "test_coverage",
		Status:           "pending",
		Confidence:       0.0,
		Evidence:         make(map[string]interface{}),
	}

	// Extract keywords from task
	keywords := extractKeywords(task.Title + " " + task.Description)
	if len(keywords) == 0 {
		verification.Status = "failed"
		return verification, nil
	}

	// Search for test files
	testFiles := []string{}
	err := filepath.Walk(codebasePath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		// Check if it's a test file
		isTestFile := strings.Contains(path, "_test.") ||
			strings.Contains(path, ".test.") ||
			strings.Contains(path, ".spec.") ||
			strings.Contains(path, "__tests__")

		if !isTestFile {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		contentStr := strings.ToLower(string(content))
		matches := 0
		for _, keyword := range keywords {
			if strings.Contains(contentStr, strings.ToLower(keyword)) {
				matches++
			}
		}

		if matches > 0 {
			relativePath, _ := filepath.Rel(codebasePath, path)
			testFiles = append(testFiles, relativePath)
		}

		return nil
	})

	if err != nil {
		return verification, fmt.Errorf("failed to scan test files: %w", err)
	}

	// Calculate confidence
	confidence := 0.0
	if len(testFiles) > 0 {
		confidence = 0.9 // High confidence if test files found
		verification.Status = "verified"
	} else {
		confidence = 0.0
		verification.Status = "failed"
	}

	verification.Confidence = confidence
	verification.Evidence["test_files"] = testFiles

	// Calculate test coverage
	coverage, err := calculateTestCoverage(testFiles, codebasePath, keywords)
	if err != nil {
		LogError(ctx, "Failed to calculate test coverage: %v", err)
		coverage = 0.0
	}
	verification.Evidence["coverage"] = coverage

	return verification, nil
}

// calculateTestCoverage calculates test coverage based on test files found
func calculateTestCoverage(testFiles []string, codebasePath string, keywords []string) (float64, error) {
	if len(testFiles) == 0 {
		return 0.0, nil
	}

	// Try to use real test coverage tools first
	// Detect language from test files
	language := detectLanguageFromTestFiles(testFiles)

	// Attempt to get real coverage based on language
	var coverage float64
	var err error

	switch language {
	case "go":
		coverage, err = getGoTestCoverage(codebasePath)
		if err == nil {
			return coverage, nil
		}
	case "javascript", "typescript":
		coverage, err = getJSTestCoverage(codebasePath)
		if err == nil {
			return coverage, nil
		}
	case "python":
		coverage, err = getPythonTestCoverage(codebasePath)
		if err == nil {
			return coverage, nil
		}
	}

	// Fallback to heuristic if test tools unavailable
	return calculateTestCoverageHeuristic(testFiles, codebasePath, keywords)
}

// detectLanguageFromTestFiles detects the primary language from test files
func detectLanguageFromTestFiles(testFiles []string) string {
	for _, testFile := range testFiles {
		ext := filepath.Ext(testFile)
		switch ext {
		case ".go":
			return "go"
		case ".js", ".jsx":
			return "javascript"
		case ".ts", ".tsx":
			return "typescript"
		case ".py":
			return "python"
		}
	}
	return "unknown"
}

// getGoTestCoverage runs go test -coverprofile and parses coverage
func getGoTestCoverage(codebasePath string) (float64, error) {
	// Run go test with coverage
	cmd := exec.Command("go", "test", "-coverprofile=coverage.out", "./...")
	cmd.Dir = codebasePath
	err := cmd.Run()
	if err != nil {
		return 0.0, fmt.Errorf("go test failed: %w", err)
	}

	// Parse coverage.out file
	coverageFile := filepath.Join(codebasePath, "coverage.out")
	data, err := os.ReadFile(coverageFile)
	if err != nil {
		return 0.0, fmt.Errorf("failed to read coverage file: %w", err)
	}

	// Parse coverage percentage from coverage.out
	// Format: mode: set
	// Format: file.go:line.column,line.column count
	lines := strings.Split(string(data), "\n")
	if len(lines) == 0 {
		return 0.0, fmt.Errorf("empty coverage file")
	}

	// Count total statements and covered statements
	totalStmts := 0
	coveredStmts := 0

	for _, line := range lines {
		if strings.HasPrefix(line, "mode:") {
			continue
		}
		if line == "" {
			continue
		}

		// Parse line: file.go:start.column,end.column count
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		count, err := strconv.Atoi(parts[len(parts)-1])
		if err != nil {
			continue
		}

		// Extract statement range (simplified: assume 1 statement per range)
		totalStmts++
		if count > 0 {
			coveredStmts++
		}
	}

	if totalStmts == 0 {
		return 0.0, nil
	}

	return float64(coveredStmts) / float64(totalStmts), nil
}

// getJSTestCoverage attempts to get coverage from jest/nyc
func getJSTestCoverage(codebasePath string) (float64, error) {
	// Try jest first
	cmd := exec.Command("npm", "test", "--", "--coverage", "--coverageReporters=text-summary")
	cmd.Dir = codebasePath
	output, err := cmd.CombinedOutput()
	if err == nil {
		// Parse coverage from output
		// Look for "All files" line with percentage
		outputStr := string(output)
		lines := strings.Split(outputStr, "\n")
		for _, line := range lines {
			if strings.Contains(line, "All files") {
				// Extract percentage: "All files | 85.5 | ..."
				parts := strings.Split(line, "|")
				if len(parts) >= 2 {
					percentStr := strings.TrimSpace(parts[1])
					percent, err := strconv.ParseFloat(percentStr, 64)
					if err == nil {
						return percent / 100.0, nil
					}
				}
			}
		}
	}

	// Try nyc
	cmd = exec.Command("npx", "nyc", "npm", "test")
	cmd.Dir = codebasePath
	output, err = cmd.CombinedOutput()
	if err == nil {
		// Parse nyc output
		outputStr := string(output)
		lines := strings.Split(outputStr, "\n")
		for _, line := range lines {
			if strings.Contains(line, "All files") {
				parts := strings.Split(line, "|")
				if len(parts) >= 2 {
					percentStr := strings.TrimSpace(parts[1])
					percent, err := strconv.ParseFloat(percentStr, 64)
					if err == nil {
						return percent / 100.0, nil
					}
				}
			}
		}
	}

	return 0.0, fmt.Errorf("no coverage tool available")
}

// getPythonTestCoverage attempts to get coverage from coverage.py
func getPythonTestCoverage(codebasePath string) (float64, error) {
	// Try coverage.py
	cmd := exec.Command("coverage", "run", "-m", "pytest")
	cmd.Dir = codebasePath
	err := cmd.Run()
	if err != nil {
		return 0.0, fmt.Errorf("coverage run failed: %w", err)
	}

	// Get coverage report
	cmd = exec.Command("coverage", "report")
	cmd.Dir = codebasePath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0.0, fmt.Errorf("coverage report failed: %w", err)
	}

	// Parse coverage percentage from output
	// Format: "TOTAL ... XX%"
	outputStr := string(output)
	lines := strings.Split(outputStr, "\n")
	for _, line := range lines {
		if strings.Contains(line, "TOTAL") {
			// Extract percentage
			re := regexp.MustCompile(`(\d+)%`)
			matches := re.FindStringSubmatch(line)
			if len(matches) >= 2 {
				percent, err := strconv.ParseFloat(matches[1], 64)
				if err == nil {
					return percent / 100.0, nil
				}
			}
		}
	}

	return 0.0, fmt.Errorf("failed to parse coverage")
}

// calculateTestCoverageHeuristic uses heuristic approach when test tools unavailable
func calculateTestCoverageHeuristic(testFiles []string, codebasePath string, keywords []string) (float64, error) {
	// Map test files to source files
	sourceFileMap := make(map[string]bool)

	// For each test file, try to find corresponding source files
	for _, testFile := range testFiles {
		// Remove test file extensions and paths
		baseName := testFile
		baseName = strings.TrimSuffix(baseName, "_test.go")
		baseName = strings.TrimSuffix(baseName, ".test.js")
		baseName = strings.TrimSuffix(baseName, ".test.ts")
		baseName = strings.TrimSuffix(baseName, ".spec.js")
		baseName = strings.TrimSuffix(baseName, ".spec.ts")
		baseName = strings.TrimSuffix(baseName, "_test.py")

		// Try to find corresponding source file
		dir := filepath.Dir(testFile)
		ext := filepath.Ext(testFile)

		// Map test extensions to source extensions
		extMap := map[string]string{
			".go": ".go",
			".js": ".js",
			".ts": ".ts",
			".py": ".py",
		}

		sourceExt := extMap[ext]
		if sourceExt == "" {
			sourceExt = ext
		}

		// Look for source file with same base name
		baseNameOnly := filepath.Base(baseName)
		sourceFile := filepath.Join(dir, baseNameOnly+sourceExt)
		fullPath := filepath.Join(codebasePath, sourceFile)

		if _, err := os.Stat(fullPath); err == nil {
			sourceFileMap[sourceFile] = true
		}
	}

	// Calculate coverage based on:
	// 1. Number of test files found
	// 2. Whether corresponding source files exist
	// 3. Number of keywords covered in tests

	if len(sourceFileMap) > 0 {
		// If we found source files with tests, coverage is high
		if len(testFiles) > 1 {
			return 0.9, nil // Multiple test files = high coverage
		}
		return 0.8, nil // Single test file = good coverage
	}

	// If no source files found but test files exist, moderate coverage
	if len(testFiles) > 0 {
		return 0.6, nil
	}

	return 0.0, nil
}

// verifyIntegration verifies if integration requirements are met
func verifyIntegration(ctx context.Context, task *Task, codebasePath string) (TaskVerification, error) {
	verification := TaskVerification{
		TaskID:           task.ID,
		VerificationType: "integration",
		Status:           "pending",
		Confidence:       0.0,
		Evidence:         make(map[string]interface{}),
	}

	// Check for integration keywords
	integrationKeywords := []string{"api", "integration", "service", "external", "third-party", "sdk"}
	taskText := strings.ToLower(task.Title + " " + task.Description)

	hasIntegrationKeyword := false
	for _, keyword := range integrationKeywords {
		if strings.Contains(taskText, keyword) {
			hasIntegrationKeyword = true
			break
		}
	}

	if !hasIntegrationKeyword {
		// No integration requirement, skip
		verification.Confidence = 1.0
		verification.Status = "verified"
		verification.Evidence["skipped"] = true
		return verification, nil
	}

	// Extract keywords from task
	keywords := extractKeywords(task.Title + " " + task.Description)

	// Check for actual integration code
	integrationFiles, err := findIntegrationCode(ctx, codebasePath, keywords)
	if err != nil {
		LogError(ctx, "Failed to find integration code: %v", err)
		// Fallback to moderate confidence
		verification.Confidence = 0.5
		verification.Status = "pending"
		verification.Evidence["integration_required"] = true
		verification.Evidence["integration_files"] = []string{}
		return verification, nil
	}

	// Calculate confidence based on found integration patterns
	if len(integrationFiles) > 0 {
		verification.Confidence = 0.8
		verification.Status = "verified"
	} else {
		verification.Confidence = 0.3
		verification.Status = "pending"
	}

	verification.Evidence["integration_required"] = true
	verification.Evidence["integration_files"] = integrationFiles

	return verification, nil
}

// IntegrationEvidence contains detailed evidence of integration code found
type IntegrationEvidence struct {
	Files           []string `json:"files"`
	Functions       []string `json:"functions"`
	IntegrationType string   `json:"integration_type"` // "REST", "GraphQL", "gRPC", "WebSocket", "Middleware", "Event", "Unknown"
	ImportPaths     []string `json:"import_paths"`
	ConfigFiles     []string `json:"config_files"`
	ASTMatched      bool     `json:"ast_matched"`   // Whether AST found matches
	RegexMatched    bool     `json:"regex_matched"` // Whether regex found matches
}

// findIntegrationCode searches for actual integration code patterns using hybrid AST + regex approach
func findIntegrationCode(ctx context.Context, codebasePath string, keywords []string) ([]string, error) {
	var integrationFiles []string

	// Enhanced patterns for integration code
	integrationPatterns := []*regexp.Regexp{
		// HTTP clients
		regexp.MustCompile(`(?i)(http\.Client|resty|axios|fetch|requests\.|urllib|httpx)`),
		// API endpoints - fixed regex patterns with proper escaping
		regexp.MustCompile(`(?i)(api\.|endpoint|/api/|/v\d+/|\.(post|get|put|delete|patch)\()`),
		// Service clients
		regexp.MustCompile(`(?i)(client\.|service\.|sdk\.|(Client|Service)\(`),
		// External libraries
		regexp.MustCompile(`(?i)(import.*http|from.*requests|require.*axios|import.*fetch)`),
		// GraphQL patterns
		regexp.MustCompile(`(?i)(graphql|gql|apollo|relay|query|mutation|subscription|gql\()`),
		// gRPC patterns
		regexp.MustCompile(`(?i)(grpc|protobuf|\.proto|rpc|service.*pb|pb\.)`),
		// WebSocket patterns
		regexp.MustCompile(`(?i)(websocket|ws://|wss://|socket\.io|ws\.|WebSocket)`),
		// Middleware patterns
		regexp.MustCompile(`(?i)(middleware|use\(|app\.use|router\.use|express\.|gin\.|mux\.)`),
		// Event handler patterns
		regexp.MustCompile(`(?i)(on\(|addEventListener|emit\(|publish\(|subscribe|event\.|EventEmitter)`),
	}

	err := filepath.Walk(codebasePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Log error but continue processing
			LogWarn(ctx, "Error accessing path %s: %v", path, err)
			return nil
		}

		if info.IsDir() {
			return nil
		}

		// Skip test files
		if strings.Contains(path, "_test.") || strings.Contains(path, ".test.") {
			return nil
		}

		// Check file extension
		ext := filepath.Ext(path)
		supportedExts := map[string]bool{
			".go": true, ".js": true, ".ts": true, ".py": true,
			".java": true, ".jsx": true, ".tsx": true,
		}
		if !supportedExts[ext] {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			LogWarn(ctx, "Failed to read file %s: %v", path, err)
			return nil // Continue with next file
		}

		contentStr := string(content)

		// Detect language from file extension
		language := detectLanguageFromExtension(ext)

		// Initialize evidence for this file
		evidence := &IntegrationEvidence{
			Files:           []string{},
			Functions:       []string{},
			IntegrationType: "Unknown",
			ImportPaths:     []string{},
			ConfigFiles:     []string{},
			ASTMatched:      false,
			RegexMatched:    false,
		}

		// Try AST-based analysis first for supported languages
		astMatched := false
		if language != "unknown" {
			funcs, imports, integrationTypes, err := findSymbolsWithAST(contentStr, language, keywords, path)
			if err != nil {
				// Log AST failure but continue with regex fallback
				LogDebug(ctx, "AST analysis failed for %s (language: %s): %v, falling back to regex", path, language, err)
			} else {
				// AST found matches
				if len(funcs) > 0 || len(imports) > 0 {
					astMatched = true
					evidence.ASTMatched = true
					evidence.Functions = funcs
					evidence.ImportPaths = imports
					if len(integrationTypes) > 0 {
						evidence.IntegrationType = strings.Join(integrationTypes, ", ")
					}
				}
			}
		}

		// Fallback to regex if AST didn't find anything or language not supported
		regexMatched := false
		detectedIntegrationType := "Unknown"

		if !astMatched {
			for _, pattern := range integrationPatterns {
				if pattern.MatchString(contentStr) {
					regexMatched = true
					evidence.RegexMatched = true

					// Detect integration type from pattern
					patternStr := pattern.String()
					if strings.Contains(patternStr, "graphql") || strings.Contains(patternStr, "gql") {
						detectedIntegrationType = "GraphQL"
					} else if strings.Contains(patternStr, "grpc") {
						detectedIntegrationType = "gRPC"
					} else if strings.Contains(patternStr, "websocket") || strings.Contains(patternStr, "socket") {
						detectedIntegrationType = "WebSocket"
					} else if strings.Contains(patternStr, "middleware") {
						detectedIntegrationType = "Middleware"
					} else if strings.Contains(patternStr, "event") || strings.Contains(patternStr, "emit") {
						detectedIntegrationType = "Event"
					} else if strings.Contains(patternStr, "http") || strings.Contains(patternStr, "api") {
						detectedIntegrationType = "REST"
					}
					break
				}
			}

			if detectedIntegrationType != "Unknown" && evidence.IntegrationType == "Unknown" {
				evidence.IntegrationType = detectedIntegrationType
			}
		}

		// Also check if keywords appear in file
		hasKeywords := false
		contentLower := strings.ToLower(contentStr)
		for _, keyword := range keywords {
			if strings.Contains(contentLower, strings.ToLower(keyword)) {
				hasKeywords = true
				break
			}
		}

		// File matches if (AST found matches OR regex found matches) AND keywords present
		if (astMatched || regexMatched) && hasKeywords {
			relativePath, err := filepath.Rel(codebasePath, path)
			if err != nil {
				relativePath = path // Fallback to full path
			}
			integrationFiles = append(integrationFiles, relativePath)
			evidence.Files = []string{relativePath}

			// Log successful match with details
			LogDebug(ctx, "Integration code found in %s: AST=%v, Regex=%v, Type=%s",
				relativePath, astMatched, regexMatched, evidence.IntegrationType)
		}

		return nil
	})

	if err != nil {
		LogError(ctx, "Error walking codebase: %v", err)
		return integrationFiles, fmt.Errorf("failed to scan codebase: %w", err)
	}

	return integrationFiles, nil
}

// detectLanguageFromExtension detects language from file extension
func detectLanguageFromExtension(ext string) string {
	langMap := map[string]string{
		".go":   "go",
		".js":   "javascript",
		".jsx":  "javascript",
		".ts":   "typescript",
		".tsx":  "typescript",
		".py":   "python",
		".java": "java",
	}
	if lang, ok := langMap[ext]; ok {
		return lang
	}
	return "unknown"
}

// findSymbolsWithAST uses AST to find function/class definitions and imports matching keywords
func findSymbolsWithAST(code string, language string, keywords []string, filePath string) ([]string, []string, []string, error) {
	var matchedFunctions []string
	var matchedImports []string
	var integrationTypes []string

	// Build keyword map for fast lookup
	keywordMap := make(map[string]bool)
	for _, kw := range keywords {
		keywordMap[strings.ToLower(kw)] = true
	}

	// Get parser for language
	parser, err := getParser(language)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("unsupported language for AST: %s", language)
	}

	// Parse code
	ctx := context.Background()
	tree, err := parser.ParseCtx(ctx, nil, []byte(code))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("AST parse error: %w", err)
	}

	if tree == nil {
		return nil, nil, nil, fmt.Errorf("AST parse returned nil tree")
	}
	defer tree.Close()

	rootNode := tree.RootNode()
	if rootNode == nil {
		return nil, nil, nil, fmt.Errorf("AST root node is nil")
	}

	// Extract symbols and imports using AST traversal
	traverseAST(rootNode, func(node *sitter.Node) bool {
		// Check for function definitions matching keywords
		var funcName string
		var isFunction bool

		switch language {
		case "go":
			if node.Type() == "function_declaration" || node.Type() == "method_declaration" {
				for i := 0; i < int(node.ChildCount()); i++ {
					child := node.Child(i)
					if child != nil && (child.Type() == "identifier" || child.Type() == "field_identifier") {
						funcName = code[child.StartByte():child.EndByte()]
						isFunction = true
						break
					}
				}
			}
		case "javascript", "typescript":
			if node.Type() == "function_declaration" || node.Type() == "function" {
				for i := 0; i < int(node.ChildCount()); i++ {
					child := node.Child(i)
					if child != nil && (child.Type() == "identifier" || child.Type() == "property_identifier") {
						funcName = code[child.StartByte():child.EndByte()]
						isFunction = true
						break
					}
				}
			}
		case "python":
			if node.Type() == "function_definition" {
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

		// Check if function name matches keywords
		if isFunction && funcName != "" {
			funcNameLower := strings.ToLower(funcName)
			for keyword := range keywordMap {
				if strings.Contains(funcNameLower, keyword) || strings.Contains(keyword, funcNameLower) {
					matchedFunctions = append(matchedFunctions, funcName)
					break
				}
			}
		}

		// Check for imports matching integration patterns
		switch language {
		case "go":
			if node.Type() == "import_declaration" || node.Type() == "import_spec_list" {
				importText := strings.ToLower(code[node.StartByte():node.EndByte()])
				// Check for integration-related imports
				if strings.Contains(importText, "http") || strings.Contains(importText, "grpc") ||
					strings.Contains(importText, "graphql") || strings.Contains(importText, "websocket") {
					matchedImports = append(matchedImports, importText)
					// Detect integration type
					if strings.Contains(importText, "graphql") {
						integrationTypes = appendIfNotExists(integrationTypes, "GraphQL")
					}
					if strings.Contains(importText, "grpc") {
						integrationTypes = appendIfNotExists(integrationTypes, "gRPC")
					}
					if strings.Contains(importText, "websocket") {
						integrationTypes = appendIfNotExists(integrationTypes, "WebSocket")
					}
					if strings.Contains(importText, "http") {
						integrationTypes = appendIfNotExists(integrationTypes, "REST")
					}
				}
			}
		case "javascript", "typescript":
			if node.Type() == "import_statement" {
				importText := strings.ToLower(code[node.StartByte():node.EndByte()])
				if strings.Contains(importText, "axios") || strings.Contains(importText, "fetch") ||
					strings.Contains(importText, "graphql") || strings.Contains(importText, "apollo") ||
					strings.Contains(importText, "socket") || strings.Contains(importText, "grpc") {
					matchedImports = append(matchedImports, importText)
					// Detect integration type
					if strings.Contains(importText, "graphql") || strings.Contains(importText, "apollo") {
						integrationTypes = appendIfNotExists(integrationTypes, "GraphQL")
					}
					if strings.Contains(importText, "grpc") {
						integrationTypes = appendIfNotExists(integrationTypes, "gRPC")
					}
					if strings.Contains(importText, "socket") {
						integrationTypes = appendIfNotExists(integrationTypes, "WebSocket")
					}
					if strings.Contains(importText, "axios") || strings.Contains(importText, "fetch") {
						integrationTypes = appendIfNotExists(integrationTypes, "REST")
					}
				}
			}
		case "python":
			if node.Type() == "import_statement" || node.Type() == "import_from_statement" {
				importText := strings.ToLower(code[node.StartByte():node.EndByte()])
				if strings.Contains(importText, "requests") || strings.Contains(importText, "httpx") ||
					strings.Contains(importText, "graphql") || strings.Contains(importText, "grpc") ||
					strings.Contains(importText, "websocket") {
					matchedImports = append(matchedImports, importText)
					// Detect integration type
					if strings.Contains(importText, "graphql") {
						integrationTypes = appendIfNotExists(integrationTypes, "GraphQL")
					}
					if strings.Contains(importText, "grpc") {
						integrationTypes = appendIfNotExists(integrationTypes, "gRPC")
					}
					if strings.Contains(importText, "websocket") {
						integrationTypes = appendIfNotExists(integrationTypes, "WebSocket")
					}
					if strings.Contains(importText, "requests") || strings.Contains(importText, "httpx") {
						integrationTypes = appendIfNotExists(integrationTypes, "REST")
					}
				}
			}
		}

		return true
	})

	return matchedFunctions, matchedImports, integrationTypes, nil
}

// calculateOverallConfidence calculates overall confidence from individual verifications
func calculateOverallConfidence(verifications []TaskVerification, factors VerificationFactors) float64 {
	var totalConfidence float64
	var totalWeight float64

	for _, verification := range verifications {
		var weight float64
		switch verification.VerificationType {
		case "code_existence":
			weight = factors.CodeExistence
		case "code_usage":
			weight = factors.CodeUsage
		case "test_coverage":
			weight = factors.TestCoverage
		case "integration":
			weight = factors.Integration
		default:
			weight = 0.0
		}

		totalConfidence += verification.Confidence * weight
		totalWeight += weight
	}

	if totalWeight == 0 {
		return 0.0
	}

	return totalConfidence / totalWeight
}

// determineTaskStatus determines task status based on confidence and verifications
func determineTaskStatus(confidence float64, verifications []TaskVerification) string {
	if confidence >= 0.8 {
		return "completed"
	} else if confidence >= 0.5 {
		return "in_progress"
	} else {
		return "pending"
	}
}

// buildEvidenceMap builds evidence map from verifications
func buildEvidenceMap(verifications []TaskVerification) map[string]interface{} {
	evidence := make(map[string]interface{})
	for _, verification := range verifications {
		evidence[verification.VerificationType] = map[string]interface{}{
			"status":     verification.Status,
			"confidence": verification.Confidence,
			"evidence":   verification.Evidence,
		}
	}
	return evidence
}

// storeVerification stores a verification result in the database
func storeVerification(ctx context.Context, verification TaskVerification) error {
	verificationID := uuid.New().String()
	now := time.Now()

	// Check if verification exists
	checkQuery := `SELECT id FROM task_verifications WHERE task_id = $1 AND verification_type = $2`
	var existingID string
	row := queryRowWithTimeout(ctx, checkQuery, verification.TaskID, verification.VerificationType)
	err := row.Scan(&existingID)

	// Initialize verifiedAt before using it in queries
	var verifiedAt *time.Time
	if verification.VerifiedAt != nil {
		verifiedAt = verification.VerifiedAt
	} else if verification.Status == "verified" {
		now := time.Now()
		verifiedAt = &now
	}

	var query string
	var args []interface{}

	if err == nil && existingID != "" {
		// Update existing
		query = `
			UPDATE task_verifications 
			SET status = $1, confidence = $2, evidence = $3, 
			    retry_count = retry_count + 1, verified_at = $4
			WHERE id = $5
		`
		evidenceJSON, _ := json.Marshal(verification.Evidence)
		args = []interface{}{
			verification.Status, verification.Confidence, string(evidenceJSON),
			verifiedAt, existingID,
		}
	} else {
		// Insert new
		query = `
			INSERT INTO task_verifications (
				id, task_id, verification_type, status, confidence, evidence, 
				retry_count, verified_at, created_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`
		evidenceJSON, _ := json.Marshal(verification.Evidence)
		args = []interface{}{
			verificationID, verification.TaskID, verification.VerificationType,
			verification.Status, verification.Confidence, string(evidenceJSON),
			verification.RetryCount, verifiedAt, now,
		}
	}

	_, err = execWithTimeout(ctx, query, args...)

	return err
}

// getCachedVerification retrieves cached verification results
func getCachedVerification(ctx context.Context, taskID string) (*VerifyTaskResponse, error) {
	query := `
		SELECT verification_type, status, confidence, evidence, verified_at
		FROM task_verifications
		WHERE task_id = $1
		ORDER BY created_at DESC
	`

	rows, err := queryWithTimeout(ctx, query, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get verifications: %w", err)
	}
	defer rows.Close()

	verifications := []TaskVerification{}
	var overallConfidence float64

	for rows.Next() {
		var verification TaskVerification
		var evidenceJSON string
		var verifiedAt sql.NullTime

		err := rows.Scan(
			&verification.VerificationType,
			&verification.Status,
			&verification.Confidence,
			&evidenceJSON,
			&verifiedAt,
		)
		if err != nil {
			continue
		}

		json.Unmarshal([]byte(evidenceJSON), &verification.Evidence)
		if verifiedAt.Valid {
			verification.VerifiedAt = &verifiedAt.Time
		}

		verifications = append(verifications, verification)
		overallConfidence += verification.Confidence
	}

	if len(verifications) > 0 {
		overallConfidence /= float64(len(verifications))
	}

	status := determineTaskStatus(overallConfidence, verifications)
	evidence := buildEvidenceMap(verifications)

	return &VerifyTaskResponse{
		TaskID:            taskID,
		OverallConfidence: overallConfidence,
		Verifications:     verifications,
		Status:            status,
		Evidence:          evidence,
	}, nil
}
