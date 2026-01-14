// Phase 12: Change Detection Module
// Detects changes when documents are re-ingested and generates change requests

package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"
)

// ChangeType represents the type of change detected
// ChangeType and ChangeRequest are defined in types.go

// KnowledgeItemChange represents a detected change in a knowledge item
type KnowledgeItemChange struct {
	Type    ChangeType             `json:"type"`
	OldItem KnowledgeItem          `json:"old_item,omitempty"`
	NewItem KnowledgeItem          `json:"new_item,omitempty"`
	Diff    map[string]interface{} `json:"diff,omitempty"`
}

// compareKnowledgeItems compares two knowledge items and determines the change type
func compareKnowledgeItems(oldItem, newItem KnowledgeItem) (ChangeType, map[string]interface{}) {
	// Calculate content hash for comparison
	oldHash := calculateContentHash(oldItem)
	newHash := calculateContentHash(newItem)

	if oldHash == newHash {
		return ChangeUnchanged, nil
	}

	diff := make(map[string]interface{})

	// Compare title
	if oldItem.Title != newItem.Title {
		diff["title"] = map[string]interface{}{
			"old": oldItem.Title,
			"new": newItem.Title,
		}
	}

	// Compare content
	if oldItem.Content != newItem.Content {
		diff["content"] = map[string]interface{}{
			"old": oldItem.Content,
			"new": newItem.Content,
		}
	}

	// Compare type
	if oldItem.Type != newItem.Type {
		diff["type"] = map[string]interface{}{
			"old": oldItem.Type,
			"new": newItem.Type,
		}
	}

	return ChangeModified, diff
}

// calculateContentHash calculates SHA256 hash of knowledge item content
func calculateContentHash(item KnowledgeItem) string {
	content := fmt.Sprintf("%s|%s|%s", item.Title, item.Content, item.Type)
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:])
}

// detectChanges compares old and new knowledge items and detects changes
func detectChanges(ctx context.Context, documentID string, newItems []KnowledgeItem) ([]KnowledgeItemChange, error) {
	var changes []KnowledgeItemChange

	// Load existing knowledge items for this document
	oldItems, err := loadKnowledgeItemsForDocument(ctx, documentID)
	if err != nil {
		LogError(ctx, "Failed to load existing knowledge items for document %s: %v", documentID, err)
		return nil, fmt.Errorf("failed to load existing knowledge items for document %s: %w", documentID, err)
	}

	// Create maps for easier lookup
	oldItemsMap := make(map[string]KnowledgeItem)
	for _, item := range oldItems {
		hash := calculateContentHash(item)
		oldItemsMap[hash] = item
		// Also index by title for fallback matching
		oldItemsMap[item.Title] = item
	}

	newItemsMap := make(map[string]KnowledgeItem)
	for _, item := range newItems {
		hash := calculateContentHash(item)
		newItemsMap[hash] = item
		newItemsMap[item.Title] = item
	}

	// Find new items (in new but not in old)
	for _, newItem := range newItems {
		newHash := calculateContentHash(newItem)
		if _, exists := oldItemsMap[newHash]; !exists {
			// Check by title as fallback
			if oldItem, existsByTitle := oldItemsMap[newItem.Title]; existsByTitle {
				// Item exists but content changed
				changeType, diff := compareKnowledgeItems(oldItem, newItem)
				if changeType != ChangeUnchanged {
					changes = append(changes, KnowledgeItemChange{
						Type:    changeType,
						OldItem: oldItem,
						NewItem: newItem,
						Diff:    diff,
					})
				}
			} else {
				// Truly new item
				changes = append(changes, KnowledgeItemChange{
					Type:    ChangeNew,
					NewItem: newItem,
				})
			}
		}
	}

	// Find removed items (in old but not in new)
	for _, oldItem := range oldItems {
		oldHash := calculateContentHash(oldItem)
		if _, exists := newItemsMap[oldHash]; !exists {
			// Check by title as fallback
			if _, existsByTitle := newItemsMap[oldItem.Title]; !existsByTitle {
				// Item was removed
				changes = append(changes, KnowledgeItemChange{
					Type:    ChangeRemoved,
					OldItem: oldItem,
				})
			}
		}
	}

	return changes, nil
}

// loadKnowledgeItemsForDocument loads existing knowledge items for a document
func loadKnowledgeItemsForDocument(ctx context.Context, documentID string) ([]KnowledgeItem, error) {
	query := `
		SELECT id, document_id, type, title, content, confidence, source_page, status, created_at
		FROM knowledge_items
		WHERE document_id = $1
		ORDER BY created_at DESC
	`

	ctx, cancel := context.WithTimeout(ctx, getQueryTimeout())
	defer cancel()

	rows, err := db.QueryContext(ctx, query, documentID)
	if err != nil {
		LogError(ctx, "Failed to query knowledge items for document %s: %v", documentID, err)
		return nil, fmt.Errorf("failed to query knowledge items for document %s: %w", documentID, err)
	}
	defer rows.Close()

	var items []KnowledgeItem
	for rows.Next() {
		var item KnowledgeItem
		var sourcePage sql.NullInt32

		err := rows.Scan(
			&item.ID, &item.DocumentID, &item.Type, &item.Title, &item.Content,
			&item.Confidence, &sourcePage, &item.Status, &item.CreatedAt,
		)
		if err != nil {
			LogWarn(ctx, "Error scanning knowledge item: %v", err)
			continue
		}

		if sourcePage.Valid {
			item.SourcePage = int(sourcePage.Int32)
		}

		items = append(items, item)
	}

	return items, nil
}

// generateChangeRequest creates a change request from a knowledge item change
func generateChangeRequest(ctx context.Context, projectID string, change KnowledgeItemChange) (*ChangeRequest, error) {
	cr := &ChangeRequest{
		ProjectID: projectID,
		Type:      change.Type,
		Status:    "pending_approval",
		CreatedAt: time.Now(),
	}

	// Generate CR-XXX ID
	crID, err := generateChangeRequestID(ctx, projectID)
	if err != nil {
		LogError(ctx, "Failed to generate change request ID for project %s: %v", projectID, err)
		return nil, fmt.Errorf("failed to generate change request ID for project %s: %w", projectID, err)
	}
	cr.ID = crID

	// Set knowledge item ID
	if change.OldItem.ID != "" {
		cr.KnowledgeItemID = &change.OldItem.ID
	} else if change.NewItem.ID != "" {
		cr.KnowledgeItemID = &change.NewItem.ID
	}

	// Set current and proposed states
	if change.OldItem.ID != "" {
		cr.CurrentState = knowledgeItemToMap(change.OldItem)
	}
	if change.NewItem.ID != "" || change.Type == ChangeNew {
		cr.ProposedState = knowledgeItemToMap(change.NewItem)
	}

	// If it's a removal, proposed state is empty
	if change.Type == ChangeRemoved {
		cr.ProposedState = make(map[string]interface{})
	}

	return cr, nil
}

// knowledgeItemToMap converts a knowledge item to a map for JSON storage
func knowledgeItemToMap(item KnowledgeItem) map[string]interface{} {
	return map[string]interface{}{
		"id":         item.ID,
		"type":       item.Type,
		"title":      item.Title,
		"content":    item.Content,
		"confidence": item.Confidence,
		"status":     item.Status,
	}
}

// generateChangeRequestID generates a unique change request ID (CR-XXX)
func generateChangeRequestID(ctx context.Context, projectID string) (string, error) {
	// Use sequence table for atomic ID generation
	query := `
		INSERT INTO change_request_sequences (project_id) 
		VALUES ($1) 
		RETURNING sequence_number
	`

	ctx, cancel := context.WithTimeout(ctx, getQueryTimeout())
	defer cancel()

	var seqNum int
	err := queryRowWithTimeout(ctx, query, projectID).Scan(&seqNum)
	if err != nil {
		return "", fmt.Errorf("failed to generate change request ID: %w", err)
	}

	return fmt.Sprintf("CR-%03d", seqNum), nil
}

// storeChangeRequest stores a change request in the database
func storeChangeRequest(ctx context.Context, cr *ChangeRequest) error {
	currentStateJSON, err := marshalJSONB(cr.CurrentState)
	if err != nil {
		return fmt.Errorf("failed to marshal current state: %w", err)
	}

	proposedStateJSON, err := marshalJSONB(cr.ProposedState)
	if err != nil {
		return fmt.Errorf("failed to marshal proposed state: %w", err)
	}

	query := `
		INSERT INTO change_requests (
			id, project_id, knowledge_item_id, type, current_state, proposed_state, status, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	queryCtx, cancel := context.WithTimeout(ctx, getQueryTimeout())
	defer cancel()

	_, err = execWithTimeout(queryCtx, query,
		cr.ID,
		cr.ProjectID,
		cr.KnowledgeItemID,
		string(cr.Type),
		currentStateJSON,
		proposedStateJSON,
		cr.Status,
		cr.CreatedAt,
	)

	if err != nil {
		LogError(ctx, "Failed to store change request %s for project %s: %v", cr.ID, cr.ProjectID, err)
		return fmt.Errorf("failed to store change request %s: %w", cr.ID, err)
	}

	LogInfo(ctx, "Successfully stored change request %s (type: %s, project: %s)", cr.ID, cr.Type, cr.ProjectID)
	return nil
}

// processChangeDetectionForDocument processes change detection after knowledge extraction
// This function should be called after knowledge items are extracted/updated for a document
func processChangeDetectionForDocument(ctx context.Context, documentID string, projectID string) ([]string, error) {
	// Load newly extracted knowledge items
	newItems, err := loadKnowledgeItemsForDocument(ctx, documentID)
	if err != nil {
		return nil, fmt.Errorf("failed to load knowledge items: %w", err)
	}

	// Detect changes
	changes, err := detectChanges(ctx, documentID, newItems)
	if err != nil {
		return nil, fmt.Errorf("failed to detect changes: %w", err)
	}

	var changeRequestIDs []string

	// Generate change requests for each change
	for _, change := range changes {
		cr, err := generateChangeRequest(ctx, projectID, change)
		if err != nil {
			LogError(ctx, "Failed to generate change request: %v", err)
			continue
		}

		// Store change request
		err = storeChangeRequest(ctx, cr)
		if err != nil {
			LogError(ctx, "Failed to store change request: %v", err)
			continue
		}

		changeRequestIDs = append(changeRequestIDs, cr.ID)
		LogInfo(ctx, "Created change request %s for document %s", cr.ID, documentID)
	}

	return changeRequestIDs, nil
}
