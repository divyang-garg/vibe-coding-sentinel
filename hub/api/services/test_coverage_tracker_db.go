// Test Coverage Tracker - Database Functions
// Handles database operations for test coverage
// Complies with CODING_STANDARDS.md: Business Services max 400 lines

package services

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"

	"sentinel-hub-api/pkg/database"
)

// getTestRequirementsForRules gets test requirements for given knowledge item IDs
func getTestRequirementsForRules(ctx context.Context, knowledgeItemIDs []string) ([]TestRequirement, error) {
	if len(knowledgeItemIDs) == 0 {
		// Get all test requirements
		query := `SELECT id, knowledge_item_id, rule_title, requirement_type, description, 
		                 code_function, priority, created_at, updated_at
		          FROM test_requirements
		          ORDER BY created_at DESC`
		rows, err := database.QueryWithTimeout(ctx, db, query)
		if err != nil {
			return nil, fmt.Errorf("failed to query test requirements: %w", err)
		}
		defer rows.Close()

		var requirements []TestRequirement
		for rows.Next() {
			var req TestRequirement
			var codeFunction sql.NullString
			err := rows.Scan(
				&req.ID, &req.KnowledgeItemID, &req.RuleTitle, &req.RequirementType,
				&req.Description, &codeFunction, &req.Priority,
				&req.CreatedAt, &req.UpdatedAt,
			)
			if err != nil {
				log.Printf("Error scanning test requirement: %v", err)
				continue
			}
			if codeFunction.Valid {
				req.CodeFunction = codeFunction.String
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
	rows, err := database.QueryWithTimeout(ctx, db, query, knowledgeItemIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to query test requirements: %w", err)
	}
	defer rows.Close()

	var requirements []TestRequirement
	for rows.Next() {
		var req TestRequirement
		var codeFunction sql.NullString
		err := rows.Scan(
			&req.ID, &req.KnowledgeItemID, &req.RuleTitle, &req.RequirementType,
			&req.Description, &codeFunction, &req.Priority,
			&req.CreatedAt, &req.UpdatedAt,
		)
		if err != nil {
			log.Printf("Error scanning test requirement: %v", err)
			continue
		}
		if codeFunction.Valid {
			req.CodeFunction = codeFunction.String
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
		// Convert test files slice to PostgreSQL array format
		testFilesArray := "{" + strings.Join(cov.TestFiles, ",") + "}"

		_, err := database.ExecWithTimeout(ctx, db, query,
			cov.ID, cov.TestRequirementID, cov.KnowledgeItemID, cov.CoveragePercentage,
			testFilesArray, cov.LastUpdated, cov.CreatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to save test coverage %s: %w", cov.ID, err)
		}
	}

	return nil
}
