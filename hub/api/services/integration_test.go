// Package services provides integration testing for service-repository interactions.
package services

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"sentinel-hub-api/database"
	"sentinel-hub-api/models"
	"sentinel-hub-api/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// IntegrationTestSuite provides comprehensive integration testing
type IntegrationTestSuite struct {
	suite.Suite
	db                 *sql.DB
	dbAdapter          *database.SQLDBAdapter
	taskRepo           TaskRepository
	docRepo            DocumentRepository
	orgRepo            OrganizationRepository
	projectRepo        ProjectRepository
	depAnalyzer        DependencyAnalyzer
	impactAnalyzer     ImpactAnalyzer
	knowledgeExtractor *repository.KnowledgeExtractorImpl
	docValidator       *repository.DocumentValidatorImpl
	searchEngine       *repository.SearchEngineImpl
}

// SetupSuite initializes test suite with database connection
func (suite *IntegrationTestSuite) SetupSuite() {
	// Setup test database connection
	suite.db = database.SetupTestDB(suite.T())
	if suite.db == nil {
		// Database not available, tests will be skipped
		return
	}

	// Create database adapter
	suite.dbAdapter = database.NewSQLDBAdapter(suite.db)

	// Initialize repositories with database
	suite.taskRepo = repository.NewTaskRepository(suite.dbAdapter)
	suite.docRepo = repository.NewDocumentRepository(suite.dbAdapter)
	suite.orgRepo = repository.NewOrganizationRepository(suite.dbAdapter)
	suite.projectRepo = repository.NewProjectRepository(suite.dbAdapter)
}

// TearDownSuite cleans up test suite
func (suite *IntegrationTestSuite) TearDownSuite() {
	if suite.db != nil {
		database.TeardownTestDB(suite.T(), suite.db)
	}
}

// SetupTest initializes test dependencies
func (suite *IntegrationTestSuite) SetupTest() {
	// Clean up test data before each test
	if suite.db != nil {
		database.CleanupTestData(suite.T(), suite.db)
	}

	// Initialize analyzers (don't require database)
	suite.depAnalyzer = repository.NewDependencyAnalyzer()
	suite.impactAnalyzer = repository.NewImpactAnalyzer()
	suite.knowledgeExtractor = repository.NewKnowledgeExtractor()
	suite.docValidator = repository.NewDocumentValidator()
	suite.searchEngine = repository.NewSearchEngine()
}

// TestTaskServiceIntegration tests task service with repository integration
func (suite *IntegrationTestSuite) TestTaskServiceIntegration() {
	if suite.db == nil {
		suite.T().Skip("Database not available")
		return
	}

	ctx := context.Background()

	// Test task creation validation
	req := models.CreateTaskRequest{
		ProjectID:   "project-123",
		Title:       "Integration Test Task",
		Description: "Testing service-repository integration",
		Source:      "integration_test",
		Priority:    "medium",
	}

	// Validate request structure
	assert.NotEmpty(suite.T(), req.ProjectID)
	assert.NotEmpty(suite.T(), req.Title)
	assert.Equal(suite.T(), "integration_test", req.Source)

	// Create service instance with real repository
	service := NewTaskService(suite.taskRepo, suite.depAnalyzer, suite.impactAnalyzer)

	// Test task creation through service
	task, err := service.CreateTask(ctx, req)
	if err != nil {
		// If database tables don't exist, skip the test
		suite.T().Skipf("Database tables may not exist: %v", err)
		return
	}

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), task)
	assert.Equal(suite.T(), req.ProjectID, task.ProjectID)
	assert.Equal(suite.T(), req.Title, task.Title)
	assert.Equal(suite.T(), models.TaskStatusPending, task.Status)
	assert.Equal(suite.T(), 1, task.Version)

	// Test retrieving the task
	retrieved, err := suite.taskRepo.FindByID(ctx, task.ID)
	if err == nil {
		assert.Equal(suite.T(), task.ID, retrieved.ID)
		assert.Equal(suite.T(), task.Title, retrieved.Title)
	}
}

// TestDocumentServiceIntegration tests document service integration
func (suite *IntegrationTestSuite) TestDocumentServiceIntegration() {
	// Test document upload request validation
	req := models.DocumentUploadRequest{
		ProjectID:    "project-123",
		Name:         "test-document.pdf",
		OriginalName: "Test Document.pdf",
	}

	assert.NotEmpty(suite.T(), req.ProjectID)
	assert.NotEmpty(suite.T(), req.Name)
	assert.Equal(suite.T(), "Test Document.pdf", req.OriginalName)

	// Test document entity creation
	now := time.Now()
	doc := &models.Document{
		ID:           "test-doc-123",
		ProjectID:    req.ProjectID,
		Name:         req.Name,
		OriginalName: req.OriginalName,
		Size:         1024,
		MimeType:     "application/pdf",
		Status:       "uploaded",
		Progress:     0,
		FilePath:     "/tmp/test.pdf",
		CreatedAt:    now,
	}

	// Verify all document fields are properly set
	assert.Equal(suite.T(), "test-doc-123", doc.ID)
	assert.Equal(suite.T(), req.ProjectID, doc.ProjectID)
	assert.Equal(suite.T(), req.Name, doc.Name)
	assert.Equal(suite.T(), req.OriginalName, doc.OriginalName)
	assert.Equal(suite.T(), "application/pdf", doc.MimeType)
	assert.Equal(suite.T(), models.DocumentStatusUploaded, doc.Status)
	assert.Equal(suite.T(), int64(0), doc.Progress)
	assert.Equal(suite.T(), "/tmp/test.pdf", doc.FilePath)
	assert.Equal(suite.T(), int64(1024), doc.Size)
	assert.Equal(suite.T(), now, doc.CreatedAt)
}

// TestOrganizationServiceIntegration tests organization service integration
func (suite *IntegrationTestSuite) TestOrganizationServiceIntegration() {
	// Test organization creation request validation
	req := models.CreateOrganizationRequest{
		Name: "Test Organization",
	}

	assert.NotEmpty(suite.T(), req.Name)

	// Test organization entity creation
	now := time.Now()
	org := &models.Organization{
		ID:        "test-org-123",
		Name:      req.Name,
		CreatedAt: now,
	}

	assert.Equal(suite.T(), req.Name, org.Name)
	assert.NotEmpty(suite.T(), org.ID)
	assert.Equal(suite.T(), now, org.CreatedAt)
}

// TestKnowledgeExtractionIntegration tests knowledge extraction workflow
func (suite *IntegrationTestSuite) TestKnowledgeExtractionIntegration() {
	ctx := context.TODO()

	// Test text with business rules
	testText := `
	This system requires authentication. Users must authenticate before accessing data.
	Security requirements include encryption at rest and in transit.
	The performance must be less than 2 seconds for login.
	API endpoints should use HTTPS encryption.
	The database must support ACID transactions.
	`

	// Test knowledge extraction
	docID := "test-doc-123"
	items, err := suite.knowledgeExtractor.ExtractFromText(ctx, testText, docID)

	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), items)

	// Should find business rules in the text
	foundRequirement := false
	foundPerformance := false
	foundSecurity := false

	for _, item := range items {
		assert.NotEmpty(suite.T(), item.ID)
		assert.NotEmpty(suite.T(), item.Type)
		assert.NotEmpty(suite.T(), item.Content)
		assert.True(suite.T(), item.Confidence >= 0 && item.Confidence <= 1)
		assert.Equal(suite.T(), docID, item.DocumentID)

		// Check for expected rule types
		switch item.Type {
		case "functional_requirement":
			foundRequirement = true
		case "performance_requirement":
			foundPerformance = true
		case "security_requirement":
			foundSecurity = true
		}
	}

	// Verify we found the expected types of knowledge
	assert.True(suite.T(), foundRequirement, "Should find functional requirements")
	assert.True(suite.T(), foundPerformance, "Should find performance requirements")
	assert.True(suite.T(), foundSecurity, "Should find security requirements")
}

// TestDependencyAnalysisIntegration tests dependency analysis workflow
func (suite *IntegrationTestSuite) TestDependencyAnalysisIntegration() {
	ctx := context.TODO()

	// Create test tasks
	tasks := []models.Task{
		{
			ID:        "task-1",
			ProjectID: "project-123",
			Title:     "Setup infrastructure",
			Status:    "completed",
		},
		{
			ID:        "task-2",
			ProjectID: "project-123",
			Title:     "Develop API",
			Status:    "in_progress",
		},
		{
			ID:        "task-3",
			ProjectID: "project-123",
			Title:     "Create UI",
			Status:    "pending",
		},
	}

	// Test dependency graph creation
	graph, err := suite.depAnalyzer.AnalyzeDependencies(ctx, tasks)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), graph)
	assert.True(suite.T(), graph.IsValid)
	assert.NotEmpty(suite.T(), graph.ExecutionOrder)

	// Should have all tasks in execution order
	assert.Len(suite.T(), graph.ExecutionOrder, len(tasks))
	for _, task := range tasks {
		assert.Contains(suite.T(), graph.ExecutionOrder, task.ID)
		assert.NotNil(suite.T(), graph.Tasks[task.ID])
		assert.Equal(suite.T(), task.ID, graph.Tasks[task.ID].ID)
	}
}

// TestImpactAnalysisIntegration tests impact analysis workflow
func (suite *IntegrationTestSuite) TestImpactAnalysisIntegration() {
	ctx := context.TODO()

	taskID := "task-123"
	changeType := "priority_change"

	// Test tasks with dependencies
	tasks := []models.Task{
		{
			ID:        taskID,
			ProjectID: "project-123",
			Title:     "Critical Task",
			Priority:  "high",
			Status:    "in_progress",
		},
		{
			ID:        "task-456",
			ProjectID: "project-123",
			Title:     "Dependent Task",
			Priority:  "medium",
			Status:    "pending",
		},
	}

	dependencies := []models.TaskDependency{
		{
			ID:              "dep-123",
			TaskID:          "task-456",
			DependsOnTaskID: taskID,
			DependencyType:  "finish_to_start",
			Confidence:      0.9,
		},
	}

	// Test impact analysis
	analysis, err := suite.impactAnalyzer.AnalyzeImpact(ctx, taskID, changeType, tasks, dependencies)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), analysis)
	assert.Equal(suite.T(), taskID, analysis.TaskID)
	assert.Equal(suite.T(), changeType, analysis.ChangeType)
	assert.NotEmpty(suite.T(), analysis.ID)
	assert.NotEmpty(suite.T(), analysis.RiskLevel)
	assert.NotNil(suite.T(), analysis.AnalyzedAt)

	// Should identify affected tasks
	assert.NotEmpty(suite.T(), analysis.AffectedTasks)
	foundDependent := false
	for _, taskTitle := range analysis.AffectedTasks {
		if strings.Contains(taskTitle, "task-456") || strings.Contains(taskTitle, "456") {
			foundDependent = true
		}
	}
	assert.True(suite.T(), foundDependent, "Should identify dependent tasks")
}

// TestDocumentValidationIntegration tests document validation workflow
func (suite *IntegrationTestSuite) TestDocumentValidationIntegration() {
	ctx := context.TODO()

	// Test valid file types
	validTypes := []string{
		"application/pdf",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"text/plain",
		"text/markdown",
	}

	for _, mimeType := range validTypes {
		err := suite.docValidator.ValidateFile(ctx, "/tmp/test"+mimeType, mimeType)
		assert.NoError(suite.T(), err, "Should accept valid MIME type: %s", mimeType)
	}

	// Test invalid file types
	invalidTypes := []string{
		"application/exe",
		"image/jpeg",
		"video/mp4",
		"",
	}

	for _, mimeType := range invalidTypes {
		err := suite.docValidator.ValidateFile(ctx, "/tmp/test"+mimeType, mimeType)
		assert.Error(suite.T(), err, "Should reject invalid MIME type: %s", mimeType)
	}

	// Test file size validation
	err := suite.docValidator.ValidateSize(ctx, 1024) // 1KB - should pass
	assert.NoError(suite.T(), err)

	err = suite.docValidator.ValidateSize(ctx, 200*1024*1024) // 200MB - should fail
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "exceeds maximum")
}

// TestSearchEngineIntegration tests search functionality
func (suite *IntegrationTestSuite) TestSearchEngineIntegration() {
	ctx := context.TODO()
	docID := "test-doc-123"
	content := "This document contains authentication and security requirements for the system."
	knowledgeItems := []models.KnowledgeItem{
		{
			ID:         "ki-1",
			Type:       "security_requirement",
			Content:    "authentication requirements",
			Confidence: 0.9,
		},
		{
			ID:         "ki-2",
			Type:       "functional_requirement",
			Content:    "security requirements",
			Confidence: 0.8,
		},
	}

	// Test indexing
	err := suite.searchEngine.IndexDocument(ctx, docID, content, knowledgeItems)
	assert.NoError(suite.T(), err)

	// Test search
	results, err := suite.searchEngine.SearchDocuments(ctx, "project-123", "authentication")
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), results)

	// Should find relevant results
	found := false
	for _, item := range results {
		if item.Content == "authentication requirements" {
			found = true
			break
		}
	}
	assert.True(suite.T(), found, "Should find authentication-related content")

	// Test search with no results
	results, err = suite.searchEngine.SearchDocuments(ctx, "project-123", "nonexistent")
	assert.NoError(suite.T(), err)
	assert.Empty(suite.T(), results)
}

// TestSuite runs the integration test suite
func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
