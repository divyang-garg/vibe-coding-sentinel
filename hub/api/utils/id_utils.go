// ID Utilities
// Standardized ID validation and generation functions

package utils

import (
	"fmt"

	"github.com/google/uuid"
)

// ValidateEntityID validates a UUID and returns a standardized error
func ValidateEntityID(id string, entityType string) error {
	if err := ValidateUUID(id); err != nil {
		return fmt.Errorf("invalid %s ID: %w", entityType, err)
	}
	return nil
}

// ValidateTaskID validates a task ID
func ValidateTaskID(id string) error {
	return ValidateEntityID(id, "task")
}

// ValidateChangeRequestID validates a change request ID
func ValidateChangeRequestID(id string) error {
	return ValidateEntityID(id, "change request")
}

// ValidateKnowledgeItemID validates a knowledge item ID
func ValidateKnowledgeItemID(id string) error {
	return ValidateEntityID(id, "knowledge item")
}

// ValidateComprehensiveValidationID validates a comprehensive validation ID
func ValidateComprehensiveValidationID(id string) error {
	// Comprehensive validation uses validation_id (VARCHAR), not UUID
	// Just check it's not empty
	if id == "" {
		return fmt.Errorf("invalid comprehensive validation ID: cannot be empty")
	}
	return nil
}

// ValidateTestRequirementID validates a test requirement ID
func ValidateTestRequirementID(id string) error {
	return ValidateEntityID(id, "test requirement")
}

// GenerateEntityID generates a new UUID for entities
func GenerateEntityID() string {
	return uuid.New().String()
}
