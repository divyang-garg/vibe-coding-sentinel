// Test Validator - Validation Check Functions
// Validates test structure, assertions, and completeness
// Complies with CODING_STANDARDS.md: Business Services max 400 lines

package services

import (
	"fmt"
	"strings"
)

// validateTestStructure validates the structure of a test function
func validateTestStructure(testCode string, language string) (bool, []string) {
	var issues []string
	isValid := true

	lines := strings.Split(testCode, "\n")
	hasAssertion := false

	// Look for test structure patterns based on language
	switch language {
	case "go":
		// Go tests: setup, execution, assertion
		for _, line := range lines {
			lineLower := strings.ToLower(line)
			if strings.Contains(lineLower, "assert") || strings.Contains(lineLower, "if") ||
				strings.Contains(lineLower, "require.") || strings.Contains(lineLower, "assert.") {
				hasAssertion = true
			}
		}

		if !hasAssertion {
			issues = append(issues, "Missing assertions - test does not verify expected behavior")
			isValid = false
		}

	case "javascript", "typescript":
		// JS/TS tests: setup, execution, assertion
		for _, line := range lines {
			lineLower := strings.ToLower(line)
			if strings.Contains(lineLower, "expect(") || strings.Contains(lineLower, "assert(") {
				hasAssertion = true
			}
		}

		if !hasAssertion {
			issues = append(issues, "Missing assertions - test does not verify expected behavior")
			isValid = false
		}

	case "python":
		// Python tests: setup, execution, assertion
		for _, line := range lines {
			lineLower := strings.ToLower(line)
			if strings.Contains(lineLower, "assert ") || strings.Contains(lineLower, "self.assert") {
				hasAssertion = true
			}
		}

		if !hasAssertion {
			issues = append(issues, "Missing assertions - test does not verify expected behavior")
			isValid = false
		}
	}

	// Check for shared state (test isolation)
	if strings.Contains(testCode, "global") || strings.Contains(testCode, "static") {
		issues = append(issues, "Potential shared state - test may not be isolated")
		isValid = false
	}

	return isValid, issues
}

// analyzeAssertions analyzes assertions in test code
func analyzeAssertions(testCode string, language string) (bool, []string) {
	var issues []string
	isValid := true

	lines := strings.Split(testCode, "\n")
	hasStrongAssertion := false

	// Look for strong assertions (not just null checks)
	for _, line := range lines {
		lineLower := strings.ToLower(line)

		// Weak assertions (just checking non-null)
		if strings.Contains(lineLower, "!= nil") || strings.Contains(lineLower, "!== null") ||
			strings.Contains(lineLower, "is not none") {
			// Check if there are stronger assertions nearby
			hasStrongAssertion = false
		}

		// Strong assertions (checking actual values)
		if strings.Contains(lineLower, "==") || strings.Contains(lineLower, "===") ||
			strings.Contains(lineLower, "equals") || strings.Contains(lineLower, "equal") {
			hasStrongAssertion = true
		}
	}

	if !hasStrongAssertion {
		issues = append(issues, "Weak assertions detected - test only checks for null/non-null, not actual values")
		isValid = false
	}

	return isValid, issues
}

// checkCompleteness checks if test covers all requirements
func checkCompleteness(testCode string, testRequirement TestRequirement) (bool, []string) {
	var issues []string
	isComplete := true

	testCodeLower := strings.ToLower(testCode)
	requirementDescLower := strings.ToLower(testRequirement.Description)

	// Extract keywords from requirement
	keywords := extractKeywords(requirementDescLower)

	// Check if test code mentions requirement keywords
	matchedKeywords := 0
	for _, keyword := range keywords {
		if strings.Contains(testCodeLower, keyword) {
			matchedKeywords++
		}
	}

	// If less than 50% of keywords match, test may not cover requirement
	if len(keywords) > 0 && float64(matchedKeywords)/float64(len(keywords)) < 0.5 {
		issues = append(issues, fmt.Sprintf("Test may not fully cover requirement: only %d/%d keywords matched", matchedKeywords, len(keywords)))
		isComplete = false
	}

	// Check requirement type coverage
	switch testRequirement.RequirementType {
	case "happy_path":
		if !strings.Contains(testCodeLower, "success") && !strings.Contains(testCodeLower, "valid") {
			issues = append(issues, "Happy path test may be missing - no success/valid scenarios found")
			isComplete = false
		}
	case "error_case":
		if !strings.Contains(testCodeLower, "error") && !strings.Contains(testCodeLower, "fail") &&
			!strings.Contains(testCodeLower, "invalid") && !strings.Contains(testCodeLower, "exception") {
			issues = append(issues, "Error case test may be missing - no error/failure scenarios found")
			isComplete = false
		}
	case "edge_case":
		if !strings.Contains(testCodeLower, "edge") && !strings.Contains(testCodeLower, "boundary") &&
			!strings.Contains(testCodeLower, "limit") && !strings.Contains(testCodeLower, "max") &&
			!strings.Contains(testCodeLower, "min") {
			issues = append(issues, "Edge case test may be missing - no boundary/limit scenarios found")
			isComplete = false
		}
	}

	return isComplete, issues
}
