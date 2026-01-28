// Dependency Detection Analysis Functions
// Implements detection logic for explicit, implicit, integration, and feature dependencies
// Complies with CODING_STANDARDS.md: Business Services max 400 lines

package services

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"sentinel-hub-api/pkg/database"

	"github.com/google/uuid"
)

// detectExplicitDependencies parses explicit dependencies from task description
func detectExplicitDependencies(ctx context.Context, task *Task) ([]TaskDependency, error) {
	var dependencies []TaskDependency

	// Pattern: "DEPENDS: TASK-123, TASK-456" or "Depends on: TASK-123"
	dependsPattern := regexp.MustCompile(`(?i)(?:depends|dependency)[\s:]+(?:on|upon)?[\s:]*([A-Z0-9-,\s]+)`)

	text := task.Title + " " + task.Description
	matches := dependsPattern.FindStringSubmatch(text)

	if len(matches) > 1 {
		depList := matches[1]
		// Split by comma and clean up
		taskRefs := strings.Split(depList, ",")

		for _, ref := range taskRefs {
			ref = strings.TrimSpace(ref)
			if ref == "" {
				continue
			}

			// Try to find task by ID or title
			depTaskID, err := findTaskByReference(ctx, task.ProjectID, ref)
			if err != nil {
				LogError(ctx, "Failed to find task %s: %v", ref, err)
				continue
			}

			if depTaskID == "" {
				continue
			}

			dependency := TaskDependency{
				ID:              uuid.New().String(),
				TaskID:          task.ID,
				DependsOnTaskID: depTaskID,
				DependencyType:  "explicit",
				Confidence:      0.95, // High confidence for explicit dependencies
			}

			dependencies = append(dependencies, dependency)
		}
	}

	return dependencies, nil
}

// detectImplicitDependencies detects dependencies through code analysis
func detectImplicitDependencies(ctx context.Context, task *Task, codebasePath string) ([]TaskDependency, error) {
	var dependencies []TaskDependency

	// Extract keywords from task
	keywords := extractKeywords(task.Title + " " + task.Description)
	if len(keywords) == 0 {
		return dependencies, nil
	}

	// Get all tasks in the project
	allTasks, err := getAllProjectTasks(ctx, task.ProjectID)
	if err != nil {
		return dependencies, fmt.Errorf("failed to get project tasks: %w", err)
	}

	// For each other task, check if it's related
	for _, otherTask := range allTasks {
		if otherTask.ID == task.ID {
			continue
		}

		// Check if keywords overlap
		otherKeywords := extractKeywords(otherTask.Title + " " + otherTask.Description)
		overlap := calculateKeywordOverlap(keywords, otherKeywords)

		if overlap > 0.3 { // 30% keyword overlap suggests dependency
			// Check if other task's code is referenced in this task's file
			hasCodeReference := false
			if task.FilePath != "" {
				hasCodeReference = checkCodeReference(codebasePath, task.FilePath, &otherTask)
			}

			confidence := overlap
			if hasCodeReference {
				confidence = 0.7 // Higher confidence if code reference found
			}

			if confidence > 0.3 {
				dependency := TaskDependency{
					ID:              uuid.New().String(),
					TaskID:          task.ID,
					DependsOnTaskID: otherTask.ID,
					DependencyType:  "implicit",
					Confidence:      confidence,
				}
				dependencies = append(dependencies, dependency)
			}
		}
	}

	return dependencies, nil
}

// detectIntegrationDependencies detects integration dependencies
func detectIntegrationDependencies(ctx context.Context, task *Task, codebasePath string) ([]TaskDependency, error) {
	var dependencies []TaskDependency

	// Validate codebasePath for security (graceful degradation if invalid)
	if codebasePath != "" {
		if err := ValidateDirectory(codebasePath); err != nil {
			// Log warning but continue with database-only detection
			LogWarn(ctx, "Invalid codebase path for integration dependency detection: %v", err)
		}
	}

	// Check if task mentions integration keywords
	integrationKeywords := []string{"api", "integration", "service", "external", "third-party", "sdk", "client"}
	taskText := strings.ToLower(task.Title + " " + task.Description)

	hasIntegrationKeyword := false
	for _, keyword := range integrationKeywords {
		if strings.Contains(taskText, keyword) {
			hasIntegrationKeyword = true
			break
		}
	}

	// If task has FilePath, validate it's within codebase and check for integration patterns
	hasCodebaseIntegrationEvidence := false
	if codebasePath != "" && task.FilePath != "" {
		// Validate path is within codebase (security check)
		fullPath := filepath.Join(codebasePath, task.FilePath)
		relPath, err := filepath.Rel(codebasePath, fullPath)
		if err == nil && !strings.HasPrefix(relPath, "..") {
			// Path is valid, check if file exists and contains integration patterns
			if content, err := os.ReadFile(fullPath); err == nil {
				contentStr := strings.ToLower(string(content))
				for _, keyword := range integrationKeywords {
					if strings.Contains(contentStr, keyword) {
						hasCodebaseIntegrationEvidence = true
						break
					}
				}
			}
		}
	}

	// If no keyword match and no codebase evidence, return early
	if !hasIntegrationKeyword && !hasCodebaseIntegrationEvidence {
		return dependencies, nil
	}

	// Query comprehensive analysis for integration-related features
	query := `
		SELECT DISTINCT validation_id, feature
		FROM comprehensive_validations
		WHERE project_id = $1 
		AND (
			LOWER(feature) LIKE ANY(ARRAY['%api%', '%integration%', '%service%', '%external%', '%sdk%', '%client%'])
			OR LOWER(findings::text) LIKE ANY(ARRAY['%api%', '%integration%', '%service%', '%external%', '%sdk%', '%client%'])
		)
	`

	rows, err := database.QueryWithTimeout(ctx, db, query, task.ProjectID)
	if err != nil {
		// If table doesn't exist or query fails, return empty (graceful degradation)
		return dependencies, nil
	}
	defer rows.Close()

	validationFeatures := make(map[string]string) // validation_id -> feature
	for rows.Next() {
		var validationID, feature string
		if err := rows.Scan(&validationID, &feature); err == nil {
			validationFeatures[validationID] = feature
		}
	}

	// Check if task keywords match any integration features
	keywords := extractKeywords(task.Title + " " + task.Description)
	for validationID, feature := range validationFeatures {
		featureKeywords := extractKeywords(feature)
		overlap := calculateKeywordOverlap(keywords, featureKeywords)

		// Boost confidence if codebase evidence exists
		if hasCodebaseIntegrationEvidence {
			// Increase confidence by 20%, cap at 1.0
			overlap = overlap * 1.2
			if overlap > 1.0 {
				overlap = 1.0
			}
		}

		if overlap > 0.3 {
			// Find tasks linked to this comprehensive analysis
			linkQuery := `
				SELECT task_id FROM task_links
				WHERE link_type = 'comprehensive_analysis' AND linked_id = $1
			`
			linkRows, err := database.QueryWithTimeout(ctx, db, linkQuery, validationID)
			if err == nil {
				defer linkRows.Close()
				for linkRows.Next() {
					var depTaskID string
					if err := linkRows.Scan(&depTaskID); err == nil && depTaskID != task.ID {
						dependency := TaskDependency{
							ID:              uuid.New().String(),
							TaskID:          task.ID,
							DependsOnTaskID: depTaskID,
							DependencyType:  "integration",
							Confidence:      overlap,
						}
						dependencies = append(dependencies, dependency)
					}
				}
			}
		}
	}

	return dependencies, nil
}

// detectFeatureDependencies detects feature-level dependencies
func detectFeatureDependencies(ctx context.Context, task *Task, codebasePath string) ([]TaskDependency, error) {
	var dependencies []TaskDependency

	// Validate codebasePath for security (graceful degradation if invalid)
	if codebasePath != "" {
		if err := ValidateDirectory(codebasePath); err != nil {
			// Log warning but continue with database-only detection
			LogWarn(ctx, "Invalid codebase path for feature dependency detection: %v", err)
		}
	}

	// Query comprehensive analysis for feature dependencies
	query := `
		SELECT validation_id, feature, checklist
		FROM comprehensive_validations
		WHERE project_id = $1
	`

	rows, err := database.QueryWithTimeout(ctx, db, query, task.ProjectID)
	if err != nil {
		// If table doesn't exist or query fails, return empty (graceful degradation)
		return dependencies, nil
	}
	defer rows.Close()

	type FeatureInfo struct {
		ValidationID string
		Feature      string
		Checklist    string
	}

	var features []FeatureInfo
	for rows.Next() {
		var fi FeatureInfo
		var checklist sql.NullString
		if err := rows.Scan(&fi.ValidationID, &fi.Feature, &checklist); err == nil {
			if checklist.Valid {
				fi.Checklist = checklist.String
			}
			features = append(features, fi)
		}
	}

	// Extract keywords from task
	keywords := extractKeywords(task.Title + " " + task.Description)

	// If task has FilePath, validate and check for feature patterns in codebase
	hasCodebaseFeatureEvidence := false
	var codebaseFeatureKeywords []string
	if codebasePath != "" && task.FilePath != "" {
		// Validate path is within codebase (security check)
		fullPath := filepath.Join(codebasePath, task.FilePath)
		relPath, err := filepath.Rel(codebasePath, fullPath)
		if err == nil && !strings.HasPrefix(relPath, "..") {
			// Path is valid, check if file exists
			if content, err := os.ReadFile(fullPath); err == nil {
				// Extract feature-related keywords from code
				codebaseFeatureKeywords = extractKeywords(string(content))
				hasCodebaseFeatureEvidence = len(codebaseFeatureKeywords) > 0
			}
		}
	}

	// Find matching features
	for _, feature := range features {
		featureKeywords := extractKeywords(feature.Feature + " " + feature.Checklist)

		// Combine task keywords with codebase keywords if available
		allKeywords := keywords
		if hasCodebaseFeatureEvidence {
			allKeywords = append(allKeywords, codebaseFeatureKeywords...)
		}

		overlap := calculateKeywordOverlap(allKeywords, featureKeywords)

		// Boost confidence if codebase evidence found
		if hasCodebaseFeatureEvidence {
			// Increase confidence by 15%, cap at 1.0
			overlap = overlap * 1.15
			if overlap > 1.0 {
				overlap = 1.0
			}
		}

		if overlap > 0.3 {
			// Find tasks linked to this comprehensive analysis
			linkQuery := `
				SELECT task_id FROM task_links
				WHERE link_type = 'comprehensive_analysis' AND linked_id = $1
			`
			linkRows, err := database.QueryWithTimeout(ctx, db, linkQuery, feature.ValidationID)
			if err == nil {
				defer linkRows.Close()
				for linkRows.Next() {
					var depTaskID string
					if err := linkRows.Scan(&depTaskID); err == nil && depTaskID != task.ID {
						dependency := TaskDependency{
							ID:              uuid.New().String(),
							TaskID:          task.ID,
							DependsOnTaskID: depTaskID,
							DependencyType:  "feature",
							Confidence:      overlap,
						}
						dependencies = append(dependencies, dependency)
					}
				}
			}
		}
	}

	return dependencies, nil
}
