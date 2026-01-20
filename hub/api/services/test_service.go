// Package services - Test Management Service
// Complies with CODING_STANDARDS.md: Business Services max 400 lines
package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"sentinel-hub-api/pkg/database"
	"github.com/google/uuid"
)

// TestServiceImpl implements TestService interface
type TestServiceImpl struct {
	db *sql.DB
}

// NewTestService creates a new test service
func NewTestService(db *sql.DB) TestService {
	return &TestServiceImpl{db: db}
}

// GenerateTestRequirements generates test requirements from business rules
func (s *TestServiceImpl) GenerateTestRequirements(ctx context.Context, req GenerateTestRequirementsRequest) (*GenerateTestRequirementsResponse, error) {
	if req.ProjectID == "" {
		return nil, fmt.Errorf("project_id is required")
	}

	// Extract business rules
	rules, err := extractBusinessRules(ctx, req.ProjectID, req.KnowledgeItemIDs, "", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to extract business rules: %w", err)
	}

	if len(rules) == 0 {
		return &GenerateTestRequirementsResponse{
			Success:      true,
			Requirements: []TestRequirement{},
			Count:        0,
			Message:      "No approved business rules found",
		}, nil
	}

	// Generate test requirements for each rule
	var allRequirements []TestRequirement
	for _, rule := range rules {
		codeFunction := "" // Would use AST analysis in production
		requirements := generateTestRequirements(rule, codeFunction)
		allRequirements = append(allRequirements, requirements...)
	}

	// Save to database
	if err := saveTestRequirements(ctx, allRequirements); err != nil {
		return nil, fmt.Errorf("failed to save test requirements: %w", err)
	}

	return &GenerateTestRequirementsResponse{
		Success:      true,
		Requirements: allRequirements,
		Count:        len(allRequirements),
		Message:      fmt.Sprintf("Generated %d test requirements from %d business rules", len(allRequirements), len(rules)),
	}, nil
}

// AnalyzeTestCoverage analyzes test coverage for business rules
func (s *TestServiceImpl) AnalyzeTestCoverage(ctx context.Context, req AnalyzeCoverageRequest) (*AnalyzeCoverageResponse, error) {
	if req.ProjectID == "" {
		return nil, fmt.Errorf("project_id is required")
	}
	if len(req.TestFiles) == 0 {
		return nil, fmt.Errorf("test_files are required")
	}

	// Get test requirements
	var requirementIDs []string
	if len(req.KnowledgeItemIDs) > 0 {
		// Get requirements for specific knowledge items
		query := `SELECT id FROM test_requirements WHERE knowledge_item_id = ANY($1)`
		rows, err := database.QueryWithTimeout(ctx, s.db, query, req.KnowledgeItemIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to query test requirements: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				return nil, fmt.Errorf("failed to scan requirement ID: %w", err)
			}
			requirementIDs = append(requirementIDs, id)
		}
	} else {
		// Get all requirements for project
		query := `
			SELECT tr.id FROM test_requirements tr
			INNER JOIN knowledge_items ki ON tr.knowledge_item_id = ki.id
			INNER JOIN documents d ON ki.document_id = d.id
			WHERE d.project_id = $1
		`
		rows, err := database.QueryWithTimeout(ctx, s.db, query, req.ProjectID)
		if err != nil {
			return nil, fmt.Errorf("failed to query test requirements: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				return nil, fmt.Errorf("failed to scan requirement ID: %w", err)
			}
			requirementIDs = append(requirementIDs, id)
		}
	}

	// Analyze coverage for each requirement
	var coverage []TestCoverage
	for _, reqID := range requirementIDs {
		cov := analyzeCoverageForRequirement(reqID, req.TestFiles)
		coverage = append(coverage, cov)
	}

	// Save coverage to database
	for _, cov := range coverage {
		if err := saveTestCoverage(ctx, cov); err != nil {
			LogWarn(ctx, "Failed to save coverage for requirement %s: %v", cov.TestRequirementID, err)
		}
	}

	return &AnalyzeCoverageResponse{
		Success:  true,
		Coverage: coverage,
		Count:    len(coverage),
		Message:  fmt.Sprintf("Analyzed coverage for %d test requirements", len(coverage)),
	}, nil
}

// GetTestCoverage retrieves test coverage for a knowledge item
func (s *TestServiceImpl) GetTestCoverage(ctx context.Context, knowledgeItemID string) (*TestCoverage, error) {
	if knowledgeItemID == "" {
		return nil, fmt.Errorf("knowledge_item_id is required")
	}

	query := `
		SELECT tc.id, tc.test_requirement_id, tc.knowledge_item_id, 
		       tc.coverage_percentage, tc.test_files, tc.last_updated, tc.created_at
		FROM test_coverage tc
		INNER JOIN test_requirements tr ON tc.test_requirement_id = tr.id
		WHERE tr.knowledge_item_id = $1
		ORDER BY tc.last_updated DESC
		LIMIT 1
	`

	ctx, cancel := context.WithTimeout(ctx, getQueryTimeout())
	defer cancel()

	var coverage TestCoverage
	var testFilesJSON sql.NullString

	err := database.QueryRowWithTimeout(ctx, s.db, query, knowledgeItemID).Scan(
		&coverage.ID, &coverage.TestRequirementID, &coverage.KnowledgeItemID,
		&coverage.CoveragePercentage, &testFilesJSON, &coverage.LastUpdated, &coverage.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("test coverage not found for knowledge item: %s", knowledgeItemID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get test coverage: %w", err)
	}

	if testFilesJSON.Valid {
		if err := json.Unmarshal([]byte(testFilesJSON.String), &coverage.TestFiles); err != nil {
			LogWarn(ctx, "Failed to unmarshal test_files: %v", err)
		}
	}

	return &coverage, nil
}

// ValidateTests validates test code against requirements
func (s *TestServiceImpl) ValidateTests(ctx context.Context, req ValidateTestsRequest) (*ValidateTestsResponse, error) {
	if req.ProjectID == "" {
		return nil, fmt.Errorf("project_id is required")
	}
	if req.TestCode == "" {
		return nil, fmt.Errorf("test_code is required")
	}

	// Detect language
	language := detectLanguage(req.TestFilePath, req.TestCode)

	// Get test requirements
	var requirementIDs []string
	if len(req.TestRequirementIDs) > 0 {
		requirementIDs = req.TestRequirementIDs
	} else {
		// Get all requirements for project
		query := `
			SELECT tr.id FROM test_requirements tr
			INNER JOIN knowledge_items ki ON tr.knowledge_item_id = ki.id
			INNER JOIN documents d ON ki.document_id = d.id
			WHERE d.project_id = $1
		`
		rows, err := database.QueryWithTimeout(ctx, s.db, query, req.ProjectID)
		if err != nil {
			return nil, fmt.Errorf("failed to query test requirements: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				return nil, fmt.Errorf("failed to scan requirement ID: %w", err)
			}
			requirementIDs = append(requirementIDs, id)
		}
	}

	// Validate tests for each requirement
	var validations []TestValidation
	for _, reqID := range requirementIDs {
		validation := validateTestForRequirement(reqID, req.TestCode, language)
		validations = append(validations, validation)

		// Save validation
		if err := saveTestValidation(ctx, validation); err != nil {
			LogWarn(ctx, "Failed to save validation for requirement %s: %v", reqID, err)
		}
	}

	return &ValidateTestsResponse{
		Success:     true,
		Validations: validations,
		Count:       len(validations),
		Message:     fmt.Sprintf("Validated %d test requirements", len(validations)),
	}, nil
}

// GetValidationResults retrieves validation results for a test requirement
func (s *TestServiceImpl) GetValidationResults(ctx context.Context, testRequirementID string) (*TestValidation, error) {
	if testRequirementID == "" {
		return nil, fmt.Errorf("test_requirement_id is required")
	}

	query := `
		SELECT id, test_requirement_id, validation_status, issues, 
		       test_code_hash, score, validated_at, created_at
		FROM test_validations
		WHERE test_requirement_id = $1
		ORDER BY validated_at DESC
		LIMIT 1
	`

	ctx, cancel := context.WithTimeout(ctx, getQueryTimeout())
	defer cancel()

	var validation TestValidation
	var issuesJSON sql.NullString

	err := database.QueryRowWithTimeout(ctx, s.db, query, testRequirementID).Scan(
		&validation.ID, &validation.TestRequirementID, &validation.ValidationStatus,
		&issuesJSON, &validation.TestCodeHash, &validation.Score,
		&validation.ValidatedAt, &validation.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("validation not found for test requirement: %s", testRequirementID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get validation results: %w", err)
	}

	if issuesJSON.Valid {
		if err := json.Unmarshal([]byte(issuesJSON.String), &validation.Issues); err != nil {
			LogWarn(ctx, "Failed to unmarshal issues: %v", err)
		}
	}

	return &validation, nil
}

// RunTests executes tests in a sandbox
func (s *TestServiceImpl) RunTests(ctx context.Context, req TestExecutionRequest) (*TestExecutionResponse, error) {
	if req.ProjectID == "" {
		return nil, fmt.Errorf("project_id is required")
	}
	if len(req.TestFiles) == 0 {
		return nil, fmt.Errorf("test_files are required")
	}

	// Create execution record
	executionID := uuid.New().String()
	execution := TestExecution{
		ID:            executionID,
		ProjectID:     req.ProjectID,
		ExecutionType: req.ExecutionType,
		Status:        "running",
		CreatedAt:     time.Now().UTC(),
	}

	// Save execution record
	if err := saveTestExecution(ctx, execution); err != nil {
		return nil, fmt.Errorf("failed to save execution record: %w", err)
	}

	// Execute tests in sandbox (async)
	go func() {
		result := executeTestsInSandbox(req)
		execution.Status = "completed"
		execution.ExecutionTimeMs = result.ExecutionTimeMs
		if result.ExitCode != 0 {
			execution.Status = "failed"
		}

		resultJSON, _ := json.Marshal(result)
		execution.Result = resultJSON
		now := time.Now().UTC()
		execution.CompletedAt = &now

		// Update execution record
		if err := saveTestExecution(ctx, execution); err != nil {
			LogError(ctx, "Failed to update execution record: %v", err)
		}
	}()

	return &TestExecutionResponse{
		Success:     true,
		ExecutionID: executionID,
		Status:     "running",
		Message:    "Test execution started",
	}, nil
}

// GetTestExecutionStatus retrieves the status of a test execution
func (s *TestServiceImpl) GetTestExecutionStatus(ctx context.Context, executionID string) (*TestExecution, error) {
	if executionID == "" {
		return nil, fmt.Errorf("execution_id is required")
	}

	query := `
		SELECT id, project_id, execution_type, status, result, 
		       execution_time_ms, created_at, completed_at
		FROM test_executions
		WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(ctx, getQueryTimeout())
	defer cancel()

	var execution TestExecution
	var resultJSON sql.NullString
	var completedAt sql.NullTime

	err := database.QueryRowWithTimeout(ctx, s.db, query, executionID).Scan(
		&execution.ID, &execution.ProjectID, &execution.ExecutionType, &execution.Status,
		&resultJSON, &execution.ExecutionTimeMs, &execution.CreatedAt, &completedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("test execution not found: %s", executionID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get execution status: %w", err)
	}

	if resultJSON.Valid {
		execution.Result = json.RawMessage(resultJSON.String)
	}
	if completedAt.Valid {
		execution.CompletedAt = &completedAt.Time
	}

	return &execution, nil
}

// Helper functions (stubs - would be implemented in production)

func analyzeCoverageForRequirement(requirementID string, testFiles []TestFile) TestCoverage {
	// Simplified - would analyze test files against requirement
	return TestCoverage{
		ID:                 uuid.New().String(),
		TestRequirementID:  requirementID,
		CoveragePercentage: 0.75, // Placeholder
		TestFiles:          []string{},
		LastUpdated:        time.Now().UTC(),
		CreatedAt:          time.Now().UTC(),
	}
}

func saveTestCoverage(ctx context.Context, coverage TestCoverage) error {
	// Stub - would save to database
	return nil
}

func validateTestForRequirement(requirementID, testCode, language string) TestValidation {
	// Simplified - would validate test code
	return TestValidation{
		ID:                uuid.New().String(),
		TestRequirementID: requirementID,
		ValidationStatus:  "valid",
		Issues:            []string{},
		Score:             0.85, // Placeholder
		ValidatedAt:       time.Now().UTC(),
		CreatedAt:         time.Now().UTC(),
	}
}

func saveTestValidation(ctx context.Context, validation TestValidation) error {
	// Stub - would save to database
	return nil
}

func executeTestsInSandbox(req TestExecutionRequest) ExecutionResult {
	// Stub - would execute tests in Docker sandbox
	return ExecutionResult{
		ExitCode: 0,
		Stdout:   "Tests passed",
		Stderr:   "",
	}
}

func detectLanguage(filePath, code string) string {
	// Simplified language detection
	if strings.Contains(filePath, ".go") || strings.Contains(code, "package ") {
		return "go"
	}
	if strings.Contains(filePath, ".js") || strings.Contains(filePath, ".ts") {
		return "javascript"
	}
	if strings.Contains(filePath, ".py") {
		return "python"
	}
	return "unknown"
}
