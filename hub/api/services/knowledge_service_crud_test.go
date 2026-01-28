// Package services - Knowledge Service CRUD Tests
// Complies with CODING_STANDARDS.md: Test coverage requirements (80%+)
package services

import (
	"context"
	"testing"

	"sentinel-hub-api/database"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestListKnowledgeItems tests listing knowledge items
func TestListKnowledgeItems(t *testing.T) {
	t.Run("invalid_project_id", func(t *testing.T) {
		service := &KnowledgeServiceImpl{db: nil}
		ctx := context.Background()
		req := ListKnowledgeItemsRequest{ProjectID: ""}

		items, err := service.ListKnowledgeItems(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, items)
		assert.Contains(t, err.Error(), "project_id is required")
	})

	t.Run("with_existing_items", func(t *testing.T) {
		if testing.Short() {
			t.Skip("Skipping database test in short mode")
		}

		db := database.SetupTestDB(t)
		defer database.TeardownTestDB(t, db)
		defer database.CleanupTestData(t, db)

		setupKnowledgeItemsTable(t, db)

		service := NewKnowledgeService(db).(*KnowledgeServiceImpl)
		ctx := context.Background()

		// Setup: Create test organization, project and document
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

		// Create test items
		item1ID := uuid.New().String()
		item2ID := uuid.New().String()
		_, err = db.ExecContext(ctx, `
			INSERT INTO knowledge_items (id, document_id, project_id, type, title, content, status, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, NOW()),
			       ($8, $9, $10, $11, $12, $13, $14, NOW())
		`, item1ID, documentID, projectID, "business_rule", "Rule 1", "Content 1", "approved",
			item2ID, documentID, projectID, "entity", "Entity 1", "Content 2", "approved")
		require.NoError(t, err)

		req := ListKnowledgeItemsRequest{
			ProjectID: projectID,
		}

		items, err := service.ListKnowledgeItems(ctx, req)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(items), 2)
	})

	t.Run("with_type_filter", func(t *testing.T) {
		if testing.Short() {
			t.Skip("Skipping database test in short mode")
		}

		db := database.SetupTestDB(t)
		defer database.TeardownTestDB(t, db)
		defer database.CleanupTestData(t, db)

		setupKnowledgeItemsTable(t, db)

		service := NewKnowledgeService(db).(*KnowledgeServiceImpl)
		ctx := context.Background()

		// Setup: Create test organization, project and document
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

		// Create test items
		item1ID := uuid.New().String()
		item2ID := uuid.New().String()
		_, err = db.ExecContext(ctx, `
			INSERT INTO knowledge_items (id, document_id, project_id, type, title, content, status, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, NOW()),
			       ($8, $9, $10, $11, $12, $13, $14, NOW())
		`, item1ID, documentID, projectID, "business_rule", "Rule 1", "Content 1", "approved",
			item2ID, documentID, projectID, "entity", "Entity 1", "Content 2", "approved")
		require.NoError(t, err)

		req := ListKnowledgeItemsRequest{
			ProjectID: projectID,
			Type:      "business_rule",
		}

		items, err := service.ListKnowledgeItems(ctx, req)
		require.NoError(t, err)
		assert.Equal(t, 1, len(items))
		assert.Equal(t, "business_rule", items[0].Type)
	})
}

// TestCreateKnowledgeItem tests creating knowledge items
func TestCreateKnowledgeItem(t *testing.T) {
	t.Run("invalid_document_id", func(t *testing.T) {
		if testing.Short() {
			t.Skip("Skipping database test in short mode")
		}

		db := database.SetupTestDB(t)
		defer database.TeardownTestDB(t, db)
		defer database.CleanupTestData(t, db)

		setupKnowledgeItemsTable(t, db)

		service := NewKnowledgeService(db).(*KnowledgeServiceImpl)
		ctx := context.Background()

		item := KnowledgeItem{
			ID:         uuid.New().String(),
			DocumentID: "non-existent",
			Type:       "business_rule",
			Title:      "Test",
			Content:    "Content",
		}

		created, err := service.CreateKnowledgeItem(ctx, item)
		assert.Error(t, err)
		assert.Nil(t, created)
	})

	t.Run("successful_creation", func(t *testing.T) {
		if testing.Short() {
			t.Skip("Skipping database test in short mode")
		}

		db := database.SetupTestDB(t)
		defer database.TeardownTestDB(t, db)
		defer database.CleanupTestData(t, db)

		setupKnowledgeItemsTable(t, db)

		service := NewKnowledgeService(db).(*KnowledgeServiceImpl)
		ctx := context.Background()

		// Setup: Create test organization, project and document
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

		item := KnowledgeItem{
			ID:         uuid.New().String(),
			DocumentID: documentID,
			Type:       "business_rule",
			Title:      "Test Rule",
			Content:    "Test Content",
			Status:     "pending",
		}

		created, err := service.CreateKnowledgeItem(ctx, item)
		require.NoError(t, err)
		assert.NotNil(t, created)
		assert.Equal(t, item.ID, created.ID)
		assert.Equal(t, item.Title, created.Title)
	})
}

// TestGetKnowledgeItem tests retrieving a knowledge item
func TestGetKnowledgeItem(t *testing.T) {
	t.Run("not_found", func(t *testing.T) {
		if testing.Short() {
			t.Skip("Skipping database test in short mode")
		}

		db := database.SetupTestDB(t)
		defer database.TeardownTestDB(t, db)
		defer database.CleanupTestData(t, db)

		setupKnowledgeItemsTable(t, db)

		service := NewKnowledgeService(db).(*KnowledgeServiceImpl)
		ctx := context.Background()

		item, err := service.GetKnowledgeItem(ctx, "non-existent")
		assert.Error(t, err)
		assert.Nil(t, item)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("successful_retrieval", func(t *testing.T) {
		if testing.Short() {
			t.Skip("Skipping database test in short mode")
		}

		db := database.SetupTestDB(t)
		defer database.TeardownTestDB(t, db)
		defer database.CleanupTestData(t, db)

		setupKnowledgeItemsTable(t, db)

		service := NewKnowledgeService(db).(*KnowledgeServiceImpl)
		ctx := context.Background()

		// Setup: Create test organization, project and document
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

		// Create test item
		itemID := uuid.New().String()
		_, err = db.ExecContext(ctx, `
			INSERT INTO knowledge_items (id, document_id, project_id, type, title, content, status, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		`, itemID, documentID, projectID, "business_rule", "Test Rule", "Test Content", "approved")
		require.NoError(t, err)

		item, err := service.GetKnowledgeItem(ctx, itemID)
		require.NoError(t, err)
		assert.NotNil(t, item)
		assert.Equal(t, itemID, item.ID)
		assert.Equal(t, "Test Rule", item.Title)
	})
}

// TestUpdateKnowledgeItem tests updating knowledge items
func TestUpdateKnowledgeItem(t *testing.T) {
	t.Run("not_found", func(t *testing.T) {
		if testing.Short() {
			t.Skip("Skipping database test in short mode")
		}

		db := database.SetupTestDB(t)
		defer database.TeardownTestDB(t, db)
		defer database.CleanupTestData(t, db)

		setupKnowledgeItemsTable(t, db)

		service := NewKnowledgeService(db).(*KnowledgeServiceImpl)
		ctx := context.Background()

		item := KnowledgeItem{
			ID:      "non-existent",
			Title:   "Updated",
			Content: "Updated Content",
		}

		updated, err := service.UpdateKnowledgeItem(ctx, "non-existent", item)
		assert.Error(t, err)
		assert.Nil(t, updated)
	})

	t.Run("successful_update", func(t *testing.T) {
		if testing.Short() {
			t.Skip("Skipping database test in short mode")
		}

		db := database.SetupTestDB(t)
		defer database.TeardownTestDB(t, db)
		defer database.CleanupTestData(t, db)

		setupKnowledgeItemsTable(t, db)

		service := NewKnowledgeService(db).(*KnowledgeServiceImpl)
		ctx := context.Background()

		// Setup: Create test organization, project and document
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

		// Create test item
		itemID := uuid.New().String()
		_, err = db.ExecContext(ctx, `
			INSERT INTO knowledge_items (id, document_id, project_id, type, title, content, status, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		`, itemID, documentID, projectID, "business_rule", "Original", "Original Content", "approved")
		require.NoError(t, err)

		item := KnowledgeItem{
			Title:   "Updated",
			Content: "Updated Content",
		}

		updated, err := service.UpdateKnowledgeItem(ctx, itemID, item)
		require.NoError(t, err)
		assert.NotNil(t, updated)
		assert.Equal(t, "Updated", updated.Title)
		assert.Equal(t, "Updated Content", updated.Content)
	})
}

// TestDeleteKnowledgeItem tests deleting knowledge items
func TestDeleteKnowledgeItem(t *testing.T) {
	t.Run("not_found", func(t *testing.T) {
		if testing.Short() {
			t.Skip("Skipping database test in short mode")
		}

		db := database.SetupTestDB(t)
		defer database.TeardownTestDB(t, db)
		defer database.CleanupTestData(t, db)

		setupKnowledgeItemsTable(t, db)

		service := NewKnowledgeService(db).(*KnowledgeServiceImpl)
		ctx := context.Background()

		err := service.DeleteKnowledgeItem(ctx, "non-existent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("successful_deletion", func(t *testing.T) {
		if testing.Short() {
			t.Skip("Skipping database test in short mode")
		}

		db := database.SetupTestDB(t)
		defer database.TeardownTestDB(t, db)
		defer database.CleanupTestData(t, db)

		setupKnowledgeItemsTable(t, db)

		service := NewKnowledgeService(db).(*KnowledgeServiceImpl)
		ctx := context.Background()

		// Setup: Create test organization, project and document
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

		// Create test item
		itemID := uuid.New().String()
		_, err = db.ExecContext(ctx, `
			INSERT INTO knowledge_items (id, document_id, project_id, type, title, content, status, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		`, itemID, documentID, projectID, "business_rule", "Test Rule", "Test Content", "approved")
		require.NoError(t, err)

		err = service.DeleteKnowledgeItem(ctx, itemID)
		require.NoError(t, err)

		// Verify deletion
		_, err = service.GetKnowledgeItem(ctx, itemID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}
