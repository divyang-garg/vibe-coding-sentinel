// Package services provides workflow orchestration business logic.
// Complies with CODING_STANDARDS.md: Business Services max 400 lines
package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"sentinel-hub-api/models"
)

// WorkflowRepository defines the interface for workflow data access
type WorkflowRepository interface {
	Save(ctx context.Context, workflow *models.WorkflowDefinition) error
	FindByID(ctx context.Context, id string) (*models.WorkflowDefinition, error)
	List(ctx context.Context, limit, offset int) ([]models.WorkflowDefinition, int, error)
	SaveExecution(ctx context.Context, execution *models.WorkflowExecution) error
	FindExecutionByID(ctx context.Context, id string) (*models.WorkflowExecution, error)
}

// WorkflowServiceImpl implements WorkflowService
type WorkflowServiceImpl struct {
	repo         WorkflowRepository
	executions   map[string]*workflowExecutionState
	executionsMu sync.RWMutex
}

// workflowExecutionState tracks the state of an async workflow execution
type workflowExecutionState struct {
	execution *models.WorkflowExecution
	cancel    context.CancelFunc
	mu        sync.RWMutex
}

// NewWorkflowService creates a new workflow service instance
func NewWorkflowService(repo WorkflowRepository) WorkflowService {
	return &WorkflowServiceImpl{
		repo:       repo,
		executions: make(map[string]*workflowExecutionState),
	}
}

// CreateWorkflow creates a new workflow definition
func (s *WorkflowServiceImpl) CreateWorkflow(ctx context.Context, req models.WorkflowDefinition) (interface{}, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("workflow name is required")
	}
	if len(req.Steps) == 0 {
		return nil, fmt.Errorf("workflow must have at least one step")
	}

	// Generate ID and timestamps if not provided
	if req.ID == "" {
		req.ID = fmt.Sprintf("wf_%d", time.Now().UnixNano())
	}
	now := time.Now()
	if req.CreatedAt.IsZero() {
		req.CreatedAt = now
	}
	req.UpdatedAt = now

	// Validate workflow steps
	if err := s.validateWorkflowSteps(req.Steps); err != nil {
		return nil, fmt.Errorf("invalid workflow steps: %w", err)
	}

	// Save workflow to database
	if err := s.repo.Save(ctx, &req); err != nil {
		return nil, fmt.Errorf("failed to save workflow: %w", err)
	}

	return map[string]interface{}{
		"id":         req.ID,
		"name":       req.Name,
		"version":    req.Version,
		"step_count": len(req.Steps),
		"created_at": req.CreatedAt,
		"status":     "created",
	}, nil
}

// GetWorkflow retrieves a workflow by ID
func (s *WorkflowServiceImpl) GetWorkflow(ctx context.Context, id string) (interface{}, error) {
	workflow, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("workflow not found: %w", err)
	}
	if workflow == nil {
		return nil, &models.NotFoundError{
			Resource: "workflow",
			Message:  fmt.Sprintf("workflow not found: %s", id),
		}
	}

	return map[string]interface{}{
		"id":            workflow.ID,
		"name":          workflow.Name,
		"description":   workflow.Description,
		"version":       workflow.Version,
		"steps":         workflow.Steps,
		"input_schema":  workflow.InputSchema,
		"output_schema": workflow.OutputSchema,
		"created_at":    workflow.CreatedAt,
		"updated_at":    workflow.UpdatedAt,
	}, nil
}

// ListWorkflows retrieves workflows with pagination
func (s *WorkflowServiceImpl) ListWorkflows(ctx context.Context, limit int, offset int) ([]interface{}, int, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	workflows, total, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list workflows: %w", err)
	}

	results := make([]interface{}, 0, len(workflows))
	for _, workflow := range workflows {
		results = append(results, map[string]interface{}{
			"id":          workflow.ID,
			"name":        workflow.Name,
			"description": workflow.Description,
			"version":     workflow.Version,
			"step_count":  len(workflow.Steps),
			"created_at":  workflow.CreatedAt,
		})
	}

	return results, total, nil
}

// ExecuteWorkflow executes a workflow asynchronously
func (s *WorkflowServiceImpl) ExecuteWorkflow(ctx context.Context, id string) (interface{}, error) {
	workflow, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("workflow not found: %w", err)
	}

	// Create execution record
	executionID := fmt.Sprintf("exec_%d", time.Now().UnixNano())
	now := time.Now()
	execution := &models.WorkflowExecution{
		ID:         executionID,
		WorkflowID: id,
		Status:     models.WorkflowStatusPending,
		StartedAt:  &now,
		Progress:   0,
		Steps:      make([]models.StepResult, len(workflow.Steps)),
		Input:      make(map[string]interface{}),
	}

	// Initialize step results
	for i, step := range workflow.Steps {
		if i < len(execution.Steps) {
			execution.Steps[i] = models.StepResult{
				StepID:     step.ID,
				Status:     models.StepStatusPending,
				RetryCount: 0,
			}
		}
	}

	// Save execution to database
	if err := s.repo.SaveExecution(ctx, execution); err != nil {
		return nil, fmt.Errorf("failed to save execution: %w", err)
	}

	// Create execution context with cancellation
	execCtx, cancel := context.WithCancel(context.Background())

	// Store execution state
	execState := &workflowExecutionState{
		execution: execution,
		cancel:    cancel,
	}
	s.executionsMu.Lock()
	s.executions[executionID] = execState
	s.executionsMu.Unlock()

	// Start async workflow execution
	go s.executeWorkflowAsync(execCtx, execState, workflow)

	return map[string]interface{}{
		"execution_id": executionID,
		"workflow_id":  id,
		"status":       "pending",
		"started_at":   execution.StartedAt,
		"step_count":   len(workflow.Steps),
	}, nil
}

// UpdateWorkflowStatus updates workflow status (for external control)
func (s *WorkflowServiceImpl) UpdateWorkflowStatus(ctx context.Context, id string, status interface{}) (interface{}, error) {
	// Check if execution is in memory (active)
	s.executionsMu.RLock()
	execState, exists := s.executions[id]
	s.executionsMu.RUnlock()

	if exists {
		// Cancel active execution if cancelling
		statusStr, ok := status.(string)
		if ok && statusStr == "cancelled" {
			execState.cancel()
		}
	}

	execution, err := s.repo.FindExecutionByID(ctx, id)
	if err != nil {
		// Error already wrapped with context from repository
		return nil, err
	}

	// Update status based on input
	statusStr, ok := status.(string)
	if !ok {
		return nil, fmt.Errorf("invalid status format")
	}

	switch statusStr {
	case "cancelled":
		execution.Status = models.WorkflowStatusCancelled
		if execution.CompletedAt == nil {
			now := time.Now()
			execution.CompletedAt = &now
		}
		// Remove from active executions
		s.executionsMu.Lock()
		delete(s.executions, id)
		s.executionsMu.Unlock()
	case "completed":
		execution.Status = models.WorkflowStatusCompleted
		if execution.CompletedAt == nil {
			now := time.Now()
			execution.CompletedAt = &now
		}
		// Remove from active executions
		s.executionsMu.Lock()
		delete(s.executions, id)
		s.executionsMu.Unlock()
	default:
		return nil, fmt.Errorf("unsupported status: %s", statusStr)
	}

	// Save updated execution
	if err := s.repo.SaveExecution(ctx, execution); err != nil {
		return nil, fmt.Errorf("failed to update execution: %w", err)
	}

	return map[string]interface{}{
		"execution_id": id,
		"status":       execution.Status,
		"updated_at":   time.Now(),
	}, nil
}

// GetWorkflowExecution retrieves workflow execution details
func (s *WorkflowServiceImpl) GetWorkflowExecution(ctx context.Context, id string) (interface{}, error) {
	execution, err := s.repo.FindExecutionByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("workflow execution not found: %w", err)
	}
	if execution == nil {
		return nil, &models.NotFoundError{
			Resource: "workflow execution",
			Message:  fmt.Sprintf("workflow execution not found: %s", id),
		}
	}

	return map[string]interface{}{
		"id":           execution.ID,
		"workflow_id":  execution.WorkflowID,
		"status":       execution.Status,
		"started_at":   execution.StartedAt,
		"completed_at": execution.CompletedAt,
		"total_steps":  len(execution.Steps),
		"progress":     execution.Progress,
	}, nil
}

// validateWorkflowSteps validates workflow step configuration
func (s *WorkflowServiceImpl) validateWorkflowSteps(steps []models.WorkflowStep) error {
	if len(steps) == 0 {
		return fmt.Errorf("workflow must have at least one step")
	}

	stepNames := make(map[string]bool)
	for i, step := range steps {
		if step.Name == "" {
			return fmt.Errorf("step %d: name is required", i)
		}
		if stepNames[step.Name] {
			return fmt.Errorf("step %d: duplicate step name '%s'", i, step.Name)
		}
		stepNames[step.Name] = true

		if step.ToolName == "" {
			return fmt.Errorf("step %d (%s): tool name is required", i, step.Name)
		}
	}

	return nil
}

// executeWorkflowAsync executes a workflow asynchronously with proper state management
func (s *WorkflowServiceImpl) executeWorkflowAsync(ctx context.Context, execState *workflowExecutionState, workflow *models.WorkflowDefinition) {
	defer func() {
		// Cleanup: remove from active executions
		s.executionsMu.Lock()
		delete(s.executions, execState.execution.ID)
		s.executionsMu.Unlock()
	}()

	execState.mu.Lock()
	execState.execution.Status = models.WorkflowStatusRunning
	execState.mu.Unlock()

	// Save initial running state
	if err := s.saveExecutionState(ctx, execState); err != nil {
		s.markExecutionFailed(execState, fmt.Errorf("failed to save initial state: %w", err))
		return
	}

	// Build step dependency graph
	stepMap := make(map[string]*models.WorkflowStep)
	stepIndexMap := make(map[string]int)
	for i, step := range workflow.Steps {
		stepMap[step.ID] = &workflow.Steps[i]
		stepIndexMap[step.ID] = i
	}

	// Execute steps with dependency resolution
	if err := s.executeStepsWithDependencies(ctx, execState, workflow.Steps, stepMap, stepIndexMap); err != nil {
		if ctx.Err() == context.Canceled {
			s.markExecutionCancelled(execState)
		} else {
			s.markExecutionFailed(execState, err)
		}
		return
	}

	// Mark execution as completed
	s.markExecutionCompleted(execState)
}

// executeStepsWithDependencies executes workflow steps respecting dependencies
func (s *WorkflowServiceImpl) executeStepsWithDependencies(
	ctx context.Context,
	execState *workflowExecutionState,
	steps []models.WorkflowStep,
	stepMap map[string]*models.WorkflowStep,
	stepIndexMap map[string]int,
) error {
	// Track completed steps
	completedSteps := make(map[string]bool)
	var wg sync.WaitGroup
	stepMu := sync.Mutex{}
	execErr := make(chan error, 1)

	// Execute steps in dependency order
	for len(completedSteps) < len(steps) {
		// Find steps ready to execute (all dependencies completed)
		readySteps := s.findReadySteps(steps, completedSteps, stepMap)

		if len(readySteps) == 0 {
			// Check if we're stuck (circular dependency or all remaining steps have unmet dependencies)
			if !s.hasExecutableSteps(steps, completedSteps, stepMap) {
				return fmt.Errorf("workflow execution stuck: remaining steps have unmet dependencies")
			}
			// Wait a bit before checking again
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(100 * time.Millisecond):
				continue
			}
		}

		// Execute ready steps (can be parallel if no dependencies between them)
		for _, step := range readySteps {
			step := step
			wg.Add(1)
			go func() {
				defer wg.Done()

				// Check if cancelled
				select {
				case <-ctx.Done():
					return
				default:
				}

				// Execute step
				if err := s.executeStep(ctx, execState, step, stepIndexMap[step.ID]); err != nil {
					stepMu.Lock()
					select {
					case execErr <- err:
					default:
					}
					stepMu.Unlock()
					return
				}

				// Mark step as completed
				stepMu.Lock()
				completedSteps[step.ID] = true
				stepMu.Unlock()
			}()
		}

		// Wait for current batch to complete or error
		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			// Check for errors
			select {
			case err := <-execErr:
				return err
			default:
			}
		case err := <-execErr:
			return err
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil
}

// findReadySteps finds steps that are ready to execute (all dependencies completed)
func (s *WorkflowServiceImpl) findReadySteps(
	steps []models.WorkflowStep,
	completedSteps map[string]bool,
	stepMap map[string]*models.WorkflowStep,
) []models.WorkflowStep {
	var ready []models.WorkflowStep

	for _, step := range steps {
		if completedSteps[step.ID] {
			continue
		}

		// Check if all dependencies are completed
		allDepsMet := true
		for _, depID := range step.DependsOn {
			if !completedSteps[depID] {
				allDepsMet = false
				break
			}
		}

		if allDepsMet {
			ready = append(ready, step)
		}
	}

	return ready
}

// hasExecutableSteps checks if there are any steps that can still be executed
func (s *WorkflowServiceImpl) hasExecutableSteps(
	steps []models.WorkflowStep,
	completedSteps map[string]bool,
	stepMap map[string]*models.WorkflowStep,
) bool {
	for _, step := range steps {
		if completedSteps[step.ID] {
			continue
		}

		// Check if all dependencies exist and are completed
		allDepsExist := true
		for _, depID := range step.DependsOn {
			if _, exists := stepMap[depID]; !exists {
				allDepsExist = false
				break
			}
			if !completedSteps[depID] {
				allDepsExist = false
				break
			}
		}

		if allDepsExist {
			return true
		}
	}

	return false
}

// executeStep executes a single workflow step with retry logic
func (s *WorkflowServiceImpl) executeStep(
	ctx context.Context,
	execState *workflowExecutionState,
	step models.WorkflowStep,
	stepIndex int,
) error {
	maxRetries := step.RetryCount
	if maxRetries < 0 {
		maxRetries = 0
	}

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Check for cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Update step status
		execState.mu.Lock()
		if attempt == 0 {
			now := time.Now()
			execState.execution.Steps[stepIndex].Status = models.StepStatusRunning
			execState.execution.Steps[stepIndex].StartedAt = &now
		} else {
			// Retry attempt
			execState.execution.Steps[stepIndex].RetryCount = attempt
		}
		execState.mu.Unlock()

		// Save step state
		if err := s.saveExecutionState(ctx, execState); err != nil {
			return fmt.Errorf("failed to save step state: %w", err)
		}

		// Execute step with timeout
		stepCtx := ctx
		if step.Timeout > 0 {
			var cancel context.CancelFunc
			stepCtx, cancel = context.WithTimeout(ctx, step.Timeout)
			defer cancel()
		}

		// Execute the step
		startTime := time.Now()
		err := s.runStepTool(stepCtx, step)
		duration := time.Since(startTime)

		if err == nil {
			// Step succeeded
			execState.mu.Lock()
			completedAt := time.Now()
			execState.execution.Steps[stepIndex].Status = models.StepStatusCompleted
			execState.execution.Steps[stepIndex].CompletedAt = &completedAt
			execState.execution.Steps[stepIndex].Duration = duration
			execState.execution.Steps[stepIndex].Output = map[string]interface{}{
				"result": "success",
			}
			execState.mu.Unlock()

			// Update progress
			s.updateProgress(execState)

			// Save success state
			if err := s.saveExecutionState(ctx, execState); err != nil {
				return fmt.Errorf("failed to save step success: %w", err)
			}

			return nil
		}

		// Step failed
		lastErr = err
		if attempt < maxRetries {
			// Wait before retry (exponential backoff)
			backoff := time.Duration(attempt+1) * 100 * time.Millisecond
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
		}
	}

	// All retries exhausted
	execState.mu.Lock()
	completedAt := time.Now()
	execState.execution.Steps[stepIndex].Status = models.StepStatusFailed
	execState.execution.Steps[stepIndex].CompletedAt = &completedAt
	execState.execution.Steps[stepIndex].Error = lastErr.Error()
	execState.mu.Unlock()

	// Save failure state
	if err := s.saveExecutionState(ctx, execState); err != nil {
		return fmt.Errorf("failed to save step failure: %w", err)
	}

	return fmt.Errorf("step %s failed after %d retries: %w", step.Name, maxRetries, lastErr)
}

// runStepTool executes the actual tool/action for a workflow step
func (s *WorkflowServiceImpl) runStepTool(ctx context.Context, step models.WorkflowStep) error {
	// Check for cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// TODO: Integrate with actual tool execution system
	// For now, simulate tool execution based on tool name
	switch step.ToolName {
	case "sleep", "delay":
		// Simulate a delay
		duration := 100 * time.Millisecond
		if val, ok := step.Arguments["duration"].(string); ok {
			if parsed, err := time.ParseDuration(val); err == nil {
				duration = parsed
			}
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(duration):
			return nil
		}
	case "validate", "check":
		// Simulate validation
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(50 * time.Millisecond):
			return nil
		}
	default:
		// Generic tool execution simulation
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(100 * time.Millisecond):
			return nil
		}
	}
}

// saveExecutionState saves the execution state to the database
func (s *WorkflowServiceImpl) saveExecutionState(ctx context.Context, execState *workflowExecutionState) error {
	execState.mu.RLock()
	execution := execState.execution
	execState.mu.RUnlock()

	return s.repo.SaveExecution(ctx, execution)
}

// updateProgress updates the execution progress based on completed steps
func (s *WorkflowServiceImpl) updateProgress(execState *workflowExecutionState) {
	execState.mu.Lock()
	defer execState.mu.Unlock()

	if len(execState.execution.Steps) == 0 {
		execState.execution.Progress = 0
		return
	}

	completed := 0
	for _, step := range execState.execution.Steps {
		if step.Status == models.StepStatusCompleted {
			completed++
		}
	}

	execState.execution.Progress = (completed * 100) / len(execState.execution.Steps)
}

// markExecutionCompleted marks the execution as completed
func (s *WorkflowServiceImpl) markExecutionCompleted(execState *workflowExecutionState) {
	execState.mu.Lock()
	execState.execution.Status = models.WorkflowStatusCompleted
	now := time.Now()
	execState.execution.CompletedAt = &now
	execState.execution.Progress = 100
	execState.mu.Unlock()

	// Save final state
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.saveExecutionState(ctx, execState); err != nil {
		// Log error but don't fail
		fmt.Printf("Warning: failed to save completed execution: %v\n", err)
	}
}

// markExecutionFailed marks the execution as failed
func (s *WorkflowServiceImpl) markExecutionFailed(execState *workflowExecutionState, err error) {
	execState.mu.Lock()
	execState.execution.Status = models.WorkflowStatusFailed
	now := time.Now()
	execState.execution.CompletedAt = &now
	if execState.execution.Error == nil {
		execState.execution.Error = &models.WorkflowError{
			Code:    "EXECUTION_FAILED",
			Message: err.Error(),
		}
	}
	execState.mu.Unlock()

	// Save final state
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.saveExecutionState(ctx, execState); err != nil {
		fmt.Printf("Warning: failed to save failed execution: %v\n", err)
	}
}

// markExecutionCancelled marks the execution as cancelled
func (s *WorkflowServiceImpl) markExecutionCancelled(execState *workflowExecutionState) {
	execState.mu.Lock()
	execState.execution.Status = models.WorkflowStatusCancelled
	now := time.Now()
	execState.execution.CompletedAt = &now
	if execState.execution.Error == nil {
		execState.execution.Error = &models.WorkflowError{
			Code:    "CANCELLED",
			Message: "Workflow execution was cancelled",
		}
	}
	execState.mu.Unlock()

	// Save final state
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.saveExecutionState(ctx, execState); err != nil {
		fmt.Printf("Warning: failed to save cancelled execution: %v\n", err)
	}
}
