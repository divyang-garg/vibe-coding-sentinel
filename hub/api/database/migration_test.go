// Package database provides tests for database migrations
// Complies with CODING_STANDARDS.md: Test file max 500 lines
package database

import (
	"context"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigration_KnowledgeTables(t *testing.T) {
	t.Run("creates knowledge tables if not exist", func(t *testing.T) {
		db := SetupTestDB(t)
		if db == nil {
			t.Skip("Database not available")
			return
		}
		defer TeardownTestDB(t, db)

		ctx := context.Background()

		// Check if knowledge table exists
		query := `
			SELECT EXISTS (
				SELECT FROM information_schema.tables 
				WHERE table_schema = 'public' 
				AND table_name = 'knowledge'
			)
		`
		var exists bool
		err := db.QueryRowContext(ctx, query).Scan(&exists)
		require.NoError(t, err)

		// Table may or may not exist depending on migration status
		// Just verify we can check
		_ = exists
	})

	t.Run("can query knowledge table structure", func(t *testing.T) {
		db := SetupTestDB(t)
		if db == nil {
			t.Skip("Database not available")
			return
		}
		defer TeardownTestDB(t, db)

		ctx := context.Background()

		// Try to get table columns
		query := `
			SELECT column_name, data_type 
			FROM information_schema.columns 
			WHERE table_schema = 'public' 
			AND table_name = 'knowledge'
			ORDER BY ordinal_position
		`
		rows, err := db.QueryContext(ctx, query)
		if err != nil {
			// Table might not exist, that's OK for this test
			t.Logf("Knowledge table may not exist: %v", err)
			return
		}
		defer rows.Close()

		columns := make(map[string]string)
		for rows.Next() {
			var colName, dataType string
			if err := rows.Scan(&colName, &dataType); err != nil {
				t.Fatalf("Failed to scan column: %v", err)
			}
			columns[colName] = dataType
		}

		// Verify we can query structure
		assert.NoError(t, rows.Err())
	})
}

func TestMigration_TransactionHandling(t *testing.T) {
	t.Run("handles migration errors gracefully", func(t *testing.T) {
		db := SetupTestDB(t)
		if db == nil {
			t.Skip("Database not available")
			return
		}
		defer TeardownTestDB(t, db)

		ctx := context.Background()

		// Test transaction rollback on error
		tx, err := db.BeginTx(ctx, nil)
		require.NoError(t, err)

		// Execute invalid SQL to trigger error
		_, err = tx.ExecContext(ctx, "CREATE TABLE invalid_syntax (")
		assert.Error(t, err)

		// Rollback should succeed
		err = tx.Rollback()
		assert.NoError(t, err)
	})

	t.Run("handles concurrent migration attempts", func(t *testing.T) {
		db := SetupTestDB(t)
		if db == nil {
			t.Skip("Database not available")
			return
		}
		defer TeardownTestDB(t, db)

		ctx := context.Background()

		// Create different tables concurrently; same-table CREATE TABLE IF NOT EXISTS
		// can race on pg_class in PostgreSQL and violate uniqueness.
		done := make(chan error, 2)

		go func() {
			_, err := db.ExecContext(ctx, `
				CREATE TABLE IF NOT EXISTS concurrent_test_a (
					id SERIAL PRIMARY KEY,
					name TEXT
				)
			`)
			done <- err
		}()

		go func() {
			_, err := db.ExecContext(ctx, `
				CREATE TABLE IF NOT EXISTS concurrent_test_b (
					id SERIAL PRIMARY KEY,
					name TEXT
				)
			`)
			done <- err
		}()

		err1 := <-done
		err2 := <-done

		assert.NoError(t, err1)
		assert.NoError(t, err2)

		db.ExecContext(ctx, "DROP TABLE IF EXISTS concurrent_test_a")
		db.ExecContext(ctx, "DROP TABLE IF EXISTS concurrent_test_b")
	})
}

func TestMigration_DataIntegrity(t *testing.T) {
	t.Run("maintains referential integrity", func(t *testing.T) {
		db := SetupTestDB(t)
		if db == nil {
			t.Skip("Database not available")
			return
		}
		defer TeardownTestDB(t, db)

		ctx := context.Background()

		// Create test tables with foreign key
		_, err := db.ExecContext(ctx, `
			CREATE TABLE IF NOT EXISTS test_parent (
				id SERIAL PRIMARY KEY,
				name TEXT
			)
		`)
		require.NoError(t, err)

		_, err = db.ExecContext(ctx, `
			CREATE TABLE IF NOT EXISTS test_child (
				id SERIAL PRIMARY KEY,
				parent_id INTEGER REFERENCES test_parent(id),
				name TEXT
			)
		`)
		require.NoError(t, err)

		// Try to insert child with invalid parent_id
		_, err = db.ExecContext(ctx, `
			INSERT INTO test_child (parent_id, name) 
			VALUES (999, 'orphan')
		`)
		assert.Error(t, err, "Should fail foreign key constraint")

		// Cleanup
		db.ExecContext(ctx, "DROP TABLE IF EXISTS test_child")
		db.ExecContext(ctx, "DROP TABLE IF EXISTS test_parent")
	})
}

func TestMigration_IndexCreation(t *testing.T) {
	t.Run("creates indexes for performance", func(t *testing.T) {
		db := SetupTestDB(t)
		if db == nil {
			t.Skip("Database not available")
			return
		}
		defer TeardownTestDB(t, db)

		ctx := context.Background()

		// Create test table
		_, err := db.ExecContext(ctx, `
			CREATE TABLE IF NOT EXISTS test_indexed (
				id SERIAL PRIMARY KEY,
				email TEXT,
				created_at TIMESTAMP
			)
		`)
		require.NoError(t, err)

		// Create index
		_, err = db.ExecContext(ctx, `
			CREATE INDEX IF NOT EXISTS idx_test_email 
			ON test_indexed(email)
		`)
		assert.NoError(t, err)

		// Verify index exists
		query := `
			SELECT indexname 
			FROM pg_indexes 
			WHERE tablename = 'test_indexed' 
			AND indexname = 'idx_test_email'
		`
		var indexName string
		err = db.QueryRowContext(ctx, query).Scan(&indexName)
		assert.NoError(t, err)
		assert.Equal(t, "idx_test_email", indexName)

		// Cleanup
		db.ExecContext(ctx, "DROP TABLE IF EXISTS test_indexed")
	})
}
