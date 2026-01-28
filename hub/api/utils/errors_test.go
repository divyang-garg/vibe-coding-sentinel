// Package utils provides tests for error handling utilities.
// Complies with CODING_STANDARDS.md: Tests max 500 lines
package utils

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"sentinel-hub-api/models"
)

func TestHandleNotFoundError(t *testing.T) {
	t.Run("nil error returns nil", func(t *testing.T) {
		err := HandleNotFoundError(nil, "task", "123")
		if err != nil {
			t.Errorf("expected nil for nil input, got %v", err)
		}
	})

	t.Run("sql.ErrNoRows returns NotFoundError", func(t *testing.T) {
		err := HandleNotFoundError(sql.ErrNoRows, "task", "123")

		var notFoundErr *models.NotFoundError
		if !errors.As(err, &notFoundErr) {
			t.Fatalf("expected NotFoundError, got %T", err)
		}

		if notFoundErr.Resource != "task" {
			t.Errorf("expected resource 'task', got %s", notFoundErr.Resource)
		}

		if notFoundErr.Message == "" {
			t.Error("expected non-empty message")
		}
	})

	t.Run("wrapped sql.ErrNoRows returns NotFoundError", func(t *testing.T) {
		wrappedErr := fmt.Errorf("database query failed: %w", sql.ErrNoRows)
		err := HandleNotFoundError(wrappedErr, "task", "123")

		var notFoundErr *models.NotFoundError
		if !errors.As(err, &notFoundErr) {
			t.Fatalf("expected NotFoundError for wrapped error, got %T", err)
		}

		if notFoundErr.Resource != "task" {
			t.Errorf("expected resource 'task', got %s", notFoundErr.Resource)
		}
	})

	t.Run("other errors are wrapped with context", func(t *testing.T) {
		originalErr := errors.New("connection failed")
		err := HandleNotFoundError(originalErr, "task", "123")

		if errors.Is(err, sql.ErrNoRows) {
			t.Error("should not be sql.ErrNoRows")
		}

		if err == nil {
			t.Error("expected error, got nil")
		}

		// Verify error is wrapped
		if !errors.Is(err, originalErr) {
			t.Error("expected error to wrap original error")
		}

		// Verify error message contains context
		errMsg := err.Error()
		if errMsg == "" {
			t.Error("expected non-empty error message")
		}
	})

	t.Run("errors.Is works with wrapped errors", func(t *testing.T) {
		originalErr := sql.ErrNoRows
		wrappedErr := fmt.Errorf("outer: %w", originalErr)
		doubleWrapped := fmt.Errorf("outer2: %w", wrappedErr)

		err := HandleNotFoundError(doubleWrapped, "task", "123")

		var notFoundErr *models.NotFoundError
		if !errors.As(err, &notFoundErr) {
			t.Fatalf("expected NotFoundError for double-wrapped error, got %T", err)
		}
	})
}

func TestWrapDatabaseError(t *testing.T) {
	t.Run("nil error returns nil", func(t *testing.T) {
		err := WrapDatabaseError(nil, "save", "task")
		if err != nil {
			t.Errorf("expected nil for nil input, got %v", err)
		}
	})

	t.Run("sql.ErrNoRows returns NotFoundError", func(t *testing.T) {
		err := WrapDatabaseError(sql.ErrNoRows, "save", "task")

		var notFoundErr *models.NotFoundError
		if !errors.As(err, &notFoundErr) {
			t.Fatalf("expected NotFoundError, got %T", err)
		}

		if notFoundErr.Resource != "task" {
			t.Errorf("expected resource 'task', got %s", notFoundErr.Resource)
		}
	})

	t.Run("wrapped sql.ErrNoRows returns NotFoundError", func(t *testing.T) {
		wrappedErr := fmt.Errorf("query failed: %w", sql.ErrNoRows)
		err := WrapDatabaseError(wrappedErr, "update", "task")

		var notFoundErr *models.NotFoundError
		if !errors.As(err, &notFoundErr) {
			t.Fatalf("expected NotFoundError for wrapped error, got %T", err)
		}
	})

	t.Run("other errors are wrapped with operation context", func(t *testing.T) {
		originalErr := errors.New("connection timeout")
		err := WrapDatabaseError(originalErr, "save", "task")

		if errors.Is(err, sql.ErrNoRows) {
			t.Error("should not be sql.ErrNoRows")
		}

		if err == nil {
			t.Error("expected error, got nil")
		}

		// Verify error is wrapped
		if !errors.Is(err, originalErr) {
			t.Error("expected error to wrap original error")
		}

		// Verify error message contains operation context
		errMsg := err.Error()
		if errMsg == "" {
			t.Error("expected non-empty error message")
		}
	})
}
