// Phase 12: Change Request Manager Module
// Manages change request workflow (CRUD operations, approval/rejection)

package services

import (
	"context"
	"database/sql"
	"fmt"

	"sentinel-hub-api/pkg/database"
)

// getChangeRequest retrieves a change request by ID
func getChangeRequest(ctx context.Context, changeRequestID string) (*ChangeRequest, error) {
	// Validate change request ID
	if err := ValidateUUID(changeRequestID); err != nil {
		return nil, fmt.Errorf("invalid change request ID: %w", err)
	}

	query := `
		SELECT id, project_id, knowledge_item_id, type, current_state, proposed_state, 
		       status, implementation_status, implementation_notes, impact_analysis,
		       created_at, approved_by, approved_at, rejected_by, rejected_at, rejection_reason
		FROM change_requests
		WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(ctx, getQueryTimeout())
	defer cancel()

	var cr ChangeRequest
	var currentStateJSON, proposedStateJSON, impactAnalysisJSON sql.NullString
	var knowledgeItemID sql.NullString
	var approvedBy, rejectedBy, rejectionReason, implNotes sql.NullString
	var approvedAt, rejectedAt sql.NullTime

	err := database.QueryRowWithTimeout(ctx, db, query, changeRequestID).Scan(
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

	// Unmarshal JSONB fields using helper
	if currentStateJSON.Valid {
		cr.CurrentState = make(map[string]interface{})
		if err := unmarshalJSONB(currentStateJSON.String, &cr.CurrentState); err != nil {
			return nil, fmt.Errorf("failed to unmarshal current state: %w", err)
		}
	}
	if proposedStateJSON.Valid {
		cr.ProposedState = make(map[string]interface{})
		if err := unmarshalJSONB(proposedStateJSON.String, &cr.ProposedState); err != nil {
			return nil, fmt.Errorf("failed to unmarshal proposed state: %w", err)
		}
	}
	if impactAnalysisJSON.Valid {
		cr.ImpactAnalysis = make(map[string]interface{})
		if err := unmarshalJSONB(impactAnalysisJSON.String, &cr.ImpactAnalysis); err != nil {
			return nil, fmt.Errorf("failed to unmarshal impact_analysis: %w", err)
		}
	}

	return &cr, nil
}

// listChangeRequests retrieves change requests with filters and pagination
func listChangeRequests(ctx context.Context, projectID string, statusFilter string, limit int, offset int) ([]ChangeRequest, int, error) {
	// Build query
	query := `
		SELECT id, project_id, knowledge_item_id, type, current_state, proposed_state, 
		       status, implementation_status, created_at
		FROM change_requests
		WHERE project_id = $1
	`
	args := []interface{}{projectID}
	argIndex := 2

	if statusFilter != "" {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, statusFilter)
		argIndex++
	}

	query += " ORDER BY created_at DESC"

	// Get total count
	countQuery := `
		SELECT COUNT(*) FROM change_requests WHERE project_id = $1
	`
	if statusFilter != "" {
		countQuery += " AND status = $2"
	}

	var total int
	err := database.QueryRowWithTimeout(ctx, db, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count change requests for project %s: %w", projectID, err)
	}

	// Add pagination
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, limit)
		argIndex++
		if offset > 0 {
			query += fmt.Sprintf(" OFFSET $%d", argIndex)
			args = append(args, offset)
		}
	}

	ctx, cancel := context.WithTimeout(ctx, getQueryTimeout())
	defer cancel()

	rows, err := database.QueryWithTimeout(ctx, db, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query change requests: %w", err)
	}
	defer rows.Close()

	var requests []ChangeRequest
	for rows.Next() {
		var cr ChangeRequest
		var currentStateJSON, proposedStateJSON sql.NullString
		var knowledgeItemID sql.NullString

		err := rows.Scan(
			&cr.ID, &cr.ProjectID, &knowledgeItemID, &cr.Type,
			&currentStateJSON, &proposedStateJSON, &cr.Status,
			&cr.CreatedAt,
		)
		if err != nil {
			LogWarn(ctx, "Error scanning change request: %v", err)
			continue
		}

		if knowledgeItemID.Valid {
			cr.KnowledgeItemID = &knowledgeItemID.String
		}

		// Unmarshal JSONB fields using helper
		if currentStateJSON.Valid {
			cr.CurrentState = make(map[string]interface{})
			if err := unmarshalJSONB(currentStateJSON.String, &cr.CurrentState); err != nil {
				LogWarn(ctx, "Failed to unmarshal current state: %v", err)
			}
		}
		if proposedStateJSON.Valid {
			cr.ProposedState = make(map[string]interface{})
			if err := unmarshalJSONB(proposedStateJSON.String, &cr.ProposedState); err != nil {
				LogWarn(ctx, "Failed to unmarshal proposed state: %v", err)
			}
		}

		requests = append(requests, cr)
	}

	return requests, total, nil
}

// approveChangeRequest approves a change request and applies changes
func approveChangeRequest(ctx context.Context, changeRequestID string, approvedBy string) error {
	// Start transaction
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Load change request
	cr, err := getChangeRequest(ctx, changeRequestID)
	if err != nil {
		return err
	}

	// Check if already approved/rejected
	if cr.Status == "approved" {
		return fmt.Errorf("change request already approved")
	}
	if cr.Status == "rejected" {
		return fmt.Errorf("change request already rejected")
	}

	// Update change request status
	updateQuery := `
		UPDATE change_requests 
		SET status = 'approved', approved_by = $1, approved_at = NOW()
		WHERE id = $2
	`
	_, err = tx.ExecContext(ctx, updateQuery, approvedBy, changeRequestID)
	if err != nil {
		return fmt.Errorf("failed to update change request: %w", err)
	}

	// Apply changes based on type
	if cr.Type == ChangeModified {
		// Update existing knowledge item
		if cr.KnowledgeItemID != nil && *cr.KnowledgeItemID != "" && cr.ProposedState != nil {
			updateKIQuery := `
				UPDATE knowledge_items
				SET title = $1, content = $2
				WHERE id = $3
			`
			title := ""
			content := ""
			if t, ok := cr.ProposedState["title"].(string); ok {
				title = t
			}
			if c, ok := cr.ProposedState["content"].(string); ok {
				content = c
			}
			_, err = tx.ExecContext(ctx, updateKIQuery, title, content, cr.KnowledgeItemID)
			if err != nil {
				return fmt.Errorf("failed to update knowledge item: %w", err)
			}
		}
	} else if cr.Type == ChangeNew {
		// Create new knowledge item
		if cr.ProposedState != nil {
			// Extract document_id from change request context or use project default
			// Try to find a document_id from existing knowledge items in the project
			var documentID string
			docQuery := `
				SELECT DISTINCT document_id 
				FROM knowledge_items 
				WHERE project_id = (SELECT project_id FROM projects WHERE id = $1)
				LIMIT 1
			`
			err := tx.QueryRowContext(ctx, docQuery, cr.ProjectID).Scan(&documentID)
			if err != nil {
				// If no document found, we'll need to create a placeholder or require document_id
				// For now, use NULL document_id (if allowed) or create a default document
				LogWarn(ctx, "No document_id found for project, creating knowledge item without document_id")
			}

			title := ""
			content := ""
			kiType := "rule"
			if t, ok := cr.ProposedState["title"].(string); ok {
				title = t
			}
			if c, ok := cr.ProposedState["content"].(string); ok {
				content = c
			}
			if ty, ok := cr.ProposedState["type"].(string); ok {
				kiType = ty
			}

			var newKIID string
			insertKIQuery := `
				INSERT INTO knowledge_items (document_id, project_id, type, title, content, status, created_at)
				VALUES ($1, $2, $3, $4, $5, 'approved', NOW())
				RETURNING id
			`
			err = tx.QueryRowContext(ctx, insertKIQuery, documentID, cr.ProjectID, kiType, title, content).Scan(&newKIID)
			if err != nil {
				return fmt.Errorf("failed to create knowledge item: %w", err)
			}

			// Update change request with new knowledge item ID
			updateCRQuery := `
				UPDATE change_requests 
				SET knowledge_item_id = $1 
				WHERE id = $2
			`
			_, err = tx.ExecContext(ctx, updateCRQuery, newKIID, changeRequestID)
			if err != nil {
				return fmt.Errorf("failed to link knowledge item to change request: %w", err)
			}

			LogInfo(ctx, "Created new knowledge item %s for change request %s", newKIID, changeRequestID)
			// Invalidate cache for this project
			invalidateGapAnalysisCache(cr.ProjectID)
		}
	} else if cr.Type == ChangeRemoved {
		// Mark knowledge item as deprecated
		if cr.KnowledgeItemID != nil && *cr.KnowledgeItemID != "" {
			updateKIQuery := `
				UPDATE knowledge_items
				SET status = 'deprecated'
				WHERE id = $1
			`
			_, err = tx.ExecContext(ctx, updateKIQuery, cr.KnowledgeItemID)
			if err != nil {
				return fmt.Errorf("failed to deprecate knowledge item: %w", err)
			}
		}
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	LogInfo(ctx, "Change request %s approved by %s", changeRequestID, approvedBy)
	return nil
}

// rejectChangeRequest rejects a change request
func rejectChangeRequest(ctx context.Context, changeRequestID string, rejectedBy string, reason string) error {
	// Load change request to check current status
	cr, err := getChangeRequest(ctx, changeRequestID)
	if err != nil {
		return err
	}

	// Check if already approved/rejected
	if cr.Status == "approved" {
		return fmt.Errorf("change request already approved")
	}
	if cr.Status == "rejected" {
		return fmt.Errorf("change request already rejected")
	}

	// Update change request
	query := `
		UPDATE change_requests 
		SET status = 'rejected', rejected_by = $1, rejected_at = NOW(), rejection_reason = $2
		WHERE id = $3
	`

	ctx, cancel := context.WithTimeout(ctx, getQueryTimeout())
	defer cancel()

	_, err = database.ExecWithTimeout(ctx, db, query, rejectedBy, reason, changeRequestID)
	if err != nil {
		return fmt.Errorf("failed to reject change request: %w", err)
	}

	LogInfo(ctx, "Change request %s rejected by %s: %s", changeRequestID, rejectedBy, reason)
	return nil
}
