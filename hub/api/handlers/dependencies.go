// Package handlers - Dependency injection container
package handlers

import (
	"context"
	"database/sql"

	"sentinel-hub-api/pkg"
	"sentinel-hub-api/repository"
	"sentinel-hub-api/services"
)

// Dependencies holds all service dependencies for handlers
type Dependencies struct {
	DB                  *sql.DB
	TaskService         services.TaskService
	DocumentService     services.DocumentService
	OrganizationService services.OrganizationService
	WorkflowService     services.WorkflowService
	APIVersionService   services.APIVersionService
	CodeAnalysisService services.CodeAnalysisService
	ASTService          services.ASTService
	RepositoryService   services.RepositoryService
	MonitoringService   services.MonitoringService
	KnowledgeService    services.KnowledgeService
	TestService         services.TestService
}

// NewDependencies creates and wires all dependencies
func NewDependencies(db *sql.DB) *Dependencies {
	// Create database wrapper
	dbWrapper := repository.NewPostgresDatabase(db)

	// Initialize repositories
	taskRepo := repository.NewTaskRepository(dbWrapper)
	docRepo := repository.NewDocumentRepository(dbWrapper)
	orgRepo := repository.NewOrganizationRepository(dbWrapper)
	projectRepo := repository.NewProjectRepository(dbWrapper)
	workflowRepo := repository.NewWorkflowRepository(dbWrapper)
	errorReportRepo := repository.NewErrorReportRepository(dbWrapper)
	llmUsageRepo := repository.NewLLMUsageRepository(dbWrapper)

	// Set up LLM usage tracking in services package
	services.SetLLMUsageRepository(llmUsageRepo)

	// Initialize analyzers
	dependencyAnalyzer := repository.NewDependencyAnalyzer()
	impactAnalyzer := repository.NewImpactAnalyzer()
	knowledgeExtractor := repository.NewKnowledgeExtractor()
	documentValidator := repository.NewDocumentValidator()
	searchEngine := repository.NewSearchEngine()

	// Initialize logger for structured logging
	logger := pkg.NewJSONLogger(pkg.JSONLoggerConfig{
		ServiceName: "sentinel-hub-api",
		Level:       pkg.LogLevelInfo,
	})

	// Initialize services
	taskService := services.NewTaskService(taskRepo, dependencyAnalyzer, impactAnalyzer)
	docService := services.NewDocumentService(docRepo, knowledgeExtractor, documentValidator, searchEngine, logger)
	orgService := services.NewOrganizationService(orgRepo, projectRepo)
	workflowService := services.NewWorkflowService(workflowRepo)
	monitoringService := services.NewMonitoringService(errorReportRepo)
	apiVersionService := services.NewAPIVersionService()
	codeAnalysisService := services.NewCodeAnalysisService()
	astService := services.NewASTService()
	repositoryService := services.NewRepositoryService()
	knowledgeService := services.NewKnowledgeService(db)
	testService := services.NewTestService(db)

	return &Dependencies{
		DB:                  db,
		TaskService:         taskService,
		DocumentService:     docService,
		OrganizationService: orgService,
		WorkflowService:     workflowService,
		APIVersionService:   apiVersionService,
		CodeAnalysisService: codeAnalysisService,
		ASTService:          astService,
		RepositoryService:   repositoryService,
		MonitoringService:   monitoringService,
		KnowledgeService:    knowledgeService,
		TestService:         testService,
	}
}

// Cleanup gracefully closes all resources
func (d *Dependencies) Cleanup(ctx context.Context) {
	if d.DB != nil {
		d.DB.Close()
	}
	// Add other cleanup operations as needed
}
