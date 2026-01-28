// Test Validator - Main Types and HTTP Handlers
// Handles test validation requests and database operations
// Complies with CODING_STANDARDS.md: Business Services max 400 lines

package services

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"sentinel-hub-api/pkg/database"

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
	ProjectID          string   `json:"project_id"`
	TestRequirementIDs []string `json:"testRequirementIds,omitempty"` // Optional: specific requirements
	TestCode           string   `json:"testCode"`                     // Test code to validate
	TestFilePath       string   `json:"testFilePath"`
}

// ValidateTestsResponse represents the response
type ValidateTestsResponse struct {
	Success     bool             `json:"success"`
	Validations []TestValidation `json:"validations"`
	Count       int              `json:"count"`
	Message     string           `json:"message,omitempty"`
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

	// Detect language from file path and code
	language := detectLanguage(req.TestFilePath, req.TestCode)
	if language == "" {
		http.Error(w, "Unable to detect language from test file", http.StatusBadRequest)
		return
	}

	// Generate test code hash for caching/deduplication
	testCodeHash := calculateTestCodeHash(req.TestCode)

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
			ID:                uuid.New().String(),
			TestRequirementID: "", // No specific requirement
			ValidationStatus:  status,
			Issues:            allIssues,
			TestCodeHash:      testCodeHash,
			Score:             score,
			ValidatedAt:       now,
			CreatedAt:         now,
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
				ID:                uuid.New().String(),
				TestRequirementID: requirement.ID,
				ValidationStatus:  status,
				Issues:            allIssues,
				TestCodeHash:      testCodeHash,
				Score:             score,
				ValidatedAt:       now,
				CreatedAt:         now,
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

	row := database.QueryRowWithTimeout(ctx, db, query, testRequirementID)

	var validation TestValidation
	var issuesStr sql.NullString

	err := row.Scan(
		&validation.ID, &validation.TestRequirementID, &validation.ValidationStatus,
		&issuesStr, &validation.Score, &validation.ValidatedAt, &validation.CreatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
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

// calculateTestCodeHash creates a hash of test code for caching/deduplication
func calculateTestCodeHash(testCode string) string {
	hash := sha256.Sum256([]byte(testCode))
	return hex.EncodeToString(hash[:])
}
