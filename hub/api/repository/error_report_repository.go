// Package repository contains error report repository implementations.
// Complies with CODING_STANDARDS.md: Repositories max 350 lines
package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sentinel-hub-api/models"
	"time"
)

// ErrorReportRepositoryImpl implements error report data access
type ErrorReportRepositoryImpl struct {
	db Database
}

// NewErrorReportRepository creates a new error report repository instance
func NewErrorReportRepository(db Database) *ErrorReportRepositoryImpl {
	return &ErrorReportRepositoryImpl{db: db}
}

// Save saves an error report to the database
func (r *ErrorReportRepositoryImpl) Save(ctx context.Context, report *models.ErrorReport) error {
	var contextJSON sql.NullString
	if report.Context != nil {
		contextBytes, err := json.Marshal(report.Context)
		if err == nil {
			contextJSON = sql.NullString{String: string(contextBytes), Valid: true}
		}
	}

	query := `
		INSERT INTO error_reports (id, message, stack_trace, category, severity, context, resolved, resolved_at, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7, $8, $9)
		ON CONFLICT (id) DO UPDATE SET
			resolved = EXCLUDED.resolved,
			resolved_at = CASE WHEN EXCLUDED.resolved = true THEN CURRENT_TIMESTAMP ELSE NULL END
	`

	if report.Timestamp.IsZero() {
		report.Timestamp = time.Now()
	}

	var resolvedAt *time.Time
	if report.Resolved {
		now := time.Now()
		resolvedAt = &now
	}

	// Convert severity to string
	severityStr := "low"
	switch report.Severity {
	case models.ErrorSeverityInfo:
		severityStr = "info"
	case models.ErrorSeverityLow:
		severityStr = "low"
	case models.ErrorSeverityMedium:
		severityStr = "medium"
	case models.ErrorSeverityHigh:
		severityStr = "high"
	case models.ErrorSeverityCritical:
		severityStr = "critical"
	}

	_, err := r.db.Exec(ctx, query,
		report.ID, report.Message, report.StackTrace, report.Category,
		severityStr, contextJSON, report.Resolved, resolvedAt,
		report.Timestamp,
	)

	if err != nil {
		return fmt.Errorf("failed to save error report %s: %w", report.ID, err)
	}
	return nil
}

// FindByID retrieves an error report by ID
func (r *ErrorReportRepositoryImpl) FindByID(ctx context.Context, id string) (*models.ErrorReport, error) {
	query := `
		SELECT id, message, stack_trace, category, severity, context, resolved, resolved_at, timestamp
		FROM error_reports WHERE id = $1
	`

	var report models.ErrorReport
	var severityStr string
	var contextJSON sql.NullString
	var resolvedAt sql.NullTime

	rowErr := r.db.QueryRow(ctx, query, id).Scan(
		&report.ID, &report.Message, &report.StackTrace, &report.Category,
		&severityStr, &contextJSON, &report.Resolved, &resolvedAt,
		&report.Timestamp,
	)

	if rowErr != nil {
		return nil, rowErr
	}

	// Convert severity string to enum
	report.Severity = models.ErrorSeverity(0) // Default
	switch severityStr {
	case "info":
		report.Severity = models.ErrorSeverityInfo
	case "low":
		report.Severity = models.ErrorSeverityLow
	case "medium":
		report.Severity = models.ErrorSeverityMedium
	case "high":
		report.Severity = models.ErrorSeverityHigh
	case "critical":
		report.Severity = models.ErrorSeverityCritical
	}

	// Unmarshal context if present
	if contextJSON.Valid {
		if err := json.Unmarshal([]byte(contextJSON.String), &report.Context); err != nil {
			return nil, fmt.Errorf("failed to unmarshal context: %w", err)
		}
	}

	return &report, nil
}

// List retrieves error reports with filters
func (r *ErrorReportRepositoryImpl) List(ctx context.Context, category string, severity string, resolved *bool, limit, offset int) ([]models.ErrorReport, int, error) {
	// Build WHERE clause
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argCount := 0

	if category != "" {
		argCount++
		whereClause += fmt.Sprintf(" AND category = $%d", argCount)
		args = append(args, category)
	}

	if severity != "" {
		argCount++
		whereClause += fmt.Sprintf(" AND severity = $%d", argCount)
		args = append(args, severity)
	}

	if resolved != nil {
		argCount++
		whereClause += fmt.Sprintf(" AND resolved = $%d", argCount)
		args = append(args, *resolved)
	}

	// Count total
	countQuery := "SELECT COUNT(*) FROM error_reports " + whereClause
	var total int
	countArgs := args
	if err := r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count error reports: %w", err)
	}

	// Get paginated results
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	argCount++
	query := `
		SELECT id, message, stack_trace, category, severity, context, resolved, resolved_at, timestamp
		FROM error_reports ` + whereClause + `
		ORDER BY timestamp DESC
		LIMIT $` + fmt.Sprintf("%d", argCount) + ` OFFSET $` + fmt.Sprintf("%d", argCount+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query error reports: %w", err)
	}
	defer rows.Close()

	var reports []models.ErrorReport
	for rows.Next() {
		var report models.ErrorReport
		var severityStr string
		var contextJSON sql.NullString

		var resolvedAt sql.NullTime
		err := rows.Scan(
			&report.ID, &report.Message, &report.StackTrace, &report.Category,
			&severityStr, &contextJSON, &report.Resolved, &resolvedAt,
			&report.Timestamp,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan error report row: %w", err)
		}

		// Convert severity
		switch severityStr {
		case "info":
			report.Severity = models.ErrorSeverityInfo
		case "low":
			report.Severity = models.ErrorSeverityLow
		case "medium":
			report.Severity = models.ErrorSeverityMedium
		case "high":
			report.Severity = models.ErrorSeverityHigh
		case "critical":
			report.Severity = models.ErrorSeverityCritical
		}

		// Unmarshal context if present
		if contextJSON.Valid {
			if err := json.Unmarshal([]byte(contextJSON.String), &report.Context); err != nil {
				return nil, 0, fmt.Errorf("failed to unmarshal context: %w", err)
			}
		}

		reports = append(reports, report)
	}

	return reports, total, nil
}

// UpdateResolved updates the resolved status of an error report
func (r *ErrorReportRepositoryImpl) UpdateResolved(ctx context.Context, id string, resolved bool) error {
	var resolvedAt *time.Time
	if resolved {
		now := time.Now()
		resolvedAt = &now
	}

	query := `
		UPDATE error_reports
		SET resolved = $1, resolved_at = $2
		WHERE id = $3
	`

	_, err := r.db.Exec(ctx, query, resolved, resolvedAt, id)
	if err != nil {
		return fmt.Errorf("failed to update error report %s resolved status: %w", id, err)
	}
	return nil
}
