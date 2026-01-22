// Package validation - Task-specific validators
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package validation

import (
	"regexp"
)

// ValidateCreateTaskRequest validates task creation requests
func ValidateCreateTaskRequest(data map[string]interface{}) error {
	composite := &CompositeValidator{
		Validators: []Validator{
			&StringValidator{
				Field:     "title",
				Required:  true,
				MinLength: 1,
				MaxLength: 500,
			},
			&StringValidator{
				Field:     "description",
				Required:  false,
				MaxLength: 5000,
			},
			&StringValidator{
				Field:    "status",
				Required: true,
				Enum:     []string{"pending", "in_progress", "completed", "archived"},
			},
			&StringValidator{
				Field:    "priority",
				Required: false,
				Enum:     []string{"low", "medium", "high", "critical"},
			},
			&StringValidator{
				Field:     "source",
				Required:  false,
				MaxLength: 100,
			},
			&StringValidator{
				Field:     "file_path",
				Required:  false,
				MaxLength: 1000,
			},
		},
	}
	return composite.Validate(data)
}

// ValidateUpdateTaskRequest validates task update requests
func ValidateUpdateTaskRequest(data map[string]interface{}) error {
	composite := &CompositeValidator{
		Validators: []Validator{
			&StringValidator{
				Field:     "title",
				Required:  false,
				MinLength: 1,
				MaxLength: 500,
			},
			&StringValidator{
				Field:     "description",
				Required:  false,
				MaxLength: 5000,
			},
			&StringValidator{
				Field:    "status",
				Required: false,
				Enum:     []string{"pending", "in_progress", "completed", "archived"},
			},
			&StringValidator{
				Field:    "priority",
				Required: false,
				Enum:     []string{"low", "medium", "high", "critical"},
			},
		},
	}
	return composite.Validate(data)
}

// ValidateListTasksRequest validates task listing/filtering requests
func ValidateListTasksRequest(data map[string]interface{}) error {
	composite := &CompositeValidator{
		Validators: []Validator{
			&StringValidator{
				Field:    "status_filter",
				Required: false,
				Enum:     []string{"pending", "in_progress", "completed", "archived"},
			},
			&StringValidator{
				Field:    "priority_filter",
				Required: false,
				Enum:     []string{"low", "medium", "high", "critical"},
			},
			&StringValidator{
				Field:     "source_filter",
				Required:  false,
				MaxLength: 100,
			},
			&NumericValidator{
				Field:    "limit",
				Required: false,
				Min:      floatPtr(1),
				Max:      floatPtr(1000),
				IsInt:    true,
			},
			&NumericValidator{
				Field:    "offset",
				Required: false,
				Min:      floatPtr(0),
				IsInt:    true,
			},
		},
	}
	return composite.Validate(data)
}

// ValidateCreateProjectRequest validates project creation requests
func ValidateCreateProjectRequest(data map[string]interface{}) error {
	composite := &CompositeValidator{
		Validators: []Validator{
			&StringValidator{
				Field:     "name",
				Required:  true,
				MinLength: 1,
				MaxLength: 255,
				Pattern:   regexp.MustCompile(`^[a-zA-Z0-9\s\-_]+$`),
			},
		},
	}
	return composite.Validate(data)
}

// ValidateCreateOrganizationRequest validates organization creation requests
func ValidateCreateOrganizationRequest(data map[string]interface{}) error {
	composite := &CompositeValidator{
		Validators: []Validator{
			&StringValidator{
				Field:     "name",
				Required:  true,
				MinLength: 1,
				MaxLength: 255,
			},
			&StringValidator{
				Field:     "description",
				Required:  false,
				MaxLength: 2000,
			},
		},
	}
	return composite.Validate(data)
}

// Helper function for float pointers
func floatPtr(f float64) *float64 {
	return &f
}
