// Test Coverage Tracker - Main Types and HTTP Handlers
// Tracks test coverage for business rules
// Complies with CODING_STANDARDS.md: Business Services max 400 lines

package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"sentinel-hub-api/pkg/database"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// TestFile represents a test file with content (matching AST/Security analyzer pattern)
type TestFile struct {
	Path     string `json:"path"`
	Content  string `json:"content"`
	Language string `json:"language,omitempty"`
}

// TestCoverage represents test coverage for a business rule
type TestCoverage struct {
	ID                 string    `json:"id"`
	TestRequirementID  string    `json:"test_requirement_id"`
	KnowledgeItemID    string    `json:"knowledge_item_id"`
	CoveragePercentage float64   `json:"coverage_percentage"` // 0.0 to 1.0
	TestFiles          []string  `json:"test_files"`
	LastUpdated        time.Time `json:"last_updated"`
	CreatedAt          time.Time `json:"created_at"`
}

// AnalyzeCoverageRequest represents the request to analyze test coverage
type AnalyzeCoverageRequest struct {
	ProjectID        string     `json:"project_id"`
	TestFiles        []TestFile `json:"testFiles"`                  // Test files with content (changed from []string)
	KnowledgeItemIDs []string   `json:"knowledgeItemIds,omitempty"` // Optional: specific rules
}

// AnalyzeCoverageResponse represents the response
type AnalyzeCoverageResponse struct {
	Success  bool           `json:"success"`
	Coverage []TestCoverage `json:"coverage"`
	Count    int            `json:"count"`
	Message  string         `json:"message,omitempty"`
}

// analyzeCoverageHandler handles the API request to analyze test coverage
func analyzeCoverageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req AnalyzeCoverageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.ProjectID == "" {
		http.Error(w, "projectId is required", http.StatusBadRequest)
		return
	}
	if len(req.TestFiles) == 0 {
		http.Error(w, "testFiles are required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// Extract business rules
	rules, err := extractBusinessRules(ctx, req.ProjectID, req.KnowledgeItemIDs, "", nil)
	if err != nil {
		log.Printf("Error extracting business rules: %v", err)
		http.Error(w, fmt.Sprintf("Failed to extract business rules: %v", err), http.StatusInternalServerError)
		return
	}

	if len(rules) == 0 {
		response := AnalyzeCoverageResponse{
			Success:  true,
			Coverage: []TestCoverage{},
			Count:    0,
			Message:  "No approved business rules found",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Get test requirements for the rules
	var knowledgeItemIDs []string
	for _, rule := range rules {
		knowledgeItemIDs = append(knowledgeItemIDs, rule.ID)
	}

	testRequirements, err := getTestRequirementsForRules(ctx, knowledgeItemIDs)
	if err != nil {
		log.Printf("Error getting test requirements: %v", err)
		http.Error(w, fmt.Sprintf("Failed to get test requirements: %v", err), http.StatusInternalServerError)
		return
	}

	// Discover test files
	testFiles := discoverTestFiles(req.TestFiles)

	// Parse test files and map to business rules
	ruleToTests := make(map[string][]string)
	for _, testFile := range testFiles {
		// Use test file content directly (provided by Agent)
		testFunctions, err := parseTestFile(testFile.Path, testFile.Content)
		if err != nil {
			log.Printf("Error parsing test file %s: %v", testFile.Path, err)
			continue
		}

		// Map tests to rules
		fileRuleToTests := mapTestsToBusinessRules(testFunctions, testFile.Path, rules)
		for ruleID, tests := range fileRuleToTests {
			ruleToTests[ruleID] = append(ruleToTests[ruleID], tests...)
		}
	}

	// Calculate coverage for each rule
	var coverage []TestCoverage
	now := time.Now()

	for _, rule := range rules {
		coveragePct := calculateCoverage(rule.ID, testRequirements, ruleToTests, testFiles)
		testFilesForRule := ruleToTests[rule.ID]

		// Get test requirement ID (use first one for this rule)
		var testReqID string
		for _, req := range testRequirements {
			if req.KnowledgeItemID == rule.ID {
				testReqID = req.ID
				break
			}
		}

		if testReqID == "" {
			continue // Skip if no test requirement exists
		}

		coverage = append(coverage, TestCoverage{
			ID:                 uuid.New().String(),
			TestRequirementID:  testReqID,
			KnowledgeItemID:    rule.ID,
			CoveragePercentage: coveragePct,
			TestFiles:          testFilesForRule,
			LastUpdated:        now,
			CreatedAt:          now,
		})
	}

	// Save to database
	if err := saveTestCoverage(ctx, coverage); err != nil {
		log.Printf("Error saving test coverage: %v", err)
		http.Error(w, fmt.Sprintf("Failed to save test coverage: %v", err), http.StatusInternalServerError)
		return
	}

	response := AnalyzeCoverageResponse{
		Success:  true,
		Coverage: coverage,
		Count:    len(coverage),
		Message:  fmt.Sprintf("Analyzed coverage for %d business rules", len(coverage)),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// getCoverageHandler handles GET request to retrieve coverage for a knowledge item
func getCoverageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	knowledgeItemID := chi.URLParam(r, "knowledge_item_id")
	if knowledgeItemID == "" {
		http.Error(w, "knowledge_item_id is required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	query := `
		SELECT id, test_requirement_id, knowledge_item_id, coverage_percentage, 
		       test_files, last_updated, created_at
		FROM test_coverage
		WHERE knowledge_item_id = $1
		ORDER BY last_updated DESC
		LIMIT 1
	`

	row := database.QueryRowWithTimeout(ctx, db, query, knowledgeItemID)

	var cov TestCoverage
	var testFilesStr sql.NullString

	err := row.Scan(
		&cov.ID, &cov.TestRequirementID, &cov.KnowledgeItemID, &cov.CoveragePercentage,
		&testFilesStr, &cov.LastUpdated, &cov.CreatedAt,
	)

	if err == sql.ErrNoRows {
		http.Error(w, "Coverage not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("Error querying coverage: %v", err)
		http.Error(w, fmt.Sprintf("Failed to query coverage: %v", err), http.StatusInternalServerError)
		return
	}

	if testFilesStr.Valid {
		// Parse test files array (stored as PostgreSQL array)
		cov.TestFiles = strings.Split(strings.Trim(testFilesStr.String, "{}"), ",")
		for i, file := range cov.TestFiles {
			cov.TestFiles[i] = strings.Trim(file, "\"")
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cov)
}
