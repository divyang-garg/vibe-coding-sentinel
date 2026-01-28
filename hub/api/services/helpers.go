// Package services helper functions
// Core utility functions for services package
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"sentinel-hub-api/ast"
	"sentinel-hub-api/pkg"
	"sentinel-hub-api/pkg/database"
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

// CachedGapAnalysis represents a cached gap analysis result with expiration
type CachedGapAnalysis struct {
	Report    *GapAnalysisReport
	ExpiresAt time.Time
}

var (
	// gapAnalysisCache stores cached gap analysis reports
	// Key format: "projectID:codebasePath"
	gapAnalysisCache sync.Map // map[string]*CachedGapAnalysis
)

// Default gap analysis cache TTL
const defaultGapAnalysisCacheTTL = 5 * time.Minute

// LogWarn logs a warning message using structured logging
func LogWarn(ctx context.Context, msg string, args ...interface{}) {
	pkg.LogWarn(ctx, msg, args...)
}

// LogError logs an error message using structured logging
func LogError(ctx context.Context, msg string, args ...interface{}) {
	pkg.LogError(ctx, msg, args...)
}

// LogInfo logs an info message using structured logging
func LogInfo(ctx context.Context, msg string, args ...interface{}) {
	pkg.LogInfo(ctx, msg, args...)
}

// LogDebug logs a debug message using structured logging
func LogDebug(ctx context.Context, msg string, args ...interface{}) {
	pkg.LogDebug(ctx, msg, args...)
}

// getQueryTimeout returns the configured query timeout from database package
func getQueryTimeout() time.Duration {
	return database.DefaultTimeoutConfig.QueryTimeout
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
	if err != nil {
		return nil, utils.HandleNotFoundError(err, "task", taskID)
	}

	task.CompletedAt = completedAt
	task.VerifiedAt = verifiedAt
	task.ArchivedAt = archivedAt

	return &task, nil
}

// invalidateGapAnalysisCache invalidates the gap analysis cache for a project
// Removes all cached gap analyses for the specified project ID
func invalidateGapAnalysisCache(projectID string) {
	if projectID == "" {
		return
	}

	// Range through all cache entries and delete those matching the project ID
	gapAnalysisCache.Range(func(key, value interface{}) bool {
		cacheKey := key.(string)
		// Cache key format is "projectID:codebasePath"
		if strings.HasPrefix(cacheKey, projectID+":") {
			gapAnalysisCache.Delete(cacheKey)
		}
		return true
	})
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

// detectLanguageFromFile detects programming language from file path
// Uses AST package's DetectLanguage for consistency
func detectLanguageFromFile(filePath string) string {
	// Use AST package's DetectLanguage for consistency
	return ast.DetectLanguage("", filePath)
}

// detectTestFramework detects test framework from file path
func detectTestFramework(filePath string) string {
	lower := strings.ToLower(filePath)
	baseName := strings.ToLower(filepath.Base(filePath))
	switch {
	case strings.HasSuffix(lower, "_test.go"):
		return "go-testing"
	case strings.Contains(lower, ".test.js") || strings.Contains(lower, ".spec.js"):
		return "jest"
	case strings.Contains(lower, ".test.ts") || strings.Contains(lower, ".spec.ts"):
		return "jest"
	case strings.HasPrefix(baseName, "test_") && strings.HasSuffix(lower, ".py"):
		return "pytest"
	case strings.HasSuffix(lower, "_test.py"):
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

// getCachedGapAnalysis retrieves cached gap analysis if available and not expired
// Returns the cached report and true if found and valid, nil and false otherwise
func getCachedGapAnalysis(projectID, codebasePath string) (*GapAnalysisReport, bool) {
	if projectID == "" || codebasePath == "" {
		return nil, false
	}

	cacheKey := projectID + ":" + codebasePath

	// Load from cache
	if cached, ok := gapAnalysisCache.Load(cacheKey); ok {
		cachedAnalysis := cached.(*CachedGapAnalysis)
		now := time.Now()

		// Check if cache entry is still valid
		if now.Before(cachedAnalysis.ExpiresAt) {
			return cachedAnalysis.Report, true
		}

		// Entry expired, remove from cache
		gapAnalysisCache.Delete(cacheKey)
	}

	return nil, false
}

// setCachedGapAnalysis stores gap analysis in cache with TTL expiration
// Uses configurable TTL from ServiceConfig or default if not configured
func setCachedGapAnalysis(projectID, codebasePath string, report *GapAnalysisReport) {
	if projectID == "" || codebasePath == "" || report == nil {
		return
	}

	// Get TTL from config or use default
	config := GetConfig()
	ttl := defaultGapAnalysisCacheTTL
	if config != nil && config.Cache.GapAnalysisTTL > 0 {
		ttl = config.Cache.GapAnalysisTTL
	}

	cacheKey := projectID + ":" + codebasePath
	cached := &CachedGapAnalysis{
		Report:    report,
		ExpiresAt: time.Now().Add(ttl),
	}

	gapAnalysisCache.Store(cacheKey, cached)
}

// extractFunctionSignature extracts function signature from AST node or code
func extractFunctionSignature(node interface{}, code string, language string) string {
	if node == nil && code == "" {
		return ""
	}

	// Try to use AST package's extraction if node is available
	if node != nil {
		// Use AST package's ExtractFunctions to get function info
		// This is a simplified approach - in production, would parse node directly
		functions, err := ast.ExtractFunctions(code, language, "")
		if err == nil && len(functions) > 0 {
			// Return first function's signature
			funcInfo := functions[0]
			paramStr := "()"
			if len(funcInfo.Parameters) > 0 {
				var params []string
				for _, param := range funcInfo.Parameters {
					if param.Name != "" {
						if param.Type != "" {
							params = append(params, param.Name+" "+param.Type)
						} else {
							params = append(params, param.Name)
						}
					}
				}
				if len(params) > 0 {
					paramStr = "(" + strings.Join(params, ", ") + ")"
				}
			}
			return funcInfo.Name + paramStr
		}
	}

	// Fallback: extract from code string using simple pattern matching
	return extractSignatureFromCode(code, language)
}

// extractSignatureFromCode extracts signature from code string as fallback
func extractSignatureFromCode(code string, language string) string {
	if code == "" {
		return ""
	}

	// Use language parameter to optimize pattern matching
	// Only check patterns relevant to the specified language
	lines := strings.Split(code, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Language-specific pattern matching
		switch language {
		case "go":
			// Go: func FunctionName(params) returnType
			if strings.HasPrefix(line, "func ") {
				// Extract function declaration up to opening brace
				idx := strings.Index(line, "{")
				if idx > 0 {
					return strings.TrimSpace(line[:idx])
				}
				// If no brace, return the whole line
				return line
			}
		case "python":
			// Python: def function_name(params):
			if strings.HasPrefix(line, "def ") {
				// Extract up to colon
				idx := strings.Index(line, ":")
				if idx > 0 {
					return strings.TrimSpace(line[:idx])
				}
				return line
			}
		case "javascript", "typescript":
			// JavaScript/TypeScript: function name(params) or const name = (params) =>
			if strings.Contains(line, "function ") {
				idx := strings.Index(line, "{")
				if idx > 0 {
					return strings.TrimSpace(line[:idx])
				}
				return line
			}
			// Arrow function: const name = (params) =>
			if strings.Contains(line, "=>") {
				idx := strings.Index(line, "=>")
				if idx > 0 {
					beforeArrow := strings.TrimSpace(line[:idx])
					// Try to extract function name and params
					if strings.Contains(beforeArrow, "=") {
						parts := strings.Split(beforeArrow, "=")
						if len(parts) == 2 {
							name := strings.TrimSpace(parts[0])
							params := strings.TrimSpace(parts[1])
							return name + " = " + params
						}
					}
					return beforeArrow
				}
			}
		default:
			// Unknown language or empty - check all patterns as fallback
			// Go pattern
			if strings.HasPrefix(line, "func ") {
				idx := strings.Index(line, "{")
				if idx > 0 {
					return strings.TrimSpace(line[:idx])
				}
				return line
			}
			// Python pattern
			if strings.HasPrefix(line, "def ") {
				idx := strings.Index(line, ":")
				if idx > 0 {
					return strings.TrimSpace(line[:idx])
				}
				return line
			}
			// JavaScript/TypeScript patterns
			if strings.Contains(line, "function ") {
				idx := strings.Index(line, "{")
				if idx > 0 {
					return strings.TrimSpace(line[:idx])
				}
				return line
			}
			if strings.Contains(line, "=>") {
				idx := strings.Index(line, "=>")
				if idx > 0 {
					beforeArrow := strings.TrimSpace(line[:idx])
					if strings.Contains(beforeArrow, "=") {
						parts := strings.Split(beforeArrow, "=")
						if len(parts) == 2 {
							name := strings.TrimSpace(parts[0])
							params := strings.TrimSpace(parts[1])
							return name + " = " + params
						}
					}
					return beforeArrow
				}
			}
		}
	}

	return ""
}

// CacheConfig represents cache configuration
type CacheConfig struct {
	TaskCacheTTL    time.Duration
	VerificationTTL time.Duration
	DependencyTTL   time.Duration
	GapAnalysisTTL  time.Duration
}

// ServiceConfig represents service configuration
type ServiceConfig struct {
	Cache       CacheConfig
	Architecture ArchitectureConfig
}

// ArchitectureConfig holds thresholds for architecture analysis
type ArchitectureConfig struct {
	WarningLines  int
	CriticalLines int
	MaxLines      int
	MaxFanOut     int
}

// GetConfig returns service configuration
func GetConfig() *ServiceConfig {
	return &ServiceConfig{
		Cache: CacheConfig{
			TaskCacheTTL:    5 * time.Minute,
			VerificationTTL: 10 * time.Minute,
			DependencyTTL:   15 * time.Minute,
			GapAnalysisTTL:  5 * time.Minute,
		},
		Architecture: ArchitectureConfig{
			WarningLines:  300,
			CriticalLines: 500,
			MaxLines:      1000,
			MaxFanOut:     15,
		},
	}
}

// contextKey type for context keys
type contextKey string

// projectKey is the context key for project
const projectKey contextKey = "project"

// Handler stubs removed - these are now implemented in handlers package:
// - validateCodeHandler -> handlers.CodeAnalysisHandler.ValidateCode
// - applyFixHandler -> handlers.FixHandler.ApplyFix
// - validateLLMConfigHandler -> handlers.LLMHandler.ValidateLLMConfig
// - getCacheMetricsHandler -> handlers.MetricsHandler.GetCacheMetrics
// - getCostMetricsHandler -> handlers.MetricsHandler.GetCostMetrics
