// Test Coverage Tracker - Analysis Functions
// Parses test files and maps them to business rules
// Complies with CODING_STANDARDS.md: Business Services max 400 lines

package services

import (
	"strings"
)

// discoverTestFiles extracts test file paths and content from TestFile structs
func discoverTestFiles(testFiles []TestFile) []TestFile {
	// Return provided files (Agent sends them with content)
	return testFiles
}

// parseTestFile parses a test file to extract test functions and map them to business rules
// This is a simplified version - in production, use AST analysis
func parseTestFile(testFilePath string, testCode string) ([]string, error) {
	var testFunctions []string

	// Simple heuristic: look for test function patterns
	lines := strings.Split(testCode, "\n")
	for _, line := range lines {
		lineLower := strings.ToLower(line)

		// Go: func TestXxx(t *testing.T)
		if strings.Contains(lineLower, "func test") {
			parts := strings.Fields(line)
			for i, part := range parts {
				if part == "func" && i+1 < len(parts) {
					funcName := parts[i+1]
					// Remove receiver if present
					if !strings.Contains(funcName, "(") {
						testFunctions = append(testFunctions, funcName)
					}
				}
			}
		}

		// JavaScript/TypeScript: test('...', ...) or it('...', ...)
		if strings.Contains(lineLower, "test(") || strings.Contains(lineLower, "it(") {
			// Extract test name from string literal
			if idx := strings.Index(line, "'"); idx != -1 {
				if endIdx := strings.Index(line[idx+1:], "'"); endIdx != -1 {
					testName := line[idx+1 : idx+1+endIdx]
					testFunctions = append(testFunctions, testName)
				}
			}
		}

		// Python: def test_xxx():
		if strings.Contains(lineLower, "def test_") {
			parts := strings.Fields(line)
			for _, part := range parts {
				if strings.HasPrefix(part, "test_") {
					testFunctions = append(testFunctions, part)
					break
				}
			}
		}
	}

	return testFunctions, nil
}

// mapTestsToBusinessRules maps test functions to business rules
// This is a simplified heuristic - in production, use AST analysis and annotations
func mapTestsToBusinessRules(testFunctions []string, testFilePath string, businessRules []KnowledgeItem) map[string][]string {
	ruleToTests := make(map[string][]string)

	// Simple heuristic: match test function names with rule title keywords
	for _, rule := range businessRules {
		ruleTitleLower := strings.ToLower(rule.Title)
		keywords := extractKeywords(ruleTitleLower)

		var matchingTests []string
		for _, testFunc := range testFunctions {
			testFuncLower := strings.ToLower(testFunc)
			for _, keyword := range keywords {
				if strings.Contains(testFuncLower, keyword) {
					matchingTests = append(matchingTests, testFilePath+":"+testFunc)
					break
				}
			}
		}

		if len(matchingTests) > 0 {
			ruleToTests[rule.ID] = matchingTests
		}
	}

	return ruleToTests
}

// isRequirementCovered checks if a test requirement is covered by test files
func isRequirementCovered(req TestRequirement, testPaths []string, testFiles []TestFile) bool {
	// Extract keywords from requirement description
	keywords := extractKeywords(strings.ToLower(req.Description))
	if len(keywords) == 0 {
		return false
	}

	// Check if any test file mentions these keywords
	for _, testPath := range testPaths {
		for _, testFile := range testFiles {
			// Check if this test file matches the path
			if strings.Contains(testFile.Path, testPath) || testFile.Path == testPath {
				testContentLower := strings.ToLower(testFile.Content)
				matchedKeywords := 0
				for _, keyword := range keywords {
					if strings.Contains(testContentLower, keyword) {
						matchedKeywords++
					}
				}
				// If 70%+ keywords match, consider requirement covered
				if float64(matchedKeywords)/float64(len(keywords)) >= 0.7 {
					return true
				}
			}
		}
	}
	return false
}

// calculateCoverage calculates coverage percentage for a business rule
func calculateCoverage(ruleID string, testRequirements []TestRequirement, ruleToTests map[string][]string, testFiles []TestFile) float64 {
	// Get test requirements for this rule
	var ruleRequirements []TestRequirement
	for _, req := range testRequirements {
		if req.KnowledgeItemID == ruleID {
			ruleRequirements = append(ruleRequirements, req)
		}
	}

	if len(ruleRequirements) == 0 {
		return 0.0 // No requirements = no coverage
	}

	// Check if tests exist for this rule
	tests := ruleToTests[ruleID]
	if len(tests) == 0 {
		return 0.0 // No tests = 0% coverage
	}

	// Analyze test content to determine which requirements are covered
	coveredRequirements := 0
	for _, req := range ruleRequirements {
		if isRequirementCovered(req, tests, testFiles) {
			coveredRequirements++
		}
	}

	return float64(coveredRequirements) / float64(len(ruleRequirements))
}
