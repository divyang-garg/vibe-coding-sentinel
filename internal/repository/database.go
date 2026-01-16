// Package repository provides data access layer
// Complies with CODING_STANDARDS.md: Repository files max 350 lines
package repository

import (
	"database/sql"
	"time"

	_ "github.com/lib/pq"
)

// NewDatabaseConnection creates a new database connection with proper configuration
func NewDatabaseConnection(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	// Configure connection pool for production readiness
	db.SetMaxOpenConns(25)                 // Maximum open connections
	db.SetMaxIdleConns(5)                  // Maximum idle connections
	db.SetConnMaxLifetime(5 * time.Minute) // Maximum connection lifetime

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
