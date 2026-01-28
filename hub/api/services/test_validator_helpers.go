// Test Validator - Helper Functions
// Utility functions for validation scoring and database operations
// Complies with CODING_STANDARDS.md: Business Services max 400 lines

package services

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"

	"sentinel-hub-api/pkg/database"
)

// calculateValidationScore calculates a validation score (0.0 to 1.0)
func calculateValidationScore(structureValid bool, assertionsValid bool, completenessValid bool, issues []string) float64 {
	score := 1.0

	// Deduct points for each issue
	for _, issue := range issues {
		if strings.Contains(issue, "Missing assertions") {
			score -= 0.3
		} else if strings.Contains(issue, "Weak assertions") {
			score -= 0.2
		} else if strings.Contains(issue, "Potential shared state") {
			score -= 0.2
		} else if strings.Contains(issue, "may not fully cover") {
			score -= 0.2
		} else if strings.Contains(issue, "may be missing") {
			score -= 0.15
		} else {
			score -= 0.1
		}
	}

	// Deduct for invalid checks
	if !structureValid {
		score -= 0.2
	}
	if !assertionsValid {
		score -= 0.2
	}
	if !completenessValid {
		score -= 0.15
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
	query := `
		SELECT id, knowledge_item_id, rule_title, requirement_type, description, code_function, priority, created_at, updated_at
		FROM test_requirements
		WHERE id = $1
	`

	row := database.QueryRowWithTimeout(ctx, db, query, requirementID)

	var requirement TestRequirement
	var codeFunction sql.NullString
	err := row.Scan(
		&requirement.ID, &requirement.KnowledgeItemID, &requirement.RuleTitle,
		&requirement.RequirementType, &requirement.Description, &codeFunction,
		&requirement.Priority, &requirement.CreatedAt, &requirement.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get test requirement: %w", err)
	}

	if codeFunction.Valid {
		requirement.CodeFunction = codeFunction.String
	}

	return &requirement, nil
}

// saveTestValidation saves validation results to database
func saveTestValidation(ctx context.Context, validation TestValidation) error {
	// Convert issues slice to PostgreSQL array format
	issuesArray := "{" + strings.Join(validation.Issues, ",") + "}"

	query := `
		INSERT INTO test_validations 
		(id, test_requirement_id, validation_status, issues, test_code_hash, score, validated_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET
			validation_status = EXCLUDED.validation_status,
			issues = EXCLUDED.issues,
			score = EXCLUDED.score,
			validated_at = EXCLUDED.validated_at
	`

	_, err := database.ExecWithTimeout(ctx, db, query,
		validation.ID, validation.TestRequirementID, validation.ValidationStatus,
		issuesArray, validation.TestCodeHash, validation.Score,
		validation.ValidatedAt, validation.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save test validation: %w", err)
	}

	return nil
}

// detectLanguage detects language from file path and code content
func detectLanguage(testFilePath string, testCode string) string {
	// Try to detect from file extension first
	ext := strings.ToLower(filepath.Ext(testFilePath))
	switch ext {
	case ".go":
		return "go"
	case ".js":
		return "javascript"
	case ".ts":
		return "typescript"
	case ".py":
		return "python"
	}

	// Fallback to code content detection
	testCodeLower := strings.ToLower(testCode)
	if strings.Contains(testCodeLower, "func test") || strings.Contains(testCodeLower, "testing.t") {
		return "go"
	}
	if strings.Contains(testCodeLower, "describe(") || strings.Contains(testCodeLower, "it(") || strings.Contains(testCodeLower, "test(") {
		return "javascript"
	}
	if strings.Contains(testCodeLower, "def test_") || strings.Contains(testCodeLower, "import unittest") {
		return "python"
	}

	return ""
}
