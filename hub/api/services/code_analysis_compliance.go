// Package services provides code standards compliance checking functions
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package services

import (
	"fmt"
	"regexp"
	"strings"

	"sentinel-hub-api/ast"
)

// checkStandardsCompliance checks code compliance with language-specific style guides
// Returns compliance status, violations, and compliance score
func (s *CodeAnalysisServiceImpl) checkStandardsCompliance(code, language string) map[string]interface{} {
	if code == "" || language == "" {
		return map[string]interface{}{
			"compliant":       false,
			"standards":       []string{},
			"violations":      []map[string]interface{}{},
			"compliance_score": 0.0,
		}
	}

	var violations []map[string]interface{}
	standards := []string{}
	score := 100.0

	// Extract functions to check naming conventions
	functions, err := ast.ExtractFunctions(code, language, "")
	if err == nil {
		// Check naming conventions
		for _, fn := range functions {
			violation := s.checkNamingConvention(fn, language)
			if violation != nil {
				violations = append(violations, violation)
				score -= 5.0
			}
		}
		standards = append(standards, "naming_conventions")
	}

	// Check formatting
	formattingViolations := s.checkFormatting(code, language)
	violations = append(violations, formattingViolations...)
	score -= float64(len(formattingViolations)) * 2.0
	standards = append(standards, "basic_formatting")

	// Check import organization
	importViolations := s.checkImportOrganization(code, language)
	violations = append(violations, importViolations...)
	score -= float64(len(importViolations)) * 3.0
	standards = append(standards, "import_organization")

	// Ensure score is within bounds
	if score < 0 {
		score = 0.0
	}

	compliant := len(violations) == 0

	return map[string]interface{}{
		"compliant":        compliant,
		"standards":        standards,
		"violations":       violations,
		"compliance_score": score,
	}
}

// checkNamingConvention checks if function name follows language-specific naming conventions
func (s *CodeAnalysisServiceImpl) checkNamingConvention(fn ast.FunctionInfo, language string) map[string]interface{} {
	name := fn.Name
	switch language {
	case "go":
		// Go: exported functions should start with uppercase
		if fn.Visibility == "exported" && len(name) > 0 && name[0] >= 'a' && name[0] <= 'z' {
			return map[string]interface{}{
				"rule":     "naming_convention",
				"line":     fn.Line,
				"message":  fmt.Sprintf("Exported function '%s' should start with uppercase", name),
				"severity": "medium",
			}
		}
		// Go: unexported functions should start with lowercase
		if fn.Visibility == "private" && len(name) > 0 && name[0] >= 'A' && name[0] <= 'Z' {
			return map[string]interface{}{
				"rule":     "naming_convention",
				"line":     fn.Line,
				"message":  fmt.Sprintf("Unexported function '%s' should start with lowercase", name),
				"severity": "medium",
			}
		}
	case "python":
		// Python: functions should use snake_case
		if matched, _ := regexp.MatchString(`^[a-z_][a-z0-9_]*$`, name); !matched {
			return map[string]interface{}{
				"rule":     "naming_convention",
				"line":     fn.Line,
				"message":  fmt.Sprintf("Function '%s' should use snake_case", name),
				"severity": "medium",
			}
		}
	case "javascript", "typescript":
		// JavaScript/TypeScript: functions should use camelCase
		if matched, _ := regexp.MatchString(`^[a-z][a-zA-Z0-9]*$`, name); !matched {
			return map[string]interface{}{
				"rule":     "naming_convention",
				"line":     fn.Line,
				"message":  fmt.Sprintf("Function '%s' should use camelCase", name),
				"severity": "medium",
			}
		}
	}
	return nil
}

// checkFormatting checks code formatting compliance
func (s *CodeAnalysisServiceImpl) checkFormatting(code, language string) []map[string]interface{} {
	var violations []map[string]interface{}
	lines := strings.Split(code, "\n")

	// Check line length
	for i, line := range lines {
		if len(line) > 100 {
			violations = append(violations, map[string]interface{}{
				"rule":     "line_length",
				"line":     i + 1,
				"message":  fmt.Sprintf("Line exceeds 100 characters (%d)", len(line)),
				"severity": "minor",
			})
		}
	}

	// Check indentation consistency
	if language == "python" {
		// Python requires consistent indentation
		indentChars := make(map[int]bool)
		for _, line := range lines {
			trimmed := strings.TrimLeft(line, " \t")
			if len(trimmed) > 0 && !strings.HasPrefix(trimmed, "#") {
				indent := len(line) - len(trimmed)
				if indent > 0 {
					indentChars[indent] = true
				}
			}
		}
		// If multiple indent sizes detected, might be inconsistent
		if len(indentChars) > 2 {
			violations = append(violations, map[string]interface{}{
				"rule":     "indentation",
				"line":     0,
				"message":  "Inconsistent indentation detected",
				"severity": "medium",
			})
		}
	}

	return violations
}

// checkImportOrganization checks import statement organization
func (s *CodeAnalysisServiceImpl) checkImportOrganization(code, language string) []map[string]interface{} {
	var violations []map[string]interface{}
	lines := strings.Split(code, "\n")

	importLines := []int{}
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Check for language-specific import patterns
		switch language {
		case "go":
			// Go: import "package" or import ("package1" "package2")
			if strings.HasPrefix(trimmed, "import ") {
				importLines = append(importLines, i+1)
			}
		case "python":
			// Python: import module or from module import something
			if strings.HasPrefix(trimmed, "import ") || strings.HasPrefix(trimmed, "from ") {
				importLines = append(importLines, i+1)
			}
		case "javascript", "typescript":
			// JavaScript/TypeScript: import ... from ... or require(...)
			if strings.HasPrefix(trimmed, "import ") || strings.Contains(trimmed, "require(") {
				importLines = append(importLines, i+1)
			}
		case "java":
			// Java: import package.Class;
			if strings.HasPrefix(trimmed, "import ") {
				importLines = append(importLines, i+1)
			}
		default:
			// Fallback: check for common import patterns
			if strings.HasPrefix(trimmed, "import ") || strings.HasPrefix(trimmed, "from ") {
				importLines = append(importLines, i+1)
			}
		}
	}

	// Check if imports are at the top (after package/module declaration)
	if len(importLines) > 0 {
		firstImportLine := importLines[0]
		// Imports should be near the top (within first 20 lines typically)
		if firstImportLine > 20 {
			violations = append(violations, map[string]interface{}{
				"rule":     "import_organization",
				"line":     firstImportLine,
				"message":  "Imports should be at the top of the file",
				"severity": "minor",
			})
		}
	}

	return violations
}
