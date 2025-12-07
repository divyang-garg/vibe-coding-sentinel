// Phase 10C: Test Validator
// Validates test correctness and completeness

package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// TestValidation represents validation results for a test requirement
type TestValidation struct {
	ID                string    `json:"id"`
	TestRequirementID string    `json:"test_requirement_id"`
	ValidationStatus  string    `json:"validation_status"` // valid, invalid, incomplete
	Issues            []string  `json:"issues"`
	TestCodeHash      string    `json:"test_code_hash,omitempty"`
	Score             float64   `json:"score"` // 0.0 to 1.0
	ValidatedAt       time.Time `json:"validated_at"`
	CreatedAt         time.Time `json:"created_at"`
}

// ValidateTestsRequest represents the request to validate tests
type ValidateTestsRequest struct {
	ProjectID         string   `json:"projectId"`
	TestRequirementIDs []string `json:"testRequirementIds,omitempty"` // Optional: specific requirements
	TestCode          string   `json:"testCode"` // Test code to validate
	TestFilePath      string   `json:"testFilePath"`
}

// ValidateTestsResponse represents the response
type ValidateTestsResponse struct {
	Success    bool             `json:"success"`
	Validations []TestValidation `json:"validations"`
	Count      int              `json:"count"`
	Message    string           `json:"message,omitempty"`
}

// validateTestStructure validates the structure of a test function
func validateTestStructure(testCode string, language string) (bool, []string) {
	var issues []string
	isValid := true
	
	lines := strings.Split(testCode, "\n")
	hasAssertion := false
	
	// Look for test structure patterns based on language
	switch language {
	case "go":
		// Go tests: setup, execution, assertion
		for _, line := range lines {
			lineLower := strings.ToLower(line)
			if strings.Contains(lineLower, "assert") || strings.Contains(lineLower, "if") || 
			   strings.Contains(lineLower, "require.") || strings.Contains(lineLower, "assert.") {
				hasAssertion = true
			}
		}
		
		if !hasAssertion {
			issues = append(issues, "Missing assertions - test does not verify expected behavior")
			isValid = false
		}
		
	case "javascript", "typescript":
		// JS/TS tests: setup, execution, assertion
		for _, line := range lines {
			lineLower := strings.ToLower(line)
			if strings.Contains(lineLower, "expect(") || strings.Contains(lineLower, "assert(") {
				hasAssertion = true
			}
		}
		
		if !hasAssertion {
			issues = append(issues, "Missing assertions - test does not verify expected behavior")
			isValid = false
		}
		
	case "python":
		// Python tests: setup, execution, assertion
		for _, line := range lines {
			lineLower := strings.ToLower(line)
			if strings.Contains(lineLower, "assert ") || strings.Contains(lineLower, "self.assert") {
				hasAssertion = true
			}
		}
		
		if !hasAssertion {
			issues = append(issues, "Missing assertions - test does not verify expected behavior")
			isValid = false
		}
	}
	
	// Check for shared state (test isolation)
	if strings.Contains(testCode, "global") || strings.Contains(testCode, "static") {
		issues = append(issues, "Potential shared state - test may not be isolated")
		isValid = false
	}
	
	return isValid, issues
}

// analyzeAssertions analyzes assertions in test code
func analyzeAssertions(testCode string, language string) (bool, []string) {
	var issues []string
	isValid := true
	
	lines := strings.Split(testCode, "\n")
	hasStrongAssertion := false
	
	// Look for strong assertions (not just null checks)
	for _, line := range lines {
		lineLower := strings.ToLower(line)
		
		// Weak assertions (just checking non-null)
		if strings.Contains(lineLower, "!= nil") || strings.Contains(lineLower, "!== null") ||
		   strings.Contains(lineLower, "is not none") {
			// Check if there are stronger assertions nearby
			hasStrongAssertion = false
		}
		
		// Strong assertions (checking actual values)
		if strings.Contains(lineLower, "==") || strings.Contains(lineLower, "===") ||
		   strings.Contains(lineLower, "equals") || strings.Contains(lineLower, "equal") {
			hasStrongAssertion = true
		}
	}
	
	if !hasStrongAssertion {
		issues = append(issues, "Weak assertions detected - test only checks for null/non-null, not actual values")
		isValid = false
	}
	
	return isValid, issues
}

// checkCompleteness checks if test covers all requirements
func checkCompleteness(testCode string, testRequirement TestRequirement) (bool, []string) {
	var issues []string
	isComplete := true
	
	testCodeLower := strings.ToLower(testCode)
	requirementDescLower := strings.ToLower(testRequirement.Description)
	
	// Extract keywords from requirement
	keywords := extractKeywords(requirementDescLower)
	
	// Check if test code mentions requirement keywords
	matchedKeywords := 0
	for _, keyword := range keywords {
		if strings.Contains(testCodeLower, keyword) {
			matchedKeywords++
		}
	}
	
	// If less than 50% of keywords match, test may not cover requirement
	if len(keywords) > 0 && float64(matchedKeywords)/float64(len(keywords)) < 0.5 {
		issues = append(issues, fmt.Sprintf("Test may not fully cover requirement: only %d/%d keywords matched", matchedKeywords, len(keywords)))
		isComplete = false
	}
	
	// Check requirement type coverage
	switch testRequirement.RequirementType {
	case "happy_path":
		if !strings.Contains(testCodeLower, "success") && !strings.Contains(testCodeLower, "valid") {
			issues = append(issues, "Happy path test may be missing - no success/valid scenarios found")
			isComplete = false
		}
	case "error_case":
		if !strings.Contains(testCodeLower, "error") && !strings.Contains(testCodeLower, "fail") &&
		   !strings.Contains(testCodeLower, "invalid") && !strings.Contains(testCodeLower, "exception") {
			issues = append(issues, "Error case test may be missing - no error/failure scenarios found")
			isComplete = false
		}
	case "edge_case":
		if !strings.Contains(testCodeLower, "edge") && !strings.Contains(testCodeLower, "boundary") &&
		   !strings.Contains(testCodeLower, "limit") && !strings.Contains(testCodeLower, "max") &&
		   !strings.Contains(testCodeLower, "min") {
			issues = append(issues, "Edge case test may be missing - no boundary/limit scenarios found")
			isComplete = false
		}
	}
	
	return isComplete, issues
}

// calculateValidationScore calculates a validation score (0.0 to 1.0)
func calculateValidationScore(structureValid bool, assertionsValid bool, completenessValid bool, issues []string) float64 {
	score := 1.0
	
	// Deduct points for each issue
	deductionPerIssue := 0.1
	score -= float64(len(issues)) * deductionPerIssue
	
	// Deduct points for invalid structure
	if !structureValid {
		score -= 0.3
	}
	
	// Deduct points for weak assertions
	if !assertionsValid {
		score -= 0.2
	}
	
	// Deduct points for incomplete coverage
	if !completenessValid {
		score -= 0.2
	}
	
	// Ensure score is between 0.0 and 1.0
	if score < 0.0 {
		score = 0.0
	}
	if score > 1.0 {
		score = 1.0
	}
	
	return score
}

// getTestRequirement gets a test requirement by ID
func getTestRequirement(ctx context.Context, requirementID string) (*TestRequirement, error) {
	query := `SELECT id, knowledge_item_id, rule_title, requirement_type, description, 
	                 code_function, priority, created_at, updated_at
	          FROM test_requirements
	          WHERE id = $1`
	
	row := queryRowWithTimeout(ctx, query, requirementID)
	
	var req TestRequirement
	err := row.Scan(
		&req.ID, &req.KnowledgeItemID, &req.RuleTitle, &req.RequirementType,
		&req.Description, &req.CodeFunction, &req.Priority,
		&req.CreatedAt, &req.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("test requirement not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query test requirement: %w", err)
	}
	
	return &req, nil
}

// saveTestValidation saves validation results to database
func saveTestValidation(ctx context.Context, validation TestValidation) error {
	// Convert issues []string to JSONB
	issuesJSON, err := json.Marshal(validation.Issues)
	if err != nil {
		return fmt.Errorf("failed to marshal issues: %w", err)
	}
	
	query := `
		INSERT INTO test_validations 
		(id, test_requirement_id, validation_status, issues, test_code_hash, score, validated_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET
			validation_status = EXCLUDED.validation_status,
			issues = EXCLUDED.issues,
			test_code_hash = EXCLUDED.test_code_hash,
			score = EXCLUDED.score,
			validated_at = EXCLUDED.validated_at
	`
	
	_, err = execWithTimeout(ctx, query,
		validation.ID, validation.TestRequirementID, validation.ValidationStatus,
		issuesJSON, validation.TestCodeHash, validation.Score, validation.ValidatedAt, validation.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save test validation: %w", err)
	}
	
	return nil
}

// detectLanguage detects language from file path and code content
func detectLanguage(testFilePath string, testCode string) string {
	// First try file extension
	if ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(testFilePath), ".")); ext != "" {
		switch ext {
		case "go":
			return "go"
		case "js", "jsx":
			return "javascript"
		case "ts", "tsx":
			return "typescript"
		case "py":
			return "python"
		}
	}
	
	// Fallback: detect from code content
	testCodeLower := strings.ToLower(testCode)
	if strings.Contains(testCodeLower, "package main") || strings.Contains(testCodeLower, "func test") {
		return "go"
	}
	if strings.Contains(testCodeLower, "describe(") || strings.Contains(testCodeLower, "it(") {
		return "javascript"
	}
	if strings.Contains(testCodeLower, "def test_") {
		return "python"
	}
	
	return "go" // Ultimate fallback
}

// validateTestsHandler handles the API request to validate tests
func validateTestsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var req ValidateTestsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}
	
	// Validate required fields
	if req.ProjectID == "" {
		http.Error(w, "projectId is required", http.StatusBadRequest)
		return
	}
	if req.TestCode == "" {
		http.Error(w, "testCode is required", http.StatusBadRequest)
		return
	}
	
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	
	// Calculate test code hash for caching
	testCodeHash := fmt.Sprintf("%x", sha256.Sum256([]byte(req.TestCode)))
	
	// Detect language from file path and code content
	language := detectLanguage(req.TestFilePath, req.TestCode)
	
	// Get test requirements to validate against
	var requirements []TestRequirement
	if len(req.TestRequirementIDs) > 0 {
		// Get specific requirements
		for _, reqID := range req.TestRequirementIDs {
			req, err := getTestRequirement(ctx, reqID)
			if err != nil {
				log.Printf("Error getting test requirement %s: %v", reqID, err)
				continue
			}
			requirements = append(requirements, *req)
		}
	} else {
		// Get all requirements for project (would need to query via knowledge items)
		// For now, validate structure and assertions only
		requirements = []TestRequirement{}
	}
	
	var validations []TestValidation
	now := time.Now()
	
	// If no specific requirements, validate test structure and assertions only
	if len(requirements) == 0 {
		structureValid, structureIssues := validateTestStructure(req.TestCode, language)
		assertionsValid, assertionIssues := analyzeAssertions(req.TestCode, language)
		
		allIssues := append(structureIssues, assertionIssues...)
		score := calculateValidationScore(structureValid, assertionsValid, true, allIssues)
		
		status := "valid"
		if !structureValid || !assertionsValid {
			status = "invalid"
		}
		
			validation := TestValidation{
				ID:               uuid.New().String(),
				TestRequirementID: "", // No specific requirement
				ValidationStatus:  status,
				Issues:           allIssues,
				TestCodeHash:      testCodeHash,
				Score:            score,
				ValidatedAt:      now,
				CreatedAt:        now,
			}
		
		validations = append(validations, validation)
	} else {
		// Validate against each requirement
		for _, requirement := range requirements {
			structureValid, structureIssues := validateTestStructure(req.TestCode, language)
			assertionsValid, assertionIssues := analyzeAssertions(req.TestCode, language)
			completenessValid, completenessIssues := checkCompleteness(req.TestCode, requirement)
			
			allIssues := append(structureIssues, assertionIssues...)
			allIssues = append(allIssues, completenessIssues...)
			score := calculateValidationScore(structureValid, assertionsValid, completenessValid, allIssues)
			
			status := "valid"
			if !structureValid || !assertionsValid {
				status = "invalid"
			} else if !completenessValid {
				status = "incomplete"
			}
			
			validation := TestValidation{
				ID:               uuid.New().String(),
				TestRequirementID: requirement.ID,
				ValidationStatus:  status,
				Issues:           allIssues,
				TestCodeHash:      testCodeHash,
				Score:            score,
				ValidatedAt:      now,
				CreatedAt:        now,
			}
			
			// Save to database
			if err := saveTestValidation(ctx, validation); err != nil {
				log.Printf("Error saving validation: %v", err)
				continue
			}
			
			validations = append(validations, validation)
		}
	}
	
	response := ValidateTestsResponse{
		Success:     true,
		Validations: validations,
		Count:       len(validations),
		Message:     fmt.Sprintf("Validated %d test requirement(s)", len(validations)),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// getValidationHandler handles GET request to retrieve validation for a test requirement
func getValidationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	testRequirementID := chi.URLParam(r, "test_requirement_id")
	if testRequirementID == "" {
		http.Error(w, "test_requirement_id is required", http.StatusBadRequest)
		return
	}
	
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	
	query := `
		SELECT id, test_requirement_id, validation_status, issues, score, validated_at, created_at
		FROM test_validations
		WHERE test_requirement_id = $1
		ORDER BY validated_at DESC
		LIMIT 1
	`
	
	row := queryRowWithTimeout(ctx, query, testRequirementID)
	
	var validation TestValidation
	var issuesStr sql.NullString
	
	err := row.Scan(
		&validation.ID, &validation.TestRequirementID, &validation.ValidationStatus,
		&issuesStr, &validation.Score, &validation.ValidatedAt, &validation.CreatedAt,
	)
	
	if err == sql.ErrNoRows {
		http.Error(w, "Validation not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("Error querying validation: %v", err)
		http.Error(w, fmt.Sprintf("Failed to query validation: %v", err), http.StatusInternalServerError)
		return
	}
	
	if issuesStr.Valid {
		// Parse issues array (stored as PostgreSQL array)
		validation.Issues = strings.Split(strings.Trim(issuesStr.String, "{}"), ",")
		for i, issue := range validation.Issues {
			validation.Issues[i] = strings.Trim(issue, "\"")
		}
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(validation)
}

