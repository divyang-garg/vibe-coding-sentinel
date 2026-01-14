// Package repository provides PostgreSQL implementations of repository interfaces.
package repository

import (
	"context"
	"database/sql"
)

// PostgresDatabase implements the Database interface using PostgreSQL
type PostgresDatabase struct {
	db *sql.DB
}

// NewPostgresDatabase creates a new PostgreSQL database instance
func NewPostgresDatabase(db *sql.DB) Database {
	return &PostgresDatabase{db: db}
}

// QueryRow executes a query that returns at most one row
func (p *PostgresDatabase) QueryRow(ctx context.Context, query string, args ...interface{}) Row {
	return p.db.QueryRowContext(ctx, query, args...)
}

// Query executes a query that returns rows
func (p *PostgresDatabase) Query(ctx context.Context, query string, args ...interface{}) (Rows, error) {
	return p.db.QueryContext(ctx, query, args...)
}

// Exec executes a query without returning rows
func (p *PostgresDatabase) Exec(ctx context.Context, query string, args ...interface{}) (Result, error) {
	return p.db.ExecContext(ctx, query, args...)
}

// BeginTx begins a transaction
func (p *PostgresDatabase) BeginTx(ctx context.Context) (Transaction, error) {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &PostgresTransaction{tx: tx}, nil
}

// PostgresTransaction implements the Transaction interface
type PostgresTransaction struct {
	tx *sql.Tx
}

// QueryRow executes a query that returns at most one row within a transaction
func (t *PostgresTransaction) QueryRow(ctx context.Context, query string, args ...interface{}) Row {
	return t.tx.QueryRowContext(ctx, query, args...)
}

// Query executes a query that returns rows within a transaction
func (t *PostgresTransaction) Query(ctx context.Context, query string, args ...interface{}) (Rows, error) {
	return t.tx.QueryContext(ctx, query, args...)
}

// Exec executes a query without returning rows within a transaction
func (t *PostgresTransaction) Exec(ctx context.Context, query string, args ...interface{}) (Result, error) {
	return t.tx.ExecContext(ctx, query, args...)
}

// BeginTx begins a nested transaction (not supported in PostgreSQL, returns error)
func (t *PostgresTransaction) BeginTx(ctx context.Context) (Transaction, error) {
	return nil, sql.ErrTxDone // Nested transactions not supported
}

// Commit commits the transaction
func (t *PostgresTransaction) Commit() error {
	return t.tx.Commit()
}

// Rollback rolls back the transaction
func (t *PostgresTransaction) Rollback() error {
	return t.tx.Rollback()
}
