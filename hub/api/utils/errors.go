// Package utils provides error handling utilities for database operations.
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package utils

import (
	"database/sql"
	"errors"
	"fmt"

	"sentinel-hub-api/models"
)

// HandleNotFoundError checks if an error is sql.ErrNoRows and returns
// a structured NotFoundError if so, otherwise wraps the error with context.
//
// This function ensures consistent error handling across all database operations
// and properly handles wrapped errors using errors.Is() as required by
// CODING_STANDARDS.md Section 4.1.
//
// Example usage:
//   err := row.Scan(...)
//   if err != nil {
//       return nil, HandleNotFoundError(err, "task", id)
//   }
func HandleNotFoundError(err error, resource, id string) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, sql.ErrNoRows) {
		return &models.NotFoundError{
			Resource: resource,
			Message:  fmt.Sprintf("%s not found: %s", resource, id),
		}
	}

	return fmt.Errorf("failed to find %s %s: %w", resource, id, err)
}

// WrapDatabaseError wraps a database error with context about the operation.
// Complies with CODING_STANDARDS.md Section 4.1: Error Wrapping
//
// If the error is sql.ErrNoRows, it returns a NotFoundError.
// Otherwise, it wraps the error with operation and resource context.
//
// Example usage:
//   _, err := db.Exec(ctx, query, ...)
//   if err != nil {
//       return WrapDatabaseError(err, "save", "task")
//   }
func WrapDatabaseError(err error, operation, resource string) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, sql.ErrNoRows) {
		return &models.NotFoundError{
			Resource: resource,
			Message:  fmt.Sprintf("%s not found during %s", resource, operation),
		}
	}

	return fmt.Errorf("database error during %s on %s: %w", operation, resource, err)
}
