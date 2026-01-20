// Test Requirement Generator - Extractors
// Extracts business rules, user journeys, and entities from knowledge base
// Phase 14D: Enhanced with caching support
// Complies with CODING_STANDARDS.md: Business Services max 400 lines

package services

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"sentinel-hub-api/pkg/database"
)

// extractBusinessRules extracts approved business rules from knowledge_items table
// Phase 14D: Now supports caching via codebaseHash parameter
func extractBusinessRules(ctx context.Context, projectID string, knowledgeItemIDs []string, codebaseHash string, config *LLMConfig) ([]KnowledgeItem, error) {
	// Phase 14D: Check cache first if codebaseHash is provided
	if codebaseHash != "" && config != nil {
		cached, ok := getCachedBusinessContext(projectID, codebaseHash, config)
		if ok {
			if rules, ok := cached["rules"].([]interface{}); ok {
				// Convert []interface{} to []KnowledgeItem
				result := make([]KnowledgeItem, 0, len(rules))
				for _, r := range rules {
					if rule, ok := r.(KnowledgeItem); ok {
						result = append(result, rule)
					}
				}
				if len(result) > 0 {
					recordCacheHit(projectID)
					return result, nil
				}
			}
		}
	}

	query := `
		SELECT ki.id, ki.document_id, ki.type, ki.title, ki.content, ki.confidence, 
		       ki.source_page, ki.status, ki.approved_by, ki.approved_at, ki.created_at
		FROM knowledge_items ki
		INNER JOIN documents d ON ki.document_id = d.id
		WHERE d.project_id = $1
		  AND ki.type = 'business_rule'
		  AND ki.status = 'approved'
	`

	args := []interface{}{projectID}
	argIndex := 2

	// If specific knowledge item IDs provided, filter by them
	if len(knowledgeItemIDs) > 0 {
		query += fmt.Sprintf(" AND ki.id = ANY($%d)", argIndex)
		args = append(args, knowledgeItemIDs)
		argIndex++
	}

	query += " ORDER BY ki.created_at DESC"

	rows, err := database.QueryWithTimeout(ctx, db, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query business rules: %w", err)
	}
	defer rows.Close()

	var rules []KnowledgeItem
	for rows.Next() {
		var rule KnowledgeItem
		var approvedBy sql.NullString
		var approvedAt sql.NullTime

		err := rows.Scan(
			&rule.ID, &rule.DocumentID, &rule.Type, &rule.Title, &rule.Content,
			&rule.Confidence, &rule.SourcePage, &rule.Status,
			&approvedBy, &approvedAt, &rule.CreatedAt,
		)
		if err != nil {
			log.Printf("Error scanning business rule: %v", err)
			continue
		}

		if approvedBy.Valid {
			rule.ApprovedBy = &approvedBy.String
		}
		if approvedAt.Valid {
			rule.ApprovedAt = &approvedAt.Time
		}

		rules = append(rules, rule)
	}

	// Phase 14D: Cache the results if codebaseHash is provided
	if codebaseHash != "" && config != nil && len(rules) > 0 {
		// Convert []KnowledgeItem to []interface{} for caching
		rulesInterface := make([]interface{}, len(rules))
		for i, r := range rules {
			rulesInterface[i] = r
		}
		// Get existing cached context to preserve entities and journeys
		cached, _ := getCachedBusinessContext(projectID, codebaseHash, config)
		var entities, journeys []interface{}
		if cached != nil {
			if e, ok := cached["entities"].([]interface{}); ok {
				entities = e
			}
			if j, ok := cached["journeys"].([]interface{}); ok {
				journeys = j
			}
		}
		setCachedBusinessContext(projectID, codebaseHash, rulesInterface, entities, journeys, config)
		recordCacheMiss(projectID)
	}

	return rules, nil
}

// extractUserJourneys extracts approved user journeys from knowledge_items table
// Phase 14D: Now supports caching via codebaseHash parameter
func extractUserJourneys(ctx context.Context, projectID string, journeyIDs []string, codebaseHash string, config *LLMConfig) ([]KnowledgeItem, error) {
	// Phase 14D: Check cache first if codebaseHash is provided
	if codebaseHash != "" && config != nil {
		cached, ok := getCachedBusinessContext(projectID, codebaseHash, config)
		if ok {
			if journeys, ok := cached["journeys"].([]interface{}); ok {
				// Convert []interface{} to []KnowledgeItem
				result := make([]KnowledgeItem, 0, len(journeys))
				for _, j := range journeys {
					if journey, ok := j.(KnowledgeItem); ok {
						result = append(result, journey)
					}
				}
				if len(result) > 0 {
					recordCacheHit(projectID)
					return result, nil
				}
			}
		}
	}

	query := `
		SELECT ki.id, ki.document_id, ki.type, ki.title, ki.content, ki.confidence, 
		       ki.source_page, ki.status, ki.approved_by, ki.approved_at, ki.created_at
		FROM knowledge_items ki
		INNER JOIN documents d ON ki.document_id = d.id
		WHERE d.project_id = $1
		  AND ki.type = 'journey'
		  AND ki.status = 'approved'
	`

	args := []interface{}{projectID}
	argIndex := 2

	// If specific journey IDs provided, filter by them
	if len(journeyIDs) > 0 {
		query += fmt.Sprintf(" AND ki.id = ANY($%d)", argIndex)
		args = append(args, journeyIDs)
		argIndex++
	}

	query += " ORDER BY ki.created_at DESC"

	rows, err := database.QueryWithTimeout(ctx, db, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query user journeys: %w", err)
	}
	defer rows.Close()

	var journeys []KnowledgeItem
	for rows.Next() {
		var journey KnowledgeItem
		var approvedBy sql.NullString
		var approvedAt sql.NullTime

		err := rows.Scan(
			&journey.ID, &journey.DocumentID, &journey.Type, &journey.Title, &journey.Content,
			&journey.Confidence, &journey.SourcePage, &journey.Status,
			&approvedBy, &approvedAt, &journey.CreatedAt,
		)
		if err != nil {
			log.Printf("Error scanning user journey: %v", err)
			continue
		}

		if approvedBy.Valid {
			journey.ApprovedBy = &approvedBy.String
		}
		if approvedAt.Valid {
			journey.ApprovedAt = &approvedAt.Time
		}

		journeys = append(journeys, journey)
	}

	// Phase 14D: Cache the results if codebaseHash is provided
	if codebaseHash != "" && config != nil && len(journeys) > 0 {
		// Convert []KnowledgeItem to []interface{} for caching
		journeysInterface := make([]interface{}, len(journeys))
		for i, j := range journeys {
			journeysInterface[i] = j
		}
		// Get existing cached context to preserve rules and entities
		cached, _ := getCachedBusinessContext(projectID, codebaseHash, config)
		var rules, entities []interface{}
		if cached != nil {
			if r, ok := cached["rules"].([]interface{}); ok {
				rules = r
			}
			if e, ok := cached["entities"].([]interface{}); ok {
				entities = e
			}
		}
		setCachedBusinessContext(projectID, codebaseHash, rules, entities, journeysInterface, config)
		recordCacheMiss(projectID)
	}

	return journeys, nil
}

// extractEntities extracts approved entities from knowledge_items table
func extractEntities(ctx context.Context, projectID string, entityIDs []string) ([]KnowledgeItem, error) {
	query := `
		SELECT ki.id, ki.document_id, ki.type, ki.title, ki.content, ki.confidence, 
		       ki.source_page, ki.status, ki.approved_by, ki.approved_at, ki.created_at
		FROM knowledge_items ki
		INNER JOIN documents d ON ki.document_id = d.id
		WHERE d.project_id = $1
		  AND ki.type = 'entity'
		  AND ki.status = 'approved'
	`

	args := []interface{}{projectID}
	argIndex := 2

	// If specific entity IDs provided, filter by them
	if len(entityIDs) > 0 {
		query += fmt.Sprintf(" AND ki.id = ANY($%d)", argIndex)
		args = append(args, entityIDs)
		argIndex++
	}

	query += " ORDER BY ki.created_at DESC"

	rows, err := database.QueryWithTimeout(ctx, db, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query entities: %w", err)
	}
	defer rows.Close()

	var entities []KnowledgeItem
	for rows.Next() {
		var entity KnowledgeItem
		var approvedBy sql.NullString
		var approvedAt sql.NullTime

		err := rows.Scan(
			&entity.ID, &entity.DocumentID, &entity.Type, &entity.Title, &entity.Content,
			&entity.Confidence, &entity.SourcePage, &entity.Status,
			&approvedBy, &approvedAt, &entity.CreatedAt,
		)
		if err != nil {
			log.Printf("Error scanning entity: %v", err)
			continue
		}

		if approvedBy.Valid {
			entity.ApprovedBy = &approvedBy.String
		}
		if approvedAt.Valid {
			entity.ApprovedAt = &approvedAt.Time
		}

		entities = append(entities, entity)
	}

	return entities, nil
}
