// Package services - Knowledge Service Integration Tests
// Complies with CODING_STANDARDS.md: Integration tests for end-to-end workflows
package services

import (
	"context"
	"testing"
	"time"

	"sentinel-hub-api/database"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestKnowledgeServiceIntegration tests end-to-end knowledge service workflows
func TestKnowledgeServiceIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := database.SetupTestDB(t)
	defer database.TeardownTestDB(t, db)
	defer database.CleanupTestData(t, db)

	service := NewKnowledgeService(db).(*KnowledgeServiceImpl)
	ctx := context.Background()

	// Setup: Create test project and document
	projectID := uuid.New().String()
	documentID := uuid.New().String()

	// Create test project
	_, err := db.ExecContext(ctx, `
		INSERT INTO projects (id, organization_id, name, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
	`, projectID, uuid.New().String(), "test-project")
	require.NoError(t, err)

	// Create test document
	_, err = db.ExecContext(ctx, `
		INSERT INTO documents (id, project_id, name, original_name, file_path, mime_type, size, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
	`, documentID, projectID, "test.pdf", "test.pdf", "/tmp/test.pdf", "application/pdf", 1024, "processed")
	require.NoError(t, err)

	t.Run("security_rules_retrieval", func(t *testing.T) {
		// Create test security rule
		ruleID := uuid.New().String()
		_, err := db.ExecContext(ctx, `
			INSERT INTO knowledge_items (id, document_id, project_id, type, title, content, status, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		`, ruleID, documentID, projectID, "security_rule", "SEC-001: Test Security Rule", "Test content", "approved")
		require.NoError(t, err)

		rules, err := service.getSecurityRules(ctx, projectID)
		require.NoError(t, err)
		assert.Greater(t, len(rules), 0)
		assert.Contains(t, rules, "SEC-001")
	})

	t.Run("entity_extraction", func(t *testing.T) {
		// Create test entity
		entityID := uuid.New().String()
		_, err := db.ExecContext(ctx, `
			INSERT INTO knowledge_items (id, document_id, project_id, type, title, content, status, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		`, entityID, documentID, projectID, "entity", "Test Entity", "Entity description", "approved")
		require.NoError(t, err)

		entities, err := service.extractEntitiesSimple(ctx, projectID)
		require.NoError(t, err)
		assert.Greater(t, len(entities), 0)
		assert.Equal(t, "entity", entities[0].Type)
	})

	t.Run("user_journey_extraction", func(t *testing.T) {
		// Create test user journey
		journeyID := uuid.New().String()
		_, err := db.ExecContext(ctx, `
			INSERT INTO knowledge_items (id, document_id, project_id, type, title, content, status, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		`, journeyID, documentID, projectID, "user_journey", "Test Journey", "Journey description", "approved")
		require.NoError(t, err)

		journeys, err := service.extractUserJourneysSimple(ctx, projectID)
		require.NoError(t, err)
		assert.Greater(t, len(journeys), 0)
		assert.Equal(t, "user_journey", journeys[0].Type)
	})

	t.Run("sync_knowledge_items", func(t *testing.T) {
		// Create test items to sync
		item1ID := uuid.New().String()
		item2ID := uuid.New().String()

		_, err := db.ExecContext(ctx, `
			INSERT INTO knowledge_items (id, document_id, project_id, type, title, content, status, sync_version, sync_status, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW()),
			       ($10, $11, $12, $13, $14, $15, $16, $17, $18, NOW())
		`, item1ID, documentID, projectID, "business_rule", "Rule 1", "Content 1", "approved", 1, "pending",
			item2ID, documentID, projectID, "business_rule", "Rule 2", "Content 2", "approved", 1, "pending")
		require.NoError(t, err)

		items := []KnowledgeItem{
			{ID: item1ID},
			{ID: item2ID},
		}

		synced, failed, err := service.syncKnowledgeItems(ctx, items, false)
		require.NoError(t, err)
		assert.Equal(t, 2, len(synced))
		assert.Equal(t, 0, len(failed))

		// Verify sync metadata was updated
		var syncVersion int
		var syncStatus string
		err = db.QueryRowContext(ctx, `SELECT sync_version, sync_status FROM knowledge_items WHERE id = $1`, item1ID).Scan(&syncVersion, &syncStatus)
		require.NoError(t, err)
		assert.Equal(t, 2, syncVersion) // Incremented from 1
		assert.Equal(t, "synced", syncStatus)
	})

	t.Run("sync_conflict_detection", func(t *testing.T) {
		// Create item and simulate concurrent update
		itemID := uuid.New().String()
		_, err := db.ExecContext(ctx, `
			INSERT INTO knowledge_items (id, document_id, project_id, type, title, content, status, sync_version, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
		`, itemID, documentID, projectID, "business_rule", "Test Rule", "Content", "approved", 1)
		require.NoError(t, err)

		// Simulate concurrent update by changing version
		_, err = db.ExecContext(ctx, `UPDATE knowledge_items SET sync_version = 5 WHERE id = $1`, itemID)
		require.NoError(t, err)

		// Try to sync - should detect conflict
		err = service.updateSyncMetadata(ctx, itemID, time.Now())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "conflict")
	})
}

// TestSyncKnowledgeEndToEnd tests the complete sync workflow
func TestSyncKnowledgeEndToEnd(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := database.SetupTestDB(t)
	defer database.TeardownTestDB(t, db)
	defer database.CleanupTestData(t, db)

	service := NewKnowledgeService(db).(*KnowledgeServiceImpl)
	ctx := context.Background()

	// Setup test data
	orgID := uuid.New().String()
	projectID := uuid.New().String()
	documentID := uuid.New().String()

	_, err := db.ExecContext(ctx, `
		INSERT INTO organizations (id, name, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
	`, orgID, "test-org")
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		INSERT INTO projects (id, organization_id, name, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
	`, projectID, orgID, "test-project")
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		INSERT INTO documents (id, project_id, name, original_name, file_path, mime_type, size, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
	`, documentID, projectID, "test.pdf", "test.pdf", "/tmp/test.pdf", "application/pdf", 1024, "processed")
	require.NoError(t, err)

	// Create items to sync
	item1ID := uuid.New().String()
	item2ID := uuid.New().String()

	_, err = db.ExecContext(ctx, `
		INSERT INTO knowledge_items (id, document_id, project_id, type, title, content, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW()),
		       ($8, $9, $10, $11, $12, $13, $14, NOW())
	`, item1ID, documentID, projectID, "business_rule", "Rule 1", "Content 1", "approved",
		item2ID, documentID, projectID, "business_rule", "Rule 2", "Content 2", "approved")
	require.NoError(t, err)

	// Test sync all items
	req := SyncKnowledgeRequest{
		ProjectID: projectID,
		Force:     false,
	}

	result, err := service.SyncKnowledge(ctx, req)
	require.NoError(t, err)
	assert.Greater(t, result.SyncedCount, 0)
	assert.Equal(t, 0, result.FailedCount)

	// Test sync specific items
	req2 := SyncKnowledgeRequest{
		ProjectID:        projectID,
		KnowledgeItemIDs: []string{item1ID},
		Force:            false,
	}

	result2, err := service.SyncKnowledge(ctx, req2)
	require.NoError(t, err)
	assert.Equal(t, 1, result2.SyncedCount)
}
