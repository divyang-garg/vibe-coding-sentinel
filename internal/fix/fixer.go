// Package fix provides automatic code fixing functionality
// Complies with CODING_STANDARDS.md: Business services max 400 lines
package fix

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/divyang-garg/sentinel-hub-api/internal/config"
)

// FixOptions configures fixing behavior
type FixOptions struct {
	TargetPath string
	DryRun     bool
	Force      bool
	Pattern    string // Specific pattern to apply: "console", "debugger", "imports", "whitespace"
}

// Result represents the result of a fix operation
type Result struct {
	FixesApplied   int
	FilesModified  int
	BackupsCreated int
}

// Fix performs automatic fixes on the codebase
func Fix(opts FixOptions) (*Result, error) {
	result := &Result{}

	if opts.TargetPath == "" {
		opts.TargetPath = "."
	}

	// Create backup directory
	backupDir := ".sentinel/backups"
	if !opts.DryRun {
		if err := os.MkdirAll(backupDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create backup directory: %w", err)
		}
	}

	// Walk through files
	err := filepath.Walk(opts.TargetPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		// Skip certain directories
		if shouldSkipPath(path) {
			return nil
		}

		// Only process code files
		ext := filepath.Ext(path)
		if !isCodeFile(ext) {
			return nil
		}

		// Read file
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		originalContent := string(content)
		modifiedContent := originalContent
		modified := false

		// Apply fixes based on pattern filter
		applyAll := opts.Pattern == ""

		if applyAll || opts.Pattern == "console" {
			modifiedContent, modified = removeConsoleLogs(modifiedContent, path, &result.FixesApplied, modified)
		}
		if applyAll || opts.Pattern == "debugger" {
			modifiedContent, modified = removeDebugger(modifiedContent, path, &result.FixesApplied, modified)
		}
		if applyAll || opts.Pattern == "whitespace" {
			modifiedContent, modified = removeTrailingWhitespace(modifiedContent, path, &result.FixesApplied, modified)
		}
		if applyAll || opts.Pattern == "imports" {
			modifiedContent, modified = sortImports(modifiedContent, path, &result.FixesApplied, modified)
		}

		// Write file if modified and not dry-run
		if modified && !opts.DryRun {
			// Create backup
			backupPath := filepath.Join(backupDir, fmt.Sprintf("%s_%d.backup", filepath.Base(path), time.Now().Unix()))
			if err := config.WriteFile(backupPath, originalContent); err != nil {
				return fmt.Errorf("failed to create backup: %w", err)
			}
			result.BackupsCreated++

			// Write modified content
			if err := config.WriteFile(path, modifiedContent); err != nil {
				return fmt.Errorf("failed to write file: %w", err)
			}
			result.FilesModified++
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Record fix history
	if !opts.DryRun && result.FixesApplied > 0 {
		if err := recordFixHistory(result, opts.TargetPath); err != nil {
			// Log error but don't fail
			fmt.Fprintf(os.Stderr, "Warning: failed to record fix history: %v\n", err)
		}
	}

	return result, nil
}

// shouldSkipPath determines if a path should be skipped
func shouldSkipPath(path string) bool {
	skipDirs := []string{
		"/node_modules/", "/.git/", "/.sentinel/", "/vendor/",
		"/build/", "/dist/", "/target/", "/bin/", "/obj/",
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
		".py": true, ".java": true, ".go": true, ".cs": true,
		".php": true, ".rb": true,
	}
	return codeExts[ext]
}

// removeConsoleLogs removes console.log statements
func removeConsoleLogs(content, path string, fixesApplied *int, modified bool) (string, bool) {
	consoleLogRegex := regexp.MustCompile(`^\s*console\.(log|debug|info|warn|error)\(.*?\);\s*$`)
	if !strings.Contains(content, "console.") {
		return content, modified
	}

	lines := strings.Split(content, "\n")
	var newLines []string
	fileModified := modified

	for _, line := range lines {
		if !consoleLogRegex.MatchString(line) {
			newLines = append(newLines, line)
		} else {
			fmt.Printf("  Remove console.log from %s\n", path)
			fileModified = true
			(*fixesApplied)++
		}
	}

	return strings.Join(newLines, "\n"), fileModified
}

// removeDebugger removes debugger statements
func removeDebugger(content, path string, fixesApplied *int, modified bool) (string, bool) {
	debuggerRegex := regexp.MustCompile(`^\s*debugger;\s*$`)
	if !strings.Contains(content, "debugger") {
		return content, modified
	}

	lines := strings.Split(content, "\n")
	var newLines []string
	fileModified := modified

	for _, line := range lines {
		if !debuggerRegex.MatchString(line) {
			newLines = append(newLines, line)
		} else {
			fmt.Printf("  Remove debugger from %s\n", path)
			fileModified = true
			(*fixesApplied)++
		}
	}

	return strings.Join(newLines, "\n"), fileModified
}

// removeTrailingWhitespace removes trailing whitespace
func removeTrailingWhitespace(content, path string, fixesApplied *int, modified bool) (string, bool) {
	lines := strings.Split(content, "\n")
	var newLines []string
	fileModified := modified

	for _, line := range lines {
		trimmed := strings.TrimRight(line, " \t")
		if trimmed != line {
			fileModified = true
		}
		newLines = append(newLines, trimmed)
	}

	if fileModified && !modified {
		fmt.Printf("  Clean trailing whitespace from %s\n", path)
		(*fixesApplied)++
	}

	return strings.Join(newLines, "\n"), fileModified
}
