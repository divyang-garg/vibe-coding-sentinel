// Package router provides HTTP route configuration for the Sentinel Hub API.
// Complies with CODING_STANDARDS.md: HTTP middleware and routing logic
package router

import (
	"os"
	"strings"
	"sentinel-hub-api/handlers"
	"sentinel-hub-api/middleware"
	"sentinel-hub-api/pkg/metrics"
	"sentinel-hub-api/pkg/security"
	"sentinel-hub-api/validation"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// NewRouter creates and configures the main HTTP router with all API routes
func NewRouter(deps *handlers.Dependencies, m *metrics.Metrics) *chi.Mux {
	r := chi.NewRouter()

	// Core middleware - tracing must be first
	r.Use(middleware.TracingMiddleware())
	if m != nil {
		r.Use(middleware.MetricsMiddleware(m))
	}
	r.Use(middleware.RecoveryMiddleware())
	r.Use(middleware.RequestLoggingMiddleware())
	r.Use(middleware.SecurityHeadersMiddleware())
	
	// CORS configuration - environment-aware
	corsAllowedOrigins := getCORSAllowedOrigins()
	r.Use(middleware.CORSMiddleware(middleware.CORSMiddlewareConfig{
		AllowedOrigins: corsAllowedOrigins,
	}))
	
	r.Use(middleware.RateLimitMiddleware(100, 10)) // 100 requests, 10 per second
	
	// Authentication middleware with service integration
	logger := getLogger()
	auditLogger := security.NewAuditLogger(logger)
	r.Use(middleware.AuthMiddleware(middleware.AuthMiddlewareConfig{
		OrganizationService: deps.OrganizationService,
		SkipPaths:           []string{"/health", "/metrics"},
		Logger:              logger,
		AuditLogger:         auditLogger,
	}))

	// Health endpoints (skip auth)
	r.Group(func(r chi.Router) {
		r.Use(middleware.TracingMiddleware())
		r.Use(middleware.RequestLoggingMiddleware())
		setupHealthRoutes(r, deps)
	})

	// API v1 routes (with full security)
	setupAPIV1Routes(r, deps)

	return r
}

// setupHealthRoutes configures health check endpoints
func setupHealthRoutes(r chi.Router, deps *handlers.Dependencies) {
	healthHandler := handlers.NewHealthHandler(deps.DB)
	r.Get("/health", healthHandler.Health)
	r.Get("/health/db", healthHandler.HealthDB)
	r.Get("/health/ready", healthHandler.HealthReady)
	r.Get("/health/live", healthHandler.Health)
	// Prometheus metrics endpoint (no auth required for scraping)
	r.Handle("/metrics", promhttp.Handler())
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

	// AST Analysis endpoints
	setupASTRoutes(r, deps)

	// Repository endpoints
	setupRepositoryRoutes(r, deps)

	// Monitoring endpoints
	setupMonitoringRoutes(r, deps)

	// Knowledge endpoints
	setupKnowledgeRoutes(r, deps)

	// Hook and telemetry endpoints
	setupHookRoutes(r, deps)

	// Test management endpoints
	setupTestRoutes(r, deps)
}

// setupTaskRoutes configures task-related routes
func setupTaskRoutes(r chi.Router, deps *handlers.Dependencies) {
	taskHandler := handlers.NewTaskHandler(deps.TaskService)
	r.Route("/api/v1/tasks", func(r chi.Router) {
		// Validation middleware for POST/PUT requests
		r.Group(func(r chi.Router) {
			r.Use(middleware.ValidationMiddleware(middleware.ValidationMiddlewareConfig{
				Validator: &validation.CompositeValidator{
					Validators: []validation.Validator{
						&validation.FuncValidator{ValidateFunc: validation.ValidateCreateTaskRequest},
					},
				},
				MaxSize: 10 * 1024 * 1024, // 10MB
			}))
			r.Post("/", taskHandler.CreateTask)
		})
		r.Group(func(r chi.Router) {
			r.Use(middleware.ValidationMiddleware(middleware.ValidationMiddlewareConfig{
				Validator: &validation.CompositeValidator{
					Validators: []validation.Validator{
						&validation.FuncValidator{ValidateFunc: validation.ValidateUpdateTaskRequest},
					},
				},
				MaxSize: 10 * 1024 * 1024, // 10MB
			}))
			r.Put("/{id}", taskHandler.UpdateTask)
		})
		r.Get("/", taskHandler.ListTasks)
		r.Get("/{id}", taskHandler.GetTask)
		r.Delete("/{id}", taskHandler.DeleteTask)
		r.Post("/{id}/verify", taskHandler.VerifyTask)
		r.Get("/{id}/dependencies", taskHandler.GetTaskDependencies)
		r.Post("/{id}/dependencies", taskHandler.AddTaskDependency)
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
		r.Group(func(r chi.Router) {
			r.Use(middleware.ValidationMiddleware(middleware.ValidationMiddlewareConfig{
				Validator: &validation.CompositeValidator{
					Validators: []validation.Validator{
						&validation.FuncValidator{ValidateFunc: validation.ValidateCreateOrganizationRequest},
					},
				},
				MaxSize: 5 * 1024 * 1024, // 5MB
			}))
			r.Post("/", orgHandler.CreateOrganization)
		})
		r.Get("/{id}", orgHandler.GetOrganization)
	})
	r.Route("/api/v1/projects", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(middleware.ValidationMiddleware(middleware.ValidationMiddlewareConfig{
				Validator: &validation.CompositeValidator{
					Validators: []validation.Validator{
						&validation.FuncValidator{ValidateFunc: validation.ValidateCreateProjectRequest},
					},
				},
				MaxSize: 5 * 1024 * 1024, // 5MB
			}))
			r.Post("/", orgHandler.CreateProject)
		})
		r.Get("/", orgHandler.ListProjects)
		r.Get("/{id}", orgHandler.GetProject)
		// API Key Management endpoints
		r.Post("/{id}/api-key", orgHandler.GenerateAPIKey)
		r.Get("/{id}/api-key", orgHandler.GetAPIKeyInfo)
		r.Delete("/{id}/api-key", orgHandler.RevokeAPIKey)
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
		r.Post("/security", codeAnalysisHandler.AnalyzeSecurity)
		r.Post("/vibe", codeAnalysisHandler.AnalyzeVibe)
		r.Post("/comprehensive", codeAnalysisHandler.AnalyzeComprehensive)
		r.Post("/intent", codeAnalysisHandler.AnalyzeIntent)
		r.Post("/doc-sync", codeAnalysisHandler.AnalyzeDocSync)
		r.Post("/business-rules", codeAnalysisHandler.AnalyzeBusinessRules)
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

// setupASTRoutes configures AST analysis routes
func setupASTRoutes(r chi.Router, deps *handlers.Dependencies) {
	astHandler := handlers.NewASTHandler(deps.ASTService)
	r.Route("/api/v1/ast", func(r chi.Router) {
		r.Route("/analyze", func(r chi.Router) {
			r.Post("/", astHandler.AnalyzeAST)
			r.Post("/multi", astHandler.AnalyzeMultiFile)
			r.Post("/security", astHandler.AnalyzeSecurity)
			r.Post("/cross", astHandler.AnalyzeCrossFile)
		})
		r.Get("/supported", astHandler.GetSupportedAnalyses)
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

// setupKnowledgeRoutes configures knowledge management routes
func setupKnowledgeRoutes(r chi.Router, deps *handlers.Dependencies) {
	knowledgeHandler := handlers.NewKnowledgeHandler(deps.KnowledgeService)
	r.Route("/api/v1/knowledge", func(r chi.Router) {
		r.Post("/gap-analysis", knowledgeHandler.GapAnalysis)
		r.Get("/business", knowledgeHandler.GetBusinessContext)
		r.Post("/sync", knowledgeHandler.SyncKnowledge)
		r.Route("/items", func(r chi.Router) {
			r.Get("/", knowledgeHandler.ListKnowledgeItems)
			r.Post("/", knowledgeHandler.CreateKnowledgeItem)
			r.Get("/{id}", knowledgeHandler.GetKnowledgeItem)
			r.Put("/{id}", knowledgeHandler.UpdateKnowledgeItem)
			r.Delete("/{id}", knowledgeHandler.DeleteKnowledgeItem)
		})
	})
}

// setupHookRoutes configures hook and telemetry routes
func setupHookRoutes(r chi.Router, deps *handlers.Dependencies) {
	hookHandler := handlers.NewHookHandler(deps.DB)

	// Telemetry endpoint
	r.Post("/api/v1/telemetry/hook", hookHandler.ReportHookTelemetry)

	// Hook endpoints
	r.Route("/api/v1/hooks", func(r chi.Router) {
		r.Get("/metrics", hookHandler.GetHookMetrics)
		r.Get("/metrics/team", hookHandler.GetHookMetricsTeam)
		r.Get("/policies", hookHandler.GetHookPolicies)
		r.Post("/policies", hookHandler.UpdateHookPolicies)
		r.Get("/limits", hookHandler.GetHookLimits)
		r.Post("/baselines", hookHandler.CreateHookBaseline)
		r.Post("/baselines/{id}/review", hookHandler.ReviewHookBaseline)
	})
}

// setupTestRoutes configures test management routes
func setupTestRoutes(r chi.Router, deps *handlers.Dependencies) {
	testHandler := handlers.NewTestHandler(deps.TestService)
	r.Route("/api/v1/test", func(r chi.Router) {
		r.Route("/requirements", func(r chi.Router) {
			r.Post("/generate", testHandler.GenerateTestRequirements)
		})
		r.Route("/coverage", func(r chi.Router) {
			r.Post("/analyze", testHandler.AnalyzeTestCoverage)
			r.Get("/{knowledge_item_id}", testHandler.GetTestCoverage)
		})
		r.Route("/validations", func(r chi.Router) {
			r.Post("/validate", testHandler.ValidateTests)
			r.Get("/{test_requirement_id}", testHandler.GetValidationResults)
		})
		r.Route("/execution", func(r chi.Router) {
			r.Post("/run", testHandler.RunTests)
			r.Get("/{execution_id}", testHandler.GetTestExecutionStatus)
		})
	})
}

// getCORSAllowedOrigins returns CORS allowed origins based on environment
func getCORSAllowedOrigins() []string {
	env := os.Getenv("ENV")
	if env == "development" || env == "dev" {
		// Development: allow common local origins
		return []string{"*", "http://localhost:3000", "http://localhost:8080"}
	}
	
	// Production: get from environment variable
	originsStr := os.Getenv("CORS_ALLOWED_ORIGINS")
	if originsStr == "" {
		return []string{} // Strict: no origins allowed if not configured
	}
	
	// Parse comma-separated origins
	origins := []string{}
	for _, origin := range strings.Split(originsStr, ",") {
		origin = strings.TrimSpace(origin)
		if origin != "" {
			origins = append(origins, origin)
		}
	}
	return origins
}

// getLogger returns a logger instance for middleware
func getLogger() middleware.Logger {
	// Use pkg.JSONLogger if available, otherwise return nil
	// This allows middleware to work even without a logger
	return nil // Can be enhanced to return actual logger
}
