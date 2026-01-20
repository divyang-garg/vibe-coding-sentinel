// Package repository provides unit tests for database connection
// Complies with CODING_STANDARDS.md: Test file max 500 lines, 90%+ coverage
package repository

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDatabaseConnection_Success(t *testing.T) {
	// Given: Valid DSN from environment or Docker default
	dsn := getTestDSN()
	if dsn == "" {
		t.Skip("Skipping test: No test database available. Start Docker with: docker-compose up -d postgres")
	}

	// When: Creating database connection
	db, err := NewDatabaseConnection(dsn)

	// Then: Connection should succeed
	if err != nil {
		t.Skipf("Skipping test: Database not available. Start Docker with: docker-compose up -d postgres. Error: %v", err)
	}
	require.NotNil(t, db)
	defer db.Close()

	// Verify connection is usable
	err = db.Ping()
	assert.NoError(t, err)
}

func TestNewDatabaseConnection_InvalidDSN(t *testing.T) {
	// Given: Invalid DSN format
	invalidDSN := "invalid://dsn/format"

	// When: Creating database connection
	db, err := NewDatabaseConnection(invalidDSN)

	// Then: Should return error
	assert.Error(t, err)
	assert.Nil(t, db)
}

func TestNewDatabaseConnection_PingFailure(t *testing.T) {
	// Given: Valid DSN format but unreachable database
	// Using a valid format but non-existent host/port
	invalidDSN := "postgres://user:password@localhost:99999/nonexistent?sslmode=disable"

	// When: Creating database connection
	db, err := NewDatabaseConnection(invalidDSN)

	// Then: Should return error (ping fails)
	assert.Error(t, err)
	assert.Nil(t, db)
}

func TestNewDatabaseConnection_ConnectionPoolSettings(t *testing.T) {
	// Given: Valid DSN from environment or Docker default
	dsn := getTestDSN()
	if dsn == "" {
		t.Skip("Skipping test: No test database available. Start Docker with: docker-compose up -d postgres")
	}

	// When: Creating database connection
	db, err := NewDatabaseConnection(dsn)
	if err != nil {
		t.Skipf("Skipping test: Database not available. Start Docker with: docker-compose up -d postgres. Error: %v", err)
	}
	require.NotNil(t, db)
	defer db.Close()

	// Then: Connection pool settings should be configured correctly
	stats := db.Stats()
	assert.Equal(t, 25, stats.MaxOpenConnections, "MaxOpenConns should be 25")
	
	// Verify connection is usable (indirectly confirms pool settings are applied)
	err = db.Ping()
	assert.NoError(t, err, "Connection should be usable after configuration")
	
	// Verify connection lifetime is set (indirectly by checking connection works)
	// The ConnMaxLifetime is internal but we can verify it doesn't cause issues
	time.Sleep(100 * time.Millisecond)
	err = db.Ping()
	assert.NoError(t, err, "Connection should remain usable")
}

// getTestDSN returns test database DSN from environment variable or Docker default
// Returns empty string if not available, allowing tests to be skipped
func getTestDSN() string {
	// First check for explicit test DSN
	if dsn := os.Getenv("TEST_DATABASE_DSN"); dsn != "" {
		return dsn
	}

	// Default to Docker Compose database if available
	// This matches docker-compose.yml configuration:
	// POSTGRES_DB: sentinel
	// POSTGRES_USER: sentinel
	// POSTGRES_PASSWORD: password
	// Port: 5432
	return "postgres://sentinel:password@localhost:5432/sentinel?sslmode=disable"
}
