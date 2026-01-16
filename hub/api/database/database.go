// Package database provides database connection management
// Complies with CODING_STANDARDS.md: Database files max 200 lines
package database

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

// Init initializes the database connection
func Init(databaseURL string) *sql.DB {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)

	log.Println("✅ Database connection established")
	return db
}
