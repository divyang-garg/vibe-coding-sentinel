// Phase 12: Input Validation Helpers
// Provides validation functions for Phase 12 API endpoints

package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ValidateUUID validates a UUID string
func ValidateUUID(id string) error {
	if id == "" {
		return fmt.Errorf("ID cannot be empty")
	}
	_, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid UUID format: %w", err)
	}
	return nil
}

// ValidatePath validates a filesystem path
func ValidatePath(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}
	// Prevent path traversal
	if strings.Contains(path, "..") {
		return fmt.Errorf("path cannot contain '..'")
	}
	// Check if path exists
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return fmt.Errorf("path does not exist: %s", path)
	}
	if err != nil {
		return fmt.Errorf("failed to stat path: %w", err)
	}
	return nil
}

// ValidateDirectory validates that path is a directory
func ValidateDirectory(path string) error {
	if err := ValidatePath(path); err != nil {
		return err
	}
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", path)
	}
	return nil
}

// validateRequired validates that a required field is not empty (lowercase for hook_handler.go compatibility)
func validateRequired(fieldName string, value string) error {
	if value == "" {
		return fmt.Errorf("%s is required", fieldName)
	}
	return nil
}

// validateUUID validates a UUID string (lowercase for hook_handler.go compatibility)
func validateUUID(id string) error {
	return ValidateUUID(id)
}

// validateHookType validates that hook type is valid
func validateHookType(hookType string) error {
	validTypes := []string{"pre-commit", "post-commit", "pre-push", "post-push", "pre-merge", "post-merge"}
	for _, validType := range validTypes {
		if hookType == validType {
			return nil
		}
	}
	return fmt.Errorf("invalid hook type: %s. Valid types are: %v", hookType, validTypes)
}

// validateResult validates that result is valid
func validateResult(result string) error {
	validResults := []string{"success", "failure", "warning", "error"}
	for _, validResult := range validResults {
		if result == validResult {
			return nil
		}
	}
	return fmt.Errorf("invalid result: %s. Valid results are: %v", result, validResults)
}

// validateDate validates a date string format
func validateDate(dateStr string) error {
	if dateStr == "" {
		return nil // Empty date is allowed
	}
	// Try to parse common date formats
	formats := []string{"2006-01-02", "2006-01-02T15:04:05Z", "2006-01-02 15:04:05"}
	for _, format := range formats {
		if _, err := time.Parse(format, dateStr); err == nil {
			return nil
		}
	}
	return fmt.Errorf("invalid date format: %s. Expected formats: YYYY-MM-DD, YYYY-MM-DDTHH:MM:SSZ, YYYY-MM-DD HH:MM:SS", dateStr)
}

// validateAction validates that action is valid
func validateAction(action string) error {
	validActions := []string{"allow", "block", "warn"}
	for _, validAction := range validActions {
		if action == validAction {
			return nil
		}
	}
	return fmt.Errorf("invalid action: %s. Valid actions are: %v", action, validActions)
}

// validateDocumentContentType validates that uploaded document has acceptable content type
func validateDocumentContentType(contentType string) error {
	allowedTypes := map[string]bool{
		"application/pdf": true,
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true, // .docx
		"application/msword":       true, // .doc
		"application/vnd.ms-excel": true, // .xls
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":         true, // .xlsx
		"application/vnd.ms-powerpoint":                                             true, // .ppt
		"application/vnd.openxmlformats-officedocument.presentationml.presentation": true, // .pptx
		"text/plain":       true,
		"text/csv":         true,
		"application/rtf":  true,
		"application/json": true,
		"application/xml":  true,
		"text/xml":         true,
	}

	if !allowedTypes[contentType] {
		return fmt.Errorf("unsupported document type: %s. Allowed types: PDF, Word, Excel, PowerPoint, text, CSV, RTF, JSON, XML", contentType)
	}
	return nil
}

// validateBinaryContentType validates that uploaded binary has acceptable content type
func validateBinaryContentType(contentType string) error {
	// For binaries, we accept generic binary types and some specific formats
	allowedTypes := map[string]bool{
		"application/octet-stream":  true, // Generic binary
		"application/x-executable":  true, // Linux executables
		"application/x-msdownload":  true, // Windows executables (.exe)
		"application/x-mach-binary": true, // macOS binaries
		"application/zip":           true, // Could be packaged binary
		"application/x-tar":         true, // Could be packaged binary
		"application/gzip":          true, // Could be compressed binary
	}

	// Allow empty content type (browsers sometimes don't set it for binaries)
	if contentType == "" {
		return nil
	}

	if !allowedTypes[contentType] {
		return fmt.Errorf("unsupported binary type: %s. Allowed types: executable binaries, archives", contentType)
	}
	return nil
}

// ValidateAPIKey performs enhanced validation of API keys for security
func ValidateAPIKey(apiKey string) error {
	if apiKey == "" {
		return fmt.Errorf("API key cannot be empty")
	}

	// Check minimum and maximum length (reasonable bounds for security)
	const minLength = 20
	const maxLength = 200
	if len(apiKey) < minLength {
		return fmt.Errorf("API key too short (minimum %d characters)", minLength)
	}
	if len(apiKey) > maxLength {
		return fmt.Errorf("API key too long (maximum %d characters)", maxLength)
	}

	// Check for basic format (alphanumeric + safe special chars)
	apiKeyPattern := regexp.MustCompile(`^[a-zA-Z0-9\-_\.]+$`)
	if !apiKeyPattern.MatchString(apiKey) {
		return fmt.Errorf("API key contains invalid characters (only alphanumeric, hyphen, underscore, and dot allowed)")
	}

	// Check for common weak patterns
	weakPatterns := []string{
		"password", "admin", "test", "default", "example",
		"123456", "abcdef", "api-key", "apikey",
	}

	apiKeyLower := strings.ToLower(apiKey)
	for _, pattern := range weakPatterns {
		if strings.Contains(apiKeyLower, pattern) {
			return fmt.Errorf("API key contains weak pattern and should be regenerated")
		}
	}

	// Check for sequential characters (weak security)
	if hasSequentialChars(apiKey) {
		return fmt.Errorf("API key contains sequential characters and should be regenerated")
	}

	return nil
}

// hasSequentialChars checks for sequential alphanumeric characters
func hasSequentialChars(s string) bool {
	runes := []rune(s)
	for i := 0; i < len(runes)-2; i++ {
		if runes[i+1] == runes[i]+1 && runes[i+2] == runes[i]+2 {
			return true // Found three sequential characters
		}
	}
	return false
}
