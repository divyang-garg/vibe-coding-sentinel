// Package validation provides comprehensive input validation framework
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package validation

import (
	"fmt"
	"regexp"
	"strings"
)

// ValidationError represents a validation failure with field context
type ValidationError struct {
	Field   string
	Value   interface{}
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed for field '%s': %s", e.Field, e.Message)
}

// Validator defines the interface for all validators
type Validator interface {
	Validate(data map[string]interface{}) error
}

// StringValidator validates string fields with comprehensive rules
type StringValidator struct {
	Field      string
	Required   bool
	MinLength  int
	MaxLength  int
	Pattern    *regexp.Regexp
	Enum       []string
	AllowEmpty bool
}

// Validate implements Validator interface for string fields
func (v *StringValidator) Validate(data map[string]interface{}) error {
	value, exists := data[v.Field]

	// Handle missing or nil values
	if !exists || value == nil {
		if v.Required {
			return &ValidationError{
				Field:   v.Field,
				Message: fmt.Sprintf("%s is required", v.Field),
			}
		}
		return nil // Optional field, skip validation
	}

	// Type assertion
	str, ok := value.(string)
	if !ok {
		return &ValidationError{
			Field:   v.Field,
			Message: fmt.Sprintf("%s must be a string", v.Field),
		}
	}

	// Empty string handling
	if str == "" {
		if v.Required {
			return &ValidationError{
				Field:   v.Field,
				Message: fmt.Sprintf("%s is required and cannot be empty", v.Field),
			}
		}
		if !v.AllowEmpty {
			return &ValidationError{
				Field:   v.Field,
				Message: fmt.Sprintf("%s cannot be empty", v.Field),
			}
		}
		return nil
	}

	// Length validation
	if v.MinLength > 0 && len(str) < v.MinLength {
		return &ValidationError{
			Field:   v.Field,
			Message: fmt.Sprintf("%s must be at least %d characters", v.Field, v.MinLength),
		}
	}

	if v.MaxLength > 0 && len(str) > v.MaxLength {
		return &ValidationError{
			Field:   v.Field,
			Message: fmt.Sprintf("%s must be at most %d characters", v.Field, v.MaxLength),
		}
	}

	// Pattern validation
	if v.Pattern != nil && !v.Pattern.MatchString(str) {
		return &ValidationError{
			Field:   v.Field,
			Message: fmt.Sprintf("%s format is invalid", v.Field),
		}
	}

	// Enum validation
	if len(v.Enum) > 0 {
		found := false
		for _, allowed := range v.Enum {
			if str == allowed {
				found = true
				break
			}
		}
		if !found {
			return &ValidationError{
				Field:   v.Field,
				Message: fmt.Sprintf("%s must be one of: %s", v.Field, strings.Join(v.Enum, ", ")),
			}
		}
	}

	return nil
}

// CompositeValidator validates multiple fields using multiple validators
type CompositeValidator struct {
	Validators []Validator
}

// Validate implements Validator interface for composite validation
func (c *CompositeValidator) Validate(data map[string]interface{}) error {
	for _, validator := range c.Validators {
		if err := validator.Validate(data); err != nil {
			return err
		}
	}
	return nil
}

// NumericValidator validates numeric fields
type NumericValidator struct {
	Field    string
	Required bool
	Min      *float64
	Max      *float64
	IsInt    bool
}

// Validate implements Validator interface for numeric fields
func (v *NumericValidator) Validate(data map[string]interface{}) error {
	value, exists := data[v.Field]

	if !exists || value == nil {
		if v.Required {
			return &ValidationError{
				Field:   v.Field,
				Message: fmt.Sprintf("%s is required", v.Field),
			}
		}
		return nil
	}

	var num float64
	var ok bool

	if v.IsInt {
		var intVal int
		intVal, ok = value.(int)
		num = float64(intVal)
	} else {
		num, ok = value.(float64)
		if !ok {
			// Try int conversion
			if intVal, okInt := value.(int); okInt {
				num = float64(intVal)
				ok = true
			}
		}
	}

	if !ok {
		return &ValidationError{
			Field:   v.Field,
			Message: fmt.Sprintf("%s must be a number", v.Field),
		}
	}

	if v.Min != nil && num < *v.Min {
		return &ValidationError{
			Field:   v.Field,
			Message: fmt.Sprintf("%s must be at least %v", v.Field, *v.Min),
		}
	}

	if v.Max != nil && num > *v.Max {
		return &ValidationError{
			Field:   v.Field,
			Message: fmt.Sprintf("%s must be at most %v", v.Field, *v.Max),
		}
	}

	return nil
}

// EmailValidator validates email addresses
func EmailValidator(field string, required bool) Validator {
	emailPattern := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return &StringValidator{
		Field:     field,
		Required:  required,
		Pattern:   emailPattern,
		MaxLength: 255,
	}
}

// UUIDValidator validates UUID format
func UUIDValidator(field string, required bool) Validator {
	uuidPattern := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	return &StringValidator{
		Field:    field,
		Required: required,
		Pattern:  uuidPattern,
	}
}

// URLValidator validates URL format
func URLValidator(field string, required bool) Validator {
	urlPattern := regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
	return &StringValidator{
		Field:     field,
		Required:  required,
		Pattern:   urlPattern,
		MaxLength: 2048,
	}
}

// SanitizeString removes potentially dangerous characters from strings
func SanitizeString(input string) string {
	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")

	// Trim whitespace
	input = strings.TrimSpace(input)

	// Remove control characters (except newline and tab)
	var result strings.Builder
	for _, r := range input {
		if r >= 32 || r == '\n' || r == '\t' {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// ValidateRequestSize checks if request body size is within limits
func ValidateRequestSize(size int64, maxSize int64) error {
	if size > maxSize {
		return fmt.Errorf("request body too large: %d bytes (maximum: %d bytes)", size, maxSize)
	}
	return nil
}

// Common validation patterns
var (
	// AlphanumericPattern matches alphanumeric strings
	AlphanumericPattern = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

	// SafeStringPattern matches strings safe for database storage
	SafeStringPattern = regexp.MustCompile(`^[a-zA-Z0-9\s\-_.,!?@#$%^&*()+=\[\]{}|\\:;"'<>/]+$`)

	// NoSQLInjectionPattern ensures no SQL injection attempts
	NoSQLInjectionPattern = regexp.MustCompile(`(?i)(union|select|insert|update|delete|drop|create|alter|exec|execute|script)`)
)

// ValidateNoSQLInjection checks for SQL injection patterns
func ValidateNoSQLInjection(field, value string) error {
	if NoSQLInjectionPattern.MatchString(value) {
		return &ValidationError{
			Field:   field,
			Message: "potentially unsafe input detected",
		}
	}
	return nil
}
