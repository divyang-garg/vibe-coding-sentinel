// Phase 14A: Test Layer Analyzer
// Analyzes test files for coverage, quality, and scenario completeness

package services

import (
	"context"
	"fmt"
	"os"
	"strings"
)

// TestLayerFinding represents a finding from test layer analysis
type TestLayerFinding struct {
	Type     string `json:"type"`     // "missing_coverage", "weak_assertion", "missing_edge_case", "missing_error_case"
	Location string `json:"location"` // Test file path
	Issue    string `json:"issue"`
	Severity string `json:"severity"` // "critical", "high", "medium", "low"
}

// analyzeTestLayer analyzes test files for coverage and quality
func analyzeTestLayer(ctx context.Context, feature *DiscoveredFeature) ([]TestLayerFinding, error) {
	findings := []TestLayerFinding{}

	if feature.TestLayer == nil {
		return findings, nil
	}

	// Reuse test coverage tracker
	for _, testFile := range feature.TestLayer.TestFiles {
		// Read test file
		data, err := os.ReadFile(testFile.Path)
		if err != nil {
			LogWarn(ctx, "Failed to read test file %s: %v", testFile.Path, err)
			continue
		}

		content := string(data)

		// Check test coverage (reuse test coverage tracker logic)
		coverageFindings := checkTestCoverageForFile(content, testFile)
		findings = append(findings, coverageFindings...)

		// Analyze test scenarios
		scenarioFindings := analyzeTestScenarios(content, testFile)
		findings = append(findings, scenarioFindings...)

		// Check assertion quality
		assertionFindings := checkAssertionQuality(content, testFile)
		findings = append(findings, assertionFindings...)
	}

	return findings, nil
}

// analyzeTestScenarios analyzes test files for scenario completeness
func analyzeTestScenarios(content string, testFile TestFileInfo) []TestLayerFinding {
	findings := []TestLayerFinding{}

	// Check for happy path tests
	hasHappyPath := strings.Contains(content, "happy") || strings.Contains(content, "success") || strings.Contains(content, "should work")

	// Check for error case tests
	hasErrorCases := strings.Contains(content, "error") || strings.Contains(content, "fail") || strings.Contains(content, "throw") || strings.Contains(content, "exception")

	// Check for boundary tests
	hasBoundaryTests := strings.Contains(content, "boundary") || strings.Contains(content, "edge") || strings.Contains(content, "limit") || strings.Contains(content, "max") || strings.Contains(content, "min")

	if !hasHappyPath {
		findings = append(findings, TestLayerFinding{
			Type:     "missing_scenario",
			Location: testFile.Path,
			Issue:    "Test file may be missing happy path test cases",
			Severity: "high",
		})
	}

	if !hasErrorCases {
		findings = append(findings, TestLayerFinding{
			Type:     "missing_error_case",
			Location: testFile.Path,
			Issue:    "Test file may be missing error case test scenarios",
			Severity: "medium",
		})
	}

	if !hasBoundaryTests {
		findings = append(findings, TestLayerFinding{
			Type:     "missing_edge_case",
			Location: testFile.Path,
			Issue:    "Test file may be missing boundary/edge case tests",
			Severity: "medium",
		})
	}

	return findings
}

// checkTestCoverageForFile checks if test file has adequate coverage
func checkTestCoverageForFile(content string, testFile TestFileInfo) []TestLayerFinding {
	findings := []TestLayerFinding{}

	// Count test cases
	testCaseCount := countTestCases(content, testFile.Framework)

	if testCaseCount == 0 {
		findings = append(findings, TestLayerFinding{
			Type:     "missing_coverage",
			Location: testFile.Path,
			Issue:    "Test file appears to have no test cases",
			Severity: "critical",
		})
	} else if testCaseCount < 3 {
		findings = append(findings, TestLayerFinding{
			Type:     "missing_coverage",
			Location: testFile.Path,
			Issue:    fmt.Sprintf("Test file has only %d test case(s), may need more coverage", testCaseCount),
			Severity: "medium",
		})
	}

	return findings
}

// checkAssertionQuality checks the quality of test assertions
func checkAssertionQuality(content string, testFile TestFileInfo) []TestLayerFinding {
	findings := []TestLayerFinding{}

	// Check for weak assertions (e.g., just checking truthiness)
	hasWeakAssertions := strings.Contains(content, "expect(true)") || strings.Contains(content, "assert(true)")

	// Check for specific assertions (better quality)
	hasSpecificAssertions := strings.Contains(content, "expect(") || strings.Contains(content, "assertEqual") || strings.Contains(content, "assert.deepEqual")

	if hasWeakAssertions && !hasSpecificAssertions {
		findings = append(findings, TestLayerFinding{
			Type:     "weak_assertion",
			Location: testFile.Path,
			Issue:    "Test file contains weak assertions (e.g., expect(true)) that may not verify actual behavior",
			Severity: "low",
		})
	}

	return findings
}

// Helper functions

func countTestCases(content string, framework string) int {
	count := 0

	switch framework {
	case "jest", "mocha":
		// Count "it(" or "test(" calls
		count += strings.Count(content, "it(")
		count += strings.Count(content, "test(")
	case "pytest":
		// Count "def test_" functions
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			if strings.Contains(line, "def test_") {
				count++
			}
		}
	case "go-test":
		// Count "func Test" functions
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			if strings.Contains(line, "func Test") {
				count++
			}
		}
	default:
		// Generic counting
		count += strings.Count(content, "test(")
		count += strings.Count(content, "it(")
		count += strings.Count(content, "def test_")
		count += strings.Count(content, "func Test")
	}

	return count
}
