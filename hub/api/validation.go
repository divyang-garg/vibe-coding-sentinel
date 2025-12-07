package main

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// validateUUID validates that a string is a valid UUID format
func validateUUID(id string) error {
	if id == "" {
		return fmt.Errorf("UUID cannot be empty")
	}
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("invalid UUID format: %s", id)
	}
	return nil
}

// validateRequired validates that a required field is not empty
func validateRequired(field, value string) error {
	if value == "" {
		return fmt.Errorf("%s is required", field)
	}
	return nil
}

// validateRange validates that a numeric value is within the specified range
func validateRange(field string, value, min, max int) error {
	if value < min {
		return fmt.Errorf("%s must be >= %d (got %d)", field, min, value)
	}
	if value > max {
		return fmt.Errorf("%s must be <= %d (got %d)", field, max, value)
	}
	return nil
}

// validateHookType validates that hook_type is one of the allowed values
func validateHookType(hookType string) error {
	allowedTypes := []string{"pre-commit", "pre-push", "commit-msg"}
	for _, allowed := range allowedTypes {
		if hookType == allowed {
			return nil
		}
	}
	return fmt.Errorf("invalid hook_type: %s (must be one of: %s)", hookType, strings.Join(allowedTypes, ", "))
}

// validateResult validates that result is one of the allowed values
func validateResult(result string) error {
	allowedResults := []string{"allowed", "blocked", "overridden"}
	for _, allowed := range allowedResults {
		if result == allowed {
			return nil
		}
	}
	return fmt.Errorf("invalid result: %s (must be one of: %s)", result, strings.Join(allowedResults, ", "))
}

// validateDate validates that a date string is in a valid format (YYYY-MM-DD or RFC3339)
func validateDate(dateStr string) error {
	if dateStr == "" {
		return nil // Empty date is allowed (optional field)
	}
	
	// Try RFC3339 format first
	if _, err := time.Parse(time.RFC3339, dateStr); err == nil {
		return nil
	}
	
	// Try YYYY-MM-DD format
	if matched, _ := regexp.MatchString(`^\d{4}-\d{2}-\d{2}$`, dateStr); matched {
		if _, err := time.Parse("2006-01-02", dateStr); err == nil {
			return nil
		}
	}
	
	return fmt.Errorf("invalid date format: %s (expected YYYY-MM-DD or RFC3339)", dateStr)
}

// validateAction validates that action is one of the allowed values for baseline review
func validateAction(action string) error {
	allowedActions := []string{"approve", "reject"}
	for _, allowed := range allowedActions {
		if action == allowed {
			return nil
		}
	}
	return fmt.Errorf("invalid action: %s (must be one of: %s)", action, strings.Join(allowedActions, ", "))
}

