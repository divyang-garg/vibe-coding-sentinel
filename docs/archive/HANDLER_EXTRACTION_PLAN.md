# Handler Extraction Plan - Phase 6 Implementation

## Current Status
- ✅ Base handler structure created (`handlers/base.go`)
- ✅ Health handlers extracted (`handlers/health.go`)
- ✅ Models compile successfully
- ⏳ 177 handlers remaining in main.go

## Extraction Strategy

### Step 1: Create Handler Categories
Extract handlers by domain/functionality:

1. **Health Handlers** ✅ DONE
   - `healthHandler` → `handlers/health.go`
   - `healthDBHandler` → `handlers/health.go`
   - `healthReadyHandler` → `handlers/health.go`

2. **Task Handlers** (Priority 1)
   - Create `handlers/task.go`
   - Extract all task-related handlers (~30 handlers)
   - Wire with TaskService

3. **Document Handlers** (Priority 1)
   - Create `handlers/document.go`
   - Extract all document-related handlers (~15 handlers)
   - Wire with DocumentService

4. **Organization Handlers** (Priority 2)
   - Create `handlers/organization.go`
   - Extract org/project handlers (~10 handlers)
   - Wire with OrganizationService

5. **Workflow Handlers** (Priority 2)
   - Create `handlers/workflow.go`
   - Extract workflow handlers (~10 handlers)

6. **API Version Handlers** (Priority 3)
   - Create `handlers/api_version.go`
   - Extract API versioning handlers (~8 handlers)

7. **Error/Monitoring Handlers** (Priority 3)
   - Create `handlers/monitoring.go`
   - Extract error/monitoring handlers (~10 handlers)

8. **Code Analysis Handlers** (Priority 3)
   - Create `handlers/code_analysis.go`
   - Extract code analysis handlers (~15 handlers)

9. **Repository Handlers** (Priority 4)
   - Create `handlers/repository.go`
   - Extract repository handlers (~10 handlers)

10. **Miscellaneous Handlers** (Priority 4)
    - Create `handlers/misc.go`
    - Extract remaining handlers

## Handler Structure Template

```go
// handlers/task.go
package handlers

import (
	"net/http"
	"sentinel-hub-api/services"
)

type TaskHandler struct {
	BaseHandler
	TaskService services.TaskService
}

func NewTaskHandler(taskService services.TaskService) *TaskHandler {
	return &TaskHandler{
		TaskService: taskService,
	}
}

func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	// Extract request
	// Call service
	// Write response
}
```

## Dependency Injection Setup

Create `handlers/dependencies.go`:

```go
package handlers

import (
	"database/sql"
	"sentinel-hub-api/repository"
	"sentinel-hub-api/services"
)

type Dependencies struct {
	DB                  *sql.DB
	TaskService         services.TaskService
	DocumentService     services.DocumentService
	OrganizationService services.OrganizationService
}

func NewDependencies(db *sql.DB) *Dependencies {
	// Initialize repositories
	taskRepo := repository.NewPostgresTaskRepository(db)
	docRepo := repository.NewDocumentRepository(db)
	orgRepo := repository.NewOrganizationRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	
	// Initialize services
	taskService := services.NewTaskService(taskRepo, nil, nil)
	docService := services.NewDocumentService(docRepo, nil, nil, nil)
	orgService := services.NewOrganizationService(orgRepo, projectRepo)
	
	return &Dependencies{
		DB:                  db,
		TaskService:         taskService,
		DocumentService:     docService,
		OrganizationService: orgService,
	}
}
```

## Router Update Plan

Update main.go router setup:

```go
// main.go (reduced to < 100 lines)
func setupRouter(deps *handlers.Dependencies) *chi.Mux {
	r := chi.NewRouter()
	
	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	
	// Health
	healthHandler := handlers.NewHealthHandler(deps.DB)
	r.Get("/health", healthHandler.Health)
	r.Get("/health/db", healthHandler.HealthDB)
	r.Get("/health/ready", healthHandler.HealthReady)
	
	// Tasks
	taskHandler := handlers.NewTaskHandler(deps.TaskService)
	r.Route("/api/v1/tasks", func(r chi.Router) {
		r.Post("/", taskHandler.CreateTask)
		r.Get("/{id}", taskHandler.GetTask)
		// ... more routes
	})
	
	// Documents
	docHandler := handlers.NewDocumentHandler(deps.DocumentService)
	r.Route("/api/v1/documents", func(r chi.Router) {
		r.Post("/upload", docHandler.UploadDocument)
		// ... more routes
	})
	
	return r
}
```

## Extraction Checklist

For each handler category:
- [ ] Create handler file
- [ ] Extract handler functions
- [ ] Wire with service dependencies
- [ ] Update router in main.go
- [ ] Test compilation
- [ ] Verify handler works

## Success Criteria

- ✅ main.go < 100 lines
- ✅ All handlers in handlers/ package
- ✅ Clean compilation
- ✅ All routes working
- ✅ No duplicate handlers

## Estimated Time

- Health handlers: ✅ DONE
- Task handlers: 2 hours
- Document handlers: 1 hour
- Organization handlers: 1 hour
- Workflow handlers: 1 hour
- Remaining handlers: 4 hours
- Router update: 1 hour
- Testing: 2 hours

**Total: ~12 hours**
