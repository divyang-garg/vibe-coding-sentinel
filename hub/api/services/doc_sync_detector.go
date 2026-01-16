// Package doc_sync_detector - Code implementation detection for documentation synchronization
// Complies with CODING_STANDARDS.md: Utilities max 250 lines

package services

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// detectImplementation scans codebase for implementation evidence
func detectImplementation(feature string, codebasePath string) ImplementationEvidence {
	evidence := ImplementationEvidence{
		Feature:     feature,
		Files:       []string{},
		Functions:   []string{},
		Endpoints:   []string{},
		Tests:       []string{},
		LineNumbers: make(map[string][]int),
	}

	// Search for function definitions
	funcPattern := regexp.MustCompile(`func\s+(\w+).*?\(`)
	featureLower := strings.ToLower(feature)

	// Search in hub/api directory
	hubAPIPath := filepath.Join(codebasePath, "hub", "api")
	if files, err := findGoFiles(hubAPIPath); err == nil {
		for _, file := range files {
			if funcs, lines := findFunctionsInFile(file, funcPattern); len(funcs) > 0 {
				for i, fn := range funcs {
					fnLower := strings.ToLower(fn)
					if strings.Contains(fnLower, featureLower) || strings.Contains(featureLower, fnLower) {
						evidence.Functions = append(evidence.Functions, fn)
						evidence.Files = appendIfNotExists(evidence.Files, file)
						if evidence.LineNumbers[file] == nil {
							evidence.LineNumbers[file] = []int{}
						}
						evidence.LineNumbers[file] = append(evidence.LineNumbers[file], lines[i])
					}
				}
			}
		}
	}

	// Search for API endpoints
	endpointPattern := regexp.MustCompile(`r\.(Post|Get|Put|Delete)\(["']([^"']+)["']`)
	if endpoints, files, lines := findEndpointsInFile(filepath.Join(codebasePath, "hub", "api", "main.go"), endpointPattern); len(endpoints) > 0 {
		for i, endpoint := range endpoints {
			endpointLower := strings.ToLower(endpoint)
			if strings.Contains(endpointLower, featureLower) || strings.Contains(featureLower, endpointLower) {
				evidence.Endpoints = append(evidence.Endpoints, endpoint)
				evidence.Files = appendIfNotExists(evidence.Files, files[i])
				if evidence.LineNumbers[files[i]] == nil {
					evidence.LineNumbers[files[i]] = []int{}
				}
				evidence.LineNumbers[files[i]] = append(evidence.LineNumbers[files[i]], lines[i])
			}
		}
	}

	// Search for test files
	testPattern := regexp.MustCompile(`.*_test\.(go|sh)$`)
	if testFiles, err := findTestFiles(filepath.Join(codebasePath, "tests"), testPattern); err == nil {
		for _, testFile := range testFiles {
			testFileLower := strings.ToLower(testFile)
			if strings.Contains(testFileLower, featureLower) {
				evidence.Tests = append(evidence.Tests, testFile)
			}
		}
	}

	// Calculate confidence score
	evidence.Confidence = calculateConfidence(evidence)

	return evidence
}

// Helper functions for code detection

func findGoFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".go") {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func findFunctionsInFile(filePath string, pattern *regexp.Regexp) ([]string, []int) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, nil
	}

	lines := strings.Split(string(content), "\n")
	var funcs []string
	var lineNums []int

	for i, line := range lines {
		if matches := pattern.FindStringSubmatch(line); len(matches) > 1 {
			funcs = append(funcs, matches[1])
			lineNums = append(lineNums, i+1)
		}
	}

	return funcs, lineNums
}

func findEndpointsInFile(filePath string, pattern *regexp.Regexp) ([]string, []string, []int) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, nil, nil
	}

	lines := strings.Split(string(content), "\n")
	var endpoints []string
	var files []string
	var lineNums []int

	for i, line := range lines {
		if matches := pattern.FindStringSubmatch(line); len(matches) >= 3 {
			method := matches[1]
			path := matches[2]
			endpoint := fmt.Sprintf("%s %s", method, path)
			endpoints = append(endpoints, endpoint)
			files = append(files, filePath)
			lineNums = append(lineNums, i+1)
		}
	}

	return endpoints, files, lineNums
}

func findTestFiles(dir string, pattern *regexp.Regexp) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && pattern.MatchString(path) {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func appendIfNotExists(slice []string, item string) []string {
	for _, s := range slice {
		if s == item {
			return slice
		}
	}
	return append(slice, item)
}

func calculateConfidence(evidence ImplementationEvidence) float64 {
	score := 0.0

	// Base score for function existence
	if len(evidence.Functions) > 0 {
		score += 0.3
	}

	// Additional score for tests
	if len(evidence.Tests) > 0 {
		score += 0.3
	}

	// Additional score for API endpoints
	if len(evidence.Endpoints) > 0 {
		score += 0.3
	}

	// Additional score for multiple files
	if len(evidence.Files) > 1 {
		score += 0.1
	}

	// Cap at 1.0
	if score > 1.0 {
		score = 1.0
	}

	return score
}

// =============================================================================
// STATUS COMPARISON ENGINE (Task 3)
// =============================================================================

// determineCodeStatus determines the code implementation status based on evidence
func determineCodeStatus(evidence ImplementationEvidence) string {
	if len(evidence.Files) == 0 && len(evidence.Functions) == 0 && len(evidence.Endpoints) == 0 {
		return "MISSING"
	}

	totalEvidence := len(evidence.Files) + len(evidence.Functions) + len(evidence.Endpoints) + len(evidence.Tests)

	if totalEvidence > 0 && evidence.Confidence >= 0.8 {
		return "COMPLETE"
	} else if totalEvidence > 0 {
		return "PARTIAL"
	}

	return "MISSING"
}

// compareStatus compares documentation status with code evidence
func compareStatus(marker StatusMarker, evidence ImplementationEvidence) []Discrepancy {
	var discrepancies []Discrepancy

	// Determine code status based on evidence
	codeStatus := determineCodeStatus(evidence)

	// Check for mismatches
	if marker.Status == "COMPLETE" && codeStatus == "MISSING" {
		discrepancies = append(discrepancies, Discrepancy{
			Type:           "missing_impl",
			Phase:          marker.Phase,
			DocStatus:      marker.Status,
			CodeStatus:     codeStatus,
			Evidence:       evidence,
			Recommendation: fmt.Sprintf("Implement %s - documentation says COMPLETE but code is missing", marker.Phase),
			FilePath:       marker.FilePath,
			LineNumber:     marker.Line,
		})
	} else if marker.Status == "STUB" && codeStatus == "COMPLETE" {
		discrepancies = append(discrepancies, Discrepancy{
			Type:           "status_mismatch",
			Phase:          marker.Phase,
			DocStatus:      marker.Status,
			CodeStatus:     codeStatus,
			Evidence:       evidence,
			Recommendation: fmt.Sprintf("Update documentation status from STUB to COMPLETE for %s", marker.Phase),
			FilePath:       marker.FilePath,
			LineNumber:     marker.Line,
		})
	} else if marker.Status == "PENDING" && codeStatus == "COMPLETE" {
		discrepancies = append(discrepancies, Discrepancy{
			Type:           "status_mismatch",
			Phase:          marker.Phase,
			DocStatus:      marker.Status,
			CodeStatus:     codeStatus,
			Evidence:       evidence,
			Recommendation: fmt.Sprintf("Update documentation status from PENDING to COMPLETE for %s", marker.Phase),
			FilePath:       marker.FilePath,
			LineNumber:     marker.Line,
		})
	} else if marker.Status == "COMPLETE" && codeStatus == "PARTIAL" {
		discrepancies = append(discrepancies, Discrepancy{
			Type:           "partial_match",
			Phase:          marker.Phase,
			DocStatus:      marker.Status,
			CodeStatus:     codeStatus,
			Evidence:       evidence,
			Recommendation: fmt.Sprintf("Complete implementation for %s - documentation says COMPLETE but code is partial", marker.Phase),
			FilePath:       marker.FilePath,
			LineNumber:     marker.Line,
		})
	}

	return discrepancies
}
