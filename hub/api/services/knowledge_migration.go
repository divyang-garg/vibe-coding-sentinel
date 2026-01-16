// Fixed import structure
package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"sentinel-hub-api/pkg/database"
)

// migrateKnowledgeItemsAPI migrates existing knowledge items to structured format
// This is the API version that can be called from the Hub API endpoint
func migrateKnowledgeItemsAPI(ctx context.Context) (int, int, error) {
	// Query all knowledge items without structured_data
	rows, err := database.QueryWithTimeout(ctx, db, `
		SELECT id, document_id, type, title, content, confidence, status
		FROM knowledge_items
		WHERE structured_data IS NULL
	`)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to query knowledge items: %w", err)
	}
	defer rows.Close()

	migratedCount := 0
	skippedCount := 0

	for rows.Next() {
		var id, docID, itemType, title, content, status string
		var confidence float64

		if err := rows.Scan(&id, &docID, &itemType, &title, &content, &confidence, &status); err != nil {
			LogWarn(ctx, "Error scanning knowledge item row: %v", err)
			skippedCount++
			continue
		}

		// Create structured knowledge item from legacy content
		// Using a simplified structure that matches the database schema
		structuredItem := map[string]interface{}{
			"id":          id,
			"version":     "1.0.0",
			"type":        itemType,
			"status":      status,
			"title":       title,
			"description": content,
			"metadata": map[string]interface{}{
				"confidence": confidence,
			},
		}

		// Attempt to parse content as JSON (in case it's already structured)
		var parsed map[string]interface{}
		if err := json.Unmarshal([]byte(content), &parsed); err == nil {
			// Content is JSON, try to extract structured fields
			if spec, ok := parsed["specification"].(map[string]interface{}); ok {
				structuredItem["specification"] = spec
			}
			if testReqs, ok := parsed["test_requirements"].([]interface{}); ok {
				structuredItem["test_requirements"] = testReqs
			}
			if traceability, ok := parsed["traceability"].(map[string]interface{}); ok {
				structuredItem["traceability"] = traceability
			}
		}

		// Marshal to JSONB
		structuredJSON, err := json.Marshal(structuredItem)
		if err != nil {
			LogWarn(ctx, "Error marshaling structured item %s: %v", id, err)
			skippedCount++
			continue
		}

		// Update database
		_, err = database.ExecWithTimeout(ctx, db, `
			UPDATE knowledge_items
			SET structured_data = $1
			WHERE id = $2
		`, string(structuredJSON), id)

		if err != nil {
			LogWarn(ctx, "Error updating knowledge item %s: %v", id, err)
			skippedCount++
			continue
		}

		migratedCount++

		if migratedCount%10 == 0 {
			LogInfo(ctx, "Migrated %d items...", migratedCount)
		}
	}

	return migratedCount, skippedCount, nil
}

// migrateKnowledgeHandler handles the POST /api/v1/knowledge/migrate endpoint
func migrateKnowledgeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	project, err := getProjectFromContext(r.Context())
	if err != nil {
		LogError(r.Context(), "Failed to get project from context: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Check if user has permission (optional: add role check)
	// For now, any authenticated user can trigger migration

	LogInfo(ctx, "Knowledge migration requested by project %s", project.ID)

	// Create context with timeout for migration
	migrationCtx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()

	// Perform migration
	migratedCount, skippedCount, err := migrateKnowledgeItemsAPI(migrationCtx)
	if err != nil {
		LogError(ctx, "Migration failed: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	LogInfo(ctx, "Migration complete: %d migrated, %d skipped", migratedCount, skippedCount)

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":         true,
		"migrated_count":  migratedCount,
		"skipped_count":   skippedCount,
		"total_processed": migratedCount + skippedCount,
		"message":         fmt.Sprintf("Successfully migrated %d knowledge items", migratedCount),
	})
}
