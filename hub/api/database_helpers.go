// Database helpers for main package
// Provides database utility functions
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package main

import (
	"context"
	"database/sql"
	"time"
)

// db is the package-level database connection
var db *sql.DB

// SetDB sets the database connection
func SetDB(database *sql.DB) {
	db = database
}

// getQueryTimeout returns the default query timeout
func getQueryTimeout() time.Duration {
	return 30 * time.Second
}

// queryRowWithTimeout executes a query and returns a row with timeout
func queryRowWithTimeout(ctx context.Context, query string, args ...interface{}) *sql.Row {
	timeoutCtx, cancel := context.WithTimeout(ctx, getQueryTimeout())
	defer cancel()
	return db.QueryRowContext(timeoutCtx, query, args...)
}

// execWithTimeout executes a query with timeout
func execWithTimeout(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, getQueryTimeout())
	defer cancel()
	return db.ExecContext(timeoutCtx, query, args...)
}

// queryWithTimeout executes a query and returns rows with timeout
func queryWithTimeout(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, getQueryTimeout())
	defer cancel()
	return db.QueryContext(timeoutCtx, query, args...)
}
