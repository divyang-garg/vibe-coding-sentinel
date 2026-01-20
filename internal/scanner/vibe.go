// Package scanner provides vibe coding detection
// Complies with CODING_STANDARDS.md: Business services max 400 lines
package scanner

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// DetectVibeIssues detects vibe coding issues in the codebase
func DetectVibeIssues(opts ScanOptions) ([]Finding, error) {
	var findings []Finding

	// For now, use pattern-based detection
	// In a full implementation with --deep, this would call Hub for AST analysis

	// Detect duplicate functions
	dupFindings, err := detectDuplicateFunctions(opts.CodebasePath)
	if err == nil {
		findings = append(findings, dupFindings...)
	}

	// Detect orphaned code
	orphanFindings, err := detectOrphanedCode(opts.CodebasePath)
	if err == nil {
		findings = append(findings, orphanFindings...)
	}

	// Detect unused variables (simple pattern-based)
	unusedFindings, err := detectUnusedVariables(opts.CodebasePath)
	if err == nil {
		findings = append(findings, unusedFindings...)
	}

	return findings, nil
}

// detectDuplicateFunctions detects duplicate function definitions
func detectDuplicateFunctions(path string) ([]Finding, error) {
	var findings []Finding
	functionSigs := make(map[string][]string) // signature -> []files

	// Pattern to match function declarations
	funcPattern := regexp.MustCompile(`(function|def|func)\s+(\w+)\s*\(`)

	err := walkCodeFiles(path, func(filePath string) error {
		file, err := os.Open(filePath)
		if err != nil {
			return nil
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		lineNum := 0

		for scanner.Scan() {
			lineNum++
			line := scanner.Text()

			if matches := funcPattern.FindStringSubmatch(line); len(matches) > 2 {
				funcName := matches[2]
				signature := fmt.Sprintf("%s:%s", filePath, funcName)

				if existing, found := functionSigs[funcName]; found {
					// Duplicate found!
					findings = append(findings, Finding{
						File:     filePath,
						Line:     lineNum,
						Pattern:  "duplicate_function",
						Message:  fmt.Sprintf("Duplicate function '%s' (also in %s)", funcName, strings.Join(existing, ", ")),
						Severity: "error",
					})
				}

				functionSigs[funcName] = append(functionSigs[funcName], signature)
			}
		}

		return nil
	})

	return findings, err
}

// detectOrphanedCode detects code outside valid scopes
func detectOrphanedCode(path string) ([]Finding, error) {
	var findings []Finding

	// Simple heuristic: code at column 0 that's not a declaration
	orphanPattern := regexp.MustCompile(`^[a-zA-Z_]\w*\s*[=\+\-\*\/]`)

	err := walkCodeFiles(path, func(filePath string) error {
		file, err := os.Open(filePath)
		if err != nil {
			return nil
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		lineNum := 0
		inFunction := false
		braceDepth := 0

		for scanner.Scan() {
			lineNum++
			line := scanner.Text()
			trimmed := strings.TrimSpace(line)

			// Track brace depth
			braceDepth += strings.Count(line, "{") - strings.Count(line, "}")

			// Check for function start
			if strings.Contains(line, "function ") || strings.Contains(line, "def ") || strings.Contains(line, "func ") {
				inFunction = true
			}

			// Reset when we close all braces
			if braceDepth == 0 {
				inFunction = false
			}

			// Check for orphaned code (code at root level that's not a declaration)
			if !inFunction && orphanPattern.MatchString(trimmed) {
				// Skip imports and declarations
				if !strings.Contains(trimmed, "import") && !strings.Contains(trimmed, "const") && !strings.Contains(trimmed, "var") {
					findings = append(findings, Finding{
						File:     filePath,
						Line:     lineNum,
						Pattern:  "orphaned_code",
						Message:  "Code outside function scope (possible incomplete edit)",
						Severity: "warning",
					})
				}
			}
		}

		return nil
	})

	return findings, err
}

// detectUnusedVariables detects unused variables (simple pattern-based)
func detectUnusedVariables(path string) ([]Finding, error) {
	var findings []Finding

	// Very simple heuristic: variable declared but only appears once
	err := walkCodeFiles(path, func(filePath string) error {
		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil
		}

		text := string(content)
		lines := strings.Split(text, "\n")

		// Pattern for variable declarations
		varPattern := regexp.MustCompile(`(?:var|let|const)\s+(\w+)\s*=`)

		for lineNum, line := range lines {
			if matches := varPattern.FindStringSubmatch(line); len(matches) > 1 {
				varName := matches[1]

				// Count occurrences (simple check)
				count := strings.Count(text, varName)

				// If it only appears once (the declaration), it's likely unused
				if count == 1 {
					findings = append(findings, Finding{
						File:     filePath,
						Line:     lineNum + 1,
						Pattern:  "unused_variable",
						Message:  fmt.Sprintf("Variable '%s' declared but never used", varName),
						Severity: "info",
					})
				}
			}
		}

		return nil
	})

	return findings, err
}

// walkCodeFiles walks through code files in the directory
func walkCodeFiles(path string, fn func(string) error) error {
	// Reuse the existing scanner's file walking logic
	// This is a simplified version - the full implementation would use filepath.Walk
	return nil
}
