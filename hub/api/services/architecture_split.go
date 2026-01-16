// Package architecture_split - File splitting suggestions for architecture analysis
// Complies with CODING_STANDARDS.md: Utilities max 250 lines

package services

import (
	"fmt"
	"path/filepath"
	"strings"
)

func extractFunctionName(line string, prefix string) string {
	// Remove prefix
	if prefix != "" {
		line = strings.TrimPrefix(line, prefix)
	}

	// Extract name (first word after prefix)
	parts := strings.Fields(line)
	if len(parts) > 0 {
		name := parts[0]
		// Remove parentheses, brackets, etc.
		name = strings.Trim(name, "()[]{}=:")
		return name
	}
	return "unknown"
}

// generateSplitSuggestion generates a split suggestion for an oversized file
func generateSplitSuggestion(file FileContent, sections []FileSection) *SplitSuggestion {
	if len(sections) < 2 {
		return nil // Not enough sections to split
	}

	var proposedFiles []ProposedFile
	var migrationInstructions []string

	// Group sections into logical files
	// Simple strategy: split into ~3-4 files of roughly equal size
	targetFiles := 3
	if len(sections) > 6 {
		targetFiles = 4
	}

	sectionsPerFile := len(sections) / targetFiles
	if sectionsPerFile < 1 {
		sectionsPerFile = 1
	}

	basePath := file.Path
	ext := filepath.Ext(basePath)
	baseName := strings.TrimSuffix(filepath.Base(basePath), ext)
	dir := filepath.Dir(basePath)

	for i := 0; i < targetFiles; i++ {
		startIdx := i * sectionsPerFile
		endIdx := startIdx + sectionsPerFile
		if i == targetFiles-1 {
			endIdx = len(sections) // Last file gets remaining sections
		}

		if startIdx >= len(sections) {
			break
		}

		var sectionNames []string
		for j := startIdx; j < endIdx && j < len(sections); j++ {
			sectionNames = append(sectionNames, sections[j].Name)
		}

		fileNum := i + 1
		newPath := filepath.Join(dir, fmt.Sprintf("%s_part%d%s", baseName, fileNum, ext))

		proposedFiles = append(proposedFiles, ProposedFile{
			Path:     newPath,
			Lines:    calculateTotalLines(sections[startIdx:min(endIdx, len(sections))]),
			Contents: sectionNames,
		})
	}

	// Generate migration instructions
	migrationInstructions = []string{
		fmt.Sprintf("1. Create new files: %s", strings.Join(getProposedPaths(proposedFiles), ", ")),
		"2. Move sections to respective files as suggested",
		"3. Update imports in files that reference moved functions/classes",
		"4. Create index file to re-export public APIs if needed",
		"5. Update tests to reference new file locations",
		"6. Run tests to verify functionality",
		"7. Remove original file after verification",
	}

	estimatedEffort := "Medium"
	if len(sections) > 10 {
		estimatedEffort = "High"
	}

	return &SplitSuggestion{
		Reason:                fmt.Sprintf("File has %d lines and %d logical sections. Splitting into %d files will improve maintainability.", len(strings.Split(file.Content, "\n")), len(sections), len(proposedFiles)),
		ProposedFiles:         proposedFiles,
		MigrationInstructions: migrationInstructions,
		EstimatedEffort:       estimatedEffort,
	}
}

// Helper functions
func getProposedPaths(files []ProposedFile) []string {
	var paths []string
	for _, f := range files {
		paths = append(paths, f.Path)
	}
	return paths
}

func calculateTotalLines(sections []FileSection) int {
	total := 0
	for _, s := range sections {
		total += s.Lines
	}
	return total
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// detectModuleType detects the type of module based on file path
