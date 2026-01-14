// Package repository contains document repository implementations.
package repository

import (
	"context"
	"database/sql"
	"sentinel-hub-api/models"
	"time"
)

// DocumentRepositoryImpl implements DocumentRepository
type DocumentRepositoryImpl struct {
	db Database
}

// NewDocumentRepository creates a new document repository instance
func NewDocumentRepository(db Database) *DocumentRepositoryImpl {
	return &DocumentRepositoryImpl{db: db}
}

// Save saves a document to the database
func (r *DocumentRepositoryImpl) Save(ctx context.Context, doc *models.Document) error {
	query := `
		INSERT INTO documents (id, project_id, name, original_name, size, mime_type, status, progress, file_path, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			progress = EXCLUDED.progress,
			extracted_text = EXCLUDED.extracted_text,
			processed_at = EXCLUDED.processed_at
		WHERE documents.id = EXCLUDED.id`

	_, err := r.db.Exec(ctx, query,
		doc.ID, doc.ProjectID, doc.Name, doc.OriginalName, doc.Size, doc.MimeType,
		string(doc.Status), doc.Progress, doc.FilePath, doc.CreatedAt)

	return err
}

// FindByID retrieves a document by ID
func (r *DocumentRepositoryImpl) FindByID(ctx context.Context, id string) (*models.Document, error) {
	query := `
		SELECT id, project_id, name, original_name, size, mime_type, status, progress,
		       file_path, extracted_text, error, created_at, processed_at
		FROM documents WHERE id = $1`

	var doc models.Document
	var filePath sql.NullString
	var extractedText sql.NullString
	var err sql.NullString
	var processedAt sql.NullTime
	var statusStr string

	rowErr := r.db.QueryRow(ctx, query, id).Scan(
		&doc.ID, &doc.ProjectID, &doc.Name, &doc.OriginalName, &doc.Size, &doc.MimeType,
		&statusStr, &doc.Progress, &filePath, &extractedText, &err, &doc.CreatedAt, &processedAt)

	if rowErr != nil {
		return nil, rowErr
	}

	// Convert string enum to typed enum
	doc.Status = models.DocumentStatus(statusStr)

	if rowErr != nil {
		return nil, rowErr
	}

	// Handle nullable fields
	if filePath.Valid {
		doc.FilePath = filePath.String
	}
	if extractedText.Valid {
		doc.ExtractedText = extractedText.String
	}
	if err.Valid {
		doc.Error = err.String
	}
	if processedAt.Valid {
		doc.ProcessedAt = &processedAt.Time
	}

	return &doc, nil
}

// FindByProjectID retrieves documents by project ID
func (r *DocumentRepositoryImpl) FindByProjectID(ctx context.Context, projectID string) ([]models.Document, error) {
	query := `
		SELECT id, project_id, name, original_name, size, mime_type, status, progress,
		       file_path, extracted_text, error, created_at, processed_at
		FROM documents
		WHERE project_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []models.Document
	for rows.Next() {
		var doc models.Document
		var filePath sql.NullString
		var extractedText sql.NullString
		var docErr sql.NullString
		var processedAt sql.NullTime
		var statusStr string

		err := rows.Scan(
			&doc.ID, &doc.ProjectID, &doc.Name, &doc.OriginalName, &doc.Size, &doc.MimeType,
			&statusStr, &doc.Progress, &filePath, &extractedText, &docErr, &doc.CreatedAt, &processedAt)

		if err != nil {
			return nil, err
		}

		// Convert string enum to typed enum
		doc.Status = models.DocumentStatus(statusStr)

		// Handle nullable fields
		if filePath.Valid {
			doc.FilePath = filePath.String
		}
		if extractedText.Valid {
			doc.ExtractedText = extractedText.String
		}
		if docErr.Valid {
			doc.Error = docErr.String
		}
		if processedAt.Valid {
			doc.ProcessedAt = &processedAt.Time
		}

		docs = append(docs, doc)
	}

	return docs, nil
}

// Update updates a document in the database
func (r *DocumentRepositoryImpl) Update(ctx context.Context, doc *models.Document) error {
	return r.Save(ctx, doc)
}

// Delete marks a document as deleted (soft delete)
func (r *DocumentRepositoryImpl) Delete(ctx context.Context, id string) error {
	query := "UPDATE documents SET status = 'deleted', error = 'deleted by user' WHERE id = $1"
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// UpdateStatus updates document processing status
func (r *DocumentRepositoryImpl) UpdateStatus(ctx context.Context, id string, status string, progress int, errorMsg string) error {
	query := `
		UPDATE documents
		SET status = $1, progress = $2, error = $3, processed_at = CASE WHEN $1 = 'completed' THEN $4 ELSE processed_at END
		WHERE id = $5`

	var processedAt *time.Time
	if status == "completed" {
		now := time.Now()
		processedAt = &now
	}

	_, err := r.db.Exec(ctx, query, status, progress, errorMsg, processedAt, id)
	return err
}

// UpdateProcessedAt updates the processed timestamp
func (r *DocumentRepositoryImpl) UpdateProcessedAt(ctx context.Context, id string, processedAt *time.Time) error {
	query := "UPDATE documents SET processed_at = $1 WHERE id = $2"
	_, err := r.db.Exec(ctx, query, processedAt, id)
	return err
}
