// Parallel Scanner - Helper Functions
// File collection, scanning, and filtering functions
// Complies with CODING_STANDARDS.md: Utilities max 250 lines

package scanner

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// collectFiles collects all code files to scan
func collectFiles(codebasePath string) ([]string, error) {
	var files []string
	absPath, err := filepath.Abs(codebasePath)
	if err != nil {
		absPath = codebasePath
	}

	err = filepath.Walk(absPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we can't read
		}

		if info.IsDir() {
			if shouldSkipDir(path) {
				return filepath.SkipDir
			}
			return nil
		}

		ext := filepath.Ext(path)
		if isCodeFile(ext) {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

// scanFile scans a single file and returns findings
func scanFile(filePath string, patterns []Pattern, absPath string) []Finding {
	var findings []Finding

	content, err := os.ReadFile(filePath)
	if err != nil {
		return findings // Return empty if can't read
	}

	contentStr := string(content)

	// Calculate relative path
	relPath, _ := filepath.Rel(absPath, filePath)
	if relPath == "" {
		relPath = filePath
	}

	// Scan with regex patterns
	lines := splitLines(contentStr)
	for lineNum, line := range lines {
		for _, pattern := range patterns {
			if pattern.Regex.MatchString(line) {
				// Find column position of match
				matches := pattern.Regex.FindStringIndex(line)
				column := 0
				if len(matches) > 0 {
					column = matches[0] + 1
				}

				finding := Finding{
					Type:     pattern.Name,
					Severity: pattern.Severity,
					File:     relPath,
					Line:     lineNum + 1,
					Column:   column,
					Message:  pattern.Message,
					Pattern:  trimSpace(line),
					Code:     trimSpace(line),
				}
				findings = append(findings, finding)
			}
		}
	}

	// Also check for entropy-based secrets
	entropyFindings := detectEntropySecrets(contentStr, relPath)
	findings = append(findings, entropyFindings...)

	return findings
}

// filterBaselineParallel filters findings against baseline (called from parallel scan)
func filterBaselineParallel(result *Result) *Result {
	// Try to load baseline
	baselinePath := ".sentinel/baseline.json"
	baselineData, err := os.ReadFile(baselinePath)
	if err != nil {
		// No baseline file, return as-is
		return result
	}

	// Parse baseline - simple structure
	type BaselineEntry struct {
		File string `json:"file"`
		Line int    `json:"line"`
		Hash string `json:"hash"`
	}
	type Baseline struct {
		Version string          `json:"version"`
		Entries []BaselineEntry `json:"entries"`
	}

	var baseline Baseline
	if err := json.Unmarshal(baselineData, &baseline); err != nil {
		return result
	}

	// Create hash map for quick lookup
	baselineMap := make(map[string]bool)
	for _, entry := range baseline.Entries {
		if entry.Hash != "" {
			baselineMap[entry.Hash] = true
		}
		hash := entry.File + ":" + itoa(entry.Line)
		baselineMap[hash] = true
	}

	// Filter findings
	filtered := make([]Finding, 0)
	for _, finding := range result.Findings {
		hash := finding.File + ":" + itoa(finding.Line)
		if !baselineMap[hash] {
			filtered = append(filtered, finding)
		}
	}

	result.Findings = filtered

	// Recalculate summary and success
	result.Summary = make(map[string]int)
	result.Success = true
	for _, f := range filtered {
		result.Summary[f.Type]++
		if f.Severity == SeverityCritical {
			result.Success = false
		}
	}

	return result
}
