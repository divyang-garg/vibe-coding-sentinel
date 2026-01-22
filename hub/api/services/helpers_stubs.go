// Package services LLM and advanced stubs
// Stub implementations for LLM and advanced features
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"sentinel-hub-api/llm"
	"sentinel-hub-api/models"
	"sentinel-hub-api/repository"
)

// callLLM makes an API call to LLM provider by delegating to llm package
func callLLM(ctx context.Context, config *LLMConfig, prompt string, taskType string) (string, int, error) {
	// Convert services.LLMConfig (which is models.LLMConfig) to llm.LLMConfig
	llmConfig := &llm.LLMConfig{
		ID:       config.ID,
		Provider: config.Provider,
		APIKey:   config.APIKey,
		Model:    config.Model,
		KeyType:  config.KeyType,
	}

	// Convert CostOptimizationConfig to llm.CostOptimizationConfig
	if config.CostOptimization.UseCache || config.CostOptimization.CacheTTLHours > 0 ||
		config.CostOptimization.ProgressiveDepth || config.CostOptimization.MaxCostPerRequest > 0 {
		llmConfig.CostOptimization = &llm.CostOptimizationConfig{
			UseCache:          config.CostOptimization.UseCache,
			CacheTTLHours:     config.CostOptimization.CacheTTLHours,
			ProgressiveDepth:  config.CostOptimization.ProgressiveDepth,
			MaxCostPerRequest: config.CostOptimization.MaxCostPerRequest,
		}
	}

	return llm.CallLLM(ctx, llmConfig, prompt, taskType)
}

// calculateEstimatedCost calculates estimated cost for LLM usage
func calculateEstimatedCost(provider, model string, tokens int) float64 {
	// Pricing rates per 1K tokens (matching cost_optimization.go)
	providerRates := map[string]map[string]float64{
		"openai": {
			"gpt-4":         0.03,
			"gpt-3.5-turbo": 0.002,
			"gpt-4-turbo":   0.03,
		},
		"anthropic": {
			"claude-3-opus":   0.015,
			"claude-3-sonnet": 0.003,
			"claude-3-haiku":  0.00025,
		},
		"azure": {
			"gpt-4":         0.03,
			"gpt-3.5-turbo": 0.002,
		},
	}

	rates, exists := providerRates[provider]
	if !exists {
		return 0.0
	}

	rate, exists := rates[model]
	if !exists {
		return 0.0
	}

	// Calculate cost: (tokens / 1000) * rate
	return (float64(tokens) / 1000.0) * rate
}

// llmUsageRepo is the repository for LLM usage tracking (set via SetLLMUsageRepository)
var llmUsageRepo repository.LLMUsageRepository

// SetLLMUsageRepository sets the LLM usage repository for tracking
func SetLLMUsageRepository(repo repository.LLMUsageRepository) {
	llmUsageRepo = repo
}

// GetTrackUsageFunction returns the trackUsage function for bridge pattern
// This allows main package to set up a bridge to services.TrackUsage
// Returns a function that accepts *models.LLMUsage for type compatibility
func GetTrackUsageFunction() func(ctx context.Context, usage *models.LLMUsage) error {
	return func(ctx context.Context, usage *models.LLMUsage) error {
		// Convert *models.LLMUsage to *services.LLMUsage (both are aliases, safe conversion)
		servicesUsage := (*LLMUsage)(usage)
		return TrackUsage(ctx, servicesUsage)
	}
}

// trackUsage tracks LLM usage and persists to database if repository is available
func trackUsage(ctx context.Context, usage *LLMUsage) error {
	if usage == nil {
		return fmt.Errorf("usage cannot be nil")
	}

	// If repository is not set, skip persistence (backward compatible)
	if llmUsageRepo == nil {
		return nil
	}

	// Ensure usage has required fields
	if usage.ID == "" {
		usage.ID = fmt.Sprintf("usage_%d", time.Now().UnixNano())
	}
	if usage.CreatedAt == "" {
		usage.CreatedAt = time.Now().Format(time.RFC3339)
	}

	// Convert to models.LLMUsage
	modelsUsage := &models.LLMUsage{
		ID:            usage.ID,
		ProjectID:     usage.ProjectID,
		ValidationID:  usage.ValidationID,
		Provider:      usage.Provider,
		Model:         usage.Model,
		TokensUsed:    usage.TokensUsed,
		EstimatedCost: usage.EstimatedCost,
		CreatedAt:     usage.CreatedAt,
	}

	return llmUsageRepo.SaveUsage(ctx, modelsUsage)
}

// TrackUsage is an exported wrapper for trackUsage to allow cross-package access
// This function should be used via the bridge pattern from main package
func TrackUsage(ctx context.Context, usage *LLMUsage) error {
	return trackUsage(ctx, usage)
}

// listLLMConfigs lists LLM configurations for a project by delegating to llm package
func listLLMConfigs(ctx context.Context, projectID string) ([]*LLMConfig, error) {
	llmConfigs, err := llm.ListLLMConfigs(ctx, projectID)
	if err != nil {
		return nil, err
	}

	// Convert llm.LLMConfig to services.LLMConfig (which is models.LLMConfig)
	configs := make([]*LLMConfig, len(llmConfigs))
	for i, llmCfg := range llmConfigs {
		configs[i] = &LLMConfig{
			ID:       llmCfg.ID,
			Provider: llmCfg.Provider,
			APIKey:   llmCfg.APIKey,
			Model:    llmCfg.Model,
			KeyType:  llmCfg.KeyType,
		}
		if llmCfg.CostOptimization != nil {
			configs[i].CostOptimization = models.CostOptimizationConfig{
				UseCache:          llmCfg.CostOptimization.UseCache,
				CacheTTLHours:     llmCfg.CostOptimization.CacheTTLHours,
				ProgressiveDepth:  llmCfg.CostOptimization.ProgressiveDepth,
				MaxCostPerRequest: llmCfg.CostOptimization.MaxCostPerRequest,
			}
		}
	}

	return configs, nil
}

// getProjectFromContext extracts project ID from context
// Used by services that need to identify the current project
func getProjectFromContext(ctx context.Context) (string, error) {
	if projectID, ok := ctx.Value("project_id").(string); ok && projectID != "" {
		return projectID, nil
	}
	return "", fmt.Errorf("project ID not found in context")
}

// selectModelWithDepth selects LLM model based on analysis depth (stub)
func selectModelWithDepth(ctx context.Context, projectID string, config *LLMConfig, mode string, depth int, feature string) (string, error) {
	return config.Model, nil
}

// callLLMWithDepth calls LLM with depth-aware settings
// Enhances the prompt based on depth parameter before calling LLM
func callLLMWithDepth(ctx context.Context, config *LLMConfig, prompt string, taskType string, depth int) (string, int, error) {
	// Enhance prompt based on depth parameter
	enhancedPrompt := buildDepthAwarePrompt(prompt, depth, taskType)
	return callLLM(ctx, config, enhancedPrompt, taskType)
}

// UpdateTask updates a task using direct database query
func UpdateTask(ctx context.Context, taskID string, req UpdateTaskRequest) (*Task, error) {
	if taskID == "" {
		return nil, fmt.Errorf("task ID is required")
	}

	// Get existing task first
	task, err := GetTask(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	// Apply updates from request
	if req.Title != nil {
		task.Title = *req.Title
	}
	if req.Description != nil {
		task.Description = *req.Description
	}
	if req.FilePath != nil {
		task.FilePath = *req.FilePath
	}
	if req.LineNumber != nil {
		task.LineNumber = req.LineNumber
	}
	if req.Status != nil {
		task.Status = models.TaskStatus(*req.Status)
	}
	if req.Priority != nil {
		task.Priority = models.TaskPriority(*req.Priority)
	}
	if req.AssignedTo != nil {
		task.AssignedTo = req.AssignedTo
	}
	if req.EstimatedEffort != nil {
		task.EstimatedEffort = req.EstimatedEffort
	}
	if req.ActualEffort != nil {
		task.ActualEffort = req.ActualEffort
	}
	if req.VerificationConfidence != nil {
		task.VerificationConfidence = *req.VerificationConfidence
	}

	// Optimistic locking check
	if req.Version > 0 && task.Version != req.Version {
		return nil, fmt.Errorf("task version mismatch: expected %d, got %d", task.Version, req.Version)
	}

	// Update timestamp and version
	task.UpdatedAt = time.Now().UTC()
	task.Version++

	// Update in database (requires database connection in services package)
	// Note: This is a simplified implementation - full version should use repository pattern
	query := `
		UPDATE tasks SET
			title = $2, description = $3, file_path = $4, line_number = $5,
			status = $6, priority = $7, assigned_to = $8, estimated_effort = $9,
			actual_effort = $10, verification_confidence = $11, updated_at = $12, version = $13
		WHERE id = $1 AND version = $14`

	if db == nil {
		return nil, fmt.Errorf("database not initialized - call SetDB first")
	}

	_, err = db.ExecContext(ctx, query,
		task.ID, task.Title, task.Description, task.FilePath, task.LineNumber,
		string(task.Status), string(task.Priority), task.AssignedTo, task.EstimatedEffort,
		task.ActualEffort, task.VerificationConfidence, task.UpdatedAt, task.Version, req.Version,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	return task, nil
}

// VerifyTask verifies a task completion by analyzing the codebase
// This performs automatic verification by checking if code changes indicate task completion
func VerifyTask(ctx context.Context, taskID string, codebasePath string, forceRecheck bool) (*VerifyTaskResponse, error) {
	if taskID == "" {
		return nil, fmt.Errorf("task ID is required")
	}
	if codebasePath == "" {
		return nil, fmt.Errorf("codebase path is required")
	}

	// Edge case: Validate codebase path before proceeding
	if codebaseInfo, err := os.Stat(codebasePath); err != nil {
		return nil, fmt.Errorf("codebase path not accessible: %w", err)
	} else if !codebaseInfo.IsDir() {
		return nil, fmt.Errorf("codebase path is not a directory")
	}

	// Get task details
	task, err := GetTask(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	
	// Edge case: Task not found or nil
	if task == nil {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}

	// Check cache if not forcing recheck
	if !forceRecheck {
		if cached, found := GetCachedVerification(taskID); found {
			return cached, nil
		}
	}

	// Analyze codebase to determine completion
	confidence, evidence := analyzeTaskCompletion(ctx, task, codebasePath)

	// Create verification record
	verification := &models.TaskVerification{
		ID:               generateID(),
		TaskID:           taskID,
		Status:           determineVerificationStatus(confidence),
		VerificationType: "automatic",
		Confidence:       confidence,
		VerifiedAt:       &[]time.Time{generateTimestamp()}[0],
		Evidence:         evidence,
		CreatedAt:        generateTimestamp(),
	}

	// Update task verification confidence if improved
	if confidence > task.VerificationConfidence {
		updateReq := UpdateTaskRequest{
			VerificationConfidence: &confidence,
			Version:                task.Version,
		}
		_, err = UpdateTask(ctx, taskID, updateReq)
		if err != nil {
			LogWarn(ctx, "Failed to update task verification confidence: %v", err)
		}
	}

	// Cache the verification
	response := &VerifyTaskResponse{
		Task:         task,
		Verification: verification,
		Success:      confidence >= 0.5, // Consider verified if confidence >= 50%
	}
	SetCachedVerification(taskID, response)

	return response, nil
}

// analyzeTaskCompletion analyzes the codebase to determine if task is completed
func analyzeTaskCompletion(ctx context.Context, task *Task, codebasePath string) (float64, map[string]interface{}) {
	confidence := 0.0
	evidence := make(map[string]interface{})

	// Edge case: Nil task
	if task == nil {
		evidence["error"] = "task is nil"
		return 0.0, evidence
	}

	// Edge case: Validate codebase path exists and is a directory
	if codebasePath == "" {
		evidence["error"] = "codebase path is empty"
		return 0.0, evidence
	}
	codebaseInfo, err := os.Stat(codebasePath)
	if err != nil {
		evidence["error"] = fmt.Sprintf("codebase path not accessible: %v", err)
		return 0.0, evidence
	}
	if !codebaseInfo.IsDir() {
		evidence["error"] = "codebase path is not a directory"
		return 0.0, evidence
	}

	// Extract keywords from task title and description
	keywords := extractKeywords(task.Title + " " + task.Description)
	// Edge case: Limit keywords to prevent performance issues
	if len(keywords) > 100 {
		keywords = keywords[:100]
		evidence["keywords_truncated"] = true
	}
	evidence["keywords"] = keywords

	// Check 1: File path exists and is accessible
	if task.FilePath != "" {
		fullPath := filepath.Join(codebasePath, task.FilePath)
		
		// Edge case: Validate path stays within codebase (prevent path traversal)
		// First, get absolute paths for comparison
		codebaseAbs, err := filepath.Abs(codebasePath)
		if err != nil {
			evidence["error"] = fmt.Sprintf("failed to resolve codebase path: %v", err)
			return 0.0, evidence
		}
		
		fullPathAbs, err := filepath.Abs(fullPath)
		if err != nil {
			evidence["file_exists"] = false
			evidence["path_error"] = err.Error()
		} else {
			// Check if resolved path is within codebase
			if !strings.HasPrefix(fullPathAbs+string(filepath.Separator), codebaseAbs+string(filepath.Separator)) &&
				fullPathAbs != codebaseAbs {
				evidence["file_exists"] = false
				evidence["path_traversal_detected"] = true
			} else {
				// Resolve symlinks if any
				resolvedPath, err := filepath.EvalSymlinks(fullPath)
				if err == nil {
					resolvedAbs, _ := filepath.Abs(resolvedPath)
					if !strings.HasPrefix(resolvedAbs+string(filepath.Separator), codebaseAbs+string(filepath.Separator)) &&
						resolvedAbs != codebaseAbs {
						evidence["file_exists"] = false
						evidence["path_traversal_detected"] = true
					} else {
						// Check if file exists (use Lstat to detect symlinks)
						fileInfo, err := os.Lstat(fullPath)
						if err == nil && !fileInfo.IsDir() {
					confidence += 0.3 // File exists
					evidence["file_exists"] = true
					evidence["file_path"] = task.FilePath

					// Check if file has been modified recently (indicates work done)
					if time.Since(fileInfo.ModTime()) < 24*time.Hour {
						confidence += 0.2 // Recently modified
						evidence["recently_modified"] = true
					}

					// Edge case: Check file size before reading (max 10MB)
					const maxFileSize = 10 * 1024 * 1024 // 10MB
					if fileInfo.Size() > maxFileSize {
						evidence["file_too_large"] = true
						evidence["file_size"] = fileInfo.Size()
					} else if fileInfo.Size() > 0 {
						// Check file content for task keywords
						content, err := os.ReadFile(fullPath)
						if err == nil {
							// Edge case: Detect binary files (simple heuristic)
							if isBinaryFile(content) {
								evidence["is_binary"] = true
								// Skip keyword matching for binary files
							} else {
								contentStr := strings.ToLower(string(content))
								keywordMatches := 0
								for _, keyword := range keywords {
									// Check context cancellation
									select {
									case <-ctx.Done():
										evidence["cancelled"] = true
										return confidence, evidence
									default:
									}
									
									if len(keyword) > 0 && strings.Contains(contentStr, strings.ToLower(keyword)) {
										keywordMatches++
									}
								}
								if len(keywords) > 0 {
									keywordScore := float64(keywordMatches) / float64(len(keywords))
									confidence += keywordScore * 0.3 // Up to 30% for keyword matches
									evidence["keyword_matches"] = keywordMatches
									evidence["keyword_score"] = keywordScore
								}
							}
						} else {
							evidence["read_error"] = err.Error()
						}
					} else {
						// Edge case: Empty file
						evidence["file_empty"] = true
					}
						} else {
							evidence["file_exists"] = false
							if err != nil {
								evidence["stat_error"] = err.Error()
							}
						}
					}
				} else {
					// Symlink resolution failed, but path is valid - try direct access
					fileInfo, err := os.Lstat(fullPath)
					if err == nil && !fileInfo.IsDir() {
						confidence += 0.3
						evidence["file_exists"] = true
						evidence["file_path"] = task.FilePath
					} else {
						evidence["file_exists"] = false
						if err != nil {
							evidence["stat_error"] = err.Error()
						}
					}
				}
			}
		}
	}

	// Check 2: Search codebase for task-related code
	if len(keywords) > 0 {
		// Check context cancellation before long operation
		select {
		case <-ctx.Done():
			evidence["cancelled"] = true
			return confidence, evidence
		default:
		}
		
		codeMatches := searchCodebaseForKeywords(ctx, codebasePath, keywords)
		if codeMatches > 0 {
			confidence += 0.2 // Found related code
			evidence["codebase_matches"] = codeMatches
		}
	}

	// Check 3: Look for test files (indicates completion)
	if task.FilePath != "" {
		testPath := findTestFile(codebasePath, task.FilePath)
		if testPath != "" {
			confidence += 0.1 // Test file exists
			evidence["test_file_exists"] = true
			evidence["test_file_path"] = testPath
		}
	}

	// Cap confidence at 1.0
	if confidence > 1.0 {
		confidence = 1.0
	}

	evidence["final_confidence"] = confidence
	return confidence, evidence
}

// searchCodebaseForKeywords searches codebase for task keywords
func searchCodebaseForKeywords(ctx context.Context, codebasePath string, keywords []string) int {
	matches := 0
	// Limit search to avoid performance issues
	maxFiles := 50
	fileCount := 0
	const maxFileSize = 10 * 1024 * 1024 // 10MB

	err := filepath.Walk(codebasePath, func(path string, info os.FileInfo, err error) error {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err != nil {
			// Edge case: Permission errors - skip but don't fail
			if os.IsPermission(err) {
				return nil
			}
			return nil
		}

		// Skip hidden directories and common ignore patterns
		if info.IsDir() {
			if strings.HasPrefix(info.Name(), ".") || info.Name() == "node_modules" || info.Name() == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}

		// Only check code files
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".go" && ext != ".js" && ext != ".ts" && ext != ".py" && ext != ".java" {
			return nil
		}

		if fileCount >= maxFiles {
			return filepath.SkipDir
		}
		fileCount++

		// Edge case: Skip large files
		if info.Size() > maxFileSize {
			return nil
		}

		// Edge case: Skip binary files
		content, err := os.ReadFile(path)
		if err != nil {
			// Permission errors - skip
			if os.IsPermission(err) {
				return nil
			}
			return nil
		}

		// Edge case: Skip binary files
		if isBinaryFile(content) {
			return nil
		}

		contentStr := strings.ToLower(string(content))
		for _, keyword := range keywords {
			if len(keyword) > 0 && strings.Contains(contentStr, strings.ToLower(keyword)) {
				matches++
				break // Count file once per keyword match
			}
		}

		return nil
	})

	if err != nil {
		return 0
	}

	return matches
}

// isBinaryFile detects if content is likely binary
// Simple heuristic: check for null bytes or high ratio of non-printable chars
func isBinaryFile(content []byte) bool {
	if len(content) == 0 {
		return false
	}
	
	// Check for null byte (definitive binary indicator)
	if strings.Contains(string(content), "\x00") {
		return true
	}
	
	// Check ratio of printable characters (simple heuristic)
	// If more than 30% are non-printable (excluding common whitespace), likely binary
	nonPrintable := 0
	sampleSize := len(content)
	if sampleSize > 8192 {
		sampleSize = 8192 // Sample first 8KB
	}
	
	for i := 0; i < sampleSize; i++ {
		b := content[i]
		// Non-printable except common whitespace
		if b < 32 && b != 9 && b != 10 && b != 13 {
			nonPrintable++
		}
	}
	
	ratio := float64(nonPrintable) / float64(sampleSize)
	return ratio > 0.3
}

// findTestFile looks for a test file corresponding to the task file
func findTestFile(codebasePath string, filePath string) string {
	if filePath == "" {
		return ""
	}

	baseDir := filepath.Dir(filePath)
	baseName := filepath.Base(filePath)
	ext := filepath.Ext(baseName)
	nameWithoutExt := strings.TrimSuffix(baseName, ext)

	// Common test file patterns based on extension
	var testPatterns []string
	switch ext {
	case ".go":
		testPatterns = []string{nameWithoutExt + "_test.go"}
	case ".js":
		testPatterns = []string{nameWithoutExt + ".test.js", nameWithoutExt + ".spec.js"}
	case ".ts":
		testPatterns = []string{nameWithoutExt + ".test.ts", nameWithoutExt + ".spec.ts"}
	case ".py":
		testPatterns = []string{"test_" + nameWithoutExt + ".py", nameWithoutExt + "_test.py"}
	default:
		// Generic pattern
		testPatterns = []string{nameWithoutExt + "_test" + ext}
	}

	// Check in same directory
	for _, testName := range testPatterns {
		testPath := filepath.Join(codebasePath, baseDir, testName)
		if _, err := os.Stat(testPath); err == nil {
			relPath, _ := filepath.Rel(codebasePath, testPath)
			return relPath
		}
	}

	// Check in test directory
	testDir := filepath.Join(codebasePath, "tests")
	if _, err := os.Stat(testDir); err == nil {
		for _, testName := range testPatterns {
			testPath := filepath.Join(testDir, testName)
			if _, err := os.Stat(testPath); err == nil {
				relPath, _ := filepath.Rel(codebasePath, testPath)
				return relPath
			}
		}
	}

	// Check for test files with similar names in same directory
	sameDir := filepath.Join(codebasePath, baseDir)
	if dirInfo, err := os.Stat(sameDir); err == nil && dirInfo.IsDir() {
		foundPath := ""
		err := filepath.Walk(sameDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if info.IsDir() {
				return nil
			}
			fileName := strings.ToLower(info.Name())
			if strings.Contains(fileName, strings.ToLower(nameWithoutExt)) &&
				(strings.Contains(fileName, "test") || strings.Contains(fileName, "_test")) {
				relPath, _ := filepath.Rel(codebasePath, path)
				foundPath = relPath
				return filepath.SkipAll // Stop walking
			}
			return nil
		})
		if err == nil && foundPath != "" {
			return foundPath
		}
	}

	return ""
}

// determineVerificationStatus determines verification status based on confidence
func determineVerificationStatus(confidence float64) models.VerificationStatus {
	if confidence >= 0.8 {
		return models.VerificationStatusVerified
	} else if confidence >= 0.5 {
		return models.VerificationStatusPending
	}
	return models.VerificationStatusPending // Low confidence still pending
}

// Note: The following functions are defined in other files:
// - stringPtr -> task_completion.go
// - countTestCases -> test_analyzer.go
// - determineSeverity, checkTestCoverage, generateGapSummary, getCurrentTimestamp, isCodeFile -> gap_analyzer.go
// - extractFunctionCodeAST, extractFunctionCode, parseSemanticAnalysisResponse, estimateTokenUsage -> logic_analyzer.go
// - sendDependencyBlockingAlert -> task_completion.go
