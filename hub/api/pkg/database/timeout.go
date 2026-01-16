// Package database provides database utilities and timeout handling
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package database

import (
	"context"
	"database/sql"
	"time"
)

// DBTimeoutConfig holds timeout configuration for database operations
type DBTimeoutConfig struct {
	QueryTimeout   time.Duration
	ContextTimeout time.Duration
	HTTPTimeout    time.Duration
}

// DefaultTimeoutConfig provides sensible default timeouts
var DefaultTimeoutConfig = DBTimeoutConfig{
	QueryTimeout:   30 * time.Second,
	ContextTimeout: 30 * time.Second,
	HTTPTimeout:    10 * time.Second,
}

// QueryWithTimeout executes a query with timeout using the provided database connection
func QueryWithTimeout(ctx context.Context, db *sql.DB, query string, args ...interface{}) (*sql.Rows, error) {
	timeout := DefaultTimeoutConfig.QueryTimeout
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return db.QueryContext(ctx, query, args...)
}

// QueryRowWithTimeout executes a row query with timeout using the provided database connection
func QueryRowWithTimeout(ctx context.Context, db *sql.DB, query string, args ...interface{}) *sql.Row {
	timeout := DefaultTimeoutConfig.QueryTimeout
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return db.QueryRowContext(ctx, query, args...)
}

// ExecWithTimeout executes a command with timeout using the provided database connection
func ExecWithTimeout(ctx context.Context, db *sql.DB, query string, args ...interface{}) (sql.Result, error) {
	timeout := DefaultTimeoutConfig.QueryTimeout
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return db.ExecContext(ctx, query, args...)
}

// WithCustomTimeout executes a function with a custom timeout context
func WithCustomTimeout(ctx context.Context, timeout time.Duration, fn func(context.Context) error) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return fn(ctx)
}
