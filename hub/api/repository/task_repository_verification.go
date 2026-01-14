// Package repository - Task verification data access operations
// Complies with CODING_STANDARDS.md: Repositories max 350 lines
package repository

import (
	"context"
	"sentinel-hub-api/models"
)

// SaveVerification saves a task verification to the database
func (r *TaskRepositoryImpl) SaveVerification(ctx context.Context, verification *models.TaskVerification) error {
	query := `
		INSERT INTO task_verifications (id, task_id, verification_type, status, confidence, evidence, retry_count, verified_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := r.db.Exec(ctx, query,
		verification.ID, verification.TaskID, verification.VerificationType, verification.Status,
		verification.Confidence, verification.Evidence, verification.RetryCount, verification.VerifiedAt, verification.CreatedAt)
	return err
}

// FindVerifications retrieves verifications for a task
func (r *TaskRepositoryImpl) FindVerifications(ctx context.Context, taskID string) ([]models.TaskVerification, error) {
	query := "SELECT id, task_id, verification_type, status, confidence, evidence, retry_count, verified_at, created_at FROM task_verifications WHERE task_id = $1"
	rows, err := r.db.Query(ctx, query, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var verifications []models.TaskVerification
	for rows.Next() {
		var v models.TaskVerification
		err := rows.Scan(&v.ID, &v.TaskID, &v.VerificationType, &v.Status, &v.Confidence, &v.Evidence, &v.RetryCount, &v.VerifiedAt, &v.CreatedAt)
		if err != nil {
			return nil, err
		}
		verifications = append(verifications, v)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return verifications, nil
}
