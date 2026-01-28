// Package database provides test database utilities for integration tests
// Complies with CODING_STANDARDS.md: Test utilities max 250 lines
package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"sentinel-hub-api/repository"

	_ "github.com/lib/pq"
)

// TestDBConfig holds test database configuration
type TestDBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// DefaultTestDBConfig returns default test database configuration
func DefaultTestDBConfig() *TestDBConfig {
	return &TestDBConfig{
		Host:     getEnv("TEST_DB_HOST", "localhost"),
		Port:     getEnv("TEST_DB_PORT", "5432"), // Default to standard PostgreSQL port
		User:     getEnv("TEST_DB_USER", "sentinel"),
		Password: getEnv("TEST_DB_PASSWORD", "sentinel"),
		DBName:   getEnv("TEST_DB_NAME", "sentinel_test"),
		SSLMode:  getEnv("TEST_DB_SSLMODE", "disable"),
	}
}

// ConnectionString returns PostgreSQL connection string
func (c *TestDBConfig) ConnectionString() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
}

// SetupTestDB creates and returns a test database connection
// It will skip the test if database is not available
func SetupTestDB(t *testing.T) *sql.DB {
	config := DefaultTestDBConfig()

	db, err := sql.Open("postgres", config.ConnectionString())
	if err != nil {
		t.Skipf("Skipping integration test: failed to connect to test database: %v", err)
		return nil
	}

	// Set connection pool settings for tests
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Verify connection
	ctx, cancel := testingContext(5 * time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		t.Skipf("Skipping integration test: test database not available: %v", err)
		return nil
	}

	return db
}

// TeardownTestDB closes the test database connection
func TeardownTestDB(t *testing.T, db *sql.DB) {
	if db != nil {
		if err := db.Close(); err != nil {
			t.Logf("Warning: failed to close test database: %v", err)
		}
	}
}

// CleanupTestData removes test data from tables
func CleanupTestData(t *testing.T, db *sql.DB) {
	if db == nil {
		return
	}

	tables := []string{
		"tasks",
		"documents",
		"organizations",
		"projects",
		"api_keys",
		"users",
		"task_dependencies",
		"llm_usage",
		"knowledge_items", // Added for knowledge service tests
	}

	for _, table := range tables {
		query := fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)
		if _, err := db.Exec(query); err != nil {
			// Table might not exist, which is okay for tests
			t.Logf("Note: could not truncate table %s: %v", table, err)
		}
	}
}

// WaitForDB waits for database to be ready (for Docker startup)
func WaitForDB(config *TestDBConfig, maxWait time.Duration) error {
	deadline := time.Now().Add(maxWait)

	for time.Now().Before(deadline) {
		db, err := sql.Open("postgres", config.ConnectionString())
		if err == nil {
			ctx, cancel := testingContext(2 * time.Second)
			err = db.PingContext(ctx)
			cancel()
			db.Close()

			if err == nil {
				return nil
			}
		}
		time.Sleep(500 * time.Millisecond)
	}

	return fmt.Errorf("database not ready after %v", maxWait)
}

// Helper functions

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func testingContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	// Use a simple context for testing
	// In real code, this would use context.Background()
	ctx := context.Background()
	return context.WithTimeout(ctx, timeout)
}

// SQLDBAdapter adapts sql.DB to repository.Database interface
type SQLDBAdapter struct {
	db *sql.DB
}

// NewSQLDBAdapter creates a new adapter from sql.DB
func NewSQLDBAdapter(db *sql.DB) *SQLDBAdapter {
	return &SQLDBAdapter{db: db}
}

// QueryRow implements repository.Database interface
func (a *SQLDBAdapter) QueryRow(ctx context.Context, query string, args ...interface{}) repository.Row {
	return a.db.QueryRowContext(ctx, query, args...)
}

// Query implements repository.Database interface
func (a *SQLDBAdapter) Query(ctx context.Context, query string, args ...interface{}) (repository.Rows, error) {
	rows, err := a.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &rowsAdapter{rows: rows}, nil
}

// Exec implements repository.Database interface
func (a *SQLDBAdapter) Exec(ctx context.Context, query string, args ...interface{}) (repository.Result, error) {
	return a.db.ExecContext(ctx, query, args...)
}

// BeginTx implements repository.Database interface
func (a *SQLDBAdapter) BeginTx(ctx context.Context) (repository.Transaction, error) {
	tx, err := a.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &txAdapter{tx: tx}, nil
}

// rowsAdapter adapts sql.Rows to repository.Rows
type rowsAdapter struct {
	rows *sql.Rows
}

func (r *rowsAdapter) Next() bool {
	return r.rows.Next()
}

func (r *rowsAdapter) Scan(dest ...interface{}) error {
	return r.rows.Scan(dest...)
}

func (r *rowsAdapter) Close() error {
	return r.rows.Close()
}

func (r *rowsAdapter) Err() error {
	return r.rows.Err()
}

// txAdapter adapts sql.Tx to repository.Transaction
type txAdapter struct {
	tx *sql.Tx
}

func (t *txAdapter) QueryRow(ctx context.Context, query string, args ...interface{}) repository.Row {
	return t.tx.QueryRowContext(ctx, query, args...)
}

func (t *txAdapter) Query(ctx context.Context, query string, args ...interface{}) (repository.Rows, error) {
	rows, err := t.tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &rowsAdapter{rows: rows}, nil
}

func (t *txAdapter) Exec(ctx context.Context, query string, args ...interface{}) (repository.Result, error) {
	return t.tx.ExecContext(ctx, query, args...)
}

func (t *txAdapter) BeginTx(ctx context.Context) (repository.Transaction, error) {
	// Nested transactions not supported, return self
	return t, nil
}

func (t *txAdapter) Commit() error {
	return t.tx.Commit()
}

func (t *txAdapter) Rollback() error {
	return t.tx.Rollback()
}
