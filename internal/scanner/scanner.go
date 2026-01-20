// Package scanner provides security scanning functionality
// Complies with CODING_STANDARDS.md: Business services max 400 lines
package scanner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// getCurrentTimestamp returns current timestamp in RFC3339 format
func getCurrentTimestamp() string {
	return time.Now().Format(time.RFC3339)
}

// Scan performs a security scan on the specified codebase
func Scan(opts ScanOptions) (*Result, error) {
	// Use parallel scanning by default for better performance
	// Fall back to sequential for small codebases or when explicitly requested
	return ScanParallel(opts)
}

// ScanSequential performs sequential scanning (legacy implementation)
func ScanSequential(opts ScanOptions) (*Result, error) {
	result := &Result{
		Success:   true,
		Findings:  []Finding{},
		Summary:   make(map[string]int),
		Timestamp: getCurrentTimestamp(),
	}

	patterns := GetSecurityPatterns()

	// Determine scan directory
	scanDir := opts.CodebasePath
	if scanDir == "" {
		scanDir = "."
	}

	// Resolve absolute path
	absPath, err := filepath.Abs(scanDir)
	if err != nil {
		absPath = scanDir
	}

	if !opts.CIMode {
		fmt.Printf("ℹ️  Scanning directory: %s\n", absPath)
	}

	// Walk through codebase
	err = filepath.Walk(absPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we can't read
		}

		// Skip directories
		if info.IsDir() {
			// Skip common ignore directories
			if shouldSkipDir(path) {
				return filepath.SkipDir
			}
			return nil
		}

		// Only scan code files
		ext := filepath.Ext(path)
		if !isCodeFile(ext) {
			return nil
		}

		// Read file content
		content, err := os.ReadFile(path)
		if err != nil {
			return nil // Skip files we can't read
		}

		// Check each pattern
		lines := strings.Split(string(content), "\n")
		for lineNum, line := range lines {
			for _, pattern := range patterns {
				if pattern.Regex.MatchString(line) {
					// Calculate relative path
					relPath, _ := filepath.Rel(absPath, path)
					if relPath == "" {
						relPath = path
					}

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
						Pattern:  strings.TrimSpace(line),
						Code:     strings.TrimSpace(line),
					}
					result.Findings = append(result.Findings, finding)
					result.Summary[pattern.Name]++

					// Mark as failed if critical finding
					if pattern.Severity == SeverityCritical {
						result.Success = false
					}
				}
			}
		}

		return nil
	})

	if err != nil && !opts.CIMode {
		fmt.Printf("⚠️  Warning: Some files could not be scanned: %v\n", err)
	}

	// Apply baseline filtering if baseline exists
	result = filterBaseline(result)

	return result, nil
}

// BaselineEntry represents an accepted finding (simplified for scanner use)
type BaselineEntry struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Pattern string `json:"pattern"`
	Hash    string `json:"hash"`
}

// Baseline represents the baseline file structure
type Baseline struct {
	Version string          `json:"version"`
	Entries []BaselineEntry `json:"entries"`
}

// filterBaseline filters findings against baseline
func filterBaseline(result *Result) *Result {
	// Try to load baseline
	baselinePath := ".sentinel/baseline.json"
	baselineData, err := os.ReadFile(baselinePath)
	if err != nil {
		// No baseline file, return as-is
		return result
	}

	// Parse baseline
	var baseline Baseline
	if err := json.Unmarshal(baselineData, &baseline); err != nil {
		// Invalid baseline, return as-is
		return result
	}

	// Create hash map for quick lookup
	baselineMap := make(map[string]bool)
	for _, entry := range baseline.Entries {
		// Create hash from file:line or use stored hash
		if entry.Hash != "" {
			baselineMap[entry.Hash] = true
		}
		hash := fmt.Sprintf("%s:%d", entry.File, entry.Line)
		baselineMap[hash] = true
	}

	// Filter findings
	filtered := make([]Finding, 0)
	for _, finding := range result.Findings {
		hash := fmt.Sprintf("%s:%d", finding.File, finding.Line)
		if !baselineMap[hash] {
			filtered = append(filtered, finding)
		}
	}

	// Update result
	result.Findings = filtered

	// Recalculate summary
	result.Summary = make(map[string]int)
	result.Success = true
	for _, finding := range filtered {
		result.Summary[finding.Type]++
		if finding.Severity == SeverityCritical {
			result.Success = false
		}
	}

	return result
}

// shouldSkipDir determines if a directory should be skipped
func shouldSkipDir(path string) bool {
	skipDirs := []string{
		".git", "node_modules", "vendor", ".next", "dist", "build",
		"target", "bin", "obj", ".vscode", ".idea", ".vs",
		"coverage", ".nyc_output", "__pycache__",
	}

	pathLower := strings.ToLower(path)
	for _, skipDir := range skipDirs {
		if strings.Contains(pathLower, skipDir) {
			return true
		}
	}
	return false
}

// isCodeFile determines if a file extension represents a code file
func isCodeFile(ext string) bool {
	codeExts := map[string]bool{
		".js": true, ".ts": true, ".jsx": true, ".tsx": true,
		".go": true, ".py": true, ".java": true, ".cs": true,
		".php": true, ".rb": true, ".sql": true, ".sh": true,
		".swift": true, ".kt": true, ".scala": true,
	}
	return codeExts[ext]
}
