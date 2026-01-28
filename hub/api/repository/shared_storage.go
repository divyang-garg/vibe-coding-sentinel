// Fixed import structure
// Complies with CODING_STANDARDS.md: Repositories max 350 lines
package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"sentinel-hub-api/models"
	"sentinel-hub-api/pkg/database"
	"sentinel-hub-api/utils"
)

// Package-level database connection (set during initialization)
var db *sql.DB

// SetDB sets the database connection for the repository package
func SetDB(database *sql.DB) {
	db = database
}

// ValidateUUID validates a UUID string
func ValidateUUID(id string) error {
	return utils.ValidateUUID(id)
}

// KnowledgeItem is an alias to models.KnowledgeItem for local use
type KnowledgeItem = models.KnowledgeItem

// ComprehensiveValidation is an alias to models.ComprehensiveValidation for local use
type ComprehensiveValidation = models.ComprehensiveValidation

// TestRequirement is an alias to models.TestRequirement for local use
type TestRequirement = models.TestRequirement

// LinkType constants for task links
const (
	LinkTypeChangeRequest         = "change_request"
	LinkTypeKnowledgeItem         = "knowledge_item"
	LinkTypeComprehensiveAnalysis = "comprehensive_analysis"
	LinkTypeTestRequirement       = "test_requirement"
)

// GetChangeRequestByID retrieves a change request by ID
func GetChangeRequestByID(ctx context.Context, db *sql.DB, changeRequestID string) (*models.ChangeRequest, error) {
	if err := utils.ValidateUUID(changeRequestID); err != nil {
		return nil, fmt.Errorf("invalid change request ID: %w", err)
	}

	query := `
		SELECT id, project_id, knowledge_item_id, type, current_state, proposed_state,
		       status, implementation_status, implementation_notes, impact_analysis,
		       created_at, approved_by, approved_at, rejected_by, rejected_at, rejection_reason
		FROM change_requests
		WHERE id = $1
	`

	row := database.QueryRowWithTimeout(ctx, db, query, changeRequestID)

	var cr models.ChangeRequest
	var knowledgeItemID sql.NullString
	var currentStateJSON, proposedStateJSON, impactAnalysisJSON sql.NullString
	var approvedBy, rejectedBy, rejectionReason, implNotes sql.NullString
	var approvedAt, rejectedAt sql.NullTime

	err := row.Scan(
		&cr.ID, &cr.ProjectID, &knowledgeItemID, &cr.Type,
		&currentStateJSON, &proposedStateJSON, &cr.Status,
		&cr.ImplementationStatus, &implNotes, &impactAnalysisJSON,
		&cr.CreatedAt, &approvedBy, &approvedAt, &rejectedBy, &rejectedAt, &rejectionReason,
	)
	if err != nil {
		return nil, utils.HandleNotFoundError(err, "change request", changeRequestID)
	}

	// Handle nullable fields
	if knowledgeItemID.Valid {
		cr.KnowledgeItemID = &knowledgeItemID.String
	}
	if approvedBy.Valid {
		cr.ApprovedBy = &approvedBy.String
	}
	if approvedAt.Valid {
		cr.ApprovedAt = &approvedAt.Time
	}
	if rejectedBy.Valid {
		cr.RejectedBy = &rejectedBy.String
	}
	if rejectedAt.Valid {
		cr.RejectedAt = &rejectedAt.Time
	}
	if rejectionReason.Valid {
		cr.RejectionReason = rejectionReason.String
	}
	if implNotes.Valid {
		cr.ImplementationNotes = implNotes.String
	}

	// Unmarshal JSONB fields
	if currentStateJSON.Valid {
		cr.CurrentState = make(map[string]interface{})
		if err := json.Unmarshal([]byte(currentStateJSON.String), &cr.CurrentState); err != nil {
			return nil, fmt.Errorf("failed to unmarshal current_state: %w", err)
		}
	}
	if proposedStateJSON.Valid {
		cr.ProposedState = make(map[string]interface{})
		if err := json.Unmarshal([]byte(proposedStateJSON.String), &cr.ProposedState); err != nil {
			return nil, fmt.Errorf("failed to unmarshal proposed_state: %w", err)
		}
	}
	if impactAnalysisJSON.Valid {
		cr.ImpactAnalysis = make(map[string]interface{})
		if err := json.Unmarshal([]byte(impactAnalysisJSON.String), &cr.ImpactAnalysis); err != nil {
			return nil, fmt.Errorf("failed to unmarshal impact_analysis: %w", err)
		}
	}

	return &cr, nil
}

// GetKnowledgeItemByID retrieves a knowledge item by ID
func GetKnowledgeItemByID(ctx context.Context, knowledgeItemID string) (*models.KnowledgeItem, error) {
	if err := ValidateUUID(knowledgeItemID); err != nil {
		return nil, fmt.Errorf("invalid knowledge item ID: %w", err)
	}

	query := `
		SELECT id, document_id, type, title, content, confidence, source_page,
		       status, approved_by, approved_at, created_at, structured_data
		FROM knowledge_items
		WHERE id = $1
	`

	row := database.QueryRowWithTimeout(ctx, db, query, knowledgeItemID)

	var ki KnowledgeItem
	var approvedBy sql.NullString
	var approvedAt sql.NullTime
	var structuredDataJSON sql.NullString

	err := row.Scan(
		&ki.ID, &ki.DocumentID, &ki.Type, &ki.Title, &ki.Content,
		&ki.Confidence, &ki.SourcePage, &ki.Status,
		&approvedBy, &approvedAt, &ki.CreatedAt, &structuredDataJSON,
	)
	if err != nil {
		return nil, utils.HandleNotFoundError(err, "knowledge item", knowledgeItemID)
	}

	if approvedBy.Valid {
		ki.ApprovedBy = &approvedBy.String
	}
	if approvedAt.Valid {
		ki.ApprovedAt = &approvedAt.Time
	}
	if structuredDataJSON.Valid {
		ki.StructuredData = make(map[string]interface{})
		if err := json.Unmarshal([]byte(structuredDataJSON.String), &ki.StructuredData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal structured_data: %w", err)
		}
	}

	return &ki, nil
}

// GetComprehensiveValidationByID retrieves a comprehensive validation by validation_id
func GetComprehensiveValidationByID(ctx context.Context, validationID string) (*models.ComprehensiveValidation, error) {
	query := `
		SELECT id, project_id, validation_id, feature, mode, depth,
		       findings, summary, layer_analysis, end_to_end_flows, checklist,
		       created_at, completed_at
		FROM comprehensive_validations
		WHERE validation_id = $1
	`

	row := database.QueryRowWithTimeout(ctx, db, query, validationID)

	var cv ComprehensiveValidation
	var findingsJSON, summaryJSON, layerAnalysisJSON, endToEndFlowsJSON, checklistJSON sql.NullString
	var completedAt sql.NullTime

	err := row.Scan(
		&cv.ID, &cv.ProjectID, &cv.ValidationID, &cv.Feature, &cv.Mode, &cv.Depth,
		&findingsJSON, &summaryJSON, &layerAnalysisJSON, &endToEndFlowsJSON, &checklistJSON,
		&cv.CreatedAt, &completedAt,
	)
	if err != nil {
		return nil, utils.HandleNotFoundError(err, "comprehensive validation", validationID)
	}

	// Unmarshal JSONB fields
	if findingsJSON.Valid {
		cv.Findings = make(map[string]interface{})
		if err := json.Unmarshal([]byte(findingsJSON.String), &cv.Findings); err != nil {
			return nil, fmt.Errorf("failed to unmarshal findings: %w", err)
		}
	}
	if summaryJSON.Valid {
		cv.Summary = make(map[string]interface{})
		if err := json.Unmarshal([]byte(summaryJSON.String), &cv.Summary); err != nil {
			return nil, fmt.Errorf("failed to unmarshal summary: %w", err)
		}
	}
	if layerAnalysisJSON.Valid {
		cv.LayerAnalysis = make(map[string]interface{})
		if err := json.Unmarshal([]byte(layerAnalysisJSON.String), &cv.LayerAnalysis); err != nil {
			return nil, fmt.Errorf("failed to unmarshal layer_analysis: %w", err)
		}
	}
	if endToEndFlowsJSON.Valid {
		if err := json.Unmarshal([]byte(endToEndFlowsJSON.String), &cv.EndToEndFlows); err != nil {
			return nil, fmt.Errorf("failed to unmarshal end_to_end_flows: %w", err)
		}
	}
	if checklistJSON.Valid {
		if err := json.Unmarshal([]byte(checklistJSON.String), &cv.Checklist); err != nil {
			return nil, fmt.Errorf("failed to unmarshal checklist: %w", err)
		}
	}

	if completedAt.Valid {
		cv.CompletedAt = &completedAt.Time
	}

	return &cv, nil
}

// GetTestRequirementByID retrieves a test requirement by ID
func GetTestRequirementByID(ctx context.Context, testRequirementID string) (*models.TestRequirement, error) {
	if err := ValidateUUID(testRequirementID); err != nil {
		return nil, fmt.Errorf("invalid test requirement ID: %w", err)
	}

	query := `
		SELECT id, knowledge_item_id, rule_title, requirement_type, description,
		       code_function, priority, created_at, updated_at
		FROM test_requirements
		WHERE id = $1
	`

	row := database.QueryRowWithTimeout(ctx, db, query, testRequirementID)

	var tr TestRequirement
	var codeFunction sql.NullString

	err := row.Scan(
		&tr.ID, &tr.KnowledgeItemID, &tr.RuleTitle, &tr.RequirementType,
		&tr.Description, &codeFunction, &tr.Priority, &tr.CreatedAt, &tr.UpdatedAt,
	)
	if err != nil {
		return nil, utils.HandleNotFoundError(err, "test requirement", testRequirementID)
	}

	if codeFunction.Valid {
		tr.CodeFunction = codeFunction.String
	}

	return &tr, nil
}

// ValidateTaskLink validates that a task link references a valid entity
func ValidateTaskLink(ctx context.Context, linkType string, linkedID string) error {
	if err := ValidateUUID(linkedID); err != nil {
		// Comprehensive validation uses VARCHAR validation_id, not UUID
		if linkType == LinkTypeComprehensiveAnalysis {
			if linkedID == "" {
				return fmt.Errorf("invalid linked ID: cannot be empty")
			}
			// Check if comprehensive validation exists
			query := `SELECT EXISTS(SELECT 1 FROM comprehensive_validations WHERE validation_id = $1)`
			row := database.QueryRowWithTimeout(ctx, db, query, linkedID)
			var exists bool
			if err := row.Scan(&exists); err != nil {
				return fmt.Errorf("failed to validate link: %w", err)
			}
			if !exists {
				return fmt.Errorf("%s with ID %s does not exist", linkType, linkedID)
			}
			return nil
		}
		return fmt.Errorf("invalid linked ID: %w", err)
	}

	var query string
	switch linkType {
	case LinkTypeChangeRequest:
		query = `SELECT EXISTS(SELECT 1 FROM change_requests WHERE id = $1)`
	case LinkTypeKnowledgeItem:
		query = `SELECT EXISTS(SELECT 1 FROM knowledge_items WHERE id = $1)`
	case LinkTypeTestRequirement:
		query = `SELECT EXISTS(SELECT 1 FROM test_requirements WHERE id = $1)`
	default:
		return fmt.Errorf("invalid link type: %s", linkType)
	}

	row := database.QueryRowWithTimeout(ctx, db, query, linkedID)
	var exists bool
	if err := row.Scan(&exists); err != nil {
		return fmt.Errorf("failed to validate link: %w", err)
	}

	if !exists {
		return fmt.Errorf("%s with ID %s does not exist", linkType, linkedID)
	}

	return nil
}
