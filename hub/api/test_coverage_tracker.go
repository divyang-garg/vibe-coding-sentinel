// Phase 10B: Test Coverage Tracker
// Tracks test coverage per business rule

package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

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

// discoverTestFiles extracts test file paths and content from TestFile structs
func discoverTestFiles(testFiles []TestFile) []TestFile {
	// Return provided files (Agent sends them with content)
	return testFiles
}

// parseTestFile parses a test file to extract test functions and map them to business rules
// This is a simplified version - in production, use AST analysis
func parseTestFile(testFilePath string, testCode string) ([]string, error) {
	var testFunctions []string

	// Simple heuristic: look for test function patterns
	lines := strings.Split(testCode, "\n")
	for _, line := range lines {
		lineLower := strings.ToLower(line)

		// Go: func TestXxx(t *testing.T)
		if strings.Contains(lineLower, "func test") {
			parts := strings.Fields(line)
			for i, part := range parts {
				if part == "func" && i+1 < len(parts) {
					funcName := parts[i+1]
					// Remove receiver if present
					if !strings.Contains(funcName, "(") {
						testFunctions = append(testFunctions, funcName)
					}
				}
			}
		}

		// JavaScript/TypeScript: test('...', ...) or it('...', ...)
		if strings.Contains(lineLower, "test(") || strings.Contains(lineLower, "it(") {
			// Extract test name from string literal
			if idx := strings.Index(line, "'"); idx != -1 {
				if endIdx := strings.Index(line[idx+1:], "'"); endIdx != -1 {
					testName := line[idx+1 : idx+1+endIdx]
					testFunctions = append(testFunctions, testName)
				}
			}
		}

		// Python: def test_xxx():
		if strings.Contains(lineLower, "def test_") {
			parts := strings.Fields(line)
			for _, part := range parts {
				if strings.HasPrefix(part, "test_") {
					testFunctions = append(testFunctions, part)
					break
				}
			}
		}
	}

	return testFunctions, nil
}

// mapTestsToBusinessRules maps test functions to business rules
// This is a simplified heuristic - in production, use AST analysis and annotations
func mapTestsToBusinessRules(testFunctions []string, testFilePath string, businessRules []KnowledgeItem) map[string][]string {
	ruleToTests := make(map[string][]string)

	// Simple heuristic: match test function names with rule title keywords
	for _, rule := range businessRules {
		ruleTitleLower := strings.ToLower(rule.Title)
		keywords := extractKeywords(ruleTitleLower)

		var matchingTests []string
		for _, testFunc := range testFunctions {
			testFuncLower := strings.ToLower(testFunc)
			for _, keyword := range keywords {
				if strings.Contains(testFuncLower, keyword) {
					matchingTests = append(matchingTests, testFilePath+":"+testFunc)
					break
				}
			}
		}

		if len(matchingTests) > 0 {
			ruleToTests[rule.ID] = matchingTests
		}
	}

	return ruleToTests
}

// isRequirementCovered checks if a test requirement is covered by test files
func isRequirementCovered(req TestRequirement, testPaths []string, testFiles []TestFile) bool {
	// Extract keywords from requirement description
	keywords := extractKeywords(strings.ToLower(req.Description))
	if len(keywords) == 0 {
		return false
	}

	// Check if any test file mentions these keywords
	for _, testPath := range testPaths {
		for _, testFile := range testFiles {
			// Check if this test file matches the path
			if strings.Contains(testFile.Path, testPath) || testFile.Path == testPath {
				testContentLower := strings.ToLower(testFile.Content)
				matchedKeywords := 0
				for _, keyword := range keywords {
					if strings.Contains(testContentLower, keyword) {
						matchedKeywords++
					}
				}
				// If 70%+ keywords match, consider requirement covered
				if float64(matchedKeywords)/float64(len(keywords)) >= 0.7 {
					return true
				}
			}
		}
	}
	return false
}

// calculateCoverage calculates coverage percentage for a business rule
func calculateCoverage(ruleID string, testRequirements []TestRequirement, ruleToTests map[string][]string, testFiles []TestFile) float64 {
	// Get test requirements for this rule
	var ruleRequirements []TestRequirement
	for _, req := range testRequirements {
		if req.KnowledgeItemID == ruleID {
			ruleRequirements = append(ruleRequirements, req)
		}
	}

	if len(ruleRequirements) == 0 {
		return 0.0 // No requirements = no coverage
	}

	// Check if tests exist for this rule
	tests := ruleToTests[ruleID]
	if len(tests) == 0 {
		return 0.0 // No tests = 0% coverage
	}

	// Analyze test content to determine which requirements are covered
	coveredRequirements := 0
	for _, req := range ruleRequirements {
		if isRequirementCovered(req, tests, testFiles) {
			coveredRequirements++
		}
	}

	return float64(coveredRequirements) / float64(len(ruleRequirements))
}

// getTestRequirementsForRules gets test requirements for given knowledge item IDs
func getTestRequirementsForRules(ctx context.Context, knowledgeItemIDs []string) ([]TestRequirement, error) {
	if len(knowledgeItemIDs) == 0 {
		// Get all test requirements
		query := `SELECT id, knowledge_item_id, rule_title, requirement_type, description, 
		                 code_function, priority, created_at, updated_at
		          FROM test_requirements
		          ORDER BY created_at DESC`
		rows, err := queryWithTimeout(ctx, query)
		if err != nil {
			return nil, fmt.Errorf("failed to query test requirements: %w", err)
		}
		defer rows.Close()

		var requirements []TestRequirement
		for rows.Next() {
			var req TestRequirement
			err := rows.Scan(
				&req.ID, &req.KnowledgeItemID, &req.RuleTitle, &req.RequirementType,
				&req.Description, &req.CodeFunction, &req.Priority,
				&req.CreatedAt, &req.UpdatedAt,
			)
			if err != nil {
				log.Printf("Error scanning test requirement: %v", err)
				continue
			}
			requirements = append(requirements, req)
		}
		return requirements, nil
	}

	// Get specific test requirements
	query := `SELECT id, knowledge_item_id, rule_title, requirement_type, description, 
	                 code_function, priority, created_at, updated_at
	          FROM test_requirements
	          WHERE knowledge_item_id = ANY($1)
	          ORDER BY created_at DESC`
	rows, err := queryWithTimeout(ctx, query, knowledgeItemIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to query test requirements: %w", err)
	}
	defer rows.Close()

	var requirements []TestRequirement
	for rows.Next() {
		var req TestRequirement
		err := rows.Scan(
			&req.ID, &req.KnowledgeItemID, &req.RuleTitle, &req.RequirementType,
			&req.Description, &req.CodeFunction, &req.Priority,
			&req.CreatedAt, &req.UpdatedAt,
		)
		if err != nil {
			log.Printf("Error scanning test requirement: %v", err)
			continue
		}
		requirements = append(requirements, req)
	}
	return requirements, nil
}

// saveTestCoverage saves test coverage to database
func saveTestCoverage(ctx context.Context, coverage []TestCoverage) error {
	query := `
		INSERT INTO test_coverage 
		(id, test_requirement_id, knowledge_item_id, coverage_percentage, test_files, last_updated, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
			coverage_percentage = EXCLUDED.coverage_percentage,
			test_files = EXCLUDED.test_files,
			last_updated = EXCLUDED.last_updated
	`

	for _, cov := range coverage {
		_, err := execWithTimeout(ctx, query,
			cov.ID, cov.TestRequirementID, cov.KnowledgeItemID, cov.CoveragePercentage,
			cov.TestFiles, cov.LastUpdated, cov.CreatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to save test coverage %s: %w", cov.ID, err)
		}
	}

	return nil
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

	// Validate project ID
	if req.ProjectID == "" {
		http.Error(w, "projectId is required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	// Get business rules for this project
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

	// Get test requirements for these rules
	knowledgeItemIDs := make([]string, len(rules))
	for i, rule := range rules {
		knowledgeItemIDs[i] = rule.ID
	}

	testRequirements, err := getTestRequirementsForRules(ctx, knowledgeItemIDs)
	if err != nil {
		log.Printf("Error getting test requirements: %v", err)
		http.Error(w, fmt.Sprintf("Failed to get test requirements: %v", err), http.StatusInternalServerError)
		return
	}

	// Discover test files (extract from request)
	testFiles := discoverTestFiles(req.TestFiles)
	if len(testFiles) == 0 {
		response := AnalyzeCoverageResponse{
			Success:  true,
			Coverage: []TestCoverage{},
			Count:    0,
			Message:  "No test files provided. Please provide test files with content in the request.",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

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

	row := queryRowWithTimeout(ctx, query, knowledgeItemID)

	var cov TestCoverage
	var testFilesStr sql.NullString

	err := row.Scan(
		&cov.ID, &cov.TestRequirementID, &cov.KnowledgeItemID, &cov.CoveragePercentage,
		&testFilesStr, &cov.LastUpdated, &cov.CreatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
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
