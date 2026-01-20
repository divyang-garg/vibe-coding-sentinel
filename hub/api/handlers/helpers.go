// Package handlers helpers
// Common helper functions for HTTP handlers
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package handlers

import (
	"context"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// contextKey type for context keys
type contextKey string

// projectKey is the context key for project
const projectKey contextKey = "project"

// getProjectFromContext extracts project from context
func getProjectFromContext(ctx context.Context) (*Project, error) {
	project, ok := ctx.Value(projectKey).(*Project)
	if !ok || project == nil {
		return nil, fmt.Errorf("project not found in context")
	}
	return project, nil
}

// sanitizeString sanitizes a string for safe storage
func sanitizeString(s string, maxLen int) string {
	s = html.EscapeString(s)
	s = strings.Map(func(r rune) rune {
		if r < 32 && r != '\n' && r != '\r' && r != '\t' {
			return -1
		}
		return r
	}, s)
	if len(s) > maxLen {
		s = s[:maxLen]
	}
	return s
}

// validateDate validates a date string in YYYY-MM-DD format
func validateDate(dateStr string) error {
	if dateStr == "" {
		return nil
	}
	matched, _ := regexp.MatchString(`^\d{4}-\d{2}-\d{2}$`, dateStr)
	if !matched {
		return fmt.Errorf("invalid date format: expected YYYY-MM-DD")
	}
	_, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return fmt.Errorf("invalid date: %w", err)
	}
	return nil
}

// validateAction validates an action string
func validateAction(action string) error {
	validActions := map[string]bool{
		"create": true, "update": true, "delete": true,
		"approve": true, "reject": true, "override": true,
	}
	if !validActions[action] {
		return fmt.Errorf("invalid action: %s", action)
	}
	return nil
}

// sanitizePath sanitizes a file path
func sanitizePath(path string) string {
	path = strings.Map(func(r rune) rune {
		if r < 32 {
			return -1
		}
		return r
	}, path)
	path = strings.ReplaceAll(path, "..", "")
	return path
}

// isValidPath checks if a path is valid
func isValidPath(path string) bool {
	if strings.Contains(path, "..") {
		return false
	}
	if strings.HasPrefix(path, "/") && !strings.HasPrefix(path, "/app/") {
		return false
	}
	return true
}

// ValidateUUID validates a UUID string
func ValidateUUID(id string) error {
	if len(id) != 36 {
		return fmt.Errorf("invalid UUID length")
	}
	matched, _ := regexp.MatchString(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`, id)
	if !matched {
		return fmt.Errorf("invalid UUID format")
	}
	return nil
}

// Health check handlers
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

func healthDBHandler(w http.ResponseWriter, r *http.Request) {
	if db != nil {
		if err := db.Ping(); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"status":"error","message":"database unreachable"}`))
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok","database":"connected"}`))
}

func healthReadyHandler(w http.ResponseWriter, r *http.Request) {
	ready := true
	var issues []string
	if db != nil {
		if err := db.Ping(); err != nil {
			ready = false
			issues = append(issues, "database")
		}
	}
	w.Header().Set("Content-Type", "application/json")
	if ready {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ready":true}`))
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(fmt.Sprintf(`{"ready":false,"issues":["%s"]}`, strings.Join(issues, "\",\""))))
	}
}

// saveUploadedFile saves an uploaded file to a temporary location
func saveUploadedFile(file io.Reader, filename string) (string, error) {
	// Create temp directory if it doesn't exist
	tempDir := os.TempDir()
	uploadDir := filepath.Join(tempDir, "sentinel-uploads")
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Generate unique filename
	timestamp := time.Now().UnixNano()
	safeFilename := sanitizePath(filename)
	tempPath := filepath.Join(uploadDir, fmt.Sprintf("%d_%s", timestamp, safeFilename))

	// Create file
	dst, err := os.Create(tempPath)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer dst.Close()

	// Copy file content
	_, err = io.Copy(dst, file)
	if err != nil {
		os.Remove(tempPath)
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	return tempPath, nil
}

// detectMimeType detects MIME type from file extension
func detectMimeType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	mimeTypes := map[string]string{
		".pdf":  "application/pdf",
		".doc":  "application/msword",
		".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".txt":  "text/plain",
		".md":   "text/markdown",
		".html": "text/html",
		".htm":  "text/html",
		".json": "application/json",
		".xml":  "application/xml",
		".csv":  "text/csv",
		".xls":  "application/vnd.ms-excel",
		".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		".png":  "image/png",
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".gif":  "image/gif",
		".svg":  "image/svg+xml",
	}

	if mimeType, ok := mimeTypes[ext]; ok {
		return mimeType
	}

	// Default to binary if unknown
	return "application/octet-stream"
}
