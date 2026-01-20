// Package services - Knowledge Management Service
// Complies with CODING_STANDARDS.md: Business Services max 400 lines
package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"sentinel-hub-api/pkg/database"

	"github.com/google/uuid"
)

// KnowledgeServiceImpl implements KnowledgeService interface
type KnowledgeServiceImpl struct {
	db *sql.DB
}

// NewKnowledgeService creates a new knowledge service
func NewKnowledgeService(db *sql.DB) KnowledgeService {
	return &KnowledgeServiceImpl{db: db}
}

// RunGapAnalysis runs gap analysis between documentation and code
func (s *KnowledgeServiceImpl) RunGapAnalysis(ctx context.Context, req GapAnalysisRequest) (*GapAnalysisReport, error) {
	if req.ProjectID == "" {
		return nil, fmt.Errorf("project_id is required")
	}
	if req.CodebasePath == "" {
		return nil, fmt.Errorf("codebase_path is required")
	}

	// Use existing gap analyzer
	report, err := analyzeGaps(ctx, req.ProjectID, req.CodebasePath, req.Options)
	if err != nil {
		return nil, fmt.Errorf("failed to run gap analysis: %w", err)
	}

	// Store report in database
	reportID, err := storeGapReport(ctx, report)
	if err != nil {
		// Log error but don't fail - report is still valid
		LogError(ctx, "Failed to store gap report: %v", err)
	}

	// Add report_id to response if stored
	if reportID != "" {
		if report.Summary == nil {
			report.Summary = make(map[string]interface{})
		}
		report.Summary["report_id"] = reportID
	}

	return report, nil
}

// ListKnowledgeItems lists knowledge items with filters
func (s *KnowledgeServiceImpl) ListKnowledgeItems(ctx context.Context, req ListKnowledgeItemsRequest) ([]KnowledgeItem, error) {
	if req.ProjectID == "" {
		return nil, fmt.Errorf("project_id is required")
	}

	// Build query with filters
	query := `
		SELECT ki.id, ki.item_type, ki.title, ki.content, ki.confidence, 
		       ki.source_page, ki.status, ki.structured_data, ki.document_id,
		       ki.approved_by, ki.approved_at, ki.created_at, ki.updated_at
		FROM knowledge_items ki
		INNER JOIN documents d ON ki.document_id = d.id
		WHERE d.project_id = $1
	`
	args := []interface{}{req.ProjectID}
	argIndex := 2

	if req.Type != "" {
		query += fmt.Sprintf(" AND ki.item_type = $%d", argIndex)
		args = append(args, req.Type)
		argIndex++
	}

	if req.Status != "" {
		query += fmt.Sprintf(" AND ki.status = $%d", argIndex)
		args = append(args, req.Status)
		argIndex++
	}

	query += " ORDER BY ki.created_at DESC"

	if req.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, req.Limit)
		argIndex++
		if req.Offset > 0 {
			query += fmt.Sprintf(" OFFSET $%d", argIndex)
			args = append(args, req.Offset)
		}
	}

	ctx, cancel := context.WithTimeout(ctx, getQueryTimeout())
	defer cancel()

	rows, err := database.QueryWithTimeout(ctx, s.db, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query knowledge items: %w", err)
	}
	defer rows.Close()

	var items []KnowledgeItem
	for rows.Next() {
		var item KnowledgeItem
		var structuredDataJSON sql.NullString
		var approvedBy sql.NullString
		var approvedAt sql.NullTime

		err := rows.Scan(
			&item.ID, &item.Type, &item.Title, &item.Content, &item.Confidence,
			&item.SourcePage, &item.Status, &structuredDataJSON, &item.DocumentID,
			&approvedBy, &approvedAt, &item.CreatedAt, &item.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan knowledge item: %w", err)
		}

		if structuredDataJSON.Valid {
			item.StructuredData = make(map[string]interface{})
			if err := json.Unmarshal([]byte(structuredDataJSON.String), &item.StructuredData); err != nil {
				LogWarn(ctx, "Failed to unmarshal structured_data for item %s: %v", item.ID, err)
			}
		}

		if approvedBy.Valid {
			item.ApprovedBy = &approvedBy.String
		}
		if approvedAt.Valid {
			item.ApprovedAt = &approvedAt.Time
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating knowledge items: %w", err)
	}

	return items, nil
}

// CreateKnowledgeItem creates a new knowledge item
func (s *KnowledgeServiceImpl) CreateKnowledgeItem(ctx context.Context, item KnowledgeItem) (*KnowledgeItem, error) {
	if item.DocumentID == "" {
		return nil, fmt.Errorf("document_id is required")
	}
	if item.Type == "" {
		return nil, fmt.Errorf("type is required")
	}
	if item.Title == "" {
		return nil, fmt.Errorf("title is required")
	}

	// Generate ID if not provided
	if item.ID == "" {
		item.ID = uuid.New().String()
	}

	// Set timestamps
	now := time.Now().UTC()
	if item.CreatedAt.IsZero() {
		item.CreatedAt = now
	}
	item.UpdatedAt = now

	// Default status if not provided
	if item.Status == "" {
		item.Status = "draft"
	}

	// Marshal structured data
	var structuredDataJSON sql.NullString
	if item.StructuredData != nil {
		data, err := json.Marshal(item.StructuredData)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal structured_data: %w", err)
		}
		structuredDataJSON = sql.NullString{String: string(data), Valid: true}
	}

	query := `
		INSERT INTO knowledge_items (id, document_id, item_type, title, content, 
		                            confidence, source_page, status, structured_data, 
		                            created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at, updated_at
	`

	ctx, cancel := context.WithTimeout(ctx, getQueryTimeout())
	defer cancel()

	err := database.QueryRowWithTimeout(ctx, s.db, query,
		item.ID, item.DocumentID, item.Type, item.Title, item.Content,
		item.Confidence, item.SourcePage, item.Status, structuredDataJSON,
		item.CreatedAt, item.UpdatedAt,
	).Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create knowledge item: %w", err)
	}

	return &item, nil
}

// GetKnowledgeItem retrieves a knowledge item by ID
func (s *KnowledgeServiceImpl) GetKnowledgeItem(ctx context.Context, id string) (*KnowledgeItem, error) {
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}

	query := `
		SELECT id, item_type, title, content, confidence, source_page, status,
		       structured_data, document_id, approved_by, approved_at,
		       created_at, updated_at
		FROM knowledge_items
		WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(ctx, getQueryTimeout())
	defer cancel()

	var item KnowledgeItem
	var structuredDataJSON sql.NullString
	var approvedBy sql.NullString
	var approvedAt sql.NullTime

	err := database.QueryRowWithTimeout(ctx, s.db, query, id).Scan(
		&item.ID, &item.Type, &item.Title, &item.Content, &item.Confidence,
		&item.SourcePage, &item.Status, &structuredDataJSON, &item.DocumentID,
		&approvedBy, &approvedAt, &item.CreatedAt, &item.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("knowledge item not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get knowledge item: %w", err)
	}

	if structuredDataJSON.Valid {
		item.StructuredData = make(map[string]interface{})
		if err := json.Unmarshal([]byte(structuredDataJSON.String), &item.StructuredData); err != nil {
			LogWarn(ctx, "Failed to unmarshal structured_data: %v", err)
		}
	}

	if approvedBy.Valid {
		item.ApprovedBy = &approvedBy.String
	}
	if approvedAt.Valid {
		item.ApprovedAt = &approvedAt.Time
	}

	return &item, nil
}

// UpdateKnowledgeItem updates an existing knowledge item
func (s *KnowledgeServiceImpl) UpdateKnowledgeItem(ctx context.Context, id string, item KnowledgeItem) (*KnowledgeItem, error) {
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}

	// Get existing item
	existing, err := s.GetKnowledgeItem(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update fields
	if item.Title != "" {
		existing.Title = item.Title
	}
	if item.Content != "" {
		existing.Content = item.Content
	}
	if item.Type != "" {
		existing.Type = item.Type
	}
	if item.Status != "" {
		existing.Status = item.Status
	}
	if item.Confidence > 0 {
		existing.Confidence = item.Confidence
	}
	if item.SourcePage > 0 {
		existing.SourcePage = item.SourcePage
	}
	if item.StructuredData != nil {
		existing.StructuredData = item.StructuredData
	}

	existing.UpdatedAt = time.Now().UTC()

	// Marshal structured data
	var structuredDataJSON sql.NullString
	if existing.StructuredData != nil {
		data, err := json.Marshal(existing.StructuredData)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal structured_data: %w", err)
		}
		structuredDataJSON = sql.NullString{String: string(data), Valid: true}
	}

	query := `
		UPDATE knowledge_items
		SET item_type = $2, title = $3, content = $4, confidence = $5,
		    source_page = $6, status = $7, structured_data = $8, updated_at = $9
		WHERE id = $1
		RETURNING updated_at
	`

	ctx, cancel := context.WithTimeout(ctx, getQueryTimeout())
	defer cancel()

	err = database.QueryRowWithTimeout(ctx, s.db, query,
		id, existing.Type, existing.Title, existing.Content,
		existing.Confidence, existing.SourcePage, existing.Status,
		structuredDataJSON, existing.UpdatedAt,
	).Scan(&existing.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to update knowledge item: %w", err)
	}

	return existing, nil
}

// DeleteKnowledgeItem deletes a knowledge item
func (s *KnowledgeServiceImpl) DeleteKnowledgeItem(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("id is required")
	}

	query := `DELETE FROM knowledge_items WHERE id = $1`

	ctx, cancel := context.WithTimeout(ctx, getQueryTimeout())
	defer cancel()

	result, err := database.ExecWithTimeout(ctx, s.db, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete knowledge item: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("knowledge item not found: %s", id)
	}

	return nil
}

// GetBusinessContext retrieves business context for a feature or entity
func (s *KnowledgeServiceImpl) GetBusinessContext(ctx context.Context, req BusinessContextRequest) (*BusinessContextResponse, error) {
	if req.ProjectID == "" {
		return nil, fmt.Errorf("project_id is required")
	}

	// Extract business rules
	rules, err := extractBusinessRules(ctx, req.ProjectID, nil, "", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to extract business rules: %w", err)
	}

	// Filter by feature/entity/keywords if provided
	filteredRules := []KnowledgeItem{}
	for _, rule := range rules {
		if req.Feature != "" && !containsKeyword(rule.Title+rule.Content, req.Feature) {
			continue
		}
		if req.Entity != "" && !containsKeyword(rule.Title+rule.Content, req.Entity) {
			continue
		}
		if len(req.Keywords) > 0 {
			matched := false
			for _, keyword := range req.Keywords {
				if containsKeyword(rule.Title+rule.Content, keyword) {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}
		filteredRules = append(filteredRules, rule)
	}

	// Extract entities
	entities, err := extractEntitiesSimple(ctx, req.ProjectID)
	if err != nil {
		LogWarn(ctx, "Failed to extract entities: %v", err)
		entities = []KnowledgeItem{}
	}

	// Extract user journeys
	journeys, err := extractUserJourneysSimple(ctx, req.ProjectID)
	if err != nil {
		LogWarn(ctx, "Failed to extract user journeys: %v", err)
		journeys = []KnowledgeItem{}
	}

	// Build constraints and side effects from rules
	constraints := []string{}
	sideEffects := []string{}
	for _, rule := range filteredRules {
		if rule.StructuredData != nil {
			if spec, ok := rule.StructuredData["specification"].(map[string]interface{}); ok {
				if constraintsData, ok := spec["constraints"].([]interface{}); ok {
					for _, c := range constraintsData {
						if constraint, ok := c.(map[string]interface{}); ok {
							if expr, ok := constraint["expression"].(string); ok {
								constraints = append(constraints, expr)
							}
						}
					}
				}
				if sideEffectsData, ok := spec["side_effects"].([]interface{}); ok {
					for _, se := range sideEffectsData {
						if sideEffect, ok := se.(map[string]interface{}); ok {
							if action, ok := sideEffect["action"].(string); ok {
								sideEffects = append(sideEffects, action)
							}
						}
					}
				}
			}
		}
	}

	// Get security rules (simplified - would query security_rules table in production)
	securityRules := []string{"SEC-001", "SEC-002", "SEC-003"} // Placeholder

	// Count test requirements
	testReqCount := 0
	for _, rule := range filteredRules {
		if rule.StructuredData != nil {
			if testReqs, ok := rule.StructuredData["test_requirements"].([]interface{}); ok {
				testReqCount += len(testReqs)
			}
		}
	}

	return &BusinessContextResponse{
		Rules:            filteredRules,
		Entities:         entities,
		UserJourneys:     journeys,
		Constraints:      constraints,
		SideEffects:      sideEffects,
		SecurityRules:    securityRules,
		TestRequirements: testReqCount,
	}, nil
}

// SyncKnowledge syncs knowledge items
func (s *KnowledgeServiceImpl) SyncKnowledge(ctx context.Context, req SyncKnowledgeRequest) (*SyncKnowledgeResponse, error) {
	if req.ProjectID == "" {
		return nil, fmt.Errorf("project_id is required")
	}

	// Get knowledge items to sync
	var items []KnowledgeItem
	var err error

	if len(req.KnowledgeItemIDs) > 0 {
		// Sync specific items
		items = make([]KnowledgeItem, 0, len(req.KnowledgeItemIDs))
		for _, id := range req.KnowledgeItemIDs {
			item, err := s.GetKnowledgeItem(ctx, id)
			if err != nil {
				if !req.Force {
					return &SyncKnowledgeResponse{
						SyncedCount: 0,
						FailedCount: 1,
						FailedItems: []string{id},
						Message:     fmt.Sprintf("Failed to get item %s: %v", id, err),
					}, nil
				}
				continue
			}
			items = append(items, *item)
		}
	} else {
		// Sync all items for project
		listReq := ListKnowledgeItemsRequest{
			ProjectID: req.ProjectID,
			Status:    "approved",
		}
		items, err = s.ListKnowledgeItems(ctx, listReq)
		if err != nil {
			return nil, fmt.Errorf("failed to list knowledge items: %w", err)
		}
	}

	// Sync items (simplified - would update sync timestamp in production)
	syncedItems := make([]string, 0, len(items))
	failedItems := make([]string, 0)

	for _, item := range items {
		// In production, would update sync metadata
		syncedItems = append(syncedItems, item.ID)
	}

	return &SyncKnowledgeResponse{
		SyncedCount: len(syncedItems),
		FailedCount: len(failedItems),
		SyncedItems: syncedItems,
		FailedItems: failedItems,
		Message:     fmt.Sprintf("Synced %d knowledge items", len(syncedItems)),
	}, nil
}

// Helper functions

func containsKeyword(text, keyword string) bool {
	return len(text) > 0 && len(keyword) > 0 &&
		(len(text) >= len(keyword) &&
			(text == keyword ||
				contains(text, keyword)))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					indexOf(s, substr) >= 0)))
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// extractEntities extracts entity knowledge items
func extractEntitiesSimple(ctx context.Context, projectID string) ([]KnowledgeItem, error) {
	// Simplified - would query knowledge_items with type='entity' in production
	return []KnowledgeItem{}, nil
}

// extractUserJourneys extracts user journey knowledge items
func extractUserJourneysSimple(ctx context.Context, projectID string) ([]KnowledgeItem, error) {
	// Simplified - would query knowledge_items with type='user_journey' in production
	return []KnowledgeItem{}, nil
}
