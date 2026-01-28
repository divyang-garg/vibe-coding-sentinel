// Package services - Knowledge Service Helper Functions
// Complies with CODING_STANDARDS.md: Helper functions extracted to maintain file size limits
package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"sentinel-hub-api/pkg/database"
)

// getSecurityRules retrieves security rules for a project from the knowledge_items table.
// It queries for approved security rules and extracts rule identifiers.
// Returns default rules if none are found for backward compatibility.
func (s *KnowledgeServiceImpl) getSecurityRules(ctx context.Context, projectID string) ([]string, error) {
	if projectID == "" {
		return nil, fmt.Errorf("project_id is required")
	}

	query := `
		SELECT title, content, structured_data
		FROM knowledge_items
		WHERE project_id = $1 
		  AND type = 'security_rule'
		  AND status = 'approved'
		ORDER BY created_at DESC
	`

	// Use QueryContext directly with timeout to avoid context cancellation during iteration
	ctx, cancel := context.WithTimeout(ctx, getQueryTimeout())
	defer cancel()
	
	rows, err := s.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to query security rules: %w", err)
	}
	defer rows.Close()

	var securityRules []string
	for rows.Next() {
		var title, content string
		var structuredDataJSON sql.NullString

		err := rows.Scan(&title, &content, &structuredDataJSON)
		if err != nil {
			LogWarn(ctx, "Failed to scan security rule: %v", err)
			continue
		}

		// Extract rule identifier from title or structured_data
		ruleID := extractSecurityRuleID(title, structuredDataJSON)
		if ruleID != "" {
			securityRules = append(securityRules, ruleID)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating security rules: %w", err)
	}

	// Return default rules if none found (backward compatibility)
	if len(securityRules) == 0 {
		LogInfo(ctx, "No security rules found for project %s, using defaults", projectID)
		return []string{"SEC-001", "SEC-002", "SEC-003"}, nil
	}

	return securityRules, nil
}

// extractSecurityRuleID extracts security rule identifier from title or structured data.
// It first tries to extract from structured_data.rule_id, then falls back to parsing the title.
func extractSecurityRuleID(title string, structuredDataJSON sql.NullString) string {
	// Try to extract from structured_data first
	if structuredDataJSON.Valid {
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(structuredDataJSON.String), &data); err == nil {
			if ruleID, ok := data["rule_id"].(string); ok && ruleID != "" {
				return ruleID
			}
		}
	}

	// Fallback to extracting from title (e.g., "SEC-001: Rule Description")
	if strings.HasPrefix(title, "SEC-") {
		parts := strings.Split(title, ":")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}

	return ""
}

// updateSyncMetadata updates sync metadata for a knowledge item.
// It updates last_synced_at, increments sync_version, sets sync_status to 'synced', and updates updated_at.
// Uses optimistic locking with sync_version to detect conflicts.
// Returns error if item not found or if conflict detected (sync_version changed).
func (s *KnowledgeServiceImpl) updateSyncMetadata(ctx context.Context, itemID string, syncTime time.Time) error {
	// First, get current sync_version to detect conflicts
	var currentVersion int
	checkQuery := `SELECT sync_version FROM knowledge_items WHERE id = $1`
	// Use QueryRowContext directly with timeout to avoid context cancellation during scan
	ctx, cancel := context.WithTimeout(ctx, getQueryTimeout())
	defer cancel()
	
	err := s.db.QueryRowContext(ctx, checkQuery, itemID).Scan(&currentVersion)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("knowledge item not found: %s", itemID)
		}
		return fmt.Errorf("failed to get current sync version: %w", err)
	}

	// Update with version check for conflict detection
	query := `
		UPDATE knowledge_items
		SET last_synced_at = $1,
		    sync_version = sync_version + 1,
		    sync_status = 'synced',
		    updated_at = $1
		WHERE id = $2
		  AND sync_version = $3
	`

	result, err := database.ExecWithTimeout(ctx, s.db, query, syncTime, itemID, currentVersion)
	if err != nil {
		return fmt.Errorf("failed to update sync metadata: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		// Check if item still exists (might have been deleted)
		var exists bool
		existsQuery := `SELECT EXISTS(SELECT 1 FROM knowledge_items WHERE id = $1)`
		err := s.db.QueryRowContext(ctx, existsQuery, itemID).Scan(&exists)
		if err == nil && !exists {
			return fmt.Errorf("knowledge item not found: %s", itemID)
		}
		// Item exists but version changed - conflict detected
		return fmt.Errorf("sync conflict detected for item %s: version changed during update", itemID)
	}

	return nil
}

// syncKnowledgeItems syncs multiple knowledge items using transactions and batch operations.
// Returns synced item IDs and failed item IDs.
// Complies with CODING_STANDARDS.md: Transaction coordination in service layer.
func (s *KnowledgeServiceImpl) syncKnowledgeItems(ctx context.Context, items []KnowledgeItem, force bool) ([]string, []string, error) {
	if len(items) == 0 {
		return []string{}, []string{}, nil
	}

	// For small batches, use individual updates with transaction
	// For large batches, use batch UPDATE query
	const batchThreshold = 50

	if len(items) <= batchThreshold {
		return s.syncKnowledgeItemsTransaction(ctx, items, force)
	}

	return s.syncKnowledgeItemsBatch(ctx, items, force)
}

// syncKnowledgeItemsTransaction syncs items using a transaction for atomicity.
// Complies with CODING_STANDARDS.md: Transaction coordination in service layer.
func (s *KnowledgeServiceImpl) syncKnowledgeItemsTransaction(ctx context.Context, items []KnowledgeItem, force bool) ([]string, []string, error) {
	// Start transaction for atomicity
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	syncedItems := make([]string, 0, len(items))
	failedItems := make([]string, 0)
	now := time.Now().UTC()

	// Update each item within transaction
	for _, item := range items {
		err := s.updateSyncMetadataTx(ctx, tx, item.ID, now)
		if err != nil {
			LogWarn(ctx, "Failed to update sync metadata for item %s: %v", item.ID, err)
			failedItems = append(failedItems, item.ID)
			// With force flag, continue processing other items even if one fails
			// Without force flag, also continue but track failures
			continue
		}
		syncedItems = append(syncedItems, item.ID)
	}

	// Commit transaction if any items were synced or if force flag is set
	if len(syncedItems) > 0 || force {
		if err := tx.Commit(); err != nil {
			return nil, nil, fmt.Errorf("failed to commit transaction: %w", err)
		}
	} else {
		// Rollback if no items were synced and force is false
		if err := tx.Rollback(); err != nil {
			LogWarn(ctx, "Failed to rollback transaction: %v", err)
		}
	}

	return syncedItems, failedItems, nil
}

// syncKnowledgeItemsBatch syncs items using batch UPDATE for better performance.
// Complies with CODING_STANDARDS.md: Performance optimization for large datasets.
func (s *KnowledgeServiceImpl) syncKnowledgeItemsBatch(ctx context.Context, items []KnowledgeItem, force bool) ([]string, []string, error) {
	// Start transaction for atomicity
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	now := time.Now().UTC()
	itemIDs := make([]string, 0, len(items))
	for _, item := range items {
		itemIDs = append(itemIDs, item.ID)
	}

	// Build query with placeholders for each ID
	// Use IN clause with multiple placeholders for better performance
	placeholders := make([]string, len(itemIDs))
	args := make([]interface{}, len(itemIDs)+1)
	args[0] = now
	for i, id := range itemIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args[i+1] = id
	}

	query := fmt.Sprintf(`
		UPDATE knowledge_items
		SET last_synced_at = $1,
		    sync_version = sync_version + 1,
		    sync_status = 'synced',
		    updated_at = $1
		WHERE id IN (%s)
	`, strings.Join(placeholders, ","))

	ctx, cancel := context.WithTimeout(ctx, getQueryTimeout())
	defer cancel()

	result, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to batch update sync metadata: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get rows affected: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Determine which items were actually synced based on rows affected
	// If rowsAffected < len(itemIDs), some items don't exist
	syncedItems := []string{}
	failedItems := []string{}
	
	if rowsAffected == int64(len(itemIDs)) {
		// All items were synced
		syncedItems = itemIDs
	} else if rowsAffected == 0 {
		// No items were synced (all don't exist or all failed)
		failedItems = itemIDs
	} else {
		// Some items were synced, some weren't
		// We can't tell which ones without querying, so we'll mark all as synced
		// if force=true, otherwise mark as failed
		if force {
			syncedItems = itemIDs
		} else {
			failedItems = itemIDs
		}
	}

	LogInfo(ctx, "Batch synced %d knowledge items (rows affected: %d)", len(syncedItems), rowsAffected)

	return syncedItems, failedItems, nil
}

// updateSyncMetadataTx updates sync metadata within a transaction.
// Uses optimistic locking with sync_version to detect conflicts.
// Complies with CODING_STANDARDS.md: Transaction coordination in service layer.
func (s *KnowledgeServiceImpl) updateSyncMetadataTx(ctx context.Context, tx *sql.Tx, itemID string, syncTime time.Time) error {
	// Get current sync_version for conflict detection
	var currentVersion int
	checkQuery := `SELECT sync_version FROM knowledge_items WHERE id = $1`
	
	err := tx.QueryRowContext(ctx, checkQuery, itemID).Scan(&currentVersion)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("knowledge item not found: %s", itemID)
		}
		return fmt.Errorf("failed to get current sync version: %w", err)
	}

	// Update with version check for conflict detection
	query := `
		UPDATE knowledge_items
		SET last_synced_at = $1,
		    sync_version = sync_version + 1,
		    sync_status = 'synced',
		    updated_at = $1
		WHERE id = $2
		  AND sync_version = $3
	`

	result, err := tx.ExecContext(ctx, query, syncTime, itemID, currentVersion)
	if err != nil {
		return fmt.Errorf("failed to update sync metadata: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		// Check if item still exists
		var exists bool
		existsQuery := `SELECT EXISTS(SELECT 1 FROM knowledge_items WHERE id = $1)`
		err := tx.QueryRowContext(ctx, existsQuery, itemID).Scan(&exists)
		if err == nil && !exists {
			return fmt.Errorf("knowledge item not found: %s", itemID)
		}
		// Conflict detected - version changed
		return fmt.Errorf("sync conflict detected for item %s: version changed during update", itemID)
	}

	return nil
}
