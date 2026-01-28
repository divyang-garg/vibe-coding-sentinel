// Package services helper functions for code analysis
// Provides filesystem and git integration helpers for intent analysis
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package services

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// fileModTime represents a file with its modification time
type fileModTime struct {
	path    string
	modTime time.Time
}

// extractRecentFiles scans the filesystem for recently modified files
// Returns a list of file paths modified within the last 24 hours, sorted by modification time (newest first)
func extractRecentFiles(codebasePath string) []string {
	if codebasePath == "" {
		return []string{}
	}

	// Validate path exists
	if _, err := os.Stat(codebasePath); err != nil {
		return []string{}
	}

	var recentFiles []fileModTime
	cutoffTime := time.Now().Add(-24 * time.Hour)

	// Walk the directory tree
	err := filepath.Walk(codebasePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Skip all errors (permission errors and others) and continue walking
			return nil
		}

		// Skip directories
		if info.IsDir() {
			// Skip hidden directories and common ignore patterns
			if strings.HasPrefix(info.Name(), ".") {
				return filepath.SkipDir
			}
			// Skip common build/dependency directories
			if info.Name() == "node_modules" || info.Name() == "vendor" || info.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if file was modified recently
		if info.ModTime().After(cutoffTime) {
			// Only include code files
			ext := strings.ToLower(filepath.Ext(path))
			codeExts := map[string]bool{
				".go": true, ".js": true, ".ts": true, ".tsx": true,
				".py": true, ".java": true, ".cpp": true, ".c": true,
				".cs": true, ".rb": true, ".php": true, ".swift": true,
			}
			if codeExts[ext] {
				recentFiles = append(recentFiles, fileModTime{
					path:    path,
					modTime: info.ModTime(),
				})
			}
		}

		return nil
	})

	if err != nil {
		return []string{}
	}

	// Sort by modification time (newest first)
	sort.Slice(recentFiles, func(i, j int) bool {
		return recentFiles[i].modTime.After(recentFiles[j].modTime)
	})

	// Limit to 50 most recent files
	maxFiles := 50
	if len(recentFiles) > maxFiles {
		recentFiles = recentFiles[:maxFiles]
	}

	// Extract just the paths
	result := make([]string, len(recentFiles))
	for i, f := range recentFiles {
		result[i] = f.path
	}

	return result
}

// extractGitStatus extracts git status information from the codebase
// Returns a map containing git status, branch, commit info, and modified files
func extractGitStatus(codebasePath string) map[string]interface{} {
	result := make(map[string]interface{})

	if codebasePath == "" {
		return result
	}

	// Check if .git directory exists
	gitDir := filepath.Join(codebasePath, ".git")
	if _, err := os.Stat(gitDir); err != nil {
		// Not a git repository
		result["is_git_repo"] = false
		return result
	}

	result["is_git_repo"] = true

	// Get current branch
	if branch, err := runGitCommand(codebasePath, "rev-parse", "--abbrev-ref", "HEAD"); err == nil {
		result["branch"] = strings.TrimSpace(branch)
	}

	// Get latest commit hash
	if commit, err := runGitCommand(codebasePath, "rev-parse", "HEAD"); err == nil {
		result["commit"] = strings.TrimSpace(commit)
	}

	// Get commit message
	if message, err := runGitCommand(codebasePath, "log", "-1", "--pretty=format:%s"); err == nil {
		result["last_commit_message"] = strings.TrimSpace(message)
	}

	// Get commit date
	if date, err := runGitCommand(codebasePath, "log", "-1", "--pretty=format:%ci"); err == nil {
		result["last_commit_date"] = strings.TrimSpace(date)
	}

	// Get status summary
	if status, err := runGitCommand(codebasePath, "status", "--short"); err == nil {
		statusLines := strings.Split(strings.TrimSpace(status), "\n")
		var modified, added, deleted, untracked []string

		for _, line := range statusLines {
			if len(line) < 3 {
				continue
			}
			statusCode := line[:2]
			file := strings.TrimSpace(line[2:])

			switch {
			case strings.HasPrefix(statusCode, "??"):
				untracked = append(untracked, file)
			case strings.Contains(statusCode, "M"):
				modified = append(modified, file)
			case strings.Contains(statusCode, "A"):
				added = append(added, file)
			case strings.Contains(statusCode, "D"):
				deleted = append(deleted, file)
			}
		}

		result["modified_files"] = modified
		result["added_files"] = added
		result["deleted_files"] = deleted
		result["untracked_files"] = untracked
		result["has_changes"] = len(modified) > 0 || len(added) > 0 || len(deleted) > 0 || len(untracked) > 0
	}

	// Get remote URL if available
	if remote, err := runGitCommand(codebasePath, "config", "--get", "remote.origin.url"); err == nil {
		result["remote_url"] = strings.TrimSpace(remote)
	}

	return result
}

// runGitCommand executes a git command in the specified directory
func runGitCommand(dir, command string, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{command}, args...)...)
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git %s failed: %w", command, err)
	}
	return string(output), nil
}

// extractProjectStructure scans and extracts the project directory structure
// Returns a map containing directory tree, file counts, and language distribution
func extractProjectStructure(codebasePath string) map[string]interface{} {
	result := make(map[string]interface{})

	if codebasePath == "" {
		return result
	}

	// Validate path exists
	if _, err := os.Stat(codebasePath); err != nil {
		return result
	}

	structure := make(map[string]interface{})
	fileCounts := make(map[string]int)
	languageCounts := make(map[string]int)
	totalFiles := 0
	totalDirs := 0

	// Common code file extensions
	codeExts := map[string]string{
		".go":    "Go",
		".js":    "JavaScript",
		".ts":    "TypeScript",
		".tsx":   "TypeScript",
		".jsx":   "JavaScript",
		".py":    "Python",
		".java":  "Java",
		".cpp":   "C++",
		".c":     "C",
		".cs":    "C#",
		".rb":    "Ruby",
		".php":   "PHP",
		".swift": "Swift",
		".kt":    "Kotlin",
		".rs":    "Rust",
	}

	// Directories to skip
	skipDirs := map[string]bool{
		".git":         true,
		"node_modules": true,
		"vendor":       true,
		".vscode":      true,
		".idea":        true,
		"dist":         true,
		"build":        true,
		"target":       true,
		"bin":          true,
		"obj":          true,
	}

	err := filepath.Walk(codebasePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Skip all errors (permission errors and others) and continue walking
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(codebasePath, path)
		if err != nil {
			return nil
		}

		// Skip root directory
		if relPath == "." {
			return nil
		}

		// Skip hidden files and directories
		if strings.HasPrefix(relPath, ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if info.IsDir() {
			// Skip common build/dependency directories
			if skipDirs[info.Name()] {
				return filepath.SkipDir
			}
			totalDirs++
			return nil
		}

		// Count files by extension
		ext := strings.ToLower(filepath.Ext(path))
		if ext != "" {
			fileCounts[ext]++
			if lang, ok := codeExts[ext]; ok {
				languageCounts[lang]++
			}
		}
		totalFiles++

		return nil
	})

	if err != nil {
		return result
	}

	// Build top-level directory structure (limit depth for readability)
	topLevelDirs := make([]string, 0)
	err = filepath.Walk(codebasePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if !info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(codebasePath, path)
		if err != nil {
			return nil
		}

		// Only include top-level directories (depth 1)
		if relPath == "." {
			return nil
		}

		parts := strings.Split(relPath, string(filepath.Separator))
		if len(parts) == 1 {
			// Skip hidden and common directories
			if !strings.HasPrefix(parts[0], ".") && !skipDirs[parts[0]] {
				topLevelDirs = append(topLevelDirs, parts[0])
			}
			return filepath.SkipDir // Don't recurse into subdirectories
		}

		return nil
	})

	structure["top_level_directories"] = topLevelDirs
	structure["total_files"] = totalFiles
	structure["total_directories"] = totalDirs
	structure["file_extensions"] = fileCounts
	structure["languages"] = languageCounts

	// Determine primary language
	primaryLanguage := "Unknown"
	maxCount := 0
	for lang, count := range languageCounts {
		if count > maxCount {
			maxCount = count
			primaryLanguage = lang
		}
	}
	structure["primary_language"] = primaryLanguage

	result["structure"] = structure
	result["root_path"] = codebasePath

	return result
}

// calculateComplianceRate calculates the compliance rate for business rules
func calculateComplianceRate(rules []KnowledgeItem, findings []BusinessContextFinding) float64 {
	if len(rules) == 0 {
		return 0.0
	}
	nonCompliant := len(findings)
	compliant := len(rules) - nonCompliant
	return float64(compliant) / float64(len(rules))
}
