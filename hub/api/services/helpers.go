// Package services helper functions
// Core utility functions for services package
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"sentinel-hub-api/utils"
)

// marshalJSONB marshals a value to JSON string for JSONB storage
func marshalJSONB(v interface{}) (string, error) {
	if v == nil {
		return "null", nil
	}
	data, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return string(data), nil
}

// unmarshalJSONB unmarshals a JSON string from JSONB storage
func unmarshalJSONB(data string, v interface{}) error {
	if data == "" || data == "null" {
		return nil // Empty or null JSONB
	}
	if err := json.Unmarshal([]byte(data), v); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	return nil
}

// ValidateUUID validates a UUID string
func ValidateUUID(id string) error {
	return utils.ValidateUUID(id)
}

// Package-level database connection (set during initialization)
var db *sql.DB

// SetDB sets the database connection for the services package
func SetDB(database *sql.DB) {
	db = database
}

// LogWarn logs a warning message (stub - in production would use proper logger)
func LogWarn(ctx context.Context, msg string, args ...interface{}) {
	fmt.Printf("WARN: "+msg+"\n", args...)
}

// LogError logs an error message (stub - in production would use proper logger)
func LogError(ctx context.Context, msg string, args ...interface{}) {
	fmt.Printf("ERROR: "+msg+"\n", args...)
}

// LogInfo logs an info message (stub - in production would use proper logger)
func LogInfo(ctx context.Context, msg string, args ...interface{}) {
	fmt.Printf("INFO: "+msg+"\n", args...)
}

// getQueryTimeout returns the default query timeout (stub)
func getQueryTimeout() time.Duration {
	return 30 * time.Second
}

// extractFeatureKeywords extracts keywords from feature name
func extractFeatureKeywords(featureName string) []string {
	var keywords []string
	words := []rune(featureName)
	var current []rune
	for _, r := range words {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			current = append(current, r)
		} else {
			if len(current) > 0 {
				keywords = append(keywords, string(current))
				current = nil
			}
		}
	}
	if len(current) > 0 {
		keywords = append(keywords, string(current))
	}
	return keywords
}

// extractKeywords extracts meaningful keywords from text
func extractKeywords(text string) []string {
	words := strings.FieldsFunc(text, func(r rune) bool {
		return r == ' ' || r == '-' || r == '_' || r == '(' || r == ')' || r == '[' || r == ']'
	})

	keywords := []string{}
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true, "but": true,
		"in": true, "on": true, "at": true, "to": true, "for": true, "of": true,
		"with": true, "by": true, "is": true, "are": true, "was": true, "were": true,
		"be": true, "been": true, "have": true, "has": true, "had": true,
		"do": true, "does": true, "did": true, "will": true, "would": true,
		"should": true, "could": true, "may": true, "might": true,
	}

	for _, word := range words {
		word = strings.ToLower(strings.TrimSpace(word))
		if len(word) > 2 && !stopWords[word] {
			keywords = append(keywords, word)
		}
	}

	return keywords
}

// GetTask retrieves a task by ID using direct database query
func GetTask(ctx context.Context, taskID string) (*Task, error) {
	if taskID == "" {
		return nil, fmt.Errorf("task ID is required")
	}
	if db == nil {
		return nil, fmt.Errorf("database not initialized - call SetDB first")
	}

	query := `
		SELECT id, project_id, source, title, description, file_path, line_number,
		       status, priority, assigned_to, estimated_effort, actual_effort,
		       verification_confidence, created_at, updated_at, completed_at, verified_at, archived_at, version
		FROM tasks WHERE id = $1`

	row := db.QueryRowContext(ctx, query, taskID)

	var task Task
	var completedAt, verifiedAt, archivedAt *time.Time

	err := row.Scan(
		&task.ID, &task.ProjectID, &task.Source, &task.Title, &task.Description,
		&task.FilePath, &task.LineNumber, &task.Status, &task.Priority, &task.AssignedTo,
		&task.EstimatedEffort, &task.ActualEffort, &task.VerificationConfidence,
		&task.CreatedAt, &task.UpdatedAt, &completedAt, &verifiedAt, &archivedAt, &task.Version,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("task not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	task.CompletedAt = completedAt
	task.VerifiedAt = verifiedAt
	task.ArchivedAt = archivedAt

	return &task, nil
}

// invalidateGapAnalysisCache invalidates the gap analysis cache for a project (stub)
func invalidateGapAnalysisCache(projectID string) {
	// Stub - cache invalidation would be implemented here
}

// ListTasks lists tasks for a project using direct database query
func ListTasks(ctx context.Context, projectID string, req ListTasksRequest) (*ListTasksResponse, error) {
	if db == nil {
		return nil, fmt.Errorf("database not initialized - call SetDB first")
	}

	// Build WHERE clause
	whereClause := "WHERE project_id = $1"
	args := []interface{}{projectID}
	argCount := 1

	// Apply filters
	status := req.StatusFilter
	if status == "" {
		status = req.Status
	}
	if status != "" {
		argCount++
		whereClause += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, status)
	}

	priority := req.PriorityFilter
	if priority == "" {
		priority = req.Priority
	}
	if priority != "" {
		argCount++
		whereClause += fmt.Sprintf(" AND priority = $%d", argCount)
		args = append(args, priority)
	}

	if !req.IncludeArchived {
		whereClause += " AND archived_at IS NULL"
	}

	// Count total
	countQuery := "SELECT COUNT(*) FROM tasks " + whereClause
	var total int
	err := db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count tasks: %w", err)
	}

	// Get paginated results
	if req.Limit <= 0 {
		req.Limit = 100
	}
	argCount++
	query := `
		SELECT id, project_id, source, title, description, file_path, line_number,
		       status, priority, assigned_to, estimated_effort, actual_effort,
		       verification_confidence, created_at, updated_at, completed_at, verified_at, archived_at, version
		FROM tasks ` + whereClause + `
		ORDER BY created_at DESC
		LIMIT $` + fmt.Sprintf("%d", argCount) + ` OFFSET $` + fmt.Sprintf("%d", argCount+1)
	args = append(args, req.Limit, req.Offset)

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks: %w", err)
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		var completedAt, verifiedAt, archivedAt *time.Time

		err := rows.Scan(
			&task.ID, &task.ProjectID, &task.Source, &task.Title, &task.Description,
			&task.FilePath, &task.LineNumber, &task.Status, &task.Priority, &task.AssignedTo,
			&task.EstimatedEffort, &task.ActualEffort, &task.VerificationConfidence,
			&task.CreatedAt, &task.UpdatedAt, &completedAt, &verifiedAt, &archivedAt, &task.Version,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}

		task.CompletedAt = completedAt
		task.VerifiedAt = verifiedAt
		task.ArchivedAt = archivedAt

		tasks = append(tasks, task)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate tasks: %w", err)
	}

	hasMore := (req.Offset + len(tasks)) < total

	return &ListTasksResponse{
		Tasks:   tasks,
		Total:   total,
		Limit:   req.Limit,
		Offset:  req.Offset,
		HasMore: hasMore,
	}, nil
}

// detectLanguageFromFile detects programming language from file path (stub)
func detectLanguageFromFile(filePath string) string {
	ext := ""
	for i := len(filePath) - 1; i >= 0; i-- {
		if filePath[i] == '.' {
			ext = filePath[i:]
			break
		}
	}

	switch ext {
	case ".go":
		return "go"
	case ".js", ".jsx":
		return "javascript"
	case ".ts", ".tsx":
		return "typescript"
	case ".py":
		return "python"
	case ".java":
		return "java"
	default:
		return "unknown"
	}
}

// detectTestFramework detects test framework from file path
func detectTestFramework(filePath string) string {
	lower := strings.ToLower(filePath)
	switch {
	case strings.HasSuffix(lower, "_test.go"):
		return "go-testing"
	case strings.Contains(lower, ".test.js") || strings.Contains(lower, ".spec.js"):
		return "jest"
	case strings.Contains(lower, ".test.ts") || strings.Contains(lower, ".spec.ts"):
		return "jest"
	case strings.Contains(lower, "test_") && strings.HasSuffix(lower, ".py"):
		return "pytest"
	default:
		return "unknown"
	}
}

// Note: traverseAST and getLineColumn are now implemented in ast_bridge.go

// ValidateDirectory validates a directory path (stub - delegates to utils)
func ValidateDirectory(path string) error {
	return utils.ValidateDirectory(path)
}

// getCachedGapAnalysis retrieves cached gap analysis (stub)
func getCachedGapAnalysis(projectID, codebasePath string) (*GapAnalysisReport, bool) {
	return nil, false
}

// setCachedGapAnalysis stores gap analysis in cache (stub)
func setCachedGapAnalysis(projectID, codebasePath string, report *GapAnalysisReport) {
	// Stub - cache would be implemented here
}

// extractFunctionSignature extracts function signature from code (stub)
func extractFunctionSignature(node interface{}, code string, language string) string {
	return ""
}

// CacheConfig represents cache configuration
type CacheConfig struct {
	TaskCacheTTL    time.Duration
	VerificationTTL time.Duration
	DependencyTTL   time.Duration
}

// ServiceConfig represents service configuration
type ServiceConfig struct {
	Cache CacheConfig
}

// GetConfig returns service configuration (stub)
func GetConfig() *ServiceConfig {
	return &ServiceConfig{
		Cache: CacheConfig{
			TaskCacheTTL:    5 * time.Minute,
			VerificationTTL: 10 * time.Minute,
			DependencyTTL:   15 * time.Minute,
		},
	}
}

// contextKey type for context keys
type contextKey string

// projectKey is the context key for project
const projectKey contextKey = "project"

// Handler stubs - required by test_handlers.go
func validateCodeHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte(`{"error":"validateCodeHandler stub"}`))
}

func applyFixHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte(`{"error":"applyFixHandler stub"}`))
}

func validateLLMConfigHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte(`{"error":"validateLLMConfigHandler stub"}`))
}

func getCacheMetricsHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte(`{"error":"getCacheMetricsHandler stub"}`))
}

func getCostMetricsHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte(`{"error":"getCostMetricsHandler stub"}`))
}
