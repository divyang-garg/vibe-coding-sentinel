// Phase 14E: Task Verification Engine
// Multi-factor verification for task completion - Core orchestration
// Complies with CODING_STANDARDS.md: Utilities max 250 lines

package main

import (
	"context"
	"fmt"
	"time"

	"sentinel-hub-api/models"
	"sentinel-hub-api/services"
)

// VerificationFactors represents the weights for different verification factors
type VerificationFactors struct {
	CodeExistence float64 // 0.4
	CodeUsage     float64 // 0.3
	TestCoverage  float64 // 0.2
	Integration   float64 // 0.1
}

// DefaultVerificationFactors returns default weights
func DefaultVerificationFactors() VerificationFactors {
	return VerificationFactors{
		CodeExistence: 0.4,
		CodeUsage:     0.3,
		TestCoverage:  0.2,
		Integration:   0.1,
	}
}

// VerifyTask verifies a task using multi-factor verification
func VerifyTask(ctx context.Context, taskID string, codebasePath string, force bool) (*VerifyTaskResponse, error) {
	// Get task
	task, err := GetTask(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	// Check cache first (unless force)
	if !force {
		if cachedVerification, found := GetCachedVerification(taskID); found {
			return cachedVerification, nil
		}
	}

	// Run multi-factor verification
	factors := DefaultVerificationFactors()
	verifications := []TaskVerification{}

	// 1. Code Existence Verification
	codeExistenceVerification, err := services.VerifyCodeExistence(ctx, convertTaskToModelsTask(task), codebasePath)
	if err != nil {
		LogError(ctx, "Code existence verification failed: %v", err)
		codeExistenceVerification = models.TaskVerification{
			TaskID:           taskID,
			VerificationType: "code_existence",
			Status:           models.VerificationStatusFailed,
			Confidence:       0.0,
		}
	}
	verifications = append(verifications, convertModelsTaskVerificationToTaskVerification(codeExistenceVerification))

	// 2. Code Usage Verification
	codeUsageVerification, err := services.VerifyCodeUsage(ctx, convertTaskToModelsTask(task), codebasePath)
	if err != nil {
		LogError(ctx, "Code usage verification failed: %v", err)
		codeUsageVerification = models.TaskVerification{
			TaskID:           taskID,
			VerificationType: "code_usage",
			Status:           models.VerificationStatusFailed,
			Confidence:       0.0,
		}
	}
	verifications = append(verifications, convertModelsTaskVerificationToTaskVerification(codeUsageVerification))

	// 3. Test Coverage Verification
	testCoverageVerification, err := verifyTestCoverage(ctx, task, codebasePath)
	if err != nil {
		LogError(ctx, "Test coverage verification failed: %v", err)
		testCoverageVerification = TaskVerification{
			TaskID:           taskID,
			VerificationType: "test_coverage",
			Status:           "failed",
			Confidence:       0.0,
		}
	}
	verifications = append(verifications, testCoverageVerification)

	// 4. Integration Verification
	integrationVerification, err := verifyIntegration(ctx, task, codebasePath)
	if err != nil {
		LogError(ctx, "Integration verification failed: %v", err)
		integrationVerification = TaskVerification{
			TaskID:           taskID,
			VerificationType: "integration",
			Status:           "failed",
			Confidence:       0.0,
		}
	}
	verifications = append(verifications, integrationVerification)

	// Calculate overall confidence
	overallConfidence := calculateOverallConfidence(verifications, factors)

	// Store verifications
	for _, verification := range verifications {
		if err := storeVerification(ctx, verification); err != nil {
			LogError(ctx, "Failed to store verification: %v", err)
		}
	}

	// Update task verification status
	now := time.Now()
	query := `
		UPDATE tasks 
		SET verification_confidence = $1, verified_at = $2, updated_at = $3
		WHERE id = $4
	`
	_, err = execWithTimeout(ctx, query, overallConfidence, now, now, taskID)
	if err != nil {
		LogError(ctx, "Failed to update task verification: %v", err)
	}

	// Determine status based on confidence
	status := determineTaskStatus(overallConfidence, verifications)

	// Build evidence map
	evidence := buildEvidenceMap(verifications)

	response := &VerifyTaskResponse{
		TaskID:            taskID,
		OverallConfidence: overallConfidence,
		Verifications:     verifications,
		Status:            status,
		Evidence:          evidence,
	}

	// Cache the verification result
	SetCachedVerification(taskID, response)

	return response, nil
}

// convertTaskToModelsTask converts a Task from main package to models.Task
func convertTaskToModelsTask(task *Task) *models.Task {
	if task == nil {
		return nil
	}
	return &models.Task{
		ID:                     task.ID,
		ProjectID:              task.ProjectID,
		Source:                 task.Source,
		Title:                  task.Title,
		Description:            task.Description,
		FilePath:               task.FilePath,
		LineNumber:             task.LineNumber,
		Status:                 models.TaskStatus(task.Status),
		Priority:               models.TaskPriority(task.Priority),
		AssignedTo:             task.AssignedTo,
		EstimatedEffort:        task.EstimatedEffort,
		ActualEffort:           task.ActualEffort,
		Tags:                   task.Tags,
		VerificationConfidence: task.VerificationConfidence,
		CreatedAt:              task.CreatedAt,
		UpdatedAt:              task.UpdatedAt,
		CompletedAt:            task.CompletedAt,
		VerifiedAt:             task.VerifiedAt,
		ArchivedAt:             task.ArchivedAt,
		Version:                task.Version,
	}
}

// convertModelsTaskVerificationToTaskVerification converts models.TaskVerification to TaskVerification
func convertModelsTaskVerificationToTaskVerification(v models.TaskVerification) TaskVerification {
	return TaskVerification{
		ID:               v.ID,
		TaskID:           v.TaskID,
		VerificationType: v.VerificationType,
		Status:           string(v.Status),
		Confidence:       v.Confidence,
		Evidence:         v.Evidence,
		RetryCount:       v.RetryCount,
		VerifiedAt:       v.VerifiedAt,
		CreatedAt:        v.CreatedAt,
	}
}
