// Test Requirement Generator - Helper Functions
// Maps rules to code and generates test requirements
// Complies with CODING_STANDARDS.md: Business Services max 400 lines

package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"sentinel-hub-api/pkg/database"

	"github.com/google/uuid"
)

// mapRuleToCode attempts to map a business rule to code functions using AST analysis
// This is a simplified version - in production, this would use the AST analyzer
func mapRuleToCode(rule KnowledgeItem, projectCode []string) string {
	// Simple heuristic: look for function names that match rule title keywords
	ruleTitleLower := strings.ToLower(rule.Title)
	keywords := extractKeywords(ruleTitleLower)

	// Search for functions containing keywords
	for _, code := range projectCode {
		codeLower := strings.ToLower(code)
		for _, keyword := range keywords {
			if strings.Contains(codeLower, keyword) {
				// Try to extract function name (simplified)
				if fn := extractFunctionNameFromCode(code, keyword); fn != "" {
					return fn
				}
			}
		}
	}

	return "" // No mapping found
}

// extractFunctionNameFromCode attempts to extract a function name from code (simplified)
// TODO: Use AST analysis (Phase 6) for more accurate function extraction
func extractFunctionNameFromCode(code, keyword string) string {
	// CURRENT IMPLEMENTATION: Uses pattern matching to find function definitions
	// FUTURE ENHANCEMENT: Use AST analysis (Phase 6) for more accurate function extraction
	// This would provide better accuracy and handle complex code structures
	// Look for function definitions containing the keyword
	lines := strings.Split(code, "\n")
	for _, line := range lines {
		lineLower := strings.ToLower(line)
		if strings.Contains(lineLower, "func ") && strings.Contains(lineLower, keyword) {
			// Try to extract function name
			parts := strings.Fields(line)
			for i, part := range parts {
				if part == "func" && i+1 < len(parts) {
					funcName := parts[i+1]
					// Remove receiver if present
					if strings.Contains(funcName, "(") {
						continue
					}
					return funcName
				}
			}
		}
	}
	return ""
}

// generateTestRequirements generates test requirements for a business rule
func generateTestRequirements(rule KnowledgeItem, codeFunction string) []TestRequirement {
	var requirements []TestRequirement
	now := time.Now()

	// Generate happy path requirement
	requirements = append(requirements, TestRequirement{
		ID:              uuid.New().String(),
		KnowledgeItemID: rule.ID,
		RuleTitle:       rule.Title,
		RequirementType: "happy_path",
		Description:     fmt.Sprintf("Test normal operation: %s", rule.Content),
		CodeFunction:    codeFunction,
		Priority:        "high",
		CreatedAt:       now,
		UpdatedAt:       now,
	})

	// Generate edge case requirement
	requirements = append(requirements, TestRequirement{
		ID:              uuid.New().String(),
		KnowledgeItemID: rule.ID,
		RuleTitle:       rule.Title,
		RequirementType: "edge_case",
		Description:     fmt.Sprintf("Test boundary conditions and edge cases: %s", rule.Content),
		CodeFunction:    codeFunction,
		Priority:        "medium",
		CreatedAt:       now,
		UpdatedAt:       now,
	})

	// Generate error case requirement
	requirements = append(requirements, TestRequirement{
		ID:              uuid.New().String(),
		KnowledgeItemID: rule.ID,
		RuleTitle:       rule.Title,
		RequirementType: "error_case",
		Description:     fmt.Sprintf("Test error handling and invalid inputs: %s", rule.Content),
		CodeFunction:    codeFunction,
		Priority:        "high",
		CreatedAt:       now,
		UpdatedAt:       now,
	})

	return requirements
}

// saveTestRequirements saves test requirements to database
func saveTestRequirements(ctx context.Context, requirements []TestRequirement) error {
	query := `
		INSERT INTO test_requirements 
		(id, knowledge_item_id, rule_title, requirement_type, description, code_function, priority, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id) DO UPDATE SET
			description = EXCLUDED.description,
			code_function = EXCLUDED.code_function,
			priority = EXCLUDED.priority,
			updated_at = EXCLUDED.updated_at
	`

	for _, req := range requirements {
		_, err := database.ExecWithTimeout(ctx, db, query,
			req.ID, req.KnowledgeItemID, req.RuleTitle, req.RequirementType,
			req.Description, req.CodeFunction, req.Priority, req.CreatedAt, req.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to save test requirement %s: %w", req.ID, err)
		}
	}

	return nil
}
