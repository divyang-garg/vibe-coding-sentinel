// Phase 12: Implementation Tracking Module
// Tracks implementation status of approved change requests

package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// Implementation status constants
const (
	ImplStatusPending    = "pending"
	ImplStatusInProgress = "in_progress"
	ImplStatusCompleted  = "completed"
	ImplStatusBlocked    = "blocked"
)

// ImplementationStatus represents the current implementation status
type ImplementationStatus struct {
	Status string `json:"status"`
	Notes  string `json:"notes"`
}

// updateImplementationStatus updates the implementation status of a change request
func updateImplementationStatus(ctx context.Context, changeRequestID string, status string, notes string) error {
	// Validate status
	validStatuses := map[string]bool{
		ImplStatusPending:    true,
		ImplStatusInProgress: true,
		ImplStatusCompleted:  true,
		ImplStatusBlocked:    true,
	}
	if !validStatuses[status] {
		return fmt.Errorf("invalid implementation status: %s", status)
	}

	// Validate status transition
	currentStatus, err := getCurrentImplementationStatus(ctx, changeRequestID)
	if err != nil {
		return fmt.Errorf("failed to get current status: %w", err)
	}

	if !isValidStatusTransition(currentStatus, status) {
		return fmt.Errorf("invalid status transition from %s to %s", currentStatus, status)
	}

	// Update database
	query := `
		UPDATE change_requests 
		SET implementation_status = $1, implementation_notes = $2
		WHERE id = $3
	`

	ctx, cancel := context.WithTimeout(ctx, getQueryTimeout())
	defer cancel()

	_, err = db.ExecContext(ctx, query, status, notes, changeRequestID)
	if err != nil {
		return fmt.Errorf("failed to update implementation status: %w", err)
	}

	LogInfo(ctx, "Updated implementation status for change request %s to %s", changeRequestID, status)
	return nil
}

// getImplementationStatus retrieves the current implementation status
func getImplementationStatus(ctx context.Context, changeRequestID string) (*ImplementationStatus, error) {
	query := `
		SELECT implementation_status, implementation_notes
		FROM change_requests
		WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(ctx, getQueryTimeout())
	defer cancel()

	var status ImplementationStatus
	var notes sql.NullString

	err := queryRowWithTimeout(ctx, query, changeRequestID).Scan(&status.Status, &notes)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("change request not found: %s", changeRequestID)
		}
		return nil, fmt.Errorf("failed to get implementation status: %w", err)
	}

	if notes.Valid {
		status.Notes = notes.String
	}

	return &status, nil
}

// checkImplementationStatus automatically checks if change request is implemented
func checkImplementationStatus(ctx context.Context, changeRequestID string, projectID string, codebasePath string) (string, error) {
	// Validate inputs
	if changeRequestID == "" {
		return "", fmt.Errorf("change request ID cannot be empty")
	}
	if projectID == "" {
		return "", fmt.Errorf("project ID cannot be empty")
	}
	if codebasePath == "" {
		return "", fmt.Errorf("codebase path cannot be empty")
	}

	// Load change request
	cr, err := getChangeRequest(ctx, changeRequestID)
	if err != nil {
		return "", fmt.Errorf("failed to load change request %s: %w", changeRequestID, err)
	}
	if cr == nil {
		return "", fmt.Errorf("change request %s not found", changeRequestID)
	}

	// Only check if status is in_progress
	if cr.Status != "approved" {
		return "", fmt.Errorf("change request is not approved")
	}

	// Extract business rule from change request
	var rule KnowledgeItem
	if cr.ProposedState != nil {
		if title, ok := cr.ProposedState["title"].(string); ok {
			rule.Title = title
		}
		if content, ok := cr.ProposedState["content"].(string); ok {
			rule.Content = content
		}
	}

	// Use Phase 11 to check implementation
	evidence := detectBusinessRuleImplementation(rule, codebasePath)

	// If confidence > 0.7, mark as completed
	if evidence.Confidence > 0.7 {
		err := updateImplementationStatus(ctx, changeRequestID, ImplStatusCompleted,
			fmt.Sprintf("Auto-detected as implemented (confidence: %.2f%%)", evidence.Confidence*100))
		if err != nil {
			return "", err
		}
		return ImplStatusCompleted, nil
	}

	// Return current status
	currentStatus, err := getCurrentImplementationStatus(ctx, changeRequestID)
	if err != nil {
		return "", err
	}
	return currentStatus, nil
}

// Helper functions

func getCurrentImplementationStatus(ctx context.Context, changeRequestID string) (string, error) {
	query := `SELECT implementation_status FROM change_requests WHERE id = $1`

	ctx, cancel := context.WithTimeout(ctx, getQueryTimeout())
	defer cancel()

	var status string
	err := queryRowWithTimeout(ctx, query, changeRequestID).Scan(&status)
	if err != nil {
		return "", err
	}

	if status == "" {
		return ImplStatusPending, nil // Default
	}

	return status, nil
}

func isValidStatusTransition(from, to string) bool {
	// Define valid transitions
	transitions := map[string][]string{
		ImplStatusPending:    {ImplStatusInProgress, ImplStatusBlocked},
		ImplStatusInProgress: {ImplStatusCompleted, ImplStatusBlocked},
		ImplStatusBlocked:    {ImplStatusInProgress},
		ImplStatusCompleted:  {}, // Terminal state
	}

	allowed, exists := transitions[from]
	if !exists {
		return false
	}

	for _, allowedStatus := range allowed {
		if allowedStatus == to {
			return true
		}
	}

	return false
}
