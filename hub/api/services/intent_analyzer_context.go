// Intent Analysis Context Gathering Functions
// Collects relevant context for intent analysis (files, git status, project structure)
// Complies with CODING_STANDARDS.md: Business Services max 400 lines

package services

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ContextData contains gathered context for intent analysis
type ContextData struct {
	RecentFiles      []string               `json:"recent_files"`
	RecentErrors     []string               `json:"recent_errors"`
	GitStatus        map[string]string      `json:"git_status"`
	ProjectStructure map[string]interface{} `json:"project_structure"`
	BusinessRules    []string               `json:"business_rules"`
	CodePatterns     []string               `json:"code_patterns"`
}

// GatherContext collects relevant context for intent analysis
func GatherContext(ctx context.Context, projectID string, codebasePath string) (*ContextData, error) {
	contextData := &ContextData{
		RecentFiles:      []string{},
		RecentErrors:     []string{},
		GitStatus:        make(map[string]string),
		ProjectStructure: make(map[string]interface{}),
		BusinessRules:    []string{},
		CodePatterns:     []string{},
	}

	// Gather recent files (from git or file system)
	if codebasePath != "" {
		recentFiles, err := getRecentFiles(ctx, codebasePath)
		if err == nil {
			contextData.RecentFiles = recentFiles
		} else {
			LogWarn(ctx, "Failed to gather recent files: %v", err)
		}
	}

	// Gather git status
	if codebasePath != "" {
		gitStatus, err := getGitStatus(ctx, codebasePath)
		if err == nil {
			contextData.GitStatus = gitStatus
		} else {
			LogWarn(ctx, "Failed to gather git status: %v", err)
		}
	}

	// Gather project structure
	if codebasePath != "" {
		projectStructure, err := getProjectStructure(ctx, codebasePath)
		if err == nil {
			contextData.ProjectStructure = projectStructure
		} else {
			LogWarn(ctx, "Failed to gather project structure: %v", err)
		}
	}

	// Gather business rules from knowledge_items
	if projectID != "" {
		rules, err := extractBusinessRules(ctx, projectID, nil, "", nil)
		if err == nil {
			businessRules := make([]string, 0, len(rules))
			for _, rule := range rules {
				businessRules = append(businessRules, rule.Title)
			}
			contextData.BusinessRules = businessRules
		} else {
			LogWarn(ctx, "Failed to gather business rules: %v", err)
		}
	}

	// Gather code patterns (from recent files)
	if len(contextData.RecentFiles) > 0 {
		patterns := extractCodePatterns(contextData.RecentFiles[:minInt(5, len(contextData.RecentFiles))])
		contextData.CodePatterns = patterns
	}

	return contextData, nil
}

// getRecentFiles gets recently modified files from git or file system
func getRecentFiles(ctx context.Context, codebasePath string) ([]string, error) {
	// Try git first
	cmd := exec.CommandContext(ctx, "git", "-C", codebasePath, "log", "--name-only", "--pretty=format:", "-10")
	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		files := make([]string, 0)
		seen := make(map[string]bool)
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && !seen[line] {
				files = append(files, line)
				seen[line] = true
			}
		}
		if len(files) > 0 {
			return files[:minInt(10, len(files))], nil
		}
	}

	// Fallback: get files from directory
	files := []string{}
	err = filepath.Walk(codebasePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if !info.IsDir() && isCodeFileForIntent(path) {
			files = append(files, path)
		}
		if len(files) >= 10 {
			return filepath.SkipDir
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files[:minInt(10, len(files))], nil
}

// getGitStatus gets git status information
func getGitStatus(ctx context.Context, codebasePath string) (map[string]string, error) {
	status := make(map[string]string)

	// Get current branch
	cmd := exec.CommandContext(ctx, "git", "-C", codebasePath, "branch", "--show-current")
	output, err := cmd.Output()
	if err == nil {
		status["branch"] = strings.TrimSpace(string(output))
	}

	// Get modified files count
	cmd = exec.CommandContext(ctx, "git", "-C", codebasePath, "status", "--porcelain")
	output, err = cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		modifiedCount := 0
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				modifiedCount++
			}
		}
		status["modified_files"] = fmt.Sprintf("%d", modifiedCount)
	}

	return status, nil
}

// getProjectStructure gets project structure information
func getProjectStructure(ctx context.Context, codebasePath string) (map[string]interface{}, error) {
	// Check context cancellation before starting
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	structure := make(map[string]interface{})
	dirs := []string{"src", "lib", "app", "components", "packages", "server", "client", "api", "routes"}

	foundDirs := []string{}
	for _, dir := range dirs {
		// Check context cancellation during directory checks
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		path := filepath.Join(codebasePath, dir)
		if _, err := os.Stat(path); err == nil {
			foundDirs = append(foundDirs, dir)
		}
	}
	structure["directories"] = foundDirs

	// Get file extensions
	extensions := make(map[string]int)
	err := filepath.Walk(codebasePath, func(path string, info os.FileInfo, err error) error {
		// Check context cancellation during file walk
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if err != nil || info.IsDir() {
			return nil
		}
		ext := filepath.Ext(path)
		if ext != "" {
			extensions[ext]++
		}
		return nil
	})
	if err == nil {
		structure["file_extensions"] = extensions
	} else if ctx.Err() != nil {
		// If error is due to context cancellation, return context error
		return nil, ctx.Err()
	}

	return structure, nil
}

// extractCodePatterns extracts code patterns from file paths
func extractCodePatterns(files []string) []string {
	patterns := []string{}
	seen := make(map[string]bool)

	for _, file := range files {
		dir := filepath.Dir(file)
		if dir != "." && !seen[dir] {
			patterns = append(patterns, dir)
			seen[dir] = true
		}
	}

	return patterns
}

// isCodeFileForIntent checks if a file is a code file (local version to avoid conflict)
func isCodeFileForIntent(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	codeExts := []string{".go", ".js", ".ts", ".jsx", ".tsx", ".py", ".java", ".kt", ".swift", ".rs", ".php", ".rb"}
	for _, codeExt := range codeExts {
		if ext == codeExt {
			return true
		}
	}
	return false
}

// minInt returns the minimum of two integers (local version to avoid conflict)
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
