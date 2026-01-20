// Task Completion - Verification Functions
// Schedules and performs task verification based on triggers
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

// ScheduleVerification schedules task verification based on trigger
func ScheduleVerification(ctx context.Context, trigger string, projectID string, codebasePath string) error {
	switch trigger {
	case "on_commit":
		// Verify tasks in changed files
		return verifyTasksOnCommit(ctx, projectID, codebasePath)
	case "on_push":
		// Verify all pending tasks
		return verifyAllPendingTasks(ctx, projectID, codebasePath)
	case "manual":
		// Manual verification - no-op, handled by endpoint
		return nil
	case "scheduled":
		// Scheduled verification - verify all tasks
		return verifyAllTasks(ctx, projectID, codebasePath)
	default:
		return fmt.Errorf("unknown trigger: %s", trigger)
	}
}

// verifyTasksOnCommit verifies tasks related to changed files
func verifyTasksOnCommit(ctx context.Context, projectID string, codebasePath string) error {
	// Get changed files from git
	changedFiles, err := getChangedFilesFromGit(codebasePath)
	if err != nil {
		LogError(ctx, "Failed to get changed files from git: %v", err)
		// Fallback to verifying all pending tasks
		return verifyAllPendingTasks(ctx, projectID, codebasePath)
	}

	if len(changedFiles) == 0 {
		// No changed files, nothing to verify
		return nil
	}

	// Get all pending tasks
	req := ListTasksRequest{
		StatusFilter: "pending",
		Limit:        1000,
		Offset:       0,
	}

	response, err := ListTasks(ctx, projectID, req)
	if err != nil {
		return fmt.Errorf("failed to list tasks: %w", err)
	}

	// Verify only tasks related to changed files
	changedFilesMap := make(map[string]bool)
	for _, file := range changedFiles {
		// Normalize paths for comparison
		normalized := filepath.Clean(file)
		changedFilesMap[normalized] = true
		// Also check without leading ./
		if strings.HasPrefix(normalized, "./") {
			changedFilesMap[normalized[2:]] = true
		}
	}

	for _, task := range response.Tasks {
		// Check if task's file path matches any changed file
		shouldVerify := false
		if task.FilePath != "" {
			taskPath := filepath.Clean(task.FilePath)
			if changedFilesMap[taskPath] {
				shouldVerify = true
			} else {
				// Check if any changed file is in the same directory or matches pattern
				for changedFile := range changedFilesMap {
					if strings.Contains(taskPath, changedFile) || strings.Contains(changedFile, taskPath) {
						shouldVerify = true
						break
					}
				}
			}
		} else {
			// If no file path, verify based on keywords in changed files
			keywords := extractKeywords(task.Title + " " + task.Description)
			for _, changedFile := range changedFiles {
				fullPath := filepath.Join(codebasePath, changedFile)
				content, err := os.ReadFile(fullPath)
				if err == nil {
					contentStr := strings.ToLower(string(content))
					for _, keyword := range keywords {
						if strings.Contains(contentStr, strings.ToLower(keyword)) {
							shouldVerify = true
							break
						}
					}
					if shouldVerify {
						break
					}
				}
			}
		}

		if shouldVerify {
			_, err := VerifyTask(ctx, task.ID, codebasePath, false)
			if err != nil {
				LogError(ctx, "Failed to verify task %s: %v", task.ID, err)
				continue
			}
		}
	}

	return nil
}

// getChangedFilesFromGit gets list of changed files from git
func getChangedFilesFromGit(codebasePath string) ([]string, error) {
	// Execute git diff to get changed files
	cmd := exec.Command("git", "diff", "--name-only", "HEAD")
	cmd.Dir = codebasePath

	output, err := cmd.Output()
	if err != nil {
		// If git command fails, try git diff --cached for staged files
		cmd = exec.Command("git", "diff", "--cached", "--name-only")
		cmd.Dir = codebasePath
		output, err = cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("failed to get changed files from git: %w", err)
		}
	}

	// Parse output
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var changedFiles []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			changedFiles = append(changedFiles, line)
		}
	}

	return changedFiles, nil
}

// verifyAllPendingTasks verifies all pending tasks
func verifyAllPendingTasks(ctx context.Context, projectID string, codebasePath string) error {
	req := ListTasksRequest{
		StatusFilter: "pending",
		Limit:        1000,
		Offset:       0,
	}

	response, err := ListTasks(ctx, projectID, req)
	if err != nil {
		return fmt.Errorf("failed to list tasks: %w", err)
	}

	for _, task := range response.Tasks {
		_, err := VerifyTask(ctx, task.ID, codebasePath, false)
		if err != nil {
			LogError(ctx, "Failed to verify task %s: %v", task.ID, err)
			continue
		}
	}

	return nil
}

// verifyAllTasks verifies all tasks
func verifyAllTasks(ctx context.Context, projectID string, codebasePath string) error {
	req := ListTasksRequest{
		Limit:  1000,
		Offset: 0,
	}

	response, err := ListTasks(ctx, projectID, req)
	if err != nil {
		return fmt.Errorf("failed to list tasks: %w", err)
	}

	for _, task := range response.Tasks {
		if task.Status == "completed" {
			continue // Skip completed tasks
		}

		_, err := VerifyTask(ctx, task.ID, codebasePath, false)
		if err != nil {
			LogError(ctx, "Failed to verify task %s: %v", task.ID, err)
			continue
		}
	}

	return nil
}
