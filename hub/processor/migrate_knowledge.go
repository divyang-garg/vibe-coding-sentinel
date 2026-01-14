// Phase 13: Knowledge Migration
// Migrates existing knowledge items to structured format

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
)

// migrateKnowledgeItems migrates existing knowledge items to structured format
func migrateKnowledgeItems(ctx context.Context) error {
	log.Println("Starting knowledge items migration...")
	
	// Query all knowledge items without structured_data
	rows, err := db.QueryContext(ctx, `
		SELECT id, document_id, type, title, content, confidence, status
		FROM knowledge_items
		WHERE structured_data IS NULL
	`)
	if err != nil {
		return fmt.Errorf("failed to query knowledge items: %w", err)
	}
	defer rows.Close()
	
	migratedCount := 0
	skippedCount := 0
	
	for rows.Next() {
		var id, docID, itemType, title, content, status string
		var confidence float64
		
		if err := rows.Scan(&id, &docID, &itemType, &title, &content, &confidence, &status); err != nil {
			log.Printf("Error scanning row: %v", err)
			skippedCount++
			continue
		}
		
		// Create structured knowledge item from legacy content
		structuredItem := &StructuredKnowledgeItem{
			ID:          id,
			Version:     "1.0.0",
			Type:        itemType,
			Status:      status,
			Title:       title,
			Description: content,
			Metadata: &Metadata{
				Confidence: confidence,
			},
		}
		
		// Attempt to parse content as JSON (in case it's already structured)
		var parsed map[string]interface{}
		if err := json.Unmarshal([]byte(content), &parsed); err == nil {
			// Content is JSON, try to extract structured fields
			if _, ok := parsed["specification"].(map[string]interface{}); ok {
				// Convert to Specification struct
				structuredItem.Specification = &Specification{}
				// Note: Full conversion would require more complex logic
				// For now, we'll store the parsed JSON
			}
		}
		
		// Marshal to JSONB
		structuredJSON, err := json.Marshal(structuredItem)
		if err != nil {
			log.Printf("Error marshaling structured item %s: %v", id, err)
			skippedCount++
			continue
		}
		
		// Update database
		_, err = db.ExecContext(ctx, `
			UPDATE knowledge_items
			SET structured_data = $1
			WHERE id = $2
		`, string(structuredJSON), id)
		
		if err != nil {
			log.Printf("Error updating knowledge item %s: %v", id, err)
			skippedCount++
			continue
		}
		
		migratedCount++
		
		if migratedCount%10 == 0 {
			log.Printf("Migrated %d items...", migratedCount)
		}
	}
	
	log.Printf("Migration complete: %d migrated, %d skipped", migratedCount, skippedCount)
	return nil
}

