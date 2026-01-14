// Phase 14E: Task Detection Engine
// Detects tasks from codebase (TODO comments, Cursor markers, explicit format)

package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// DetectedTask represents a task found during scanning
type DetectedTask struct {
	Source      string
	Title       string
	Description string
	FilePath    string
	LineNumber  int
	Priority    string
	Tags        []string
}

// TaskDetector handles task detection from codebase
type TaskDetector struct {
	projectID    string
	codebasePath string
}

// NewTaskDetector creates a new task detector
func NewTaskDetector(projectID, codebasePath string) *TaskDetector {
	return &TaskDetector{
		projectID:    projectID,
		codebasePath: codebasePath,
	}
}

// DetectTasks scans the codebase for tasks
func (td *TaskDetector) DetectTasks(ctx context.Context) ([]DetectedTask, error) {
	var tasks []DetectedTask

	// Patterns for different task types
	todoPattern := regexp.MustCompile(`(?i)(?:TODO|FIXME|NOTE|HACK|XXX|BUG):\s*(.+?)(?:\n|$)`)
	cursorTaskPattern := regexp.MustCompile(`(?i)-?\s*\[([ x])\]\s*(?:Task:)?\s*(.+?)(?:\n|$)`)
	explicitTaskPattern := regexp.MustCompile(`(?i)(?:TASK|TASK-?\d+):\s*(.+?)(?:\n|$)`)
	dependsPattern := regexp.MustCompile(`(?i)DEPENDS:\s*(.+?)(?:\n|$)`)

	// File extensions to scan
	extensions := map[string]bool{
		".go": true, ".js": true, ".ts": true, ".jsx": true, ".tsx": true,
		".py": true, ".java": true, ".cpp": true, ".c": true, ".h": true,
		".cs": true, ".rb": true, ".php": true, ".swift": true, ".kt": true,
		".md": true, ".txt": true,
	}

	err := filepath.Walk(td.codebasePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we can't read
		}

		// Skip hidden directories and common ignore patterns
		if info.IsDir() {
			if strings.HasPrefix(info.Name(), ".") && info.Name() != "." {
				return filepath.SkipDir
			}
			// Skip common directories
			if info.Name() == "node_modules" || info.Name() == "vendor" ||
				info.Name() == ".git" || info.Name() == "build" || info.Name() == "dist" {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if file extension is supported
		ext := filepath.Ext(path)
		if !extensions[ext] {
			return nil
		}

		// Read file content
		content, err := os.ReadFile(path)
		if err != nil {
			return nil // Skip files we can't read
		}

		lines := strings.Split(string(content), "\n")
		relativePath, _ := filepath.Rel(td.codebasePath, path)

		// Scan for TODO/FIXME comments
		for i, line := range lines {
			lineNum := i + 1

			// TODO/FIXME pattern
			if matches := todoPattern.FindStringSubmatch(line); len(matches) > 1 {
				task := DetectedTask{
					Source:      "cursor",
					Title:       strings.TrimSpace(matches[1]),
					Description: "",
					FilePath:    relativePath,
					LineNumber:  lineNum,
					Priority:    td.detectPriority(matches[1]),
					Tags:        td.extractTags(matches[1]),
				}
				tasks = append(tasks, task)
			}

			// Cursor task marker pattern
			if matches := cursorTaskPattern.FindStringSubmatch(line); len(matches) > 2 {
				checked := strings.TrimSpace(matches[1])
				if checked == " " || checked == "x" {
					// Only process unchecked tasks
					if checked == " " {
						task := DetectedTask{
							Source:      "cursor",
							Title:       strings.TrimSpace(matches[2]),
							Description: "",
							FilePath:    relativePath,
							LineNumber:  lineNum,
							Priority:    td.detectPriority(matches[2]),
							Tags:        td.extractTags(matches[2]),
						}
						tasks = append(tasks, task)
					}
				}
			}

			// Explicit task format
			if matches := explicitTaskPattern.FindStringSubmatch(line); len(matches) > 1 {
				task := DetectedTask{
					Source:      "cursor",
					Title:       strings.TrimSpace(matches[1]),
					Description: "",
					FilePath:    relativePath,
					LineNumber:  lineNum,
					Priority:    td.detectPriority(matches[1]),
					Tags:        td.extractTags(matches[1]),
				}
				// Check for DEPENDS on next lines
				if i+1 < len(lines) {
					if depMatches := dependsPattern.FindStringSubmatch(lines[i+1]); len(depMatches) > 1 {
						// Dependencies will be handled during dependency detection
						task.Description = fmt.Sprintf("Depends on: %s", strings.TrimSpace(depMatches[1]))
					}
				}
				tasks = append(tasks, task)
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan codebase: %w", err)
	}

	return tasks, nil
}

// detectPriority extracts priority from task text
func (td *TaskDetector) detectPriority(text string) string {
	textLower := strings.ToLower(text)
	if strings.Contains(textLower, "critical") || strings.Contains(textLower, "urgent") {
		return "critical"
	}
	if strings.Contains(textLower, "high") || strings.Contains(textLower, "important") {
		return "high"
	}
	if strings.Contains(textLower, "low") || strings.Contains(textLower, "minor") {
		return "low"
	}
	return "medium"
}

// extractTags extracts tags from task text
func (td *TaskDetector) extractTags(text string) []string {
	var tags []string
	tagPattern := regexp.MustCompile(`#(\w+)`)
	matches := tagPattern.FindAllStringSubmatch(text, -1)
	for _, match := range matches {
		if len(match) > 1 {
			tags = append(tags, match[1])
		}
	}
	return tags
}

// DeduplicateTasks removes duplicate tasks based on title and file path
func DeduplicateTasks(ctx context.Context, projectID string, tasks []DetectedTask) ([]DetectedTask, error) {
	// Check existing tasks in database
	query := `
		SELECT title, file_path, line_number 
		FROM tasks 
		WHERE project_id = $1 AND status != 'completed'
	`
	rows, err := queryWithTimeout(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to query existing tasks: %w", err)
	}
	defer rows.Close()

	existingTasks := make(map[string]bool)
	for rows.Next() {
		var title, filePath string
		var lineNumber sql.NullInt64
		if err := rows.Scan(&title, &filePath, &lineNumber); err != nil {
			continue
		}
		key := fmt.Sprintf("%s:%s:%d", title, filePath, lineNumber.Int64)
		existingTasks[key] = true
	}

	// Filter out duplicates
	var uniqueTasks []DetectedTask
	seen := make(map[string]bool)

	for _, task := range tasks {
		key := fmt.Sprintf("%s:%s:%d", task.Title, task.FilePath, task.LineNumber)

		// Skip if already in database
		if existingTasks[key] {
			continue
		}

		// Skip if already seen in this batch
		if seen[key] {
			continue
		}

		seen[key] = true
		uniqueTasks = append(uniqueTasks, task)
	}

	return uniqueTasks, nil
}

// StoreDetectedTasks stores detected tasks in the database
func StoreDetectedTasks(ctx context.Context, projectID string, tasks []DetectedTask) ([]string, error) {
	var taskIDs []string

	for _, task := range tasks {
		taskID := uuid.New().String()

		query := `
			INSERT INTO tasks (
				id, project_id, source, title, description, file_path, 
				line_number, status, priority, tags, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
			ON CONFLICT DO NOTHING
		`

		now := time.Now()
		var lineNum *int
		if task.LineNumber > 0 {
			lineNum = &task.LineNumber
		}

		_, err := execWithTimeout(ctx, query,
			taskID, projectID, task.Source, task.Title, task.Description,
			task.FilePath, lineNum, "pending", task.Priority,
			pq.Array(task.Tags), now, now,
		)

		if err != nil {
			LogError(ctx, "Failed to store task %s: %v", task.Title, err)
			continue
		}

		taskIDs = append(taskIDs, taskID)
	}

	return taskIDs, nil
}
