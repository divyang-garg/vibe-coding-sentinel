// Package router provides HTTP route configuration for the Sentinel Hub API.
// Complies with CODING_STANDARDS.md: HTTP middleware and routing logic
package router

import (
	"sentinel-hub-api/handlers"
	"sentinel-hub-api/middleware"

	"github.com/go-chi/chi/v5"
)

// NewRouter creates and configures the main HTTP router with all API routes
func NewRouter(deps *handlers.Dependencies) *chi.Mux {
	r := chi.NewRouter()

	// Core middleware
	r.Use(middleware.RecoveryMiddleware())
	r.Use(middleware.RequestLoggingMiddleware())
	r.Use(middleware.SecurityHeadersMiddleware())
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.RateLimitMiddleware(100, 10)) // 100 requests, 10 per second
	r.Use(middleware.AuthMiddleware())

	// Health endpoints (skip auth)
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequestLoggingMiddleware())
		setupHealthRoutes(r, deps)
	})

	// API v1 routes (with full security)
	setupAPIV1Routes(r, deps)

	return r
}

// setupHealthRoutes configures health check endpoints
func setupHealthRoutes(r chi.Router, deps *handlers.Dependencies) {
	healthHandler := handlers.NewHealthHandler(nil)
	r.Get("/health", healthHandler.Health)
	r.Get("/health/db", healthHandler.HealthDB)
	r.Get("/health/ready", healthHandler.HealthReady)
}

// setupAPIV1Routes configures all API v1 endpoints
func setupAPIV1Routes(r chi.Router, deps *handlers.Dependencies) {
	// Task endpoints
	setupTaskRoutes(r, deps)

	// Document endpoints
	setupDocumentRoutes(r, deps)

	// Organization endpoints
	setupOrganizationRoutes(r, deps)

	// Workflow endpoints
	setupWorkflowRoutes(r, deps)

	// API Version endpoints
	setupAPIVersionRoutes(r, deps)

	// Code Analysis endpoints
	setupCodeAnalysisRoutes(r, deps)

	// Repository endpoints
	setupRepositoryRoutes(r, deps)

	// Monitoring endpoints
	setupMonitoringRoutes(r, deps)
}

// setupTaskRoutes configures task-related routes
func setupTaskRoutes(r chi.Router, deps *handlers.Dependencies) {
	taskHandler := handlers.NewTaskHandler(deps.TaskService)
	r.Route("/api/v1/tasks", func(r chi.Router) {
		r.Post("/", taskHandler.CreateTask)
		r.Get("/", taskHandler.ListTasks)
		r.Get("/{id}", taskHandler.GetTask)
		r.Put("/{id}", taskHandler.UpdateTask)
		r.Delete("/{id}", taskHandler.DeleteTask)
	})
}

// setupDocumentRoutes configures document-related routes
func setupDocumentRoutes(r chi.Router, deps *handlers.Dependencies) {
	docHandler := handlers.NewDocumentHandler(deps.DocumentService)
	r.Route("/api/v1/documents", func(r chi.Router) {
		r.Post("/upload", docHandler.UploadDocument)
		r.Get("/", docHandler.ListDocuments)
		r.Get("/{id}", docHandler.GetDocument)
		r.Get("/{id}/status", docHandler.GetDocumentStatus)
	})
}

// setupOrganizationRoutes configures organization and project routes
func setupOrganizationRoutes(r chi.Router, deps *handlers.Dependencies) {
	orgHandler := handlers.NewOrganizationHandler(deps.OrganizationService)
	r.Route("/api/v1/organizations", func(r chi.Router) {
		r.Post("/", orgHandler.CreateOrganization)
		r.Get("/{id}", orgHandler.GetOrganization)
	})
	r.Route("/api/v1/projects", func(r chi.Router) {
		r.Post("/", orgHandler.CreateProject)
		r.Get("/", orgHandler.ListProjects)
		r.Get("/{id}", orgHandler.GetProject)
	})
}

// setupWorkflowRoutes configures workflow-related routes
func setupWorkflowRoutes(r chi.Router, deps *handlers.Dependencies) {
	workflowHandler := handlers.NewWorkflowHandler(deps.WorkflowService)
	r.Route("/api/v1/workflows", func(r chi.Router) {
		r.Post("/", workflowHandler.CreateWorkflow)
		r.Get("/", workflowHandler.ListWorkflows)
		r.Get("/{id}", workflowHandler.GetWorkflow)
		r.Post("/{id}/execute", workflowHandler.ExecuteWorkflow)
	})
	r.Route("/api/v1/workflows/executions", func(r chi.Router) {
		r.Get("/{id}", workflowHandler.GetWorkflowExecution)
	})
}

// setupAPIVersionRoutes configures API versioning routes
func setupAPIVersionRoutes(r chi.Router, deps *handlers.Dependencies) {
	apiVersionHandler := handlers.NewAPIVersionHandler(deps.APIVersionService)
	r.Route("/api/v1/versions", func(r chi.Router) {
		r.Post("/", apiVersionHandler.CreateAPIVersion)
		r.Get("/", apiVersionHandler.ListAPIVersions)
		r.Get("/{id}", apiVersionHandler.GetAPIVersion)
		r.Get("/compatibility", apiVersionHandler.GetVersionCompatibility)
		r.Post("/migrations", apiVersionHandler.CreateAPIVersionMigration)
	})
}

// setupCodeAnalysisRoutes configures code analysis routes
func setupCodeAnalysisRoutes(r chi.Router, deps *handlers.Dependencies) {
	codeAnalysisHandler := handlers.NewCodeAnalysisHandler(deps.CodeAnalysisService)
	r.Route("/api/v1/analyze", func(r chi.Router) {
		r.Post("/code", codeAnalysisHandler.AnalyzeCode)
	})
	r.Route("/api/v1/lint", func(r chi.Router) {
		r.Post("/code", codeAnalysisHandler.LintCode)
	})
	r.Route("/api/v1/refactor", func(r chi.Router) {
		r.Post("/code", codeAnalysisHandler.RefactorCode)
	})
	r.Route("/api/v1/generate", func(r chi.Router) {
		r.Post("/docs", codeAnalysisHandler.GenerateDocs)
	})
	r.Route("/api/v1/validate", func(r chi.Router) {
		r.Post("/code", codeAnalysisHandler.ValidateCode)
	})
}

// setupRepositoryRoutes configures repository management routes
func setupRepositoryRoutes(r chi.Router, deps *handlers.Dependencies) {
	repoHandler := handlers.NewRepositoryHandler(deps.RepositoryService)
	r.Route("/api/v1/repositories", func(r chi.Router) {
		r.Get("/", repoHandler.ListRepositories)
		r.Get("/{id}/impact", repoHandler.GetRepositoryImpact)
		r.Get("/{id}/centrality", repoHandler.GetRepositoryCentrality)
		r.Get("/network", repoHandler.GetRepositoryNetwork)
		r.Get("/clusters", repoHandler.GetRepositoryClusters)
		r.Post("/analyze-cross-repo", repoHandler.AnalyzeCrossRepoImpact)
	})
}

// setupMonitoringRoutes configures monitoring and error handling routes
func setupMonitoringRoutes(r chi.Router, deps *handlers.Dependencies) {
	monitoringHandler := handlers.NewMonitoringHandler(deps.MonitoringService)
	r.Route("/api/v1/monitoring", func(r chi.Router) {
		r.Route("/errors", func(r chi.Router) {
			r.Get("/dashboard", monitoringHandler.GetErrorDashboard)
			r.Get("/analysis", monitoringHandler.GetErrorAnalysis)
			r.Get("/stats", monitoringHandler.GetErrorStats)
			r.Post("/classify", monitoringHandler.ClassifyError)
			r.Post("/report", monitoringHandler.ReportError)
		})
		r.Get("/health", monitoringHandler.GetHealthMetrics)
		r.Get("/performance", monitoringHandler.GetPerformanceMetrics)
	})
}
