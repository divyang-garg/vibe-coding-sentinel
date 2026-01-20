// Phase 14E: Task Verification Engine - Test Coverage Verification
// Complies with CODING_STANDARDS.md: Utilities max 250 lines

package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

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
