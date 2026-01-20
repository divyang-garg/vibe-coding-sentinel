// Package ast provides edge case detection for AST findings
// Complies with CODING_STANDARDS.md: Detection modules max 250 lines
package ast

import (
	"regexp"
	"strings"
)

// EdgeCaseResult contains detected edge cases that reduce confidence
type EdgeCaseResult struct {
	HasReflection     bool    // reflect.ValueOf, eval(), getattr()
	HasDynamicImport  bool    // import(), __import__(), require()
	HasPluginUsage    bool    // plugin.Lookup, dynamic loading
	HasGeneratedCode  bool    // Code generation markers
	ConfidencePenalty float64 // Penalty to apply to confidence (0.0-1.0)
}

// DetectEdgeCases identifies risky patterns that reduce confidence
func DetectEdgeCases(code, language string) EdgeCaseResult {
	result := EdgeCaseResult{}

	switch language {
	case "go":
		result.HasReflection = detectGoReflection(code)
		result.HasPluginUsage = detectGoPlugin(code)
		result.HasGeneratedCode = detectGeneratedCode(code)
	case "javascript", "typescript":
		result.HasReflection = detectJSReflection(code)
		result.HasDynamicImport = detectJSDynamicImport(code)
		result.HasGeneratedCode = detectGeneratedCode(code)
	case "python":
		result.HasReflection = detectPythonReflection(code)
		result.HasDynamicImport = detectPythonDynamicImport(code)
		result.HasGeneratedCode = detectGeneratedCode(code)
	default:
		// Generic detection for unknown languages
		result.HasReflection = detectGenericReflection(code)
		result.HasGeneratedCode = detectGeneratedCode(code)
	}

	// Calculate confidence penalty based on detected edge cases
	result.ConfidencePenalty = calculatePenalty(result)

	return result
}

// detectGoReflection detects Go reflection usage
func detectGoReflection(code string) bool {
	patterns := []string{
		`reflect\.ValueOf`,
		`reflect\.TypeOf`,
		`reflect\.MethodByName`,
		`reflect\.Call`,
		`\.MethodByName\(`,
		`reflect\.New`,
	}
	return matchesAnyPattern(code, patterns)
}

// detectGoPlugin detects Go plugin usage
func detectGoPlugin(code string) bool {
	patterns := []string{
		`plugin\.Lookup`,
		`plugin\.Open`,
		`\.Lookup\(`,
	}
	return matchesAnyPattern(code, patterns)
}

// detectJSReflection detects JavaScript/TypeScript reflection/dynamic code
func detectJSReflection(code string) bool {
	patterns := []string{
		`eval\s*\(`,
		`Function\s*\(`,
		`new Function`,
		`\.constructor`,
		`Object\.getOwnPropertyNames`,
		`Object\.keys`,
		`Reflect\.`,
	}
	return matchesAnyPattern(code, patterns)
}

// detectJSDynamicImport detects JavaScript/TypeScript dynamic imports
func detectJSDynamicImport(code string) bool {
	patterns := []string{
		`import\s*\(`,
		`require\s*\([^)]*\+`,
		`require\([^)]*\[`,
		`System\.import`,
		`__webpack_require__`,
	}
	return matchesAnyPattern(code, patterns)
}

// detectPythonReflection detects Python reflection/dynamic code
func detectPythonReflection(code string) bool {
	patterns := []string{
		`eval\s*\(`,
		`exec\s*\(`,
		`getattr\s*\(`,
		`setattr\s*\(`,
		`hasattr\s*\(`,
		`__getattribute__`,
		`__setattr__`,
		`inspect\.`,
	}
	return matchesAnyPattern(code, patterns)
}

// detectPythonDynamicImport detects Python dynamic imports
func detectPythonDynamicImport(code string) bool {
	patterns := []string{
		`__import__\s*\(`,
		`importlib\.`,
		`importlib\.import_module`,
		`importlib\.util\.find_spec`,
		`imp\.load_module`,
		`imp\.find_module`,
	}
	return matchesAnyPattern(code, patterns)
}

// detectGeneratedCode detects code generation markers
func detectGeneratedCode(code string) bool {
	// Check for common code generation markers
	markers := []string{
		"// Code generated",
		"// DO NOT EDIT",
		"// This file was generated",
		"# Code generated",
		"# DO NOT EDIT",
		"# This file was generated",
		"@generated",
		"<!-- Code generated",
		"<!-- DO NOT EDIT",
	}

	codeUpper := strings.ToUpper(code)
	for _, marker := range markers {
		if strings.Contains(codeUpper, strings.ToUpper(marker)) {
			return true
		}
	}

	return false
}

// detectGenericReflection detects generic reflection patterns
func detectGenericReflection(code string) bool {
	patterns := []string{
		`eval\s*\(`,
		`exec\s*\(`,
		`\.call\s*\(`,
		`\.apply\s*\(`,
	}
	return matchesAnyPattern(code, patterns)
}

// matchesAnyPattern checks if code matches any of the given regex patterns
func matchesAnyPattern(code string, patterns []string) bool {
	for _, pattern := range patterns {
		matched, err := regexp.MatchString(pattern, code)
		if err == nil && matched {
			return true
		}
	}
	return false
}

// calculatePenalty calculates confidence penalty based on detected edge cases
func calculatePenalty(result EdgeCaseResult) float64 {
	penalty := 0.0

	// Reflection usage significantly reduces confidence
	if result.HasReflection {
		penalty += 0.30
	}

	// Dynamic imports reduce confidence
	if result.HasDynamicImport {
		penalty += 0.20
	}

	// Plugin usage reduces confidence
	if result.HasPluginUsage {
		penalty += 0.25
	}

	// Generated code should never be auto-fixed
	if result.HasGeneratedCode {
		penalty += 1.0 // Maximum penalty
	}

	// Cap penalty at 1.0
	if penalty > 1.0 {
		penalty = 1.0
	}

	return penalty
}
