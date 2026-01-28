// Package services - Service interfaces and types
// Complies with CODING_STANDARDS.md: Business Services max 400 lines
package services

import (
	"context"
	"time"

	"sentinel-hub-api/models"
)

// TaskRepository defines the interface for task data operations
type TaskRepository interface {
	Save(ctx context.Context, task *models.Task) error
	FindByID(ctx context.Context, id string) (*models.Task, error)
	Update(ctx context.Context, task *models.Task) error
	Delete(ctx context.Context, id string) error
	FindByProjectID(ctx context.Context, projectID string, req models.ListTasksRequest) ([]models.Task, int, error)
	SaveDependency(ctx context.Context, dependency *models.TaskDependency) error
	FindDependencies(ctx context.Context, taskID string) ([]models.TaskDependency, error)
	FindDependents(ctx context.Context, taskID string) ([]models.TaskDependency, error)
	DeleteDependency(ctx context.Context, id string) error
	SaveVerification(ctx context.Context, verification *models.TaskVerification) error
}

// DependencyAnalyzer defines the interface for dependency analysis
type DependencyAnalyzer interface {
	AnalyzeDependencies(ctx context.Context, tasks []models.Task) (*models.TaskDependencyGraph, error)
	DetectCycles(ctx context.Context, dependencies []models.TaskDependency) ([][]string, error)
}

// ImpactAnalyzer defines the interface for impact analysis
type ImpactAnalyzer interface {
	AnalyzeImpact(ctx context.Context, taskID string, changeType string, tasks []models.Task, dependencies []models.TaskDependency) (*models.TaskImpactAnalysis, error)
}

// TaskService defines the interface for task-related business operations
type TaskService interface {
	// Core task operations
	CreateTask(ctx context.Context, req models.CreateTaskRequest) (*models.Task, error)
	GetTaskByID(ctx context.Context, id string) (*models.Task, error)
	UpdateTask(ctx context.Context, id string, req models.UpdateTaskRequest) (*models.Task, error)
	ListTasks(ctx context.Context, req models.ListTasksRequest) (*models.ListTasksResponse, error)
	DeleteTask(ctx context.Context, id string) error

	// Task verification operations
	VerifyTask(ctx context.Context, id string, req models.VerifyTaskRequest) (*models.VerifyTaskResponse, error)
	ScanTasks(ctx context.Context, projectID string) ([]models.Task, error)

	// Dependency management
	AddDependency(ctx context.Context, taskID string, req models.AddDependencyRequest) (*models.TaskDependency, error)
	GetDependencies(ctx context.Context, taskID string) (*models.DependencyGraphResponse, error)
	AnalyzeDependencies(ctx context.Context, taskID string) (*models.TaskDependencyGraph, error)

	// Impact analysis
	AnalyzeTaskImpact(ctx context.Context, taskID string, change models.TaskChange) (*models.TaskImpactAnalysis, error)
	GetTaskImpactAnalysis(ctx context.Context, taskID string) (*models.TaskImpactAnalysis, error)

	// Task execution and state management
	UpdateTaskStatus(ctx context.Context, id string, status string, version int) (*models.Task, error)
	AssignTask(ctx context.Context, id string, userID string, version int) (*models.Task, error)
	CompleteTask(ctx context.Context, id string, actualEffort *int, version int) (*models.Task, error)

	// Bulk operations
	BulkUpdateTasks(ctx context.Context, updates []models.TaskChange) ([]*models.Task, error)
	GetTaskExecutionPlan(ctx context.Context, taskIDs []string) (*models.TaskExecutionPlan, error)
}

// WorkflowService defines the interface for workflow-related business operations
type WorkflowService interface {
	CreateWorkflow(ctx context.Context, req models.WorkflowDefinition) (interface{}, error)
	GetWorkflow(ctx context.Context, id string) (interface{}, error)
	ListWorkflows(ctx context.Context, limit int, offset int) ([]interface{}, int, error)
	ExecuteWorkflow(ctx context.Context, id string) (interface{}, error)
	UpdateWorkflowStatus(ctx context.Context, id string, status interface{}) (interface{}, error)
	GetWorkflowExecution(ctx context.Context, id string) (interface{}, error)
}

// APIVersionService defines the interface for API versioning operations
type APIVersionService interface {
	CreateAPIVersion(ctx context.Context, req models.APIVersion) (*models.APIVersion, error)
	GetAPIVersion(ctx context.Context, id string) (*models.APIVersion, error)
	ListAPIVersions(ctx context.Context) ([]*models.APIVersion, error)
	GetVersionCompatibility(ctx context.Context, fromVersion, toVersion string) (interface{}, error)
	CreateVersionMigration(ctx context.Context, req models.VersionMigration) (*models.VersionMigration, error)
}

// CodeAnalysisService defines the interface for code analysis operations
type CodeAnalysisService interface {
	AnalyzeCode(ctx context.Context, req models.CodeAnalysisRequest) (interface{}, error)
	LintCode(ctx context.Context, req models.CodeLintRequest) (interface{}, error)
	RefactorCode(ctx context.Context, req models.CodeRefactorRequest) (interface{}, error)
	GenerateDocumentation(ctx context.Context, req models.DocumentationRequest) (interface{}, error)
	ValidateCode(ctx context.Context, req models.CodeValidationRequest) (interface{}, error)
	AnalyzeSecurity(ctx context.Context, req models.SecurityASTRequest) (interface{}, error)
	AnalyzeVibe(ctx context.Context, req models.CodeAnalysisRequest) (interface{}, error)
	AnalyzeComprehensive(ctx context.Context, req ComprehensiveAnalysisRequest) (interface{}, error)
	AnalyzeIntent(ctx context.Context, req IntentAnalysisRequest) (interface{}, error)
	AnalyzeDocSync(ctx context.Context, req DocSyncRequest) (interface{}, error)
	AnalyzeBusinessRules(ctx context.Context, req BusinessRulesAnalysisRequest) (interface{}, error)
}

// RepositoryService defines the interface for repository management operations
type RepositoryService interface {
	ListRepositories(ctx context.Context, language string, limit int) ([]interface{}, error)
	GetRepositoryImpact(ctx context.Context, id string) (interface{}, error)
	GetRepositoryCentrality(ctx context.Context, id string) (interface{}, error)
	GetRepositoryNetwork(ctx context.Context) (interface{}, error)
	GetRepositoryClusters(ctx context.Context) ([]interface{}, error)
	AnalyzeCrossRepoImpact(ctx context.Context, req interface{}) (interface{}, error)
}

// MonitoringService defines the interface for monitoring and error handling operations
type MonitoringService interface {
	GetErrorDashboard(ctx context.Context) (interface{}, error)
	GetErrorAnalysis(ctx context.Context) (interface{}, error)
	GetErrorStats(ctx context.Context) (interface{}, error)
	ClassifyError(ctx context.Context, req models.ErrorClassification) (interface{}, error)
	ReportError(ctx context.Context, req models.ErrorReport) error
	GetHealthMetrics(ctx context.Context) (interface{}, error)
	GetPerformanceMetrics(ctx context.Context) (interface{}, error)
}

// KnowledgeService defines the interface for knowledge management operations
type KnowledgeService interface {
	RunGapAnalysis(ctx context.Context, req GapAnalysisRequest) (*GapAnalysisReport, error)
	ListKnowledgeItems(ctx context.Context, req ListKnowledgeItemsRequest) ([]KnowledgeItem, error)
	CreateKnowledgeItem(ctx context.Context, item KnowledgeItem) (*KnowledgeItem, error)
	GetKnowledgeItem(ctx context.Context, id string) (*KnowledgeItem, error)
	UpdateKnowledgeItem(ctx context.Context, id string, item KnowledgeItem) (*KnowledgeItem, error)
	DeleteKnowledgeItem(ctx context.Context, id string) error
	GetBusinessContext(ctx context.Context, req BusinessContextRequest) (*BusinessContextResponse, error)
	SyncKnowledge(ctx context.Context, req SyncKnowledgeRequest) (*SyncKnowledgeResponse, error)
}

// TestService defines the interface for test management operations
type TestService interface {
	GenerateTestRequirements(ctx context.Context, req GenerateTestRequirementsRequest) (*GenerateTestRequirementsResponse, error)
	AnalyzeTestCoverage(ctx context.Context, req AnalyzeCoverageRequest) (*AnalyzeCoverageResponse, error)
	GetTestCoverage(ctx context.Context, knowledgeItemID string) (*TestCoverage, error)
	ValidateTests(ctx context.Context, req ValidateTestsRequest) (*ValidateTestsResponse, error)
	GetValidationResults(ctx context.Context, testRequirementID string) (*TestValidation, error)
	RunTests(ctx context.Context, req TestExecutionRequest) (*TestExecutionResponse, error)
	GetTestExecutionStatus(ctx context.Context, executionID string) (*TestExecution, error)
}

// FixService defines the interface for code fix operations
type FixService interface {
	ApplyFix(ctx context.Context, req models.ApplyFixRequest) (*models.ApplyFixResponse, error)
}

// LLMService defines the interface for LLM configuration operations
type LLMService interface {
	ValidateConfig(ctx context.Context, config models.LLMConfig) (*models.ValidateLLMConfigResponse, error)
}

// MetricsService defines the interface for metrics operations
type MetricsService interface {
	GetCacheMetrics(ctx context.Context, projectID string) (*models.CacheMetricsResponse, error)
	GetCostMetrics(ctx context.Context, projectID string) (*models.CostMetricsResponse, error)
}

// Utility functions
func generateID() string {
	return time.Now().Format("20060102150405") + "_id"
}

func generateTimestamp() time.Time {
	return time.Now().UTC()
}
