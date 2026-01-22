// Package services test detector
// Detects test files and functions matching business rule keywords
// Complies with CODING_STANDARDS.md: Utilities max 250 lines

package services

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// detectTests finds test files and functions matching business rule
// Supports multiple test frameworks: Go testing, Jest, pytest
func detectTests(codebasePath string, ruleTitle string, keywords []string) []string {
	var tests []string

	// Scan for test files
	testFiles := scanTestFiles(codebasePath)

	for _, testFile := range testFiles {
		content, err := os.ReadFile(testFile)
		if err != nil {
			continue
		}

		contentStr := string(content)
		framework := detectTestFramework(testFile)

		switch framework {
		case "go-testing":
			testFuncs := detectGoTests(contentStr, keywords)
			tests = append(tests, testFuncs...)
		case "jest":
			testFuncs := detectJestTests(contentStr, keywords)
			tests = append(tests, testFuncs...)
		case "pytest":
			testFuncs := detectPytestTests(contentStr, keywords)
			tests = append(tests, testFuncs...)
		default:
			// Try generic test detection
			testFuncs := detectGenericTests(contentStr, keywords)
			tests = append(tests, testFuncs...)
		}
	}

	return tests
}

// detectTestFramework is defined in helpers.go to avoid duplication
// scanTestFiles scans codebase for test files recursively
func scanTestFiles(codebasePath string) []string {
	var testFiles []string

	err := filepath.Walk(codebasePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we can't access
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check if file matches test patterns
		fileName := strings.ToLower(info.Name())
		ext := strings.ToLower(filepath.Ext(path))

		switch {
		case strings.HasSuffix(fileName, "_test.go"):
			testFiles = append(testFiles, path)
		case strings.Contains(fileName, ".test.") && (ext == ".js" || ext == ".ts"):
			testFiles = append(testFiles, path)
		case strings.Contains(fileName, ".spec.") && (ext == ".js" || ext == ".ts"):
			testFiles = append(testFiles, path)
		case strings.HasPrefix(fileName, "test_") && ext == ".py":
			testFiles = append(testFiles, path)
		case strings.HasSuffix(fileName, "_test.py"):
			testFiles = append(testFiles, path)
		}

		return nil
	})

	if err != nil {
		// Return empty slice on error to allow processing to continue
		return []string{}
	}

	return testFiles
}

// detectGoTests detects Go test functions
func detectGoTests(code string, keywords []string) []string {
	var tests []string

	// Pattern: func TestXxx(t *testing.T)
	pattern := regexp.MustCompile(`func\s+(Test\w+)\s*\([^)]*testing\.T`)
	matches := pattern.FindAllStringSubmatch(code, -1)

	for _, match := range matches {
		if len(match) >= 2 {
			testName := match[1]
			// Check if test name matches keywords
			if matchesKeywords(testName, keywords) {
				tests = appendIfNotExists(tests, testName)
			}
		}
	}

	return tests
}

// detectJestTests detects Jest test functions
func detectJestTests(code string, keywords []string) []string {
	var tests []string

	// Pattern: test('test name', ...) or it('test name', ...) or describe('suite', ...)
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?:test|it|describe)\s*\(\s*['"]([^'"]+)['"]`),
		regexp.MustCompile(`(?:test|it|describe)\s*\(\s*['"]([^'"]+)['"]`),
	}

	for _, pattern := range patterns {
		matches := pattern.FindAllStringSubmatch(code, -1)
		for _, match := range matches {
			if len(match) >= 2 {
				testName := match[1]
				if matchesKeywords(testName, keywords) {
					tests = appendIfNotExists(tests, testName)
				}
			}
		}
	}

	return tests
}

// detectPytestTests detects pytest test functions
func detectPytestTests(code string, keywords []string) []string {
	var tests []string

	// Pattern: def test_xxx(...)
	pattern := regexp.MustCompile(`def\s+(test_\w+)\s*\(`)
	matches := pattern.FindAllStringSubmatch(code, -1)

	for _, match := range matches {
		if len(match) >= 2 {
			testName := match[1]
			if matchesKeywords(testName, keywords) {
				tests = appendIfNotExists(tests, testName)
			}
		}
	}

	return tests
}

// detectGenericTests tries to detect tests using generic patterns
func detectGenericTests(code string, keywords []string) []string {
	var tests []string

	// Look for common test function patterns
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`function\s+test\w+\s*\(`),
		regexp.MustCompile(`const\s+test\w+\s*=`),
		regexp.MustCompile(`def\s+test\w+\s*\(`),
	}

	for _, pattern := range patterns {
		matches := pattern.FindAllStringSubmatch(code, -1)
		for _, match := range matches {
			if len(match) > 0 {
				// Extract potential test name
				testName := strings.TrimSpace(match[0])
				if matchesKeywords(testName, keywords) {
					tests = appendIfNotExists(tests, testName)
				}
			}
		}
	}

	return tests
}
