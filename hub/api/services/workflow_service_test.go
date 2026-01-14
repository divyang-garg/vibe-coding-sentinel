// Package services provides unit tests for workflow service.
// Complies with CODING_STANDARDS.md: Test files max 500 lines
package services

import (
	"context"
	"testing"
	"time"

	"sentinel-hub-api/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockWorkflowRepository is a mock implementation for testing
type MockWorkflowRepository struct {
	mock.Mock
}

func (m *MockWorkflowRepository) Save(ctx context.Context, workflow *models.WorkflowDefinition) error {
	args := m.Called(ctx, workflow)
	return args.Error(0)
}

func (m *MockWorkflowRepository) FindByID(ctx context.Context, id string) (*models.WorkflowDefinition, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.WorkflowDefinition), args.Error(1)
}

func (m *MockWorkflowRepository) FindAll(ctx context.Context, limit, offset int) ([]*models.WorkflowDefinition, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*models.WorkflowDefinition), args.Error(1)
}

func TestWorkflowServiceImpl_CreateWorkflow(t *testing.T) {
	tests := []struct {
		name    string
		req     models.WorkflowDefinition
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid workflow creation",
			req: models.WorkflowDefinition{
				Name:        "Test Workflow",
				Description: "A test workflow",
				Version:     "1.0.0",
				Steps: []models.WorkflowStep{
					{
						ID:       "step-1",
						Name:     "First Step",
						ToolName: "test-tool",
						Arguments: map[string]interface{}{
							"param1": "value1",
						},
					},
				},
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"input": map[string]interface{}{
							"type": "string",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing workflow name",
			req: models.WorkflowDefinition{
				Description: "A test workflow",
				Steps:       []models.WorkflowStep{{ID: "step-1"}},
			},
			wantErr: true,
			errMsg:  "workflow name is required",
		},
		{
			name: "no workflow steps",
			req: models.WorkflowDefinition{
				Name:        "Test Workflow",
				Description: "A test workflow",
				Steps:       []models.WorkflowStep{},
			},
			wantErr: true,
			errMsg:  "workflow must have at least one step",
		},
		{
			name: "invalid workflow step",
			req: models.WorkflowDefinition{
				Name:        "Test Workflow",
				Description: "A test workflow",
				Steps: []models.WorkflowStep{
					{
						ID:   "step-1",
						Name: "First Step",
						// Missing ToolName - should fail validation
					},
				},
			},
			wantErr: true,
			errMsg:  "tool name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewWorkflowService()

			result, err := service.CreateWorkflow(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)

				// Type assert and validate the result
				resultMap, ok := result.(map[string]interface{})
				assert.True(t, ok, "result should be a map")

				assert.Contains(t, resultMap, "id")
				assert.Contains(t, resultMap, "name")
				assert.Contains(t, resultMap, "version")
				assert.Contains(t, resultMap, "step_count")
				assert.Contains(t, resultMap, "created_at")
				assert.Contains(t, resultMap, "status")

				assert.Equal(t, tt.req.Name, resultMap["name"])
				assert.Equal(t, tt.req.Version, resultMap["version"])
				assert.Equal(t, len(tt.req.Steps), resultMap["step_count"])
				assert.Equal(t, "created", resultMap["status"])
			}
		})
	}
}

func TestWorkflowServiceImpl_GetWorkflow(t *testing.T) {
	service := NewWorkflowService()

	// First create a workflow
	workflow := models.WorkflowDefinition{
		Name:        "Test Workflow",
		Description: "A test workflow",
		Version:     "1.0.0",
		Steps: []models.WorkflowStep{
			{
				ID:       "step-1",
				Name:     "First Step",
				ToolName: "test-tool",
			},
		},
	}

	createResult, err := service.CreateWorkflow(context.Background(), workflow)
	assert.NoError(t, err)
	assert.NotNil(t, createResult)

	resultMap := createResult.(map[string]interface{})
	workflowID := resultMap["id"].(string)

	// Now test getting the workflow
	result, err := service.GetWorkflow(context.Background(), workflowID)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	resultMap = result.(map[string]interface{})
	assert.Equal(t, workflowID, resultMap["id"])
	assert.Equal(t, "Test Workflow", resultMap["name"])
	assert.Equal(t, "A test workflow", resultMap["description"])
	assert.Equal(t, "1.0.0", resultMap["version"])
	assert.NotNil(t, resultMap["steps"])
	assert.NotNil(t, resultMap["created_at"])
	assert.NotNil(t, resultMap["updated_at"])
}

func TestWorkflowServiceImpl_GetWorkflow_NotFound(t *testing.T) {
	service := NewWorkflowService()

	result, err := service.GetWorkflow(context.Background(), "non-existent-id")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "workflow not found")
}

func TestWorkflowServiceImpl_ListWorkflows(t *testing.T) {
	service := NewWorkflowService()

	// Create multiple workflows
	workflows := []models.WorkflowDefinition{
		{
			Name: "Workflow 1",
			Steps: []models.WorkflowStep{
				{ID: "step-1", Name: "Step 1", ToolName: "tool1"},
			},
		},
		{
			Name: "Workflow 2",
			Steps: []models.WorkflowStep{
				{ID: "step-1", Name: "Step 1", ToolName: "tool1"},
			},
		},
		{
			Name: "Workflow 3",
			Steps: []models.WorkflowStep{
				{ID: "step-1", Name: "Step 1", ToolName: "tool1"},
			},
		},
	}

	for _, wf := range workflows {
		_, err := service.CreateWorkflow(context.Background(), wf)
		assert.NoError(t, err)
	}

	// Test listing with limit
	result, total, err := service.ListWorkflows(context.Background(), 2, 0)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(result))
	assert.Equal(t, 3, total)

	// Validate result structure
	for _, item := range result {
		itemMap, ok := item.(map[string]interface{})
		assert.True(t, ok)
		assert.Contains(t, itemMap, "id")
		assert.Contains(t, itemMap, "name")
		assert.Contains(t, itemMap, "step_count")
		assert.Contains(t, itemMap, "created_at")
	}
}

func TestWorkflowServiceImpl_ExecuteWorkflow(t *testing.T) {
	service := NewWorkflowService()

	// Create a workflow first
	workflow := models.WorkflowDefinition{
		Name: "Executable Workflow",
		Steps: []models.WorkflowStep{
			{
				ID:       "step-1",
				Name:     "First Step",
				ToolName: "test-tool",
			},
			{
				ID:       "step-2",
				Name:     "Second Step",
				ToolName: "test-tool-2",
			},
		},
	}

	createResult, err := service.CreateWorkflow(context.Background(), workflow)
	assert.NoError(t, err)

	resultMap := createResult.(map[string]interface{})
	workflowID := resultMap["id"].(string)

	// Execute the workflow
	result, err := service.ExecuteWorkflow(context.Background(), workflowID)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	resultMap = result.(map[string]interface{})
	assert.Contains(t, resultMap, "execution_id")
	assert.Contains(t, resultMap, "workflow_id")
	assert.Contains(t, resultMap, "status")
	assert.Contains(t, resultMap, "started_at")
	assert.Contains(t, resultMap, "step_count")

	assert.Equal(t, workflowID, resultMap["workflow_id"])
	assert.Equal(t, "running", resultMap["status"])
	assert.Equal(t, 2, resultMap["step_count"])

	executionID := resultMap["execution_id"].(string)
	assert.NotEmpty(t, executionID)
}

func TestWorkflowServiceImpl_UpdateWorkflowStatus(t *testing.T) {
	service := NewWorkflowService()

	// Create and execute a workflow first
	workflow := models.WorkflowDefinition{
		Name: "Status Update Test",
		Steps: []models.WorkflowStep{
			{ID: "step-1", Name: "Step 1", ToolName: "tool1"},
		},
	}

	createResult, err := service.CreateWorkflow(context.Background(), workflow)
	assert.NoError(t, err)

	resultMap := createResult.(map[string]interface{})
	workflowID := resultMap["id"].(string)

	execResult, err := service.ExecuteWorkflow(context.Background(), workflowID)
	assert.NoError(t, err)

	execMap := execResult.(map[string]interface{})
	executionID := execMap["execution_id"].(string)

	// Wait a bit for execution to potentially complete
	time.Sleep(200 * time.Millisecond)

	// Update status to cancelled
	result, err := service.UpdateWorkflowStatus(context.Background(), executionID, "cancelled")
	assert.NoError(t, err)
	assert.NotNil(t, result)

	resultMap = result.(map[string]interface{})
	assert.Equal(t, executionID, resultMap["execution_id"])
	status := resultMap["status"].(models.WorkflowStatus)
	assert.Equal(t, models.WorkflowStatusCancelled, status)
	assert.Contains(t, resultMap, "updated_at")
}

func TestWorkflowServiceImpl_GetWorkflowExecution(t *testing.T) {
	service := NewWorkflowService()

	// Create and execute a workflow
	workflow := models.WorkflowDefinition{
		Name: "Execution Status Test",
		Steps: []models.WorkflowStep{
			{ID: "step-1", Name: "Step 1", ToolName: "tool1"},
		},
	}

	createResult, err := service.CreateWorkflow(context.Background(), workflow)
	assert.NoError(t, err)

	resultMap := createResult.(map[string]interface{})
	workflowID := resultMap["id"].(string)

	execResult, err := service.ExecuteWorkflow(context.Background(), workflowID)
	assert.NoError(t, err)
	assert.NotNil(t, execResult)

	execMap := execResult.(map[string]interface{})
	assert.Contains(t, execMap, "execution_id")

	executionID, ok := execMap["execution_id"].(string)
	assert.True(t, ok, "execution_id should be a string")
	assert.NotEmpty(t, executionID, "execution_id should not be empty")

	// Get execution status
	result, err := service.GetWorkflowExecution(context.Background(), executionID)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	resultMap = result.(map[string]interface{})
	assert.Contains(t, resultMap, "id")
	assert.Contains(t, resultMap, "workflow_id")
	assert.Equal(t, workflowID, resultMap["workflow_id"])

	// The id in the result should match the execution ID
	resultExecutionID, ok := resultMap["id"].(string)
	assert.True(t, ok)
	assert.Equal(t, executionID, resultExecutionID)
	assert.Contains(t, resultMap, "status")
	assert.Contains(t, resultMap, "started_at")
	assert.Contains(t, resultMap, "completed_at")
	assert.Contains(t, resultMap, "total_steps")
	assert.Contains(t, resultMap, "progress")
}

func TestWorkflowServiceImpl_GetWorkflowExecution_NotFound(t *testing.T) {
	service := NewWorkflowService()

	result, err := service.GetWorkflowExecution(context.Background(), "non-existent-execution")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "workflow execution not found")
}
