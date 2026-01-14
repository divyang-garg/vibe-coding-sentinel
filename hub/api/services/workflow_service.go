// Package services provides workflow orchestration business logic.
// Complies with CODING_STANDARDS.md: Business Services max 400 lines
package services

import (
	"context"
	"fmt"
	"time"

	"sentinel-hub-api/models"
)

// WorkflowServiceImpl implements WorkflowService
type WorkflowServiceImpl struct {
	// In production, this would have repositories for workflow persistence
	// For now, we'll use in-memory storage for demonstration
	workflows  map[string]*models.WorkflowDefinition
	executions map[string]*models.WorkflowExecution
	nextID     int
}

// NewWorkflowService creates a new workflow service instance
func NewWorkflowService() WorkflowService {
	return &WorkflowServiceImpl{
		workflows:  make(map[string]*models.WorkflowDefinition),
		executions: make(map[string]*models.WorkflowExecution),
		nextID:     1,
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

	// Generate ID and timestamps
	req.ID = fmt.Sprintf("wf_%d", s.nextID)
	s.nextID++
	now := time.Now()
	req.CreatedAt = now
	req.UpdatedAt = now

	// Validate workflow steps
	if err := s.validateWorkflowSteps(req.Steps); err != nil {
		return nil, fmt.Errorf("invalid workflow steps: %w", err)
	}

	// Store workflow
	s.workflows[req.ID] = &req

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
	workflow, exists := s.workflows[id]
	if !exists {
		return nil, fmt.Errorf("workflow not found")
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

	total := len(s.workflows)
	results := make([]interface{}, 0, limit)

	count := 0
	for _, workflow := range s.workflows {
		if count < offset {
			count++
			continue
		}
		if len(results) >= limit {
			break
		}

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

// ExecuteWorkflow executes a workflow
func (s *WorkflowServiceImpl) ExecuteWorkflow(ctx context.Context, id string) (interface{}, error) {
	workflow, exists := s.workflows[id]
	if !exists {
		return nil, fmt.Errorf("workflow not found")
	}

	// Create execution record
	executionID := fmt.Sprintf("exec_%d", s.nextID)
	s.nextID++

	now := time.Now()
	execution := &models.WorkflowExecution{
		ID:         executionID,
		WorkflowID: id,
		Status:     models.WorkflowStatusRunning,
		StartedAt:  &now,
		Progress:   0,
		Steps:      make([]models.StepResult, len(workflow.Steps)),
	}
	*execution.StartedAt = time.Now()

	s.executions[executionID] = execution

	// Simulate workflow execution (in production, this would be async)
	go s.executeWorkflowSteps(execution, workflow.Steps)

	return map[string]interface{}{
		"execution_id": executionID,
		"workflow_id":  id,
		"status":       "running",
		"started_at":   execution.StartedAt,
		"step_count":   len(workflow.Steps),
	}, nil
}

// UpdateWorkflowStatus updates workflow status (for external control)
func (s *WorkflowServiceImpl) UpdateWorkflowStatus(ctx context.Context, id string, status interface{}) (interface{}, error) {
	execution, exists := s.executions[id]
	if !exists {
		return nil, fmt.Errorf("workflow execution not found")
	}

	// Update status based on input
	statusStr, ok := status.(string)
	if !ok {
		return nil, fmt.Errorf("invalid status format")
	}

	switch statusStr {
	case "cancelled":
		execution.Status = models.WorkflowStatusCancelled
	case "completed":
		execution.Status = models.WorkflowStatusCompleted
		if execution.CompletedAt == nil {
			now := time.Now()
			execution.CompletedAt = &now
		}
	default:
		return nil, fmt.Errorf("unsupported status: %s", statusStr)
	}

	return map[string]interface{}{
		"execution_id": id,
		"status":       execution.Status,
		"updated_at":   time.Now(),
	}, nil
}

// GetWorkflowExecution retrieves workflow execution details
func (s *WorkflowServiceImpl) GetWorkflowExecution(ctx context.Context, id string) (interface{}, error) {
	execution, exists := s.executions[id]
	if !exists {
		return nil, fmt.Errorf("workflow execution not found")
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

// executeWorkflowSteps simulates workflow step execution
func (s *WorkflowServiceImpl) executeWorkflowSteps(execution *models.WorkflowExecution, steps []models.WorkflowStep) {
	for i, step := range steps {
		startTime := time.Now()

		// Simulate step execution time
		time.Sleep(100 * time.Millisecond)

		// Mark step as completed
		execution.Steps[i] = models.StepResult{
			StepID:      step.ID,
			Status:      models.StepStatusCompleted,
			Output:      map[string]interface{}{"result": "success"},
			StartedAt:   &startTime,
			CompletedAt: &time.Time{},
			Duration:    time.Since(startTime),
		}
		*execution.Steps[i].CompletedAt = time.Now()
		execution.Progress = (i + 1) * 100 / len(steps)
	}

	// Mark execution as completed
	execution.Status = models.WorkflowStatusCompleted
	now := time.Now()
	execution.CompletedAt = &now
}
