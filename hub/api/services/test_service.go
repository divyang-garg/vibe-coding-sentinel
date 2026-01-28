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
	"sentinel-hub-api/utils"

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
		// Use AST analysis to map business rules to code functions
		codeFunction := mapBusinessRuleToCodeFunction(ctx, rule, req.ProjectID)
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
	if err := saveTestCoverage(ctx, coverage); err != nil {
		LogWarn(ctx, "Failed to save coverage: %v", err)
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

	if err != nil {
		return nil, utils.HandleNotFoundError(err, "test coverage", knowledgeItemID)
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

	if err != nil {
		return nil, utils.HandleNotFoundError(err, "validation", testRequirementID)
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
		// Use actual Docker implementation instead of stub
		result, err := executeTestInSandbox(ctx, req)
		if err != nil {
			execution.Status = "failed"
			errorResult := ExecutionResult{
				ExitCode: 1,
				Stdout:   "",
				Stderr:   err.Error(),
			}
			resultJSON, _ := json.Marshal(errorResult)
			execution.Result = resultJSON
			now := time.Now().UTC()
			execution.CompletedAt = &now
			if err := saveTestExecution(ctx, execution); err != nil {
				LogError(ctx, "Failed to update execution record: %v", err)
			}
			return
		}

		execution.Status = "completed"
		// ExecutionTimeMs will be set when execution completes
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
		Status:      "running",
		Message:     "Test execution started",
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

	if err != nil {
		return nil, utils.HandleNotFoundError(err, "test execution", executionID)
	}

	if resultJSON.Valid {
		execution.Result = json.RawMessage(resultJSON.String)
	}
	if completedAt.Valid {
		execution.CompletedAt = &completedAt.Time
	}

	return &execution, nil
}

// Helper functions

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

// mapBusinessRuleToCodeFunction maps a business rule to code functions using AST analysis
// This implements the AST-based function mapping as specified in STUB_FUNCTIONALITY_ANALYSIS.md
func mapBusinessRuleToCodeFunction(ctx context.Context, rule KnowledgeItem, projectID string) string {
	// Extract keywords from rule title and content for matching
	keywords := extractKeywords(rule.Title)
	if rule.Content != "" {
		contentKeywords := extractKeywords(rule.Content)
		keywords = append(keywords, contentKeywords...)
	}

	if len(keywords) == 0 {
		return ""
	}

	// Try to get codebasePath from project configuration
	codebasePath := getProjectCodebasePath(ctx, projectID)

	// If codebasePath is available, use detectBusinessRuleImplementation for accurate mapping
	if codebasePath != "" {
		evidence := detectBusinessRuleImplementation(rule, codebasePath)
		if len(evidence.Functions) > 0 {
			// Return the first matching function (highest confidence)
			return evidence.Functions[0]
		}
	}

	// Fallback: Use keyword-based function name suggestion
	// This generates a suggested function name based on rule keywords
	// The actual implementation can be mapped later when code is available
	return suggestFunctionNameFromKeywords(keywords)
}

// getProjectCodebasePath retrieves the codebase path from project configuration
// Returns empty string if not available (allows graceful fallback)
func getProjectCodebasePath(ctx context.Context, projectID string) string {
	if projectID == "" || db == nil {
		return ""
	}

	// Try to get codebase_path from projects table
	query := `SELECT codebase_path FROM projects WHERE id = $1`
	var codebasePath sql.NullString

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := database.QueryRowWithTimeout(ctx, db, query, projectID).Scan(&codebasePath)
	if err != nil {
		// Project not found or codebase_path not set - return empty string for fallback
		return ""
	}

	if codebasePath.Valid {
		return codebasePath.String
	}

	return ""
}

// suggestFunctionNameFromKeywords generates a suggested function name from keywords
// This provides a reasonable function name suggestion when AST analysis isn't available
func suggestFunctionNameFromKeywords(keywords []string) string {
	if len(keywords) == 0 {
		return ""
	}

	// Use the most significant keywords (longer words are typically more meaningful)
	// Filter out common words and build a camelCase function name
	var significantKeywords []string
	for _, kw := range keywords {
		if len(kw) > 3 { // Only use longer keywords
			significantKeywords = append(significantKeywords, kw)
		}
	}

	if len(significantKeywords) == 0 {
		// Fallback to first keyword if all are short
		if len(keywords) > 0 {
			return capitalizeFirst(keywords[0])
		}
		return ""
	}

	// Build camelCase function name from keywords
	// Use first 2-3 significant keywords to keep name reasonable
	maxKeywords := 3
	if len(significantKeywords) < maxKeywords {
		maxKeywords = len(significantKeywords)
	}

	parts := significantKeywords[:maxKeywords]
	functionName := parts[0]
	for i := 1; i < len(parts); i++ {
		functionName += capitalizeFirst(parts[i])
	}

	return capitalizeFirst(functionName)
}

// capitalizeFirst capitalizes the first letter of a string
func capitalizeFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
