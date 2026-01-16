// Fixed import structure
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
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("change request not found: %s", changeRequestID)
		}
		return nil, fmt.Errorf("failed to load change request: %w", err)
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
		cr.RejectionReason = &rejectionReason.String
	}
	if implNotes.Valid {
		cr.ImplementationNotes = &implNotes.String
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
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("knowledge item not found: %s", knowledgeItemID)
		}
		return nil, fmt.Errorf("failed to load knowledge item: %w", err)
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
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("comprehensive validation not found: %s", validationID)
		}
		return nil, fmt.Errorf("failed to load comprehensive validation: %w", err)
	}

	// Unmarshal JSONB fields
	jsonFields := []struct {
		json  sql.NullString
		field *map[string]interface{}
		name  string
	}{
		{findingsJSON, &cv.Findings, "findings"},
		{summaryJSON, &cv.Summary, "summary"},
		{layerAnalysisJSON, &cv.LayerAnalysis, "layer_analysis"},
		{endToEndFlowsJSON, &cv.EndToEndFlows, "end_to_end_flows"},
		{checklistJSON, &cv.Checklist, "checklist"},
	}

	for _, jf := range jsonFields {
		if jf.json.Valid {
			*jf.field = make(map[string]interface{})
			if err := json.Unmarshal([]byte(jf.json.String), jf.field); err != nil {
				return nil, fmt.Errorf("failed to unmarshal %s: %w", jf.name, err)
			}
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
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("test requirement not found: %s", testRequirementID)
		}
		return nil, fmt.Errorf("failed to load test requirement: %w", err)
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
