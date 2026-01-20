// Package services LLM and advanced stubs
// Stub implementations for LLM and advanced features
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package services

import (
	"context"
	"fmt"
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

// VerifyTask verifies a task completion (stub)
func VerifyTask(ctx context.Context, taskID string, codebasePath string, forceRecheck bool) (*VerifyTaskResponse, error) {
	return &VerifyTaskResponse{}, nil
}

// Note: The following functions are defined in other files:
// - stringPtr -> task_completion.go
// - countTestCases -> test_analyzer.go
// - determineSeverity, checkTestCoverage, generateGapSummary, getCurrentTimestamp, isCodeFile -> gap_analyzer.go
// - extractFunctionCodeAST, extractFunctionCode, parseSemanticAnalysisResponse, estimateTokenUsage -> logic_analyzer.go
// - sendDependencyBlockingAlert -> task_completion.go
